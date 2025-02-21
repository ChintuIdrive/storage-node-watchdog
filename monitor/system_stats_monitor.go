package monitor

import (
	"ChintuIdrive/storage-node-watchdog/collector"
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/disk"
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

var highLoadStartTime time.Time
var alertLock sync.Mutex
var lastCPUAlert time.Time
var lastMemAlert time.Time
var lastAvgLoadAlert time.Time
var lastReadBytesAlert time.Time
var lastWriteBytesAlert time.Time
var lastReadCountAlert time.Time
var lastWriteCountAlert time.Time
var diskUsageAlert time.Time

type SystemStatsMonitor struct {
	ssc *collector.SystemStatsCollector
}

func NewSystemStatsMonitor(ssc *collector.SystemStatsCollector) *SystemStatsMonitor {
	return &SystemStatsMonitor{
		ssc: ssc,
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
		checkMetric(systemStats.CPUStats.AvgLoad1, HighAvgLoadThreshold, &lastAvgLoadAlert, "Avg Load1")
		checkMetric(systemStats.CPUStats.CPUUsage, CPUusageThreshold, &lastCPUAlert, "CPU Usage")
		checkMetric(systemStats.CPUStats.CPUUsage, CPUusageThreshold, &lastCPUAlert, "CPU Usage")
		checkMetric(systemStats.RAMStats.UsedPercent, MemoryUsageThreshold, &lastMemAlert, "Memory Usage")

		for disk, diskStat := range systemStats.DiskStatsMap {
			//checkMetric(diskStat.DiskUsageStat.UsedPercent, DiskUsageThreshold, &lastMemAlert, "Disk Usage on "+disk)
			checkDiskUsage(disk, diskStat.DiskUsageStat)
			checkDiskIO(disk, diskStat.DiskIOStat)
		}
		time.Sleep(15 * time.Second) // Adjust as needed
	}
}

// Check if system stats exceed thresholds and trigger alerts
func checkAlerts(loadAvg1, cpuUsage float64, memUsage float64) {
	alertLock.Lock()
	defer alertLock.Unlock()

	now := time.Now()
	// High load detection logic
	if loadAvg1 > HighAvgLoadThreshold {
		if highLoadStartTime.IsZero() {
			highLoadStartTime = time.Now() // Start tracking high load time
		} else if time.Since(highLoadStartTime) >= HighLoadDuration {
			log.Println("[ALERT] System Load1 has been high for 5+ minutes!")
			highLoadStartTime = time.Time{} // Reset timer after restart
		}
	} else {
		highLoadStartTime = time.Time{} // Reset if load is normal
	}
	// CPU Alert
	if cpuUsage > CPUusageThreshold && now.Sub(lastCPUAlert) > AlertCooldown {
		log.Printf("[ALERT] High CPU Usage: %.2f%%", cpuUsage)
		lastCPUAlert = now
	}

	// Memory Alert
	if memUsage > MemoryUsageThreshold && now.Sub(lastMemAlert) > AlertCooldown {
		log.Printf("[ALERT] High Memory Usage: %.2f%%", memUsage)
		lastMemAlert = now
	}
}

func checkMetric(metricValue float64, threshold float64, lastAlertTime *time.Time, metricName string) {
	alertLock.Lock()
	defer alertLock.Unlock()

	now := time.Now()
	if metricValue > threshold {
		if lastAlertTime.IsZero() {
			*lastAlertTime = now // Start tracking high metric usage time
		} else if now.Sub(*lastAlertTime) >= HighLoadDuration {
			log.Printf("[ALERT] High %s: %.2f%% for 5+ minutes!", metricName, metricValue)
			*lastAlertTime = time.Time{} // Reset timer after alert
		}
	} else {
		*lastAlertTime = time.Time{} // Reset if metric usage is normal
	}
}

func checkDiskIO(disk string, diskIOStat disk.IOCountersStat) {
	// Define thresholds for disk IO metrics
	const (
		ReadBytesThreshold  = 1000000000 // 1 GB
		WriteBytesThreshold = 1000000000 // 1 GB
		IOPSReadThreshold   = 1000       // 1000 IOPS
		IOPSWriteThreshold  = 1000       // 1000 IOPS
	)
	now := time.Now()
	// Check read bytes
	if diskIOStat.ReadBytes > ReadBytesThreshold && now.Sub(lastReadBytesAlert) > AlertCooldown {
		log.Printf("[ALERT] High Disk Read Bytes: %d bytes on %s", diskIOStat.ReadBytes, disk)
		lastReadBytesAlert = now
	}

	// Check write bytes
	if diskIOStat.WriteBytes > WriteBytesThreshold && now.Sub(lastWriteBytesAlert) > AlertCooldown {
		log.Printf("[ALERT] High Disk Write Bytes: %d bytes on %s", diskIOStat.WriteBytes, disk)
		lastWriteBytesAlert = now
	}

	// Check read IOPS
	if diskIOStat.ReadCount > IOPSReadThreshold && now.Sub(lastReadCountAlert) > AlertCooldown {
		log.Printf("[ALERT] High Disk Read IOPS: %d on %s", diskIOStat.ReadCount, disk)
		lastReadCountAlert = now
	}

	// Check write IOPS
	if diskIOStat.WriteCount > IOPSWriteThreshold && now.Sub(lastWriteCountAlert) > AlertCooldown {
		log.Printf("[ALERT] High Disk Write IOPS: %d on %s", diskIOStat.WriteCount, disk)
		lastWriteCountAlert = now
	}
}

func checkDiskUsage(diskName string, diskUsageStat *collector.DiskUsageStat) {
	alertLock.Lock()
	defer alertLock.Unlock()

	now := time.Now()
	if diskUsageStat.UsedPercent > DiskUsageThreshold && now.Sub(diskUsageAlert) > AlertCooldown {
		log.Printf("[ALERT] High Disk Usage on %s: %.2f%%", diskName, diskUsageStat.UsedPercent)
		diskUsageAlert = now
	}
}
