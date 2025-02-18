package api

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/dto"
	"encoding/json"
	"log"
	"net/http"
)

type S3MetricsHandler struct {
	s3MetricsCollector *collector.S3MetricCollector
	apiServerClient    *clients.APIserverClient
}

func NewS3MetricsHandler(s3MetricsCollector *collector.S3MetricCollector, apiServerClient *clients.APIserverClient) *S3MetricsHandler {
	return &S3MetricsHandler{
		s3MetricsCollector: s3MetricsCollector,
		apiServerClient:    apiServerClient,
	}
}

func (s3handler *S3MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/tenant_s3_metrics":
		s3handler.handleS3Metrics(w, r)
	case "/all_tenant_s3_metrics":
		s3handler.handleS3StatsForAllTenant(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (s3handler *S3MetricsHandler) handleS3Metrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	dns := query.Get("dns")

	if dns == "" {
		http.Error(w, "Missing dns query parameter", http.StatusBadRequest)
		return
	}
	var tenant dto.Tenant
	tenantsFromApiServer, err := s3handler.apiServerClient.GetTenatsListFromApiServer()
	if err != nil {
		log.Printf("Failed to fetch tenant list from API server: %v", err)
	}
	for _, t := range tenantsFromApiServer {
		if t.DNS == dns {
			tenant = t
			break
		}
	}
	if tenant.DNS == "" {
		http.Error(w, "Invalid tenant", http.StatusBadRequest)
		return
	}
	s3metrics, err := s3handler.s3MetricsCollector.CollectS3Metrics(tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s3metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s3handler *S3MetricsHandler) handleS3StatsForAllTenant(w http.ResponseWriter, r *http.Request) {
	// Handle another endpoint
	tenatS3StatsMap := make(map[string]*collector.S3Metrics)
	tenantsFromApiServer, err := s3handler.apiServerClient.GetTenatsListFromApiServer()
	if err != nil {
		log.Printf("Failed to fetch tenant list from API server: %v", err)
	}
	for _, t := range tenantsFromApiServer {
		s3stats, err := s3handler.s3MetricsCollector.CollectS3Metrics(t)
		if err != nil {
			//Notify it why it is not able to get the s3 metics
			log.Printf("Failed to collect S3 metrics for tenant %s: %v", t.DNS, err)
			continue
		}
		tenatS3StatsMap[t.DNS] = s3stats
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tenatS3StatsMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
