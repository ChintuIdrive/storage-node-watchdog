package monitor

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/dto"
	"log"
	"time"
)

type PrcessStatsMonitor struct {
	apiServerClient     *clients.APIserverClient
	controllerClient    *clients.ControllerClient
	procMetricCollector *collector.ProcesMetricsCollector
	s3MetricsCollector  *collector.S3MetricCollector
}

func NewPrcessStatsMonitor(pmc *collector.ProcesMetricsCollector, s3mc *collector.S3MetricCollector, ac *clients.APIserverClient, cc *clients.ControllerClient) *PrcessStatsMonitor {
	return &PrcessStatsMonitor{
		procMetricCollector: pmc,
		s3MetricsCollector:  s3mc,
		apiServerClient:     ac,
		controllerClient:    cc,
	}
}

//var monitoredProcesses = []string{"minio", "e2_node_controller_service", "trash-cleaner-service", "rclone", "kes", "vault", "load-simulator"}

func (psa *PrcessStatsMonitor) MonitorProcess() {
	for {
		processMetrics := psa.procMetricCollector.CollectProcessMetrics()
		for _, metric := range processMetrics {
			// analyze metric and notify to admin using api server api
			log.Printf("analyzing prcess:%s", metric.Name)
			log.Printf("Process: %s, PID: %d, CPU Usage: %.2f%%, Memory Usage: %.2f%%", metric.Name, metric.PID, metric.CPUUsage, metric.MemUsage)
		}

		time.Sleep(15 * time.Minute) // Adjust interval as needed
	}
}

func (psm *PrcessStatsMonitor) MonitorTenantsProcessMetrics() {

	for {
		tenantsFromApiServer, err := psm.apiServerClient.GetTenatsListFromApiServer()
		if err != nil {
			//notify watchdog not able to fetch tenantlist from api server
			log.Printf("Failed to fetch tenant list from API server: %v", err)
		}
		runningTenats := psm.procMetricCollector.CollectRunningTenantProcMetrics()
		for _, tenant := range tenantsFromApiServer {
			tenantProcessInfo, err := psm.controllerClient.GetTenantWithProcessInfo(tenant)
			if err != nil {
				//notify why it is not able to get
				log.Printf("Tenant %s from API server not found in controller tenant list", tenant.DNS)
				continue
			}
			runningTenant, found := findRunningMinioProc(*tenantProcessInfo, runningTenats)
			if !found {
				// minio process not found for tenant in running minio process list
				log.Printf("Minio process for tenant %s not found in running minio process list", tenantProcessInfo.DNS)
				continue
			}

			// analyse metrics and notify if any metric reached the threshold

			log.Printf("Tenant: %s, PID: %d, CPU Usage: %.2f%%, Memory Usage: %.2f%%", tenantProcessInfo.DNS, runningTenant.PID, runningTenant.CPUUsage, runningTenant.MemUsage)
		}
		time.Sleep(15 * time.Minute) // Adjust interval as needed
	}

}

func (psm *PrcessStatsMonitor) MonitorTenantsS3Stats() {

	for {
		tenantsFromApiServer, err := psm.apiServerClient.GetTenatsListFromApiServer()
		if err != nil {
			//notify watchdog not able to fetch tenantlist from api server
			log.Printf("Failed to fetch tenant list from API server: %v", err)
		}
		for _, tenant := range tenantsFromApiServer {
			s3stats, err := psm.s3MetricsCollector.CollectS3Metrics(tenant)
			if err != nil {
				//Notify it why it is not able to get the s3 metics
				log.Printf("Failed to collect S3 metrics for tenant %s: %v", tenant.DNS, err)
				continue
			}
			// analyze s3stats and notify if any metric reach the threshold
			log.Printf("Tenant: %s, BucketCount: %d, Time taken in bucket listing: %v", s3stats.DNS, s3stats.BucketsCount, s3stats.BucketListingDuration)
			for bucket, objMetric := range s3stats.ObjectMetricsMap {

				log.Printf("Tenant: %s, Bucket: %s, ObjectCount %d, Time taken in object listing: %v", s3stats.DNS, bucket, objMetric.ObjectsCount, objMetric.ObjecttListingDuration.Seconds())
			}
		}
		time.Sleep(15 * time.Minute) // Adjust interval as needed

	}

}

func findRunningMinioProc(tenant dto.TenatWithProcessInfo, minioMetrics []collector.TenantProcessMetrics) (collector.TenantProcessMetrics, bool) {
	for _, miniotenat := range minioMetrics {
		if tenant.ProcessID == int(miniotenat.PID) {
			return miniotenat, true
		}
	}
	return collector.TenantProcessMetrics{}, false
}
