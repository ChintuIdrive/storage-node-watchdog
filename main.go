package main

import (
	"ChintuIdrive/storage-node-watchdog/conf"
	"ChintuIdrive/storage-node-watchdog/cryption"
	"ChintuIdrive/storage-node-watchdog/monitor"
	"encoding/json"
	"log"
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
var (
	logFile *os.File
	config  *conf.Config
)

func init() {
	var err error
	if _, err := os.Stat("conf/config.json"); os.IsNotExist(err) {
		config = conf.GetDefaultConfig()
		configFile, err := os.Create("conf/config.json")
		if err != nil {
			log.Fatalf("Failed to create config file: %s", err)
		}
		defer configFile.Close()

		configData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal default config: %s", err)
		}
		configFile.Write(configData)

		// encoder := json.NewEncoder(configFile)
		// if err := encoder.Encode(defaultConfig); err != nil {
		// 	log.Fatalf("Failed to write default config to file: %s", err)
		// }
	} else {
		config, err = conf.LoadConfig("conf/config.json")
		if err != nil {
			log.Fatalf("Failed to load config: %s", err)
		}
	}

	logFile, err = os.OpenFile("watchdog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	monitor.StartMonitoring(config)
	//apiserverkey := "E8AA3FBB0F512B32"

	// tenants, err := apiserverClient.GetTenatsListFromApiServer("nc1")

	// if err == nil {
	// 	for _, tenat := range tenants {
	// 		acckey, err := controllerClient.GetAccessKeys(tenat)
	// 		ds, err := acckey.SecretKey.GetDString()
	// 		if err != nil {
	// 			log.Print(err)
	// 		} else {
	// 			log.Print(ds)
	// 		}
	// 		client := clients.NewS3Client(tenat.DNS, acckey.AccessKey, ds)
	// 		startTime := time.Now()
	// 		buckets, err := client.ListBuckets()
	// 		duration := time.Since(startTime)
	// 		log.Printf("Time taken to list buckets: %v", duration)
	// 		if err != nil {
	// 			log.Printf("Error listing buckets: %v", err)
	// 		}
	// 		for _, bucket := range buckets {
	// 			log.Printf("Bucket: %s", *bucket.Name)
	// 		}
	// 	}
}

//collector.CollectMetrics()
// Setup API routes
// ApI to start and stop the process add authentication(using private and pulic key) for this ip based filtering
// certificate based authentication used dynamic string (use timestamp)
// http.HandleFunc("/metrics", collector.SystemMetricsHandler)
// http.HandleFunc("/process_metrics", collector.ProcessMetricsHandler)

// fmt.Println("API running on :8080")
// log.Fatal(http.ListenAndServe(":8080", nil))

func main1() {
	// Example JSON
	jsonData := `{
  "accessKey": "dGVzdDMvR1RFVYLSzL0H",
  "secretKey": {
    "CString": "R33U1KHv/FyiEqgOhweIhNSSo7liM9DOHOHJBiKQLEVZZ4vTawiYr2DnaKfcPElQ"
  },
  "expiration": "0001-01-01T00:00:00",
  "StatusCode": 200
}`

	// endpoint := "x1m0.nc02.edgedrive.com"
	// accKey := "d3RhY2hkb2cvSYIcyL0H"
	// secKey := "ke5FyeGdM7mTbRs11OgxZTfjzKNWql4VZ7Ak0xHZ"

	// client := clients.NewS3Client(endpoint, accKey, secKey)
	// buckets, err := client.ListBuckets()
	// if err != nil {
	// 	log.Fatalf("Error listing buckets: %v", err)
	// }
	// for _, bucket := range buckets {
	// 	log.Printf("Bucket: %s", *bucket.Name)
	// }

	var secret cryption.SecretData
	err := json.Unmarshal([]byte(jsonData), &secret)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Get DString
	dstring, err := secret.SecretKey.GetDString()
	log.Print(dstring)
}
