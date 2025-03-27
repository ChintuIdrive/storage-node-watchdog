package collector

// import (
// 	"ChintuIdrive/storage-node-watchdog/conf"
// 	"ChintuIdrive/storage-node-watchdog/dto"
// 	"time"

// 	"github.com/shirou/gopsutil/cpu"
// 	"github.com/shirou/gopsutil/load"
// 	"github.com/shirou/gopsutil/mem"
// )

// type CpuMetrics struct {
// 	CPUUsage  dto.Metric[float64] `json:"cpu_usage"`
// 	AvgLoad1  dto.Metric[float64] `json:"avg_load"`
// 	AvgLoad5  dto.Metric[float64] `json:"avg_load5"`
// 	AvgLoad15 dto.Metric[float64] `json:"avg_load15"`
// }
// type MemoryMetrics struct {
// 	Total       dto.Metric[uint64]  `json:"total"`
// 	Used        dto.Metric[uint64]  `json:"used"`
// 	UsedPercent dto.Metric[float64] `json:"used_persent"`
// 	Free        dto.Metric[uint64]  `json:"free"`
// }

// func GetCpuStats(config *conf.Config) *CpuMetrics {

// 	cpuUsage, _ := cpu.Percent(time.Second, true)
// 	loadAvg, _ := load.Avg()
// 	sysThreshold := config.GetSystemLevelThreshold()
// 	cpustats := &CpuMetrics{
// 		CPUUsage:  dto.Metric[float64]{Name: "cpu_usage", Value: cpuUsage[0], Limit: sysThreshold.HighCPUusage},
// 		AvgLoad1:  dto.Metric[float64]{Name: "avg_load1", Value: loadAvg.Load1, Limit: sysThreshold.HighAvgLoad},
// 		AvgLoad5:  dto.Metric[float64]{Name: "avg_load5", Value: loadAvg.Load5, Limit: sysThreshold.HighAvgLoad},
// 		AvgLoad15: dto.Metric[float64]{Name: "avg_load15", Value: loadAvg.Load15, Limit: sysThreshold.HighAvgLoad},
// 	}
// 	return cpustats
// }

// func GetMemoryMetrics(config *conf.Config) *MemoryMetrics {
// 	memStats, _ := mem.VirtualMemory()
// 	sysThreshold := config.GetSystemLevelThreshold()
// 	return &MemoryMetrics{
// 		Total:       dto.Metric[uint64]{Name: "total", Value: memStats.Total},
// 		Used:        dto.Metric[uint64]{Name: "used", Value: memStats.Used},
// 		UsedPercent: dto.Metric[float64]{Name: "used_percent", Value: memStats.UsedPercent, Threshold: sysThreshold.HighCPUusage},
// 		Free:        dto.Metric[uint64]{Name: "free", Value: memStats.Free},
// 	}
// }
