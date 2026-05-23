package config

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"nas-partner/backend/internal/ddns/util"
)

// AutoGetUrl 内置自动检测公网 IP 的 API（按顺序尝试）
var AutoGetUrl = []string{
	"https://myip.ipip.net",
	"https://4.ipw.cn",
	"https://api.ipify.org",
	"https://ipv4.icanhazip.com",
	"https://checkip.amazonaws.com",
}

// maxURLFailures 自动检测API连续失败次数上限，超过后本会话内临时禁用
const maxURLFailures = 3

// urlFailCount 记录各API的连续失败次数（进程内，重启重置）
var urlFailCount struct {
	mu sync.Mutex
	m  map[string]int
}

func init() {
	urlFailCount.m = make(map[string]int)
}

// autoDetectClient 自动检测 IP 用的 HTTP 客户端，超时更短
var autoDetectClient = &http.Client{
	Timeout:   5 * time.Second,
	Transport: noProxyTransport,
}

// noProxyTransport 无代理、短超时传输层，专用于自动检测
var noProxyTransport = &http.Transport{
	DisableKeepAlives: true,
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: -1,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	TLSHandshakeTimeout:   3 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

// isURLBlocked 检查URL是否因连续失败被临时禁用
func isURLBlocked(url string) bool {
	urlFailCount.mu.Lock()
	defer urlFailCount.mu.Unlock()
	return urlFailCount.m[url] >= maxURLFailures
}

// recordURLFailure 记录一次URL调用失败
func recordURLFailure(url string) {
	urlFailCount.mu.Lock()
	defer urlFailCount.mu.Unlock()
	urlFailCount.m[url]++
	if urlFailCount.m[url] >= maxURLFailures {
		util.Log("API %s 连续 %d 次获取失败，本会话内临时禁用", url, maxURLFailures)
	}
}

// recordURLSuccess 重置URL的失败计数
func recordURLSuccess(url string) {
	urlFailCount.mu.Lock()
	defer urlFailCount.mu.Unlock()
	urlFailCount.m[url] = 0
}

// Ipv4Reg IPv4正则
var Ipv4Reg = regexp.MustCompile(`((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])`)

// Ipv6Reg IPv6正则
var Ipv6Reg = regexp.MustCompile(`([0-9A-Fa-f:.]{2,})`)

// DnsConfig 配置
type DnsConfig struct {
	Name string
	Ipv4 struct {
		Enable       bool
		GetType      string
		URL          string
		NetInterface string
		Cmd          string
		Addr         string
		Domains      []string
	}
	Ipv6 struct {
		Enable       bool
		GetType      string
		URL          string
		NetInterface string
		Cmd          string
		Addr         string
		Ipv6Reg      string
		Domains      []string
	}
	DNS           DNS
	TTL           string
}

// DNS DNS配置
type DNS struct {
	Name     string
	ID       string
	Secret   string
	ExtParam string
}

func (conf *DnsConfig) getIpv4AddrFromInterface() string {
	ipv4, _, err := GetNetInterface()
	if err != nil {
		util.Log("从网卡获得IPv4失败")
		return ""
	}

	for _, netInterface := range ipv4 {
		if netInterface.Name == conf.Ipv4.NetInterface && len(netInterface.Address) > 0 {
			return netInterface.Address[0]
		}
	}

	util.Log("从网卡中获得IPv4失败! 网卡名: %s", conf.Ipv4.NetInterface)
	return ""
}

func (conf *DnsConfig) getIpv4AddrFromUrl() string {
	client := util.CreateBoundNoProxyHTTPClient("tcp4", "")
	urls := strings.Split(conf.Ipv4.URL, ",")
	for _, url := range urls {
		url = strings.TrimSpace(url)
		resp, err := client.Get(url)
		if err != nil {
			util.Log("通过接口获取IPv4失败! 接口地址: %s", url)
			util.Log("异常信息: %s", err)
			continue
		}
		defer resp.Body.Close()
		lr := io.LimitReader(resp.Body, 1024000)
		body, err := io.ReadAll(lr)
		if err != nil {
			util.Log("异常信息: %s", err)
			continue
		}
		result := Ipv4Reg.FindString(string(body))
		if result == "" {
			util.Log("获取IPv4结果失败! 接口: %s ,返回值: %s", url, string(body))
		}
		return result
	}
	return ""
}

func findIPv6InText(text string) string {
	for _, candidate := range Ipv6Reg.FindAllString(text, -1) {
		if ip := net.ParseIP(candidate); ip != nil && ip.To4() == nil {
			return candidate
		}
	}
	return ""
}

func (conf *DnsConfig) getAddrFromCmd(addrType string) string {
	var cmd string
	var comp *regexp.Regexp
	if addrType == "IPv4" {
		cmd = conf.Ipv4.Cmd
		comp = Ipv4Reg
	} else {
		cmd = conf.Ipv6.Cmd
		comp = Ipv6Reg
	}
	if cmd == "" {
		return ""
	}

	var execCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		execCmd = exec.Command("powershell", "-Command", cmd)
	} else {
		_, err := exec.LookPath("bash")
		if err != nil {
			execCmd = exec.Command("sh", "-c", cmd)
		} else {
			execCmd = exec.Command("bash", "-c", cmd)
		}
	}

	out, err := execCmd.CombinedOutput()
	if err != nil {
		util.Log("获取%s结果失败! 未能成功执行命令：%s, 错误：%q, 退出状态码：%s", addrType, execCmd.String(), out, err)
		return ""
	}
	str := string(out)
	result := ""
	if addrType == "IPv4" {
		result = comp.FindString(str)
	} else {
		result = findIPv6InText(str)
	}
	if result == "" {
		util.Log("获取%s结果失败! 命令: %s, 标准输出: %q", addrType, execCmd.String(), str)
	}
	return result
}

