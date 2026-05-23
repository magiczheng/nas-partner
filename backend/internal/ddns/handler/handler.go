package handler

import (
	"net/http"
	"strconv"

	"nas-partner/backend/internal/ddns/config"
	ddnsmodel "nas-partner/backend/internal/ddns/model"
	"nas-partner/backend/internal/ddns/scheduler"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	list, err := ddnsmodel.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []*ddnsmodel.DDNSConfig{}
	}
	c.JSON(http.StatusOK, list)
}

func Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	cfg, err := ddnsmodel.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func Create(c *gin.Context) {
	var cfg ddnsmodel.DDNSConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := cfg.Create(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cfg)
}

func Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var cfg ddnsmodel.DDNSConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg.ID = id
	if err := cfg.Update(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := ddnsmodel.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func Toggle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	cfg, err := ddnsmodel.Toggle(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func Run(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result := scheduler.Default.RunOnce(id)
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置未找到"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func ListNetInterfaces(c *gin.Context) {
	ipv4, ipv6, err := config.GetNetInterface()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ipv4": ipv4, "ipv6": ipv6})
}

type testIPRequest struct {
	IPv4Enabled      bool   `json:"ipv4_enabled"`
	IPv4GetType      string `json:"ipv4_get_type"`
	IPv4URL          string `json:"ipv4_url"`
	IPv4NetInterface string `json:"ipv4_net_interface"`
	IPv4Cmd          string `json:"ipv4_cmd"`
	IPv4Addr         string `json:"ipv4_addr"`
	IPv6Enabled      bool   `json:"ipv6_enabled"`
	IPv6GetType      string `json:"ipv6_get_type"`
	IPv6URL          string `json:"ipv6_url"`
	IPv6NetInterface string `json:"ipv6_net_interface"`
	IPv6Cmd          string `json:"ipv6_cmd"`
	IPv6Addr         string `json:"ipv6_addr"`
}

func TestIP(c *gin.Context) {
	var req testIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dc := config.DnsConfig{}
	dc.Ipv4.Enable = req.IPv4Enabled
	dc.Ipv4.GetType = req.IPv4GetType
	dc.Ipv4.URL = req.IPv4URL
	dc.Ipv4.NetInterface = req.IPv4NetInterface
	dc.Ipv4.Cmd = req.IPv4Cmd
	dc.Ipv4.Addr = req.IPv4Addr

	dc.Ipv6.Enable = req.IPv6Enabled
	dc.Ipv6.GetType = req.IPv6GetType
	dc.Ipv6.URL = req.IPv6URL
	dc.Ipv6.NetInterface = req.IPv6NetInterface
	dc.Ipv6.Cmd = req.IPv6Cmd
	dc.Ipv6.Addr = req.IPv6Addr

	ipv4 := ""
	ipv6 := ""

	if req.IPv4Enabled {
		ipv4 = dc.GetIpv4Addr()
	}
	if req.IPv6Enabled {
		ipv6 = dc.GetIpv6Addr()
	}

	c.JSON(http.StatusOK, gin.H{"ipv4": ipv4, "ipv6": ipv6})
}

func ListRunLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	logs, err := ddnsmodel.ListRunLogsByConfigID(id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if logs == nil {
		logs = []*ddnsmodel.DDNSRunLog{}
	}
	c.JSON(http.StatusOK, logs)
}

func ClearLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := ddnsmodel.DeleteAllByConfigID(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func CleanupLogs(c *gin.Context) {
	deleted, err := ddnsmodel.DeleteOlderThan(3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func ListLatestLogs(c *gin.Context) {
	list, err := ddnsmodel.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type ConfigWithLog struct {
		*ddnsmodel.DDNSConfig
		LatestLog *ddnsmodel.DDNSRunLog `json:"latest_log"`
	}

	result := make([]*ConfigWithLog, 0, len(list))
	for _, cfg := range list {
		item := &ConfigWithLog{DDNSConfig: cfg}
		log, err := ddnsmodel.GetLatestRunLogByConfigID(cfg.ID)
		if err == nil {
			item.LatestLog = log
		}
		result = append(result, item)
	}
	c.JSON(http.StatusOK, result)
}
