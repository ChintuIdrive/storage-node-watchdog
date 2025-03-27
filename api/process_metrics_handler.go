package api

import (
	"ChintuIdrive/storage-node-watchdog/collector"
	"encoding/json"
	"net/http"
)

type ProcessMetricsHandler struct {
	processCollector *collector.ProcesMetricsCollector
}

func NewProcessMetricsHandler(processCollector *collector.ProcesMetricsCollector) *ProcessMetricsHandler {
	return &ProcessMetricsHandler{
		processCollector: processCollector,
	}
}

func (pmh *ProcessMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	processMetrics := pmh.processCollector.GetProcessMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(processMetrics)
}

type RunningTenantMetricsHandler struct {
	processCollector *collector.ProcesMetricsCollector
}

func NewRunningTenantMetricsHandler(processCollector *collector.ProcesMetricsCollector) *RunningTenantMetricsHandler {
	return &RunningTenantMetricsHandler{
		processCollector: processCollector,
	}
}

func (rtmh *RunningTenantMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runningProcMetrics := rtmh.processCollector.CollectRunningTenantProcMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runningProcMetrics)
}
