package main

import (
	"ChintuIdrive/storage-node-watchdog/api"
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/conf"
	"ChintuIdrive/storage-node-watchdog/monitor"
	"encoding/json"
	"log"
	"os"
)

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
	asc := clients.NewApiServerClient(config.ApiServerConfig)
	// tenantsFromApiServer, err := asc.GetTenatsListFromApiServer()
	// if err != nil {
	// 	//notify watchdog not able to fetch tenantlist from api server
	// 	log.Printf("Failed to fetch tenant list from API server: %v", err)
	// }
	// config.LoadS3Config(tenantsFromApiServer)
	cc := clients.NewControllerClientt(config.ControllerConfig)
	//cc.LadAccessKeys(tenantsFromApiServer)
	ssc := collector.NewSystemStatsCollector(config)
	pmc := collector.NewProcesMetricsCollector(config)
	s3mc := collector.NewS3MetricCollector(config, cc)

	monitor.StartMonitoring(config, cc, asc, ssc, pmc, s3mc)
	api.RegisterHandlers(config, cc, asc, ssc, pmc, s3mc)
}
