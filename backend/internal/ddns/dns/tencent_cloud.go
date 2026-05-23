package dns

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"nas-partner/backend/internal/ddns/config"
	"nas-partner/backend/internal/ddns/util"
)

const (
	tencentCloudEndPoint = "https://dnspod.tencentcloudapi.com"
	tencentCloudVersion  = "2021-03-23"
)

type TencentCloud struct {
	DNS        config.DNS
	Domains    config.Domains
	TTL        int
	httpClient *http.Client
}

type TencentCloudRecord struct {
	Domain     string `json:"Domain"`
	SubDomain  string `json:"SubDomain,omitempty"`
	Subdomain  string `json:"Subdomain,omitempty"`
	RecordType string `json:"RecordType"`
	RecordLine string `json:"RecordLine"`
	Value      string `json:"Value,omitempty"`
	RecordId   int64  `json:"RecordId,omitempty"`
	TTL        int    `json:"TTL,omitempty"`
}

type TencentCloudRecordListsResp struct {
	TencentCloudStatus
	Response struct {
		RecordCountInfo struct {
			TotalCount int `json:"TotalCount"`
		} `json:"RecordCountInfo"`
		RecordList []TencentCloudRecord `json:"RecordList"`
	}
}

type TencentCloudStatus struct {
	Response struct {
		Error struct {
			Code    string
			Message string
		}
	}
}

func (tc *TencentCloud) Init(dnsConf *config.DnsConfig, ipv4cache *util.IpCache, ipv6cache *util.IpCache) {
	tc.Domains.Ipv4Cache = ipv4cache
	tc.Domains.Ipv6Cache = ipv6cache
	tc.DNS = dnsConf.DNS
	tc.Domains.GetNewIp(dnsConf)
	if dnsConf.TTL == "" {
		tc.TTL = 600
	} else {
		ttl, err := strconv.Atoi(dnsConf.TTL)
		if err != nil {
			tc.TTL = 600
		} else {
			tc.TTL = ttl
		}
	}
	tc.httpClient = dnsConf.GetHTTPClient()
}

func (tc *TencentCloud) AddUpdateDomainRecords() config.Domains {
	tc.addUpdateDomainRecords("A")
	tc.addUpdateDomainRecords("AAAA")
	return tc.Domains
}

func (tc *TencentCloud) addUpdateDomainRecords(recordType string) {
	ipAddr, domains := tc.Domains.GetNewIpResult(recordType)
	if ipAddr == "" {
		return
	}

	for _, domain := range domains {
		result, err := tc.getRecordList(domain, recordType)
		if err != nil {
			util.Log("查询域名信息发生异常! %s", err)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "查询失败: " + err.Error()
			return
		}

		if result.Response.RecordCountInfo.TotalCount > 0 {
			recordSelected := result.Response.RecordList[0]
			params := domain.GetCustomParams()
			if params.Has("RecordId") {
				for i := 0; i < result.Response.RecordCountInfo.TotalCount; i++ {
					if strconv.FormatInt(result.Response.RecordList[i].RecordId, 10) == params.Get("RecordId") {
						recordSelected = result.Response.RecordList[i]
					}
				}
			}
			tc.modify(recordSelected, domain, recordType, ipAddr)
		} else {
			tc.create(domain, recordType, ipAddr)
		}
	}
}

func (tc *TencentCloud) create(domain *config.Domain, recordType string, ipAddr string) {
	record := &TencentCloudRecord{
		Domain:     domain.DomainName,
		SubDomain:  domain.GetSubDomain(),
		RecordType: recordType,
		RecordLine: tc.getRecordLine(domain),
		Value:      ipAddr,
		TTL:        tc.TTL,
	}

	var status TencentCloudStatus
	err := tc.request("CreateRecord", record, &status)
	if err != nil {
		util.Log("新增域名解析 %s 失败! 异常信息: %s", domain, err)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "创建失败: " + err.Error()
		return
	}

	if status.Response.Error.Code == "" {
		util.Log("新增域名解析 %s 成功! IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "已创建记录: " + ipAddr
	} else {
		util.Log("新增域名解析 %s 失败! 异常信息: %s", domain, status.Response.Error.Message)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "创建失败: " + status.Response.Error.Message
	}
}

func (tc *TencentCloud) modify(record TencentCloudRecord, domain *config.Domain, recordType string, ipAddr string) {
	if record.Value == ipAddr {
		util.Log("你的IP %s 没有变化, 域名 %s", ipAddr, domain)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "IP " + ipAddr + " 无变化"
		return
	}

	record.Domain = domain.DomainName
	record.SubDomain = domain.GetSubDomain()
	record.RecordType = recordType
	record.RecordLine = tc.getRecordLine(domain)
	record.Value = ipAddr
	record.TTL = tc.TTL

	var status TencentCloudStatus
	err := tc.request("ModifyRecord", record, &status)
	if err != nil {
		util.Log("更新域名解析 %s 失败! 异常信息: %s", domain, err)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "更新失败: " + err.Error()
		return
	}

	if status.Response.Error.Code == "" {
		util.Log("更新域名解析 %s 成功! IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "已更新到 " + ipAddr
	} else {
		util.Log("更新域名解析 %s 失败! 异常信息: %s", domain, status.Response.Error.Message)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "更新失败: " + status.Response.Error.Message
	}
}

func (tc *TencentCloud) getRecordList(domain *config.Domain, recordType string) (result TencentCloudRecordListsResp, err error) {
	record := TencentCloudRecord{
		Domain:     domain.DomainName,
		Subdomain:  domain.GetSubDomain(),
		RecordType: recordType,
		RecordLine: tc.getRecordLine(domain),
	}
	err = tc.request("DescribeRecordList", record, &result)
	return
}

func (tc *TencentCloud) getRecordLine(domain *config.Domain) string {
	if domain.GetCustomParams().Has("RecordLine") {
		return domain.GetCustomParams().Get("RecordLine")
	}
	return "默认"
}

func (tc *TencentCloud) request(action string, data interface{}, result interface{}) (err error) {
	jsonStr := make([]byte, 0)
	if data != nil {
		jsonStr, _ = json.Marshal(data)
	}
	req, err := http.NewRequest("POST", tencentCloudEndPoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Version", tencentCloudVersion)

	util.TencentCloudSigner(tc.DNS.ID, tc.DNS.Secret, req, action, string(jsonStr), util.DnsPod)

	client := tc.httpClient
	resp, err := client.Do(req)
	err = util.GetHTTPResponse(resp, err, result)
	return
}
