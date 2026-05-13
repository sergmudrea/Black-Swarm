package node

import (
	"runtime"
	"time"
)

// ResourceUsage holds a snapshot of system resource utilisation.
type ResourceUsage struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemoryMB    uint64    `json:"memory_mb"`
	MemoryTotal uint64    `json:"memory_total_mb"`
	Goroutines  int       `json:"goroutines"`
	Load        float64   `json:"load"` // 0.0 – 1.0 based on CPU and memory
}

// SampleResources collects the current resource usage of the node.
func SampleResources() *ResourceUsage {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	cpuPercent := estimateCPUUsage()

	usedMB := mem.Alloc / 1024 / 1024
	totalMB := mem.Sys / 1024 / 1024

	load := cpuPercent / 100.0
	if totalMB > 0 {
		memLoad := float64(usedMB) / float64(totalMB)
		if memLoad > load {
			load = memLoad
		}
	}
	if load > 1.0 {
		load = 1.0
	}

	return &ResourceUsage{
		Timestamp:   time.Now(),
		CPUPercent:  cpuPercent,
		MemoryMB:    usedMB,
		MemoryTotal: totalMB,
		Goroutines:  runtime.NumGoroutine(),
		Load:        load,
	}
}

// estimateCPUUsage returns an approximate CPU usage percentage.
// In a production implementation this would use gopsutil or similar.
func estimateCPUUsage() float64 {
	return float64(runtime.NumGoroutine()) / 100.0 * 10.0
}
