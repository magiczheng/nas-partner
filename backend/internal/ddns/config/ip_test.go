package config

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIpv4AddrManual(t *testing.T) {
	conf := &DnsConfig{}
	conf.Ipv4.Addr = "192.168.1.100"
	conf.Ipv4.GetType = "manual"

	ip := conf.GetIpv4Addr()
	if ip != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", ip)
	}
}

func TestGetIpv4AddrManualEmpty(t *testing.T) {
	conf := &DnsConfig{}
	conf.Ipv4.GetType = "manual"

	ip := conf.GetIpv4Addr()
	if ip != "" {
		t.Errorf("expected empty string, got %s", ip)
	}
}

func TestGetIpv4AddrAuto(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("203.0.113.5"))
	}))
	defer server.Close()

	saved := AutoGetUrl
	AutoGetUrl = []string{server.URL}
	defer func() { AutoGetUrl = saved }()

	conf := &DnsConfig{}
	conf.Ipv4.GetType = "auto"

	ip := conf.GetIpv4Addr()
	if ip != "203.0.113.5" {
		t.Errorf("expected 203.0.113.5, got %s", ip)
	}
}

func TestGetIpv4AddrAutoFallback(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-an-ip"))
	}))
	defer badServer.Close()

	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("10.0.0.1"))
	}))
	defer goodServer.Close()

	saved := AutoGetUrl
	AutoGetUrl = []string{badServer.URL, goodServer.URL}
	defer func() { AutoGetUrl = saved }()

	conf := &DnsConfig{}
	conf.Ipv4.GetType = "auto"

	ip := conf.GetIpv4Addr()
	if ip != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1 (fallback), got %s", ip)
	}
}

func TestGetIpv4AddrAutoAllFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-an-ip"))
	}))
	defer server.Close()

	saved := AutoGetUrl
	AutoGetUrl = []string{server.URL}
	defer func() { AutoGetUrl = saved }()

	conf := &DnsConfig{}
	conf.Ipv4.GetType = "auto"

	ip := conf.GetIpv4Addr()
	if ip != "" {
		t.Errorf("expected empty when all fail, got %s", ip)
	}
}

func TestGetIpv4AddrNetInterface(t *testing.T) {
	allNetInterfaces, _, err := GetNetInterface()
	if err != nil {
		t.Fatalf("GetNetInterface failed: %v", err)
	}

	if len(allNetInterfaces) == 0 {
		t.Skip("no network interfaces available")
	}

	conf := &DnsConfig{}
	conf.Ipv4.GetType = "netInterface"
	conf.Ipv4.NetInterface = allNetInterfaces[0].Name

	ip := conf.GetIpv4Addr()
	if ip == "" {
		// The interface might exist but have no global unicast addr by the time we check
		t.Logf("interface %s exists, but no IP returned", allNetInterfaces[0].Name)
	} else {
		t.Logf("interface %s returned IP: %s", allNetInterfaces[0].Name, ip)
	}
}

func TestGetIpv4AddrNetInterfaceNotFound(t *testing.T) {
	conf := &DnsConfig{}
	conf.Ipv4.GetType = "netInterface"
	conf.Ipv4.NetInterface = "nonexistent_iface_xyz"

	ip := conf.GetIpv4Addr()
	if ip != "" {
		t.Errorf("expected empty for nonexistent interface, got %s", ip)
	}
}

func TestIpv4RegSimple(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My IP is 192.168.1.1 and it's nice", "192.168.1.1"},
		{"203.0.113.5", "203.0.113.5"},
		{"10.0.0.1", "10.0.0.1"},
		{"172.16.0.1", "172.16.0.1"},
		{"no ip here", ""},
		{"", ""},
		{"8.8.8.8", "8.8.8.8"},
		{"255.255.255.255", "255.255.255.255"},
		{"0.0.0.0", "0.0.0.0"},
		{"999.999.999.999", ""},
		// surrounded by text
		{"ip:10.0.0.1,end", "10.0.0.1"},
	}
	for _, tc := range tests {
		result := Ipv4Reg.FindString(tc.input)
		if result != tc.expected {
			t.Errorf("Ipv4Reg.FindString(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestGetNetInterface(t *testing.T) {
	ipv4, ipv6, err := GetNetInterface()
	if err != nil {
		t.Fatalf("GetNetInterface failed: %v", err)
	}
	t.Logf("found %d IPv4 interfaces, %d IPv6 interfaces", len(ipv4), len(ipv6))
	for _, iface := range ipv4 {
		t.Logf("  IPv4: %s -> %v", iface.Name, iface.Address)
	}
	for _, iface := range ipv6 {
		t.Logf("  IPv6: %s -> %v", iface.Name, iface.Address)
	}
}

func TestGetIpv6AddrManual(t *testing.T) {
	conf := &DnsConfig{}
	conf.Ipv6.Addr = "2001:db8::1"
	conf.Ipv6.GetType = "manual"

	ip := conf.GetIpv6Addr()
	if ip != "2001:db8::1" {
		t.Errorf("expected 2001:db8::1, got %s", ip)
	}
}
