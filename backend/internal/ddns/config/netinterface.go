package config

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// AddressInfo 单个地址的详细信息
type AddressInfo struct {
	Address string `json:"address"`
	Type    string `json:"type"` // "permanent" 或 "temporary"
}

// NetInterface 本机网络
type NetInterface struct {
	Name          string        `json:"name"`
	Address       []string      `json:"address"`
	AddressDetail []AddressInfo `json:"address_detail,omitempty"`
}

// getIPv6AddressTypes 解析系统命令输出，获取 IPv6 地址类型（临时/永久）
func getIPv6AddressTypes() map[string]map[string]string {
	result := make(map[string]map[string]string)

	var out []byte
	var err error

	switch runtime.GOOS {
	case "darwin":
		out, err = exec.Command("ifconfig", "-a").Output()
		if err == nil {
			parseMacOSIfconfig(string(out), result)
		}
	case "linux":
		out, err = exec.Command("ip", "-6", "addr", "show").Output()
		if err == nil {
			parseLinuxIPAddr(string(out), result)
		}
	}

	return result
}

func parseMacOSIfconfig(output string, result map[string]map[string]string) {
	lines := strings.Split(output, "\n")
	var currentIface string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// 新网卡头："en0: flags=..."
		if line[0] != '\t' && line[0] != ' ' {
			currentIface = strings.SplitN(trimmed, ":", 2)[0]
			continue
		}

		// IPv6 地址行："inet6 <addr> prefixlen 64 autoconf secured"
		if !strings.HasPrefix(trimmed, "inet6") || strings.Contains(trimmed, "fe80") || strings.Contains(trimmed, "::1") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}
		addr := fields[1]
		addrType := "permanent"
		for _, flag := range fields[2:] {
			if flag == "temporary" {
				addrType = "temporary"
				break
			}
		}
		if _, ok := result[currentIface]; !ok {
			result[currentIface] = make(map[string]string)
		}
		result[currentIface][addr] = addrType
	}
}

func parseLinuxIPAddr(output string, result map[string]map[string]string) {
	lines := strings.Split(output, "\n")
	var currentIface string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// 网卡头："2: enp0s3: <...>"
		if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' && strings.Contains(line, ": ") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				rest := parts[1]
				if colonIdx := strings.Index(rest, ":"); colonIdx >= 0 {
					currentIface = rest[:colonIdx]
				} else if spaceIdx := strings.Index(rest, " "); spaceIdx >= 0 {
					currentIface = rest[:spaceIdx]
				}
			}
			continue
		}

		// IPv6 地址行："inet6 240e::1/64 scope global temporary dynamic"
		if !strings.HasPrefix(trimmed, "inet6") || strings.Contains(trimmed, "fe80") || strings.Contains(trimmed, "::1") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) < 2 || fields[0] != "inet6" {
			continue
		}
		addr := strings.SplitN(fields[1], "/", 2)[0]
		addrType := "permanent"
		for _, flag := range fields[2:] {
			if flag == "temporary" {
				addrType = "temporary"
				break
			}
		}
		if _, ok := result[currentIface]; !ok {
			result[currentIface] = make(map[string]string)
		}
		result[currentIface][addr] = addrType
	}
}

// GetNetInterface 获得网卡地址
func GetNetInterface() (ipv4NetInterfaces []NetInterface, ipv6NetInterfaces []NetInterface, err error) {
	allNetInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces failed, err:", err.Error())
		return ipv4NetInterfaces, ipv6NetInterfaces, err
	}

	_, ipv6Unicast, _ := net.ParseCIDR("2000::/3")
	ipv6AddrTypes := getIPv6AddressTypes()

	for i := 0; i < len(allNetInterfaces); i++ {
		if (allNetInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := allNetInterfaces[i].Addrs()
			ipv4 := []string{}
			ipv6 := []string{}
			ipv6Detail := []AddressInfo{}

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
					_, bits := ipnet.Mask.Size()
					addrStr := ipnet.IP.String()

					if bits == 128 && ipv6Unicast.Contains(ipnet.IP) {
						addrType := "permanent"
						if ifaceTypes, ok := ipv6AddrTypes[allNetInterfaces[i].Name]; ok {
							if t, ok := ifaceTypes[addrStr]; ok {
								addrType = t
							}
						}
						ipv6Detail = append(ipv6Detail, AddressInfo{
							Address: addrStr,
							Type:    addrType,
						})
					}
					if bits == 32 {
						ipv4 = append(ipv4, addrStr)
					}
				}
			}

			if len(ipv6Detail) > 0 {
				// 永久地址排在临时地址前面
				sort.SliceStable(ipv6Detail, func(i, j int) bool {
					return ipv6Detail[i].Type < ipv6Detail[j].Type // "permanent" < "temporary"
				})
				ipv6 = make([]string, len(ipv6Detail))
				for i, d := range ipv6Detail {
					ipv6[i] = d.Address
				}

				ipv6NetInterfaces = append(ipv6NetInterfaces, NetInterface{
					Name:          allNetInterfaces[i].Name,
					Address:       ipv6,
					AddressDetail: ipv6Detail,
				})
			}

			if len(ipv4) > 0 {
				ipv4NetInterfaces = append(ipv4NetInterfaces, NetInterface{
					Name:    allNetInterfaces[i].Name,
					Address: ipv4,
				})
			}
		}
	}

	return ipv4NetInterfaces, ipv6NetInterfaces, nil
}
