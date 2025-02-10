package main

import (
	"ChintuIdrive/storage-node-watchdog/collector"
	"fmt"
	"log"
	"net/http"
	"os"
)

// type StorageUsageStats struct {
// 	OsDiskUsageStats    *DiskUsageStat `json:"os-disk-usage-stats"`
// 	Data1DiskUsageStats *DiskUsageStat `json:"data1-disk-usage-stats"`
// 	Data2DiskUsageStats *DiskUsageStat `json:"data2-disk-usage-stats"`
// 	Data3DiskUsageStats *DiskUsageStat `json:"data3-disk-usage-stats"`
// 	Data4DiskUsageStats *DiskUsageStat `json:"data4-disk-usage-stats"`
// 	//metadata disk to do
// }

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

func main() {

	collector.CollectMetrics()
	// Setup API routes
	// ApI to start and stop the process add authentication(using private and pulic key) for this ip based filtering
	// certificate based authentication used dynamic string (use timestamp)
	// http.HandleFunc("/metrics", collector.SystemMetricsHandler)
	// http.HandleFunc("/process_metrics", collector.ProcessMetricsHandler)

	fmt.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
