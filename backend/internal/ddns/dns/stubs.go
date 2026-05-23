package dns

import (
	"net/http"

	"nas-partner/backend/internal/ddns/config"
	"nas-partner/backend/internal/ddns/util"
)

// BaiduCloud stub
type BaiduCloud struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *BaiduCloud) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *BaiduCloud) AddUpdateDomainRecords() config.Domains {
	util.Log("BaiduCloud 暂未实现")
	return s.Domains
}

// Porkbun stub
type Porkbun struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *Porkbun) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *Porkbun) AddUpdateDomainRecords() config.Domains {
	util.Log("Porkbun 暂未实现")
	return s.Domains
}

// GoDaddyDNS stub
type GoDaddyDNS struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *GoDaddyDNS) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *GoDaddyDNS) AddUpdateDomainRecords() config.Domains {
	util.Log("GoDaddyDNS 暂未实现")
	return s.Domains
}

// NameCheap stub
type NameCheap struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *NameCheap) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *NameCheap) AddUpdateDomainRecords() config.Domains {
	util.Log("NameCheap 暂未实现")
	return s.Domains
}

// NameSilo stub
type NameSilo struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *NameSilo) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *NameSilo) AddUpdateDomainRecords() config.Domains {
	util.Log("NameSilo 暂未实现")
	return s.Domains
}

// Vercel stub
type Vercel struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *Vercel) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *Vercel) AddUpdateDomainRecords() config.Domains {
	util.Log("Vercel 暂未实现")
	return s.Domains
}

// Dynadot stub
type Dynadot struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *Dynadot) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *Dynadot) AddUpdateDomainRecords() config.Domains {
	util.Log("Dynadot 暂未实现")
	return s.Domains
}

// Dynv6 stub
type Dynv6 struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *Dynv6) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *Dynv6) AddUpdateDomainRecords() config.Domains {
	util.Log("Dynv6 暂未实现")
	return s.Domains
}

// Spaceship stub
type Spaceship struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *Spaceship) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *Spaceship) AddUpdateDomainRecords() config.Domains {
	util.Log("Spaceship 暂未实现")
	return s.Domains
}

// Gcore stub
type Gcore struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *Gcore) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *Gcore) AddUpdateDomainRecords() config.Domains {
	util.Log("Gcore 暂未实现")
	return s.Domains
}

// NSOne stub
type NSOne struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *NSOne) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *NSOne) AddUpdateDomainRecords() config.Domains {
	util.Log("NSOne 暂未实现")
	return s.Domains
}

// ClouDNS stub
type ClouDNS struct{ DNS config.DNS; Domains config.Domains; httpClient *http.Client }

func (s *ClouDNS) Init(dnsConf *config.DnsConfig, ipv4cache, ipv6cache *util.IpCache) {
	s.DNS = dnsConf.DNS; s.Domains.Ipv4Cache = ipv4cache; s.Domains.Ipv6Cache = ipv6cache
	s.Domains.GetNewIp(dnsConf); s.httpClient = dnsConf.GetHTTPClient()
}

func (s *ClouDNS) AddUpdateDomainRecords() config.Domains {
	util.Log("ClouDNS 暂未实现")
	return s.Domains
}
