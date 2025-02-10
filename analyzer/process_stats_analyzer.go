package analyzer

import (
	"ChintuIdrive/storage-node-watchdog/collector"
	"log"
)

type PrcessStatsAnalyzer struct {
	pmc  *collector.ProcesMetricsCollector
	s3mc *collector.S3MetricCollector
}

func (psa *PrcessStatsAnalyzer) AnalyzeProcess() {
	psa.pmc.CollectProcessMetrics()
	processMetrics := psa.pmc.GetProcessMetrics()

	for _, metric := range processMetrics {
		// analyze metric and notify to admin using api server api
		log.Printf("analyzing prcess:%s", metric.Name)
		log.Printf("Process: %s, PID: %d, CPU Usage: %.2f%%, Memory Usage: %.2f%%", metric.Name, metric.PID, metric.CPUUsage, metric.MemUsage)
	}

}

func (psa *PrcessStatsAnalyzer) AnalyzeTenats() {
	psa.pmc.CollectMinioMetrics()
	minioMetrics := psa.pmc.GetMinioProcessMetrics()
	tenatListFromApiServer, err := psa.s3mc.GetTenatsListFromApiServer("nc1")
	if err != nil {
		log.Println(err.Error())
		//notify watchdog not able to fetch tenantlist from api server
	}
	tenatListFromController, err := psa.s3mc.GetTenantListFromController()
	if err != nil {
		log.Println(err.Error())
		//notify watchdog not able to fetch tenantlist from controller
	}

	for _, tenantFromApiServer := range tenatListFromApiServer {
		tenant, found := findTenant(tenantFromApiServer, tenatListFromController)
		if !found {
			// controller doesn't have minio process for the tenant
			log.Printf("Tenant %s from API server not found in controller tenant list", tenantFromApiServer.DNS)
			continue
		}

		runningMinioProc, found := findRunningMinioProc(tenant, minioMetrics)
		if !found {
			// minio process not found for tenant in running minio process list
			log.Printf("Minio process for tenant %s not found in running minio process list", tenant.DNS)
			continue
		}

		// minio process found for tenant in running minio process list
		// collect stats
		log.Printf("Tenant: %s, PID: %d, CPU Usage: %.2f%%, Memory Usage: %.2f%%", tenant.DNS, runningMinioProc.PID, runningMinioProc.CPUUsage, runningMinioProc.MemUsage)
	}

}

func findTenant(tenantFromApiServer collector.Tenant, tenatListFromController []collector.StorageServerInfo) (collector.StorageServerInfo, bool) {
	for _, tenantFromController := range tenatListFromController {
		if tenantFromApiServer.DNS == tenantFromController.DNS {
			return tenantFromController, true
		}
	}
	return collector.StorageServerInfo{}, false
}

func findRunningMinioProc(tenant collector.StorageServerInfo, minioMetrics []collector.TenantProcessMetrics) (collector.TenantProcessMetrics, bool) {
	for _, miniotenat := range minioMetrics {
		if tenant.ProcessID == int(miniotenat.PID) {
			return miniotenat, true
		}
	}
	return collector.TenantProcessMetrics{}, false
}
