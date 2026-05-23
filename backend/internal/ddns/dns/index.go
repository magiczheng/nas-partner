package dns

import (
	"nas-partner/backend/internal/ddns/config"
	ddnsmodel "nas-partner/backend/internal/ddns/model"
	"nas-partner/backend/internal/ddns/util"
)

// DNS interface
type DNS interface {
	Init(dnsConf *config.DnsConfig, ipv4cache *util.IpCache, ipv6cache *util.IpCache)
	AddUpdateDomainRecords() (domains config.Domains)
}

// RunOnce 运行一次DDNS更新
func RunOnce(cfg *ddnsmodel.DDNSConfig) config.Domains {
	dc := modelToDnsConfig(cfg)

	ipv4cache := &util.IpCache{}
	ipv6cache := &util.IpCache{}

	var dnsSelected DNS
	switch dc.DNS.Name {
	case "alidns":
		dnsSelected = &Alidns{}
	case "cloudflare":
		dnsSelected = &Cloudflare{}
	case "dnspod":
		dnsSelected = &Dnspod{}
	case "tencentcloud":
		dnsSelected = &TencentCloud{}
	case "huaweicloud":
		dnsSelected = &Huaweicloud{}
	case "callback":
		dnsSelected = &Callback{}
	case "baiducloud":
		dnsSelected = &BaiduCloud{}
	case "porkbun":
		dnsSelected = &Porkbun{}
	case "godaddy":
		dnsSelected = &GoDaddyDNS{}
	case "namecheap":
		dnsSelected = &NameCheap{}
	case "namesilo":
		dnsSelected = &NameSilo{}
	case "vercel":
		dnsSelected = &Vercel{}
	case "dynadot":
		dnsSelected = &Dynadot{}
	case "dynv6":
		dnsSelected = &Dynv6{}
	case "spaceship":
		dnsSelected = &Spaceship{}
	case "gcore":
		dnsSelected = &Gcore{}
	case "nsone":
		dnsSelected = &NSOne{}
	case "cloudns":
		dnsSelected = &ClouDNS{}
	default:
		dnsSelected = &Alidns{}
	}
	dnsSelected.Init(&dc, ipv4cache, ipv6cache)
	return dnsSelected.AddUpdateDomainRecords()
}

func modelToDnsConfig(cfg *ddnsmodel.DDNSConfig) config.DnsConfig {
	var dc config.DnsConfig
	dc.Name = cfg.Name
	dc.DNS.Name = cfg.DNSProvider
	dc.DNS.ID = cfg.AccessKeyID
	dc.DNS.Secret = cfg.AccessKeySecret
	dc.DNS.ExtParam = cfg.ExtraParams
	dc.TTL = cfg.TTL

	dc.Ipv4.Enable = cfg.IPv4Enabled
	dc.Ipv4.GetType = cfg.IPv4GetType
	dc.Ipv4.URL = cfg.IPv4URL
	dc.Ipv4.NetInterface = cfg.IPv4NetInterface
	dc.Ipv4.Cmd = cfg.IPv4Cmd
	dc.Ipv4.Addr = cfg.IPv4Addr
	dc.Ipv4.Domains = cfg.Domains

	dc.Ipv6.Enable = cfg.IPv6Enabled
	dc.Ipv6.GetType = cfg.IPv6GetType
	dc.Ipv6.URL = cfg.IPv6URL
	dc.Ipv6.NetInterface = cfg.IPv6NetInterface
	dc.Ipv6.Cmd = cfg.IPv6Cmd
	dc.Ipv6.Addr = cfg.IPv6Addr
	dc.Ipv6.Domains = cfg.Domains

	return dc
}

// GetDomains Status
func GetDomainsStatus(domains config.Domains) (config.UpdateStatusType, config.UpdateStatusType) {
	return getDomainsStatus(domains.Ipv4Domains), getDomainsStatus(domains.Ipv6Domains)
}

func getDomainsStatus(domains []*config.Domain) config.UpdateStatusType {
	successNum := 0
	for _, v := range domains {
		switch v.UpdateStatus {
		case config.UpdatedFailed:
			return config.UpdatedFailed
		case config.UpdatedSuccess:
			successNum++
		}
	}
	if successNum > 0 {
		return config.UpdatedSuccess
	}
	return config.UpdatedNothing
}
