package monitor

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/conf"
)

func StartMonitoring(config *conf.Config) {
	systemStatsMonitor := NewSystemStatsMonitor()
	go systemStatsMonitor.MonitorSystemStats(config.MonitoredDisks)

	apiserverClient := clients.NewApiServerClient()
	controllerClient := clients.NewControllerClientt(config.ControllerDNS, config.AddServiceAccountApi, config.GetTenantInfoApi)
	pmc := collector.NewProcesMetricsCollector()
	s3mc := collector.NewS3MetricCollector(controllerClient)
	processStatsMonitor := NewPrcessStatsMonitor(pmc, s3mc, apiserverClient, controllerClient)

	go processStatsMonitor.MonitorProcess(config.MonitoredProcesses)
	go processStatsMonitor.MonitorTenantsProcessMetrics(config.NodeId, config.TenantProcessName)
	go processStatsMonitor.MonitorTenantsS3Stats()
}
