package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"nas-partner/backend/internal/ddns/config"
	"nas-partner/backend/internal/ddns/util"
)

const zonesAPI = "https://api.cloudflare.com/client/v4/zones"

type Cloudflare struct {
	DNS        config.DNS
	Domains    config.Domains
	TTL        int
	httpClient *http.Client
}

type CloudflareZonesResp struct {
	CloudflareStatus
	Result []struct {
		ID     string
		Name   string
		Status string
		Paused bool
	}
}

type CloudflareRecordsResp struct {
	CloudflareStatus
	Result []CloudflareRecord
}

type CloudflareRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
	Comment string `json:"comment"`
}

type CloudflareStatus struct {
	Success  bool
	Messages []string
}

func (cf *Cloudflare) Init(dnsConf *config.DnsConfig, ipv4cache *util.IpCache, ipv6cache *util.IpCache) {
	cf.Domains.Ipv4Cache = ipv4cache
	cf.Domains.Ipv6Cache = ipv6cache
	cf.DNS = dnsConf.DNS
	cf.Domains.GetNewIp(dnsConf)
	if dnsConf.TTL == "" {
		cf.TTL = 1
	} else {
		ttl, err := strconv.Atoi(dnsConf.TTL)
		if err != nil {
			cf.TTL = 1
		} else {
			cf.TTL = ttl
		}
	}
	cf.httpClient = dnsConf.GetHTTPClient()
}

func (cf *Cloudflare) AddUpdateDomainRecords() config.Domains {
	cf.addUpdateDomainRecords("A")
	cf.addUpdateDomainRecords("AAAA")
	return cf.Domains
}

func (cf *Cloudflare) addUpdateDomainRecords(recordType string) {
	ipAddr, domains := cf.Domains.GetNewIpResult(recordType)
	if ipAddr == "" {
		return
	}

	for _, domain := range domains {
		result, err := cf.getZones(domain)
		if err != nil {
			util.Log("查询域名信息发生异常! %s", err)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "查询Zone失败: " + err.Error()
			return
		}

		if len(result.Result) == 0 {
			util.Log("在DNS服务商中未找到根域名: %s", domain.DomainName)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "未找到根域名: " + domain.DomainName
			return
		}

		params := url.Values{}
		params.Set("type", recordType)
		params.Set("name", domain.ToASCII())
		params.Set("per_page", "50")
		if c := domain.GetCustomParams().Get("comment"); c != "" {
			params.Set("comment", c)
		}

		zoneID := result.Result[0].ID

		var records CloudflareRecordsResp
		err = cf.request("GET", fmt.Sprintf(zonesAPI+"/%s/dns_records?%s", zoneID, params.Encode()), nil, &records)
		if err != nil {
			util.Log("查询域名信息发生异常! %s", err)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "查询DNS记录失败: " + err.Error()
			return
		}

		if !records.Success {
			util.Log("查询域名信息发生异常! %s", strings.Join(records.Messages, ", "))
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "查询DNS记录失败: " + strings.Join(records.Messages, ", ")
			return
		}

		if len(records.Result) > 0 {
			cf.modify(records, zoneID, domain, ipAddr)
		} else {
			cf.create(zoneID, domain, recordType, ipAddr)
		}
	}
}

func (cf *Cloudflare) create(zoneID string, domain *config.Domain, recordType string, ipAddr string) {
	record := &CloudflareRecord{
		Type:    recordType,
		Name:    domain.ToASCII(),
		Content: ipAddr,
		Proxied: false,
		TTL:     cf.TTL,
		Comment: domain.GetCustomParams().Get("comment"),
	}
	record.Proxied = domain.GetCustomParams().Get("proxied") == "true"
	var status CloudflareStatus
	err := cf.request("POST", fmt.Sprintf(zonesAPI+"/%s/dns_records", zoneID), record, &status)
	if err != nil {
		util.Log("新增域名解析 %s 失败! 异常信息: %s", domain, err)
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "创建失败: " + err.Error()
		return
	}
	if status.Success {
		util.Log("新增域名解析 %s 成功! IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
		domain.Detail = "已创建记录: " + ipAddr
	} else {
		util.Log("新增域名解析 %s 失败! 异常信息: %s", domain, strings.Join(status.Messages, ", "))
		domain.UpdateStatus = config.UpdatedFailed
		domain.Detail = "创建失败: " + strings.Join(status.Messages, ", ")
	}
}

func (cf *Cloudflare) modify(result CloudflareRecordsResp, zoneID string, domain *config.Domain, ipAddr string) {
	for _, record := range result.Result {
		if record.Content == ipAddr {
			util.Log("你的IP %s 没有变化, 域名 %s", ipAddr, domain)
			domain.UpdateStatus = config.UpdatedSuccess
			domain.Detail = "IP " + ipAddr + " 无变化"
			continue
		}
		var status CloudflareStatus
		record.Content = ipAddr
		record.TTL = cf.TTL
		if domain.GetCustomParams().Has("proxied") {
			record.Proxied = domain.GetCustomParams().Get("proxied") == "true"
		}
		err := cf.request("PUT", fmt.Sprintf(zonesAPI+"/%s/dns_records/%s", zoneID, record.ID), record, &status)
		if err != nil {
			util.Log("更新域名解析 %s 失败! 异常信息: %s", domain, err)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "更新失败: " + err.Error()
			return
		}
		if status.Success {
			util.Log("更新域名解析 %s 成功! IP: %s", domain, ipAddr)
			domain.UpdateStatus = config.UpdatedSuccess
			domain.Detail = "已更新到 " + ipAddr
		} else {
			util.Log("更新域名解析 %s 失败! 异常信息: %s", domain, strings.Join(status.Messages, ", "))
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "更新失败: " + strings.Join(status.Messages, ", ")
		}
	}
}

func (cf *Cloudflare) getZones(domain *config.Domain) (result CloudflareZonesResp, err error) {
	params := url.Values{}
	params.Set("name", domain.DomainName)
	params.Set("status", "active")
	params.Set("per_page", "50")
	err = cf.request("GET", zonesAPI+"?"+params.Encode(), nil, &result)
	return
}

func (cf *Cloudflare) request(method string, url string, data interface{}, result interface{}) (err error) {
	jsonStr := make([]byte, 0)
	if data != nil {
		jsonStr, _ = json.Marshal(data)
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+cf.DNS.Secret)
	req.Header.Set("Content-Type", "application/json")

	client := cf.httpClient
	resp, err := client.Do(req)
	err = util.GetHTTPResponse(resp, err, result)
	return
}
