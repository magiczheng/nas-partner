package system

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

type CPUInfo struct {
	Percent float64 `json:"percent"`
	Cores   int     `json:"cores"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskInfo struct {
	Path        string  `json:"path"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type SystemInfo struct {
	Hostname  string    `json:"hostname"`
	Uptime    uint64    `json:"uptime"`
	OS        string    `json:"os"`
	CPU       CPUInfo   `json:"cpu"`
	Memory    MemoryInfo `json:"memory"`
	Disk      DiskInfo  `json:"disk"`
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
}

func GetSystemInfo(c *gin.Context) {
	info := SystemInfo{}

	// Hostname + OS
	if hi, err := host.Info(); err == nil {
		info.Hostname = hi.Hostname
		info.OS = hi.OS
	}

	// Uptime in seconds
	if up, err := host.Uptime(); err == nil {
		info.Uptime = up
	}

	// CPU usage (block 500ms for a meaningful reading)
	if percents, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(percents) > 0 {
		info.CPU.Percent = percents[0]
	}
	if count, err := cpu.Counts(true); err == nil {
		info.CPU.Cores = count
	}

	// Memory
	if vm, err := mem.VirtualMemory(); err == nil {
		info.Memory = MemoryInfo{
			Total:       vm.Total,
			Used:        vm.Used,
			Available:   vm.Available,
			UsedPercent: vm.UsedPercent,
		}
	}

	// Disk usage for the data directory
	diskPath := "/app/data"
	if du, err := disk.Usage(diskPath); err == nil {
		info.Disk = DiskInfo{
			Path:        diskPath,
			Total:       du.Total,
			Used:        du.Used,
			Free:        du.Free,
			UsedPercent: du.UsedPercent,
		}
	} else if du, err := disk.Usage("/"); err == nil {
		info.Disk = DiskInfo{
			Path:        "/",
			Total:       du.Total,
			Used:        du.Used,
			Free:        du.Free,
			UsedPercent: du.UsedPercent,
		}
	}

	// Load average
	if la, err := load.Avg(); err == nil {
		info.Load1 = la.Load1
		info.Load5 = la.Load5
		info.Load15 = la.Load15
	}

	c.JSON(http.StatusOK, info)
}
