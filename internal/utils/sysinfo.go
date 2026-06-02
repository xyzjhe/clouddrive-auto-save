// internal/utils/sysinfo.go
package utils

import (
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// SysInfo 系统资源信息
type SysInfo struct {
	CPUPercent float64 `json:"cpu_percent"` // CPU 使用率 0-100
	RAMPercent float64 `json:"ram_percent"` // 内存使用率 0-100
	RAMUsedGB  float64 `json:"ram_used_gb"`  // 已用内存 GB
	RAMTotalGB float64 `json:"ram_total_gb"` // 总内存 GB
	NumCPU     int     `json:"num_cpu"`      // CPU 核心数
}

var (
	cachedCPU    float64
	cpuOnce      sync.Once
	cpuStop      chan struct{}
)

// StartCPUCollector 启动后台 CPU 采样（每 5 秒采样一次）
// 首次调用时自动启动，后续调用为空操作
func StartCPUCollector() {
	cpuOnce.Do(func() {
		cpuStop = make(chan struct{})
		go func() {
			// 首次采样：阻塞等待 1 秒获取初始值
			if percents, err := cpu.Percent(1, false); err == nil && len(percents) > 0 {
				cachedCPU = roundOneDecimal(percents[0])
			}
			slog.Info("CPU 采样器已启动", "initial_cpu", cachedCPU)

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if percents, err := cpu.Percent(0, false); err == nil && len(percents) > 0 {
						cachedCPU = roundOneDecimal(percents[0])
					}
				case <-cpuStop:
					return
				}
			}
		}()
	})
}

// GetSysInfo 采集系统资源信息（CPU 取缓存值，内存实时读取）
func GetSysInfo() SysInfo {
	info := SysInfo{
		NumCPU:     runtime.NumCPU(),
		CPUPercent: cachedCPU,
	}

	// 内存使用率（实时读取，无阻塞）
	if vm, err := mem.VirtualMemory(); err == nil {
		info.RAMPercent = roundOneDecimal(vm.UsedPercent)
		info.RAMUsedGB = roundTwoDecimal(float64(vm.Used) / 1024 / 1024 / 1024)
		info.RAMTotalGB = roundTwoDecimal(float64(vm.Total) / 1024 / 1024 / 1024)
	} else {
		slog.Debug("获取内存信息失败", "error", err)
	}

	return info
}

func roundOneDecimal(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

func roundTwoDecimal(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
