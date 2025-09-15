package service

import (
	"context"
	"math/rand"
	"time"

	"monitor-server/internal/model"
)

// MonitorService defines the interface for system monitoring operations
type MonitorService interface {
	GetCPUData(ctx context.Context) (*model.CpuData, error)
	GetMemoryData(ctx context.Context) (*model.MemoryData, error)
	GetDiskData(ctx context.Context) (*model.DiskData, error)
	GetNetworkData(ctx context.Context) (*model.NetworkData, error)
	GetSystemInfo(ctx context.Context) (*model.SystemInfo, error)
	GetProcessData(ctx context.Context, limit int, sortBy string) (*model.ProcessData, error)
	StartHistoryCollection(ctx context.Context)
	StopHistoryCollection()
}

// monitorService implements MonitorService interface
type monitorService struct {
	// TODO: Add dependencies like repository, logger, etc.
}

// NewMonitorService creates a new monitor service instance
func NewMonitorService() MonitorService {
	return &monitorService{
		// TODO: Initialize dependencies
	}
}

// Helper functions for generating mock data
func randomBetween(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func generateCPUHistory(count int) []model.CpuUsage {
	history := make([]model.CpuUsage, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		timestamp := now.Add(-time.Duration(count-i-1) * 5 * time.Second)
		history[i] = model.CpuUsage{
			Timestamp: timestamp,
			Usage:     randomBetween(10, 90),
		}
	}

	return history
}

func generateMemoryHistory(count int) []model.MemoryUsage {
	history := make([]model.MemoryUsage, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		timestamp := now.Add(-time.Duration(count-i-1) * 5 * time.Second)
		usagePercent := randomBetween(40, 85)
		history[i] = model.MemoryUsage{
			Timestamp:    timestamp,
			UsagePercent: usagePercent,
			Used:         (usagePercent / 100) * 16, // 假设总内存16GB
		}
	}

	return history
}

func generateNetworkHistory(count int) []model.NetworkUsage {
	history := make([]model.NetworkUsage, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		timestamp := now.Add(-time.Duration(count-i-1) * 5 * time.Second)
		history[i] = model.NetworkUsage{
			Timestamp:       timestamp,
			BytesSentPerSec: uint64(randomBetween(1024*100, 1024*1024*10)),   // 100KB - 10MB/s
			BytesRecvPerSec: uint64(randomBetween(1024*500, 1024*1024*50)),   // 500KB - 50MB/s
		}
	}

	return history
}

// GetCPUData retrieves current CPU monitoring data
func (s *monitorService) GetCPUData(ctx context.Context) (*model.CpuData, error) {
	temperature := randomBetween(35, 65)
	
	return &model.CpuData{
		Usage:       randomBetween(15, 85),
		Cores:       16,
		Frequency:   randomBetween(2400, 4200),
		Temperature: &temperature,
		Model:       "AMD Ryzen 7 7950X",
		History:     generateCPUHistory(20),
	}, nil
}

// GetMemoryData retrieves current memory monitoring data
func (s *monitorService) GetMemoryData(ctx context.Context) (*model.MemoryData, error) {
	total := 16.0
	used := randomBetween(6, 12)
	free := total - used
	available := free + randomBetween(1, 3)

	return &model.MemoryData{
		Total:         total,
		Used:          used,
		Free:          free,
		Available:     available,
		UsagePercent:  (used / total) * 100,
		SwapTotal:     4.0,
		SwapUsed:      randomBetween(0, 1),
		History:       generateMemoryHistory(20),
	}, nil
}

// GetDiskData retrieves current disk monitoring data
func (s *monitorService) GetDiskData(ctx context.Context) (*model.DiskData, error) {
	disks := []model.DiskInfo{
		{
			Device:       "/dev/nvme0n1p1",
			MountPoint:   "/",
			Filesystem:   "ext4",
			Total:        500,
			Used:         randomBetween(200, 400),
			Free:         0,
			UsagePercent: 0,
		},
		{
			Device:       "/dev/nvme0n1p2",
			MountPoint:   "/home",
			Filesystem:   "ext4",
			Total:        1000,
			Used:         randomBetween(300, 700),
			Free:         0,
			UsagePercent: 0,
		},
		{
			Device:       "/dev/sda1",
			MountPoint:   "/var",
			Filesystem:   "xfs",
			Total:        2000,
			Used:         randomBetween(800, 1500),
			Free:         0,
			UsagePercent: 0,
		},
	}

	// Calculate free space and usage percent
	var totalCapacity, totalUsed, totalFree float64
	for i := range disks {
		disks[i].Free = disks[i].Total - disks[i].Used
		disks[i].UsagePercent = (disks[i].Used / disks[i].Total) * 100
		
		totalCapacity += disks[i].Total
		totalUsed += disks[i].Used
		totalFree += disks[i].Free
	}

	return &model.DiskData{
		Disks:         disks,
		TotalCapacity: totalCapacity,
		TotalUsed:     totalUsed,
		TotalFree:     totalFree,
	}, nil
}

// GetNetworkData retrieves current network monitoring data
func (s *monitorService) GetNetworkData(ctx context.Context) (*model.NetworkData, error) {
	interfaces := []model.NetworkInterface{
		{
			Name:        "eth0",
			BytesSent:   uint64(randomInt(1024*1024*100, 1024*1024*1000)), // 100MB - 1GB
			BytesRecv:   uint64(randomInt(1024*1024*500, 1024*1024*5000)), // 500MB - 5GB
			PacketsSent: uint64(randomInt(10000, 100000)),
			PacketsRecv: uint64(randomInt(50000, 500000)),
			Speed:       1000, // 1Gbps
			IsUp:        true,
		},
		{
			Name:        "wlan0",
			BytesSent:   uint64(randomInt(1024*1024*50, 1024*1024*200)),
			BytesRecv:   uint64(randomInt(1024*1024*100, 1024*1024*800)),
			PacketsSent: uint64(randomInt(5000, 20000)),
			PacketsRecv: uint64(randomInt(10000, 80000)),
			Speed:       300, // 300Mbps
			IsUp:        rand.Float64() > 0.3, // 70% 概率启用
		},
		{
			Name:        "lo",
			BytesSent:   uint64(randomInt(1024*1024*10, 1024*1024*50)),
			BytesRecv:   uint64(randomInt(1024*1024*10, 1024*1024*50)),
			PacketsSent: uint64(randomInt(1000, 5000)),
			PacketsRecv: uint64(randomInt(1000, 5000)),
			Speed:       0, // 回环接口
			IsUp:        true,
		},
	}

	var totalBytesSent, totalBytesRecv uint64
	for _, iface := range interfaces {
		totalBytesSent += iface.BytesSent
		totalBytesRecv += iface.BytesRecv
	}

	return &model.NetworkData{
		Interfaces:     interfaces,
		TotalBytesSent: totalBytesSent,
		TotalBytesRecv: totalBytesRecv,
		History:        generateNetworkHistory(20),
	}, nil
}

// GetSystemInfo retrieves system information
func (s *monitorService) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	uptime := uint64(randomInt(3600, 86400*30)) // 1小时到30天
	bootTime := time.Now().Add(-time.Duration(uptime) * time.Second)

	return &model.SystemInfo{
		Hostname:  "ubuntu-server-01",
		Platform:  "Linux",
		OS:        "Ubuntu 22.04.3 LTS",
		Arch:      "x86_64",
		Uptime:    uptime,
		BootTime:  bootTime,
		Processes: randomInt(200, 450),
		LoadAverage: []float64{
			randomBetween(0.2, 1.5),
			randomBetween(0.3, 1.8),
			randomBetween(0.4, 2.2),
		},
	}, nil
}

