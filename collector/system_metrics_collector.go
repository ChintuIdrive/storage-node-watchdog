package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
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

//var monitoreddisks = []string{"/", "/data1", "/data2", "/data3", "/data4"}

type SyatemStatsCollector struct {
	systemStats *SystemStats
}

// Collect system metrics
func (smc *SyatemStatsCollector) CollectSystemMetrics(monitoreddisks []string) *SystemStats {
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

	smc.systemStats = &SystemStats{
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
		smc.systemStats.DiskStatsMap[diskName] = diskStat
		//diskUsageMap[diskName] =

	}

	statsLock.Unlock()

	// log.Printf("Disk Read: %d MB, Disk Write: %d MB, Root FS Free: %d GB",
	// 	totalRead/1024/1024, totalWrite/1024/1024, rootFsStats.Free/1024/1024/1024)

	// Check for alerts

	return smc.systemStats

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

// API Handler: Get system stats
func (smc *SyatemStatsCollector) SystemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(smc.systemStats)
}
