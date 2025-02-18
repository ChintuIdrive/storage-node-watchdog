package api

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/conf"
	"net/http"
)

func RegisterHandlers(config *conf.Config, cc *clients.ControllerClient, asc *clients.APIserverClient,
	ssc *collector.SystemStatsCollector, pmc *collector.ProcesMetricsCollector, s3mc *collector.S3MetricCollector) {
	systemMetricsHandler := NewSystemMetricsHandler(ssc)
	http.Handle("/system_metrics", systemMetricsHandler)

	processMetricsHandler := NewProcessMetricsHandler(pmc)
	http.Handle("/process_metrics", processMetricsHandler)

	runningTenantMetricsHandler := NewRunningTenantMetricsHandler(pmc)
	http.Handle("/running_tenant_metrics", runningTenantMetricsHandler)

	s3handler := NewS3MetricsHandler(s3mc, asc)
	http.Handle("/tenant_s3_metrics", s3handler)
	http.Handle("/all_tenant_s3_metrics", s3handler)

	http.ListenAndServe(":8080", nil)
}
