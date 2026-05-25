package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
)

var dockerHost string

var (
	prevNetRx   = make(map[string]uint64)
	prevNetTx   = make(map[string]uint64)
	prevNetTime = make(map[string]time.Time)
	netCacheMu  sync.Mutex
)

func SetDockerHost(host string) {
	dockerHost = host
}

type ContainerInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	State       string   `json:"state"`
	Ports       []string `json:"ports"`
	CPUPercent  float64  `json:"cpu_percent"`
	MemoryUsage int64    `json:"memory_usage"`
	MemoryLimit int64    `json:"memory_limit"`
	NetworkRx   int64    `json:"network_rx"`
	NetworkTx   int64    `json:"network_tx"`
}

type containerStatsJSON struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
}

func ListContainers(c *gin.Context) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法连接 Docker"})
		return
	}
	defer cli.Close()

	result, err := cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取容器列表失败"})
		return
	}

	type containerEntry struct {
		ctID string
		info ContainerInfo
	}

	entries := make([]containerEntry, len(result.Items))
	for i, ct := range result.Items {
		name := ""
		if len(ct.Names) > 0 {
			name = strings.TrimPrefix(ct.Names[0], "/")
		}

		ports := make([]string, 0)
		for _, p := range ct.Ports {
			if p.PublicPort > 0 {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}

		id := ct.ID
		if len(id) > 12 {
			id = id[:12]
		}

		entries[i] = containerEntry{
			ctID: ct.ID,
			info: ContainerInfo{
				ID:     id,
				Name:   name,
				Status: ct.Status,
				State:  string(ct.State),
				Ports:  ports,
			},
		}
	}

	var wg sync.WaitGroup
	for i := range entries {
		wg.Add(1)
		go func(e *containerEntry) {
			defer wg.Done()
			resp, err := cli.ContainerStats(context.Background(), e.ctID, client.ContainerStatsOptions{
				Stream:                false,
				IncludePreviousSample: true,
			})
			if err != nil {
				return
			}
			defer resp.Body.Close()

			var s containerStatsJSON
			if json.NewDecoder(resp.Body).Decode(&s) != nil {
				return
			}

			e.info.MemoryUsage = int64(s.MemoryStats.Usage)
			e.info.MemoryLimit = int64(s.MemoryStats.Limit)

			cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage - s.PreCPUStats.CPUUsage.TotalUsage)
			sysDelta := float64(s.CPUStats.SystemCPUUsage - s.PreCPUStats.SystemCPUUsage)
			if sysDelta > 0 && s.CPUStats.OnlineCPUs > 0 {
				e.info.CPUPercent = (cpuDelta / sysDelta) * float64(s.CPUStats.OnlineCPUs) * 100
			}

			var rx, tx uint64
			for _, net := range s.Networks {
				rx += net.RxBytes
				tx += net.TxBytes
			}

			netCacheMu.Lock()
			now := time.Now()
			if prevRx, ok := prevNetRx[e.ctID]; ok {
				elapsed := now.Sub(prevNetTime[e.ctID]).Seconds()
				if elapsed > 0 {
					e.info.NetworkRx = int64(float64(rx-prevRx) / elapsed)
					e.info.NetworkTx = int64(float64(tx-prevNetTx[e.ctID]) / elapsed)
				}
			}
			prevNetRx[e.ctID] = rx
			prevNetTx[e.ctID] = tx
			prevNetTime[e.ctID] = now
			netCacheMu.Unlock()
		}(&entries[i])
	}
	wg.Wait()

	list := make([]ContainerInfo, len(entries))
	for i, e := range entries {
		list[i] = e.info
	}

	c.JSON(http.StatusOK, list)
}