// GetProcessData retrieves process monitoring data
func (s *monitorService) GetProcessData(ctx context.Context, limit int, sortBy string) (*model.ProcessData, error) {
	processNames := []string{
		"systemd", "kthreadd", "ksoftirqd/0", "migration/0", "rcu_gp",
		"rcu_par_gp", "kworker/0:0H", "mm_percpu_wq", "ksoftirqd/1",
		"migration/1", "rcu_sched", "watchdog/0", "sshd", "nginx",
		"mysql", "redis-server", "docker", "containerd", "node",
		"python3", "bash", "vim", "htop", "firefox", "code",
	}

	processes := make([]model.ProcessInfo, 0, len(processNames))
	runningCount := 0
	sleepingCount := 0

	for i, name := range processNames {
		status := "running"
		if rand.Float64() > 0.8 { // 20% 概率为睡眠
			status = "sleeping"
			sleepingCount++
		} else {
			runningCount++
		}

		createTime := time.Now().Add(-time.Duration(randomInt(60, 86400)) * time.Second)

		processes = append(processes, model.ProcessInfo{
			PID:           int32(1000 + i*100 + randomInt(1, 99)),
			Name:          name,
			CPUPercent:    randomBetween(0, 25),
			MemoryPercent: float32(randomBetween(0.1, 15)),
			MemoryMB:      randomBetween(10, 500),
			Status:        status,
			CreateTime:    createTime,
			Cmdline:       "/usr/bin/" + name,
		})
	}

	// Limit the results
	if limit > 0 && limit < len(processes) {
		processes = processes[:limit]
	}

	return &model.ProcessData{
		Processes:         processes,
		TotalProcesses:    len(processNames),
		RunningProcesses:  runningCount,
		SleepingProcesses: sleepingCount,
	}, nil
}

// StartHistoryCollection starts collecting historical data in background
func (s *monitorService) StartHistoryCollection(ctx context.Context) {
	// TODO: Implement background goroutine to collect historical data
	// This should run periodically and store data in a circular buffer
}

// StopHistoryCollection stops the historical data collection
func (s *monitorService) StopHistoryCollection() {
	// TODO: Implement cleanup for background collection
}