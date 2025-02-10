package collector

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
)

type ProcessMetrics struct {
	Name             string  `json:"name"`
	PID              int32   `json:"pid"`
	IsTenant         bool    `json:"is_tenant"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemUsage         float32 `json:"memory_usage"`
	ConnectionsCount int     `json:"connections_count"`
}

type TenantProcessMetrics struct {
	ProcessMetrics
	DNS string `json:"dns"`
}

type ProcesMetricsCollector struct {
	processStats      []ProcessMetrics
	tenatProcessStats []TenantProcessMetrics
}

// Processes to monitor
var monitoredProcesses = []string{"minio", "e2_node_controller_service", "trash-cleaner-service", "rclone", "kes", "vault", "load-simulator"}

// Collect per-process metrics
func (pmc *ProcesMetricsCollector) CollectProcessMetrics() {

	for {
		processList, _ := process.Processes()
		var metrics []ProcessMetrics
		//var tenantMetrics []*TenantProcessMetrics

		for _, proc := range processList {
			name, _ := proc.Name()
			for _, monitored := range monitoredProcesses {
				if name == monitored {

					cpuPercent, _ := proc.CPUPercent()
					//memInfo, _ := proc.MemoryInfo()
					memPercent, _ := proc.MemoryPercent()
					connections, _ := proc.Connections()
					processMetrics := ProcessMetrics{
						Name:             name,
						PID:              proc.Pid,
						CPUUsage:         cpuPercent,
						MemUsage:         memPercent,
						ConnectionsCount: len(connections),
					}
					metrics = append(metrics, processMetrics)
					// if processMetrics.Name == "minio" {
					// 	processMetrics.IsTenant = true

					// 	tenatProcess := &TenantProcessMetrics{
					// 		ProcessMetrics: *processMetrics,
					// 		DNS:            "ghhh",
					// 	}
					// 	tenantMetrics = append(tenantMetrics, tenatProcess)
					// } else {
					// 	metrics = append(metrics, processMetrics)
					// }

				}
			}
		}

		statsLock.Lock()
		pmc.processStats = metrics
		//pmc.tenatProcessStats = tenantMetrics

		statsLock.Unlock()

		time.Sleep(15 * time.Second) // Adjust interval as needed
	}
}

func (pmc *ProcesMetricsCollector) GetProcessMetrics() []ProcessMetrics {
	return pmc.processStats
}

func (pmc *ProcesMetricsCollector) CollectMinioMetrics() {
	processList, _ := process.Processes()
	var tenantMetrics []TenantProcessMetrics

	for _, proc := range processList {
		name, _ := proc.Name()
		if strings.ToLower(name) == "minio" {
			cpuPercent, _ := proc.CPUPercent()
			//memInfo, _ := proc.MemoryInfo()
			memPercent, _ := proc.MemoryPercent()
			connections, _ := proc.Connections()
			processMetrics := &ProcessMetrics{
				Name:             name,
				PID:              proc.Pid,
				CPUUsage:         cpuPercent,
				MemUsage:         memPercent,
				ConnectionsCount: len(connections),
			}
			processMetrics.IsTenant = true

			tenatProcess := TenantProcessMetrics{
				ProcessMetrics: *processMetrics,
			}
			tenantMetrics = append(tenantMetrics, tenatProcess)
		}
	}
	statsLock.Lock()
	pmc.tenatProcessStats = tenantMetrics
	statsLock.Unlock()

}
func (pmc *ProcesMetricsCollector) GetMinioProcessMetrics() []TenantProcessMetrics {
	return pmc.tenatProcessStats
}

// API Handler: Get process metrics
// if cpuusage and memusage are crossing threshold for 5 min interva then raise the alarm use ticker to check the threshold
func (pmc *ProcesMetricsCollector) ProcessMetricsHandler(w http.ResponseWriter, r *http.Request) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pmc.processStats)
}
