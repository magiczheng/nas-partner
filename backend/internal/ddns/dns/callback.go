package dns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nas-partner/backend/internal/ddns/config"
	"nas-partner/backend/internal/ddns/util"
)

type Callback struct {
	DNS        config.DNS
	Domains    config.Domains
	TTL        string
	lastIpv4   string
	lastIpv6   string
	httpClient *http.Client
	ipv4Enable bool
	ipv6Enable bool
}

func (cb *Callback) Init(dnsConf *config.DnsConfig, ipv4cache *util.IpCache, ipv6cache *util.IpCache) {
	cb.Domains.Ipv4Cache = ipv4cache
	cb.Domains.Ipv6Cache = ipv6cache
	cb.lastIpv4 = ipv4cache.Addr
	cb.lastIpv6 = ipv6cache.Addr
	cb.ipv4Enable = dnsConf.Ipv4.Enable
	cb.ipv6Enable = dnsConf.Ipv6.Enable

	cb.DNS = dnsConf.DNS
	cb.Domains.GetNewIp(dnsConf)
	if dnsConf.TTL == "" {
		cb.TTL = "600"
	} else {
		cb.TTL = dnsConf.TTL
	}
	cb.httpClient = dnsConf.GetHTTPClient()
}

func (cb *Callback) AddUpdateDomainRecords() config.Domains {
	if cb.ipv4Enable {
		cb.addUpdateDomainRecords("A")
	}
	if cb.ipv6Enable {
		cb.addUpdateDomainRecords("AAAA")
	}
	return cb.Domains
}

func (cb *Callback) addUpdateDomainRecords(recordType string) {
	ipAddr, domains := cb.Domains.GetNewIpResult(recordType)
	if ipAddr == "" {
		return
	}

	if recordType == "A" {
		if cb.lastIpv4 == ipAddr {
			util.Log("你的IPv4未变化, 未触发 %s 请求", "Callback")
			return
		}
	} else {
		if cb.lastIpv6 == ipAddr {
			util.Log("你的IPv6未变化, 未触发 %s 请求", "Callback")
			return
		}
	}

	for _, domain := range domains {
		method := "GET"
		postPara := ""
		contentType := "application/x-www-form-urlencoded"
		if cb.DNS.Secret != "" {
			method = "POST"
			postPara = cb.replacePara(cb.DNS.Secret, ipAddr, domain, recordType, cb.TTL)
			if json.Valid([]byte(postPara)) {
				contentType = "application/json"
			}
		}
		requestURL := cb.replacePara(cb.DNS.ID, ipAddr, domain, recordType, cb.TTL)
		u, err := url.Parse(requestURL)
		if err != nil {
			util.Log("Callback的URL不正确")
			domain.Detail = "Callback URL 格式错误"
			return
		}
		req, err := http.NewRequest(method, u.String(), strings.NewReader(postPara))
		if err != nil {
			util.Log("异常信息: %v", err)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "请求失败: " + err.Error()
			return
		}
		req.Header.Add("content-type", contentType)

		resp, err := cb.httpClient.Do(req)
		body, err := util.GetHTTPResponseOrg(resp, err)
		if err == nil {
			util.Log("Callback调用成功, 域名: %s, IP: %s, 返回数据: %s", domain, ipAddr, string(body))
			domain.UpdateStatus = config.UpdatedSuccess
			domain.Detail = "已更新到 " + ipAddr
		} else {
			util.Log("Callback调用失败, 异常信息: %v", err)
			domain.UpdateStatus = config.UpdatedFailed
			domain.Detail = "调用失败: " + err.Error()
		}
	}
}

func (cb *Callback) replacePara(orgPara, ipAddr string, domain *config.Domain, recordType string, ttl string) string {
	params := map[string]string{
		"ip":         ipAddr,
		"domain":     domain.String(),
		"recordType": recordType,
		"ttl":        ttl,
		"ipv4Addr":   cb.Domains.Ipv4Addr,
		"ipv6Addr":   cb.Domains.Ipv6Addr,
		"timestamp":  strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}

	for k, v := range domain.GetCustomParams() {
		if len(v) == 1 {
			params[k] = v[0]
		}
	}

	oldnew := make([]string, 0, len(params)*2)
	for k, v := range params {
		k = fmt.Sprintf("#{%s}", k)
		oldnew = append(oldnew, k, v)
	}

	return strings.NewReplacer(oldnew...).Replace(orgPara)
}
