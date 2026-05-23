package dns

import (
	"net/http"
	"net/url"

	"nas-partner/backend/internal/ddns/config"
	"nas-partner/backend/internal/ddns/util"
)

const (
	recordListAPI   = "https://dnsapi.cn/Record.List"
	recordModifyURL = "https://dnsapi.cn/Record.Modify"
	recordCreateAPI = "https://dnsapi.cn/Record.Create"
)

type Dnspod struct {
	DNS        config.DNS
	Domains    config.Domains
	TTL        string
	httpClient *http.Client
}

type DnspodRecord struct {
	ID      string
	Name    string
	Type    string
	Value   string
	Enabled string
}

type DnspodRecordListResp struct {
	DnspodStatus
	Records []DnspodRecord
}

type DnspodStatus struct {
	Status struct {
		Code    string
		Message string
	}
}

func (dnspod *Dnspod) Init(dnsConf *config.DnsConfig, ipv4cache *util.IpCache, ipv6cache *util.IpCache) {
	dnspod.Domains.Ipv4Cache = ipv4cache
	dnspod.Domains.Ipv6Cache = ipv6cache
	dnspod.DNS = dnsConf.DNS
	dnspod.Domains.GetNewIp(dnsConf)
	if dnsConf.TTL == "" {
		dnspod.TTL = "600"
	} else {
		dnspod.TTL = dnsConf.TTL
	}
	dnspod.httpClient = dnsConf.GetHTTPClient()
}

func (dnspod *Dnspod) AddUpdateDomainRecords() config.Domains {
	dnspod.addUpdateDomainRecords("A")
	dnspod.addUpdateDomainRecords("AAAA")
	return dnspod.Domains
}

func (dnspod *Dnspod) addUpdateDomainRecords(recordType string) {
	ipAddr, domains := dnspod.Domains.GetNewIpResult(recordType)
	if ipAddr == "" {
		return
	}

	for _, domain := range domains {
		result, err := dnspod.getRecordList(domain, recordType)
		if err != nil {
			util.Log("查询域名信息发生异常! %s", err)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "查询失败: " + err.Error()
			return
		}

		if len(result.Records) > 0 {
			recordSelected := result.Records[0]
			params := domain.GetCustomParams()
			if params.Has("record_id") {
				for i := 0; i < len(result.Records); i++ {
					if result.Records[i].ID == params.Get("record_id") {
						recordSelected = result.Records[i]
					}
				}
			}
			dnspod.modify(recordSelected, domain, recordType, ipAddr)
		} else {
			dnspod.create(domain, recordType, ipAddr)
		}
	}
}

func (dnspod *Dnspod) create(domain *config.Domain, recordType string, ipAddr string) {
	params := domain.GetCustomParams()
	params.Set("login_token", dnspod.DNS.ID+","+dnspod.DNS.Secret)
	params.Set("domain", domain.DomainName)
	params.Set("sub_domain", domain.GetSubDomain())
	params.Set("record_type", recordType)
	params.Set("value", ipAddr)
	params.Set("ttl", dnspod.TTL)
	params.Set("format", "json")
	if !params.Has("record_line") {
		params.Set("record_line", "默认")
	}

	status, err := dnspod.request(recordCreateAPI, params)
	if err != nil {
		util.Log("新增域名解析 %s 失败! 异常信息: %s", domain, err)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "创建失败: " + err.Error()
		return
	}

	if status.Status.Code == "1" {
		util.Log("新增域名解析 %s 成功! IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "已创建记录: " + ipAddr
	} else {
		util.Log("新增域名解析 %s 失败! 异常信息: %s", domain, status.Status.Message)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "创建失败: " + status.Status.Message
	}
}

func (dnspod *Dnspod) modify(record DnspodRecord, domain *config.Domain, recordType string, ipAddr string) {
	if record.Value == ipAddr {
		util.Log("你的IP %s 没有变化, 域名 %s", ipAddr, domain)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "IP " + ipAddr + " 无变化"
		return
	}

	params := domain.GetCustomParams()
	params.Set("login_token", dnspod.DNS.ID+","+dnspod.DNS.Secret)
	params.Set("domain", domain.DomainName)
	params.Set("sub_domain", domain.GetSubDomain())
	params.Set("record_type", recordType)
	params.Set("value", ipAddr)
	params.Set("ttl", dnspod.TTL)
	params.Set("format", "json")
	params.Set("record_id", record.ID)
	if !params.Has("record_line") {
		params.Set("record_line", "默认")
	}

	status, err := dnspod.request(recordModifyURL, params)
	if err != nil {
		util.Log("更新域名解析 %s 失败! 异常信息: %s", domain, err)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "更新失败: " + err.Error()
		return
	}

	if status.Status.Code == "1" {
		util.Log("更新域名解析 %s 成功! IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "已更新到 " + ipAddr
	} else {
		util.Log("更新域名解析 %s 失败! 异常信息: %s", domain, status.Status.Message)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "更新失败: " + status.Status.Message
	}
}

func (dnspod *Dnspod) request(apiAddr string, values url.Values) (status DnspodStatus, err error) {
	client := dnspod.httpClient
	resp, err := client.PostForm(apiAddr, values)
	err = util.GetHTTPResponse(resp, err, &status)
	return
}

func (dnspod *Dnspod) getRecordList(domain *config.Domain, typ string) (result DnspodRecordListResp, err error) {
	params := domain.GetCustomParams()
	params.Set("login_token", dnspod.DNS.ID+","+dnspod.DNS.Secret)
	params.Set("domain", domain.DomainName)
	params.Set("record_type", typ)
	params.Set("sub_domain", domain.GetSubDomain())
	params.Set("format", "json")

	client := dnspod.httpClient
	resp, err := client.PostForm(recordListAPI, params)
	err = util.GetHTTPResponse(resp, err, &result)
	return
}
