package dns

import (
	"bytes"
	"net/http"
	"net/url"

	"nas-partner/backend/internal/ddns/config"
	"nas-partner/backend/internal/ddns/util"
)

const alidnsEndpoint = "https://alidns.aliyuncs.com/"

type Alidns struct {
	DNS        config.DNS
	Domains    config.Domains
	TTL        string
	httpClient *http.Client
}

type AlidnsRecord struct {
	DomainName string
	RecordID   string
	Value      string
}

type AlidnsSubDomainRecords struct {
	TotalCount    int
	DomainRecords struct {
		Record []AlidnsRecord
	}
}

type AlidnsResp struct {
	RecordID  string
	RequestID string
}

func (ali *Alidns) Init(dnsConf *config.DnsConfig, ipv4cache *util.IpCache, ipv6cache *util.IpCache) {
	ali.Domains.Ipv4Cache = ipv4cache
	ali.Domains.Ipv6Cache = ipv6cache
	ali.DNS = dnsConf.DNS
	ali.Domains.GetNewIp(dnsConf)
	if dnsConf.TTL == "" {
		ali.TTL = "600"
	} else {
		ali.TTL = dnsConf.TTL
	}
	ali.httpClient = dnsConf.GetHTTPClient()
}

func (ali *Alidns) AddUpdateDomainRecords() config.Domains {
	ali.addUpdateDomainRecords("A")
	ali.addUpdateDomainRecords("AAAA")
	return ali.Domains
}

func (ali *Alidns) addUpdateDomainRecords(recordType string) {
	ipAddr, domains := ali.Domains.GetNewIpResult(recordType)
	if ipAddr == "" {
		return
	}

	for _, domain := range domains {
		var records AlidnsSubDomainRecords
		params := domain.GetCustomParams()
		params.Set("Action", "DescribeSubDomainRecords")
		params.Set("DomainName", domain.DomainName)
		params.Set("SubDomain", domain.GetFullDomain())
		params.Set("Type", recordType)
		err := ali.request(params, &records)

		if err != nil {
			util.Log("查询域名信息发生异常! %s", err)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "查询失败: " + err.Error()
			return
		}

		if records.TotalCount > 0 {
			recordSelected := records.DomainRecords.Record[0]
			if params.Has("RecordId") {
				for i := 0; i < len(records.DomainRecords.Record); i++ {
					if records.DomainRecords.Record[i].RecordID == params.Get("RecordId") {
						recordSelected = records.DomainRecords.Record[i]
					}
				}
			}
			ali.modify(recordSelected, domain, recordType, ipAddr)
		} else {
			ali.create(domain, recordType, ipAddr)
		}
	}
}

func (ali *Alidns) create(domain *config.Domain, recordType string, ipAddr string) {
	params := domain.GetCustomParams()
	params.Set("Action", "AddDomainRecord")
	params.Set("DomainName", domain.DomainName)
	params.Set("RR", domain.GetSubDomain())
	params.Set("Type", recordType)
	params.Set("Value", ipAddr)
	params.Set("TTL", ali.TTL)

	var result AlidnsResp
	err := ali.request(params, &result)
	if err != nil {
		util.Log("新增域名解析 %s 失败! 异常信息: %s", domain, err)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "创建失败: " + err.Error()
		return
	}

	if result.RecordID != "" {
		util.Log("新增域名解析 %s 成功! IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "已创建记录: " + ipAddr
	} else {
		util.Log("新增域名解析 %s 失败! 异常信息: %s", domain, "返回RecordId为空")
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "创建失败: 返回RecordId为空"
	}
}

func (ali *Alidns) modify(recordSelected AlidnsRecord, domain *config.Domain, recordType string, ipAddr string) {
	if recordSelected.Value == ipAddr {
		util.Log("你的IP %s 没有变化, 域名 %s", ipAddr, domain)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "IP " + ipAddr + " 无变化"
		return
	}

	params := domain.GetCustomParams()
	params.Set("Action", "UpdateDomainRecord")
	params.Set("RR", domain.GetSubDomain())
	params.Set("RecordId", recordSelected.RecordID)
	params.Set("Type", recordType)
	params.Set("Value", ipAddr)
	params.Set("TTL", ali.TTL)

	var result AlidnsResp
	err := ali.request(params, &result)
	if err != nil {
		util.Log("更新域名解析 %s 失败! 异常信息: %s", domain, err)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "更新失败: " + err.Error()
		return
	}

	if result.RecordID != "" {
		util.Log("更新域名解析 %s 成功! IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "已更新到 " + ipAddr
	} else {
		util.Log("更新域名解析 %s 失败! 异常信息: %s", domain, "返回RecordId为空")
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "更新失败: 返回RecordId为空"
	}
}

func (ali *Alidns) request(params url.Values, result interface{}) (err error) {
	method := http.MethodGet
	util.AliyunSigner(ali.DNS.ID, ali.DNS.Secret, &params, method, "2015-01-09")

	req, err := http.NewRequest(method, alidnsEndpoint, bytes.NewBuffer(nil))
	if err != nil {
		return
	}
	req.URL.RawQuery = params.Encode()

	client := ali.httpClient
	resp, err := client.Do(req)
	err = util.GetHTTPResponse(resp, err, result)
	return
}
