package scheduler

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"nas-partner/backend/internal/ddns/dns"
	ddnsmodel "nas-partner/backend/internal/ddns/model"
)

type Scheduler struct {
	mu     sync.Mutex
	jobs   map[int64]*job
	stopCh chan struct{}
}

type job struct {
	cfg     *ddnsmodel.DDNSConfig
	ticker  *time.Ticker
	stopCh  chan struct{}
	running bool
}

var Default = &Scheduler{
	jobs:   make(map[int64]*job),
	stopCh: make(chan struct{}),
}

// Start 启动所有启用的DDNS配置
func (s *Scheduler) Start() {
	s.cleanupOldLogs()
	go s.loop()
}

// cleanupOldLogs 清理3天前的执行日志
func (s *Scheduler) cleanupOldLogs() {
	n, err := ddnsmodel.DeleteOlderThan(3)
	if err != nil {
		log.Printf("DDNS: 清理旧日志失败: %v", err)
		return
	}
	if n > 0 {
		log.Printf("DDNS: 已清理 %d 条过期日志", n)
	}
}

// loop 定时加载所有启用的配置并启动任务
func (s *Scheduler) loop() {
	// initial load
	s.syncJobs()

	// reload every 30 seconds to catch new configs
	ticker := time.NewTicker(30 * time.Second)
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ticker.C:
			s.syncJobs()
		case <-cleanupTicker.C:
			s.cleanupOldLogs()
		case <-s.stopCh:
			return
		}
	}
}

// syncJobs 同步启用的任务
func (s *Scheduler) syncJobs() {
	configs, err := ddnsmodel.List()
	if err != nil {
		log.Printf("DDNS scheduler: failed to list configs: %v", err)
		return
	}

	enabledIDs := make(map[int64]bool)
	for _, c := range configs {
		if !c.Enabled {
			continue
		}
		enabledIDs[c.ID] = true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop jobs that are no longer enabled
	for id, j := range s.jobs {
		if !enabledIDs[id] {
			j.stop()
			delete(s.jobs, id)
		}
	}

	// Start jobs that are enabled but not running
	for _, c := range configs {
		if !c.Enabled {
			continue
		}
		if _, ok := s.jobs[c.ID]; ok {
			continue
		}
		j := &job{
			cfg:    c,
			stopCh: make(chan struct{}),
		}
		s.jobs[c.ID] = j
		go j.run()
	}
}

// RunOnce 立即运行指定的配置，返回执行结果
func (s *Scheduler) RunOnce(id int64) *ddnsmodel.DDNSRunLog {
	cfg, err := ddnsmodel.GetByID(id)
	if err != nil {
		log.Printf("DDNS scheduler: config %d not found: %v", id, err)
		return nil
	}
	return s.execute(cfg)
}

// execute 执行一次DDNS更新，返回执行结果
func (s *Scheduler) execute(cfg *ddnsmodel.DDNSConfig) *ddnsmodel.DDNSRunLog {
	log.Printf("DDNS: 开始执行 %s (ID: %d)", cfg.Name, cfg.ID)
	domains := dns.RunOnce(cfg)

	v4Status, v6Status := dns.GetDomainsStatus(domains)

	// Build detailed message from domain results
	type ipLine struct {
		addr string
		domains []string
	}
	var lines []string
	var v4Info, v6Info ipLine
	if cfg.IPv4Enabled {
		addr := domains.Ipv4Addr
		if addr == "" {
			addr = "获取失败"
		}
		for _, d := range domains.Ipv4Domains {
			detail := string(d.UpdateStatus)
			if d.Detail != "" {
				detail = d.Detail
			}
			v4Info.domains = append(v4Info.domains, fmt.Sprintf("%s→%s", d, detail))
		}
		v4Info.addr = addr
	} else {
		v4Info.addr = "未启用"
	}
	if cfg.IPv6Enabled {
		addr := domains.Ipv6Addr
		if addr == "" {
			addr = "获取失败"
		}
		for _, d := range domains.Ipv6Domains {
			detail := string(d.UpdateStatus)
			if d.Detail != "" {
				detail = d.Detail
			}
			v6Info.domains = append(v6Info.domains, fmt.Sprintf("%s→%s", d, detail))
		}
		v6Info.addr = addr
	} else {
		v6Info.addr = "未启用"
	}

	// Format: one line per family, compact
	if v4Info.addr != "" {
		s := fmt.Sprintf("V4:%s", v4Info.addr)
		if len(v4Info.domains) > 0 {
			s += " | " + strings.Join(v4Info.domains, ", ")
		}
		lines = append(lines, s)
	}
	if v6Info.addr != "" {
		s := fmt.Sprintf("V6:%s", v6Info.addr)
		if len(v6Info.domains) > 0 {
			s += " | " + strings.Join(v6Info.domains, ", ")
		}
		lines = append(lines, s)
	}
	message := strings.Join(lines, "\n")

	// Determine overall status
	status := ""
	switch {
	case v4Status == "成功" || v6Status == "成功":
		status = "成功"
	case v4Status == "失败" || v6Status == "失败":
		status = "失败"
		if message == "" {
			message = "更新失败"
		}
	default:
		status = "未改变"
	}

	logEntry, err := ddnsmodel.CreateRunLog(cfg.ID, status, message, domains.Ipv4Addr, domains.Ipv6Addr)
	if err != nil {
		log.Printf("DDNS: 写入执行日志失败 %s: %v", cfg.Name, err)
	}

	// 更新当前 IP（成功才写，失败保留旧值）
	if v4Status == "成功" && domains.Ipv4Addr != "" {
		cfg.CurrentIPv4Addr = domains.Ipv4Addr
	}
	if v6Status == "成功" && domains.Ipv6Addr != "" {
		cfg.CurrentIPv6Addr = domains.Ipv6Addr
	}
	if v4Status == "成功" || v6Status == "成功" {
		if err = cfg.UpdateCurrentIPs(); err != nil {
			log.Printf("DDNS: 更新当前IP失败 %s: %v", cfg.Name, err)
		}
	}

	log.Printf("DDNS: %s (ID: %d) 执行完毕, 状态: %s", cfg.Name, cfg.ID, status)
	return logEntry
}

func (j *job) run() {
	log.Printf("DDNS job: 启动 %s (间隔: %ds)", j.cfg.Name, j.cfg.Interval)

	// Run immediately on start
	Default.execute(j.cfg)

	interval := time.Duration(j.cfg.Interval) * time.Second
	if interval < 60*time.Second {
		interval = 60 * time.Second
	}
	j.ticker = time.NewTicker(interval)
	j.running = true

	for {
		select {
		case <-j.ticker.C:
			// Reload config to get latest state
			cfg, err := ddnsmodel.GetByID(j.cfg.ID)
			if err != nil {
				log.Printf("DDNS job: 获取配置失败 %d: %v", j.cfg.ID, err)
				continue
			}
			if !cfg.Enabled {
				return
			}
			j.cfg = cfg
			Default.execute(j.cfg)
		case <-j.stopCh:
			return
		}
	}
}

func (j *job) stop() {
	if j.ticker != nil {
		j.ticker.Stop()
	}
	j.running = false
	select {
	case j.stopCh <- struct{}{}:
	default:
	}
}

// Stop 停止所有任务
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, j := range s.jobs {
		j.stop()
		delete(s.jobs, id)
	}
	close(s.stopCh)
}

// IsRunning 检查指定配置是否正在运行
func (s *Scheduler) IsRunning(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	return ok && j.running
}
