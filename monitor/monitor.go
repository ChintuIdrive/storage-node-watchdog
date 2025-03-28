package monitor

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/conf"
	"ChintuIdrive/storage-node-watchdog/dto"
)

func StartMonitoring(config *conf.Config, cc *clients.ControllerClient, asc *clients.APIserverClient,
	ssc *collector.SystemStatsCollector, pmc *collector.ProcesMetricsCollector, s3mc *collector.S3MetricCollector) {
	SystemMetrics := &dto.SystemMetrics{
		ResourceMetrics: conf.GetResourceMetrics(),
		DiskMetrics:     conf.GetDiskMetrics(config.MonitoredDisks),
	}
	systemStatsMonitor := NewSystemStatsMonitor(ssc, asc, config, SystemMetrics)
	go systemStatsMonitor.MonitorSystemStats()

	processStatsMonitor := NewPrcessStatsMonitor(config, pmc, s3mc, asc, cc)

	go processStatsMonitor.MonitorProcess()
	go processStatsMonitor.MonitorTenantsProcessMetrics()
	//go processStatsMonitor.MonitorTenantsS3Stats()

}
