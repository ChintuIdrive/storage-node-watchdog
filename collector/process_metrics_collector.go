package collector

import (
	"ChintuIdrive/storage-node-watchdog/conf"
	"strings"

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
	config *conf.Config
	//processStats      []ProcessMetrics
	//tenatProcessStats []TenantProcessMetrics
}

func NewProcesMetricsCollector(config *conf.Config) *ProcesMetricsCollector {
	return &ProcesMetricsCollector{
		config: config,
	}
}

// Processes to monitor
//var monitoredProcesses = []string{"minio", "e2_node_controller_service", "trash-cleaner-service", "rclone", "kes", "vault", "load-simulator"}

// Collect per-process metrics
func (pmc *ProcesMetricsCollector) CollectProcessMetrics() []ProcessMetrics {
	monitoredProcesses := pmc.config.GetProcessToMonitor()
	//for {
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
				processMetrics := ProcessMetrics{
					Name:             name,
					PID:              proc.Pid,
					CPUUsage:         cpuPercent,
					MemUsage:         memPercent,
					ConnectionsCount: len(connections),
				}
				metrics = append(metrics, processMetrics)
			}
		}
	}

	// statsLock.Lock()
	// pmc.processStats = metrics
	// statsLock.Unlock()

	return metrics
	//time.Sleep(15 * time.Second) // Adjust interval as needed
	//}
}

// func (pmc *ProcesMetricsCollector) GetProcessMetrics() []ProcessMetrics {
// 	return pmc.processStats
// }

func (pmc *ProcesMetricsCollector) CollectRunningTenantProcMetrics() []TenantProcessMetrics {
	tenatProcessName := pmc.config.TenantProcessName
	processList, _ := process.Processes()
	var tenantMetrics []TenantProcessMetrics

	for _, proc := range processList {
		name, _ := proc.Name()
		if strings.ToLower(name) == tenatProcessName {
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
	// statsLock.Lock()
	// pmc.tenatProcessStats = tenantMetrics
	// statsLock.Unlock()
	return tenantMetrics
}
