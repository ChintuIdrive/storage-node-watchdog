package api

import (
	"ChintuIdrive/storage-node-watchdog/collector"
	"encoding/json"
	"net/http"
)

type SystemMetricsHandler struct {
	systemCollector *collector.SystemStatsCollector
}

func NewSystemMetricsHandler(systemCollector *collector.SystemStatsCollector) *SystemMetricsHandler {
	return &SystemMetricsHandler{
		systemCollector: systemCollector,
	}
}

func (smh *SystemMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	systemMetrics := smh.systemCollector.CollectSystemMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemMetrics)
}