func (conf *DnsConfig) getIpv4AddrManual() string {
	return conf.Ipv4.Addr
}

func (conf *DnsConfig) getIpv4AddrAuto() string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type result struct {
		url string
		ip  string
	}
	first := make(chan result, 1)

	var wg sync.WaitGroup
	for _, url := range AutoGetUrl {
		if isURLBlocked(url) {
			continue
		}
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				return
			}
			resp, err := autoDetectClient.Do(req)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				util.Log("通过接口获取IPv4失败! 接口地址: %s", u)
				util.Log("异常信息: %s", err)
				recordURLFailure(u)
				return
			}
			body, err := io.ReadAll(io.LimitReader(resp.Body, 1024000))
			resp.Body.Close()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				util.Log("异常信息: %s", err)
				recordURLFailure(u)
				return
			}
			ip := Ipv4Reg.FindString(string(body))
			if ip == "" {
				if ctx.Err() != nil {
					return
				}
				util.Log("获取IPv4结果失败! 接口: %s ,返回值: %s", u, string(body))
				recordURLFailure(u)
				return
			}
			recordURLSuccess(u)
			select {
			case first <- result{u, ip}:
			default:
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(first)
	}()

	r, ok := <-first
	if ok {
		return r.ip
	}
	return ""
}

func (conf *DnsConfig) GetIpv4Addr() string {
	switch conf.Ipv4.GetType {
	case "netInterface":
		return conf.getIpv4AddrFromInterface()
	case "auto":
		return conf.getIpv4AddrAuto()
	case "manual":
		return conf.getIpv4AddrManual()
	default:
		log.Println("IPv4's get IP method is unknown")
		return ""
	}
}

func (conf *DnsConfig) getIpv6AddrFromInterface() string {
	_, ipv6, err := GetNetInterface()
	if err != nil {
		util.Log("从网卡获得IPv6失败")
		return ""
	}

	// 解析格式："en0|240e:..."（前端选择的具体地址），或旧格式 "en0"（使用第一个地址）
	ifaceName := conf.Ipv6.NetInterface
	specificAddr := ""
	if parts := strings.SplitN(ifaceName, "|", 2); len(parts) == 2 {
		ifaceName = parts[0]
		specificAddr = parts[1]
	}

	for _, netInterface := range ipv6 {
		if netInterface.Name == ifaceName && len(netInterface.Address) > 0 {
			// 如果指定了具体地址，优先使用
			if specificAddr != "" {
				for _, addr := range netInterface.Address {
					if addr == specificAddr {
						return addr
					}
				}
				util.Log("未找到指定的IPv6地址 %s, 使用第一个地址", specificAddr)
				return netInterface.Address[0]
			}

			if conf.Ipv6.Ipv6Reg != "" {
				if match, err := regexp.MatchString("@\\d", conf.Ipv6.Ipv6Reg); err == nil && match {
					num, err := strconv.Atoi(conf.Ipv6.Ipv6Reg[1:])
					if err == nil {
						if num > 0 {
							if num <= len(netInterface.Address) {
								return netInterface.Address[num-1]
							}
							util.Log("未找到第 %d 个IPv6地址! 将使用第一个IPv6地址", num)
							return netInterface.Address[0]
						}
						util.Log("IPv6匹配表达式 %s 不正确! 最小从1开始", conf.Ipv6.Ipv6Reg)
						return ""
					}
				}
				util.Log("IPv6将使用正则表达式 %s 进行匹配", conf.Ipv6.Ipv6Reg)
				for i := 0; i < len(netInterface.Address); i++ {
					matched, err := regexp.MatchString(conf.Ipv6.Ipv6Reg, netInterface.Address[i])
					if matched && err == nil {
						util.Log("匹配成功! 匹配到地址: %s", netInterface.Address[i])
						return netInterface.Address[i]
					}
				}
				util.Log("没有匹配到任何一个IPv6地址, 将使用第一个地址")
			}
			return netInterface.Address[0]
		}
	}

	util.Log("从网卡中获得IPv6失败! 网卡名: %s", conf.Ipv6.NetInterface)
	return ""
}

func (conf *DnsConfig) getIpv6AddrFromUrl() string {
	client := util.CreateBoundNoProxyHTTPClient("tcp6", "")
	urls := strings.Split(conf.Ipv6.URL, ",")
	for _, url := range urls {
		url = strings.TrimSpace(url)
		resp, err := client.Get(url)
		if err != nil {
			util.Log("通过接口获取IPv6失败! 接口地址: %s", url)
			util.Log("异常信息: %s", err)
			continue
		}

		defer resp.Body.Close()
		lr := io.LimitReader(resp.Body, 1024000)
		body, err := io.ReadAll(lr)
		if err != nil {
			util.Log("异常信息: %s", err)
			continue
		}
		result := findIPv6InText(string(body))
		if result == "" {
			util.Log("获取IPv6结果失败! 接口: %s ,返回值: %s", url, string(body))
		}
		return result
	}
	return ""
}

func (conf *DnsConfig) getIpv6AddrManual() string {
	return conf.Ipv6.Addr
}

func (conf *DnsConfig) GetIpv6Addr() (result string) {
	switch conf.Ipv6.GetType {
	case "netInterface":
		return conf.getIpv6AddrFromInterface()
	case "manual":
		return conf.getIpv6AddrManual()
	default:
		log.Println("IPv6's get IP method is unknown, supported: netInterface, manual")
		return ""
	}
}

func (conf *DnsConfig) GetHTTPClient() *http.Client {
	return util.CreateHTTPClient()
}
