package monitor

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/conf"
)

func StartMonitoring(config *conf.Config, cc *clients.ControllerClient, asc *clients.APIserverClient,
	ssc *collector.SystemStatsCollector, pmc *collector.ProcesMetricsCollector, s3mc *collector.S3MetricCollector) {

	// systemStatsMonitor := NewSystemStatsMonitor(ssc, asc, config)
	// go systemStatsMonitor.MonitorSystemStats()

	processStatsMonitor := NewPrcessStatsMonitor(config, pmc, s3mc, asc, cc)

	//go processStatsMonitor.MonitorProcess()
	go processStatsMonitor.MonitorTenantsProcessMetrics()
	//go processStatsMonitor.MonitorTenantsS3Stats()

}
