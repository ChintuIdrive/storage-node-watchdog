package collector

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/process"
)

type ProcessMetrics struct {
	Name             string  `json:"name"`
	PID              int32   `json:"pid"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemUsage         float32 `json:"memory_usage"`
	ConnectionsCount int     `json:"connections_count"`
}

// Processes to monitor
var monitoredProcesses = []string{"minio", "e2_node_controller_service", "trash-cleaner-service", "rclone", "kes", "vault", "load-simulator"}

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
					//memInfo, _ := proc.MemoryInfo()
					memPercent, _ := proc.MemoryPercent()
					connections, _ := proc.Connections()
					metrics = append(metrics, ProcessMetrics{
						Name:             name,
						PID:              proc.Pid,
						CPUUsage:         cpuPercent,
						MemUsage:         memPercent,
						ConnectionsCount: len(connections),
					})
				}
			}
		}

		statsLock.Lock()
		processStats = metrics
		for _, metric := range metrics {
			log.Printf("Process: %s, PID: %d, CPU Usage: %.2f%%, Memory Usage: %.2f%%", metric.Name, metric.PID, metric.CPUUsage, metric.MemUsage)
		}
		statsLock.Unlock()

		time.Sleep(15 * time.Second) // Adjust interval as needed
	}
}

// API Handler: Get process metrics
// if cpuusage and memusage are crossing threshold for 5 min interva then raise the alarm use ticker to check the threshold
func ProcessMetricsHandler(w http.ResponseWriter, r *http.Request) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(processStats)
}
