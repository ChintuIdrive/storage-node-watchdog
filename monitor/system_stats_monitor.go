package monitor

import (
	"ChintuIdrive/storage-node-watchdog/collector"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Alert thresholds
const (
	HighLoadThreshold = 2.0
	HighLoadDuration  = 1 * time.Minute
	CPUThreshold      = 10.0             // Alert if CPU > 80%
	MemoryThreshold   = 90.0             // Alert if RAM usage > 90%
	AlertCooldown     = 60 * time.Second // Cooldown period for alerts
)

var highLoadStartTime time.Time
var alertLock sync.Mutex
var lastCPUAlert time.Time
var lastMemAlert time.Time

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
		log.Printf("System Stats")
		log.Printf("CPU Usage: %.2f%%, RAM Usage:  %.2f%%, Total RAM: %d MB, Active Connections: %d",
			systemStats.CPUStats.CPUUsage, systemStats.RAMStats.UsedPercent, systemStats.RAMStats.Total, systemStats.ActiveConnCount)
		log.Printf("Load Avg (1m): %.2f, (5m): %.2f, (15m): %.2f",
			systemStats.CPUStats.AvgLoad1, systemStats.CPUStats.AvgLoad5, systemStats.CPUStats.AvgLoad15)
		checkAlerts(systemStats.CPUStats.AvgLoad1, systemStats.CPUStats.CPUUsage, systemStats.RAMStats.UsedPercent)
		time.Sleep(15 * time.Second) // Adjust as needed
	}
}
func (ssm *SystemStatsMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	systemMetrics := ssm.ssc.CollectSystemMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemMetrics)
}

// Check if system stats exceed thresholds and trigger alerts
func checkAlerts(loadAvg1, cpuUsage float64, memUsage float64) {
	alertLock.Lock()
	defer alertLock.Unlock()

	now := time.Now()
	// High load detection logic
	if loadAvg1 > HighLoadThreshold {
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
	if cpuUsage > CPUThreshold && now.Sub(lastCPUAlert) > AlertCooldown {
		log.Printf("[ALERT] High CPU Usage: %.2f%%", cpuUsage)
		lastCPUAlert = now
	}

	// Memory Alert
	if memUsage > MemoryThreshold && now.Sub(lastMemAlert) > AlertCooldown {
		log.Printf("[ALERT] High Memory Usage: %.2f%%", memUsage)
		lastMemAlert = now
	}
}
