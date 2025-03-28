package monitor

import (
	"ChintuIdrive/storage-node-watchdog/actions"
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/conf"
	"ChintuIdrive/storage-node-watchdog/dto"
	"encoding/json"
	"log"
	"time"
)

// Alert thresholds
const (
	HighAvgLoadThreshold = 2.0
	HighLoadDuration     = 1 * time.Minute
	CPUusageThreshold    = 10.0 // Alert if CPU > 80%
	MemoryUsageThreshold = 90.0 // Alert if RAM usage > 90%
	DiskUsageThreshold   = 90
	AlertCooldown        = 60 * time.Second // Cooldown period for alerts
)

type SystemStatsMonitor struct {
	asc           *clients.APIserverClient
	ssc           *collector.SystemStatsCollector
	systemMetrics *dto.SystemMetrics
	config        *conf.Config
}

func NewSystemStatsMonitor(ssc *collector.SystemStatsCollector, asc *clients.APIserverClient, config *conf.Config, systemMetrics *dto.SystemMetrics) *SystemStatsMonitor {
	return &SystemStatsMonitor{
		ssc:           ssc,
		asc:           asc,
		config:        config,
		systemMetrics: systemMetrics,
	}
}

func (ssm *SystemStatsMonitor) MonitorSystemStats() {
	for {
		systemStats := ssm.ssc.CollectSystemMetrics()
		// Log the system metrics
		// log.Printf("System Stats")
		// log.Printf("CPU Usage: %.2f%%, RAM Usage:  %.2f%%, Total RAM: %d MB, Active Connections: %d",
		// 	systemStats.CPUStats.CPUUsage, systemStats.RAMStats.UsedPercent, systemStats.RAMStats.Total, systemStats.ActiveConnCount)
		// log.Printf("Load Avg (1m): %.2f, (5m): %.2f, (15m): %.2f",
		// 	systemStats.CPUStats.AvgLoad1, systemStats.CPUStats.AvgLoad5, systemStats.CPUStats.AvgLoad15)

		for _, metric := range ssm.systemMetrics.ResourceMetrics {
			switch metric.Name {
			case "avg_load1":
				// Handle avg_load1
				metric.Value = systemStats.CPUStats.AvgLoad1
				notify, msg := metric.MonitorThresholdWithDuration()
				if notify {
					ssm.notifySystemUsageMetric(metric, msg)
				}
			case "cpu_usage":
				// Handle CPU usage
				metric.Value = systemStats.CPUStats.CPUUsage
				notify, msg := metric.MonitorThresholdWithDuration()
				if notify {
					ssm.notifySystemUsageMetric(metric, msg)
				}
			case "memory_usage":
				// Handle Memory usage
				metric.Value = systemStats.RAMStats.UsedPercent
				notify, msg := metric.MonitorThresholdWithDuration()
				if notify {
					ssm.notifySystemUsageMetric(metric, msg)
				}
			// case "disk_usage":
			// 	for disk, diskStat := range systemStats.DiskStatsMap {
			// 		metric.Value=diskStat.DiskUsageStat.UsedPercent
			// 	}
			default:
				// Handle unknown metric names
			}
		}
		for disk, diskStat := range systemStats.DiskStatsMap {

			diskWithMetric, available := findDiskMetric(disk, ssm.systemMetrics.DiskMetrics)
			if !available {
				log.Printf("provide metric configuration for %s", disk)
				continue
			} else {
				diskWithMetric.DiskUsage.Value = diskStat.DiskUsageStat.UsedPercent
				notify, msg := diskWithMetric.DiskUsage.MonitorImmediateThreshold(disk)
				if notify {
					ssm.notifySystemUsageMetric(diskWithMetric.DiskUsage, msg)
				}
				for _, iometric := range diskWithMetric.IoMetrics {
					switch iometric.Name {
					case "read_bytes":
						iometric.Value = diskStat.DiskIOStat.ReadBytes
						notify, msg := iometric.MonitorImmediateThreshold(disk)
						if notify {
							ssm.notifySystemIoMetric(iometric, msg)
						}
					case "write_bytes":
						iometric.Value = diskStat.DiskIOStat.WriteBytes
						notify, msg := iometric.MonitorImmediateThreshold(disk)
						if notify {
							ssm.notifySystemIoMetric(iometric, msg)
						}
					case "read_count":
						iometric.Value = diskStat.DiskIOStat.ReadCount
						notify, msg := iometric.MonitorImmediateThreshold(disk)
						if notify {
							ssm.notifySystemIoMetric(iometric, msg)
						}

					case "write_count":
						iometric.Value = diskStat.DiskIOStat.WriteCount
						notify, msg := iometric.MonitorImmediateThreshold(disk)
						if notify {
							ssm.notifySystemIoMetric(iometric, msg)
						}

					default:
						//
					}
				}
			}

		}
		time.Sleep(15 * time.Second) // Adjust as needed
	}
}

func findDiskMetric(diskName string, disks []*dto.DiskMetrics) (*dto.DiskMetrics, bool) {

	for _, diskWithMetric := range disks {
		if diskWithMetric.Name == diskName {
			return diskWithMetric, true
		}
	}

	return &dto.DiskMetrics{}, false
}

func (ssm *SystemStatsMonitor) notifySystemUsageMetric(metric *dto.Metric[float64], msg string) {
	log.Printf("[ACTION] Notify for %s", metric.Name)
	sysNotification := actions.SystemNotification[float64]{
		Type:      actions.SystemMetric,
		NodeId:    ssm.config.ApiServerConfig.NodeId,
		TimeStamp: time.Now().Format(time.RFC3339),
		Actions:   []actions.Action{actions.Notify},
		Metric:    metric,
		Message:   msg,
	}

	log.Printf("System Notification: %v", sysNotification)
	payload, err := json.Marshal(sysNotification)
	if err != nil {
		log.Printf("Error marshalling system notification: %v", err)
		return
	}

	ssm.asc.Notify(payload)
}

func (ssm *SystemStatsMonitor) notifySystemIoMetric(metric *dto.Metric[uint64], msg string) {
	log.Printf("[ACTION] Notify for %s", metric.Name)
	sysNotification := actions.SystemNotification[uint64]{
		Type:      actions.SystemMetric,
		NodeId:    ssm.config.ApiServerConfig.NodeId,
		TimeStamp: time.Now().Format(time.RFC3339),
		Actions:   []actions.Action{actions.Notify},
		Metric:    metric,
		Message:   msg,
	}

	log.Printf("System Notification: %v", sysNotification)
	payload, err := json.Marshal(sysNotification)
	if err != nil {
		log.Printf("Error marshalling system notification: %v", err)
		return
	}

	ssm.asc.Notify(payload)
}
