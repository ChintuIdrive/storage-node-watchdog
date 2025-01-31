package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Processes to monitor
var monitoredProcesses = []string{"minio", "e2_node_controller", "trash_cleaner", "rclone"}

// Structs for JSON response
type SystemStats struct {
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     uint64    `json:"memory_usage"`
	TotalMemory     uint64    `json:"total_memory"`
	ActiveConnCount int       `json:"active_connections"`
	LastUpdated     time.Time `json:"last_updated"`
}

type ProcessMetrics struct {
	Name     string  `json:"name"`
	PID      int32   `json:"pid"`
	CPUUsage float64 `json:"cpu_usage"`
	MemUsage uint64  `json:"memory_usage"`
}

// Global variables
var (
	systemStats  SystemStats
	processStats []ProcessMetrics
	statsLock    sync.RWMutex
)

// Collect system metrics
func collectSystemMetrics() {
	for {
		cpuUsage, _ := cpu.Percent(time.Second, false)
		memStats, _ := mem.VirtualMemory()
		connCount, _ := getActiveConnections()

		statsLock.Lock()
		systemStats = SystemStats{
			CPUUsage:        cpuUsage[0],
			MemoryUsage:     memStats.Used,
			TotalMemory:     memStats.Total,
			ActiveConnCount: connCount,
			LastUpdated:     time.Now(),
		}
		statsLock.Unlock()

		time.Sleep(5 * time.Second) // Adjust as needed
	}
}

// Get active network connections count
func getActiveConnections() (int, error) {
	connections, err := net.Connections("all")
	if err != nil {
		return 0, err
	}
	return len(connections), nil
}

// Collect per-process metrics
func collectProcessMetrics() {
	for {
		processList, _ := process.Processes()
		var metrics []ProcessMetrics

		for _, proc := range processList {
			name, _ := proc.Name()
			for _, monitored := range monitoredProcesses {
				if name == monitored {
					cpuPercent, _ := proc.CPUPercent()
					memInfo, _ := proc.MemoryInfo()
					metrics = append(metrics, ProcessMetrics{
						Name:     name,
						PID:      proc.Pid,
						CPUUsage: cpuPercent,
						MemUsage: memInfo.RSS,
					})
				}
			}
		}

		statsLock.Lock()
		processStats = metrics
		statsLock.Unlock()

		time.Sleep(5 * time.Second) // Adjust interval as needed
	}
}

// API Handler: Get system stats
func systemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemStats)
}

// API Handler: Get process metrics
func processMetricsHandler(w http.ResponseWriter, r *http.Request) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(processStats)
}

func main() {
	// Start background collectors
	go collectSystemMetrics()
	go collectProcessMetrics()

	// Setup API routes
	http.HandleFunc("/metrics", systemMetricsHandler)
	http.HandleFunc("/process_metrics", processMetricsHandler)

	fmt.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
