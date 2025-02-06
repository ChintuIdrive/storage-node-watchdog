package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Processes to monitor
var monitoredProcesses = []string{"minio", "e2_node_controller_service", "trash-cleaner-service", "rclone", "kes", "vault", "load-simulator"}
var monitoreddisks = []string{"/", "/data1", "/data2", "/data3", "/data4"}

// Alert thresholds
const (
	HighLoadThreshold = 2.0
	HighLoadDuration  = 1 * time.Minute
	CPUThreshold      = 10.0             // Alert if CPU > 80%
	MemoryThreshold   = 90.0             // Alert if RAM usage > 90%
	AlertCooldown     = 60 * time.Second // Cooldown period for alerts
)

// Structs for JSON response
type SystemStats struct {
	CPUStats        *CpuStats             `json:"cpu_stats"`
	RAMStats        MemoryStats           `json:"ram_stats"`
	DiskStatsMap    map[string]*DiskStats `json:"disk-stats-map"`
	ActiveConnCount int                   `json:"active_connections"`
	LastUpdated     time.Time             `json:"last_updated"`
}

type CpuStats struct {
	CoreCount int     `json:"core_count"`
	CPUUsage  float64 `json:"cpu_usage"`
	AvgLoad1  float64 `json:"avg_load"`
	AvgLoad5  float64 `json:"avg_load5"`
	AvgLoad15 float64 `json:"avg_load15"`
}

type MemoryStats struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_persent"`
	Free        uint64  `json:"free"`
}

type DiskStats struct {
	DiskUsageStat *DiskUsageStat      `json:"disk-usage-stats"`
	DiskIOStat    disk.IOCountersStat `json:"disk-io-stats"`
}
type DiskUsageStat struct {
	Device      string  `json:"device"`
	Path        string  `json:"path"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type StorageUsageStats struct {
	OsDiskUsageStats    *DiskUsageStat `json:"os-disk-usage-stats"`
	Data1DiskUsageStats *DiskUsageStat `json:"data1-disk-usage-stats"`
	Data2DiskUsageStats *DiskUsageStat `json:"data2-disk-usage-stats"`
	Data3DiskUsageStats *DiskUsageStat `json:"data3-disk-usage-stats"`
	Data4DiskUsageStats *DiskUsageStat `json:"data4-disk-usage-stats"`
	//metadata disk to do
}
type ProcessMetrics struct {
	Name             string  `json:"name"`
	PID              int32   `json:"pid"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemUsage         float32 `json:"memory_usage"`
	ConnectionsCount int     `json:"connections_count"`
}

// Log file setup
var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("watchdog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var highLoadStartTime time.Time

// Global variables
var (
	systemStats  SystemStats
	processStats []ProcessMetrics
	statsLock    sync.RWMutex
	lastCPUAlert time.Time
	lastMemAlert time.Time
	alertLock    sync.Mutex
)

// Collect system metrics
func collectSystemMetrics() {
	for {
		cpuUsage, _ := cpu.Percent(time.Second, true)
		memStats, _ := mem.VirtualMemory()
		connCount, _ := getActiveConnections()
		loadAvg, _ := load.Avg()
		diskStats, _ := disk.IOCounters()
		//rootFsStats, _ := disk.Usage("/")
		corescount, _ := cpu.Counts(true)
		log.Printf("number of core %d", corescount)
		infoStats, _ := cpu.Info()
		for _, infostat := range infoStats {
			log.Printf("CPU Info: %v", infostat.String())
		}

		statsLock.Lock()

		systemStats = SystemStats{
			CPUStats: &CpuStats{
				CPUUsage:  cpuUsage[0],
				AvgLoad1:  loadAvg.Load1,
				AvgLoad5:  loadAvg.Load5,
				AvgLoad15: loadAvg.Load15,
				CoreCount: corescount,
			},
			RAMStats: MemoryStats{
				Total:       memStats.Total,
				Used:        memStats.Used,
				UsedPercent: memStats.UsedPercent,
				Free:        memStats.Free,
			},
			DiskStatsMap: make(map[string]*DiskStats),
			// TotalRead:       totalRead,
			// TotalWrite:      totalWrite,
			ActiveConnCount: connCount,
			LastUpdated:     time.Now(),
		}
		// Disk I/O
		//diskUsageMap := make(map[string]*DiskUsageStat)
		for _, diskName := range monitoreddisks {
			diskUsageStat, err := disk.Usage(diskName)
			if err != nil {
				log.Printf("Error getting disk usage for %s: %v", diskName, err)
				continue
			}
			devicePath, _ := getDeviceForMount(diskName)
			deviceName := filepath.Base(devicePath) // Extracts "vdb1"
			diskiotat := diskStats[deviceName]
			diskUsageStats := &DiskUsageStat{
				Device:      devicePath,
				Path:        diskUsageStat.Path,
				Fstype:      diskUsageStat.Fstype,
				Total:       diskUsageStat.Total,
				Free:        diskUsageStat.Free,
				Used:        diskUsageStat.Used,
				UsedPercent: diskUsageStat.UsedPercent,
			}

			diskStat := &DiskStats{
				DiskUsageStat: diskUsageStats,
				DiskIOStat:    diskiotat,
			}
			systemStats.DiskStatsMap[diskName] = diskStat
			//diskUsageMap[diskName] =

		}

		statsLock.Unlock()
		// Log the system metrics
		log.Printf("System Stats")
		log.Printf("CPU Usage: %.2f%%, RAM Usage:  %.2f%%, Total RAM: %d MB, Active Connections: %d",
			systemStats.CPUStats.CPUUsage, systemStats.RAMStats.UsedPercent, systemStats.RAMStats.Total, connCount)
		log.Printf("Load Avg (1m): %.2f, (5m): %.2f, (15m): %.2f",
			systemStats.CPUStats.AvgLoad1, systemStats.CPUStats.AvgLoad5, systemStats.CPUStats.AvgLoad15)
		// log.Printf("Disk Read: %d MB, Disk Write: %d MB, Root FS Free: %d GB",
		// 	totalRead/1024/1024, totalWrite/1024/1024, rootFsStats.Free/1024/1024/1024)

		// Check for alerts
		checkAlerts(systemStats.CPUStats.AvgLoad1, systemStats.CPUStats.CPUUsage, systemStats.RAMStats.UsedPercent)
		time.Sleep(15 * time.Second) // Adjust as needed
	}
}

func getDeviceForMount(mountPoint string) (string, error) {
	partitions, err := disk.Partitions(false) // Get all mounted filesystems
	if err != nil {
		return "", err
	}

	for _, partition := range partitions {
		if partition.Mountpoint == mountPoint {
			return partition.Device, nil
		}
	}

	return "", fmt.Errorf("mount point %s not found", mountPoint)
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
					proc.Connections()
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

// API Handler: Get system stats
func systemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemStats)
}

// API Handler: Get process metrics
// if cpuusage and memusage are crossing threshold for 5 min interva then raise the alarm use ticker to check the threshold
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
	// ApI to start and stop the process add authentication(using private and pulic key) for this ip based filtering
	// certificate based authentication used dynamic string (use timestamp)
	http.HandleFunc("/metrics", systemMetricsHandler)
	http.HandleFunc("/process_metrics", processMetricsHandler)

	fmt.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
