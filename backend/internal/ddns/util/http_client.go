package util

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"
)

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

var defaultTransport = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	DialContext:           dialer.DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func CreateHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: defaultTransport,
	}
}

func GetLocalAddrFromInterface(ifaceName string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.IsGlobalUnicast() {
			return ipNet.IP.String(), nil
		}
	}
	return "", nil
}

func CreateHTTPClientWithInterface(ifaceName string) *http.Client {
	if ifaceName == "" {
		return CreateHTTPClient()
	}
	localIP, err := GetLocalAddrFromInterface(ifaceName)
	if err != nil {
		log.Printf("绑定网卡失败, 将使用默认网卡. 网卡: %s, 错误: %v", ifaceName, err)
		return CreateHTTPClient()
	}
	localAddr := &net.TCPAddr{IP: net.ParseIP(localIP)}
	boundDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		LocalAddr: localAddr,
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           boundDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

var noProxyTcp4Transport = &http.Transport{
	DisableKeepAlives: true,
	DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp4", address)
	},
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

var noProxyTcp6Transport = &http.Transport{
	DisableKeepAlives: true,
	DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp6", address)
	},
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func CreateNoProxyHTTPClient(network string) *http.Client {
	if network == "tcp6" {
		return &http.Client{
			Timeout:   30 * time.Second,
			Transport: noProxyTcp6Transport,
		}
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: noProxyTcp4Transport,
	}
}

func CreateBoundNoProxyHTTPClient(network, ifaceName string) *http.Client {
	if ifaceName == "" {
		return CreateNoProxyHTTPClient(network)
	}
	localIP, err := getLocalAddrFromInterfaceByNetwork(ifaceName, network)
	if err != nil {
		log.Printf("绑定网卡失败, 将使用默认网卡. 网卡: %s, 错误: %v", ifaceName, err)
		return CreateNoProxyHTTPClient(network)
	}
	localAddrIP := net.ParseIP(localIP)
	if localAddrIP == nil {
		log.Printf("绑定网卡失败, 将使用默认网卡. 网卡: %s, 网络: %s, 错误: 本地IP无效: %s", ifaceName, network, localIP)
		return CreateNoProxyHTTPClient(network)
	}
	localAddr := &net.TCPAddr{IP: localAddrIP}
	if network == "tcp6" && ifaceName != "" {
		localAddr.Zone = ifaceName
	}
	boundDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		LocalAddr: localAddr,
	}
	transport := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, _, address string) (net.Conn, error) {
			return boundDialer.DialContext(ctx, network, address)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

func getLocalAddrFromInterfaceByNetwork(ifaceName, network string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || !ipNet.IP.IsGlobalUnicast() {
			continue
		}
		if isIPMatchedNetwork(ipNet.IP, network) {
			return ipNet.IP.String(), nil
		}
	}
	return "", nil
}

func isIPMatchedNetwork(ip net.IP, network string) bool {
	switch network {
	case "tcp4":
		return ip.To4() != nil
	case "tcp6":
		return ip.To16() != nil && ip.To4() == nil
	default:
		return true
	}
}
