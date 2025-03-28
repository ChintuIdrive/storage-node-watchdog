package monitor

import (
	"ChintuIdrive/storage-node-watchdog/actions"
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/collector"
	"ChintuIdrive/storage-node-watchdog/conf"
	"ChintuIdrive/storage-node-watchdog/dto"
	"encoding/json"
	"log"
	"slices"
	"time"
)

type PrcessStatsMonitor struct {
	config              *conf.Config
	apiServerClient     *clients.APIserverClient
	controllerClient    *clients.ControllerClient
	procMetricCollector *collector.ProcesMetricsCollector
	s3MetricsCollector  *collector.S3MetricCollector
}

func NewPrcessStatsMonitor(config *conf.Config, pmc *collector.ProcesMetricsCollector, s3mc *collector.S3MetricCollector, ac *clients.APIserverClient, cc *clients.ControllerClient) *PrcessStatsMonitor {
	return &PrcessStatsMonitor{
		config:              config,
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
		for _, proc := range processMetrics {
			// analyze metric and notify to admin using api server api
			log.Printf("analyzing prcess:%s", proc.Name)
			log.Printf("Process: %s, PID: %d, CPU Usage: %.2f%%, Memory Usage: %.2f%%", proc.Name, proc.PID, proc.CPUusage.Value, proc.MemUsage.Value)
			notify, msg := proc.CPUusage.MonitorThresholdWithDuration()
			if notify {
				psa.notifyProcessUsage(proc.PID, proc.Name, msg, proc.CPUusage)
			}
			notify, msg = proc.MemUsage.MonitorThresholdWithDuration()
			if notify {
				psa.notifyProcessUsage(proc.PID, proc.Name, msg, proc.MemUsage)
			}
			notify, msg = proc.ConnectionsCount.MonitorThresholdWithDuration()
			if notify {
				psa.notifyConnectionsCount(proc.PID, proc.Name, msg, proc.ConnectionsCount)
			}

		}

		time.Sleep(20 * time.Second) // Adjust interval as needed
	}
}

func (psm *PrcessStatsMonitor) MonitorTenantsProcessMetrics() {

	for {
		tenantsFromApiServer, err := psm.apiServerClient.GetTenatsListFromApiServer()
		if err != nil {
			//notify watchdog not able to fetch tenantlist from api server
			log.Printf("Failed to fetch tenant list from API server: %v", err)
		}
		tenatsFromController := psm.procMetricCollector.CollectRunningTenantProcMetrics()
		if len(tenantsFromApiServer) > len(tenatsFromController) {
			// get tenants which is down in controller
		} else if len(tenantsFromApiServer) < len(tenatsFromController) {
			// get running minio process which is not assigned to any dns/tenant
		} else {
			//match tenant with runnining minio process
		}

		for _, tenant := range tenantsFromApiServer {
			tenantProcessInfo, err := psm.controllerClient.GetTenantWithProcessInfo(tenant)
			if err != nil {
				//notify why it is not able to get
				log.Printf("Tenant %s from API server not found in controller tenant list", tenant.DNS)
				continue
			}
			runningTenant, found := findRunningMinioProc(*tenantProcessInfo, tenatsFromController)

			if !found {
				// minio process not found for tenant in running minio process list
				log.Printf("Minio process for tenant %s not found in running minio process list", tenantProcessInfo.DNS)
				continue
			}
			runningTenant.DNS = tenantProcessInfo.DNS
			var monTenatProc *dto.MonitoredTenantProcess
			available := slices.ContainsFunc(psm.config.TenantProcessList, func(tp *dto.MonitoredTenantProcess) bool {
				if tp.DNS == runningTenant.DNS {
					monTenatProc = tp
					return true
				}
				return false
			})
			if available {
				// analyse metrics and notify if any metric reached the threshold
				monTenatProc.Name = runningTenant.Name
				monTenatProc.PID = runningTenant.PID
				monTenatProc.CPUusage.Value = runningTenant.CPUUsage
				notify, msg := monTenatProc.CPUusage.MonitorThresholdWithDuration()
				if notify {
					psm.NotifyTenantProcessUsage(runningTenant, monTenatProc.CPUusage, msg)
				}
				monTenatProc.MemUsage.Value = float64(runningTenant.MemUsage)
				notify, msg = monTenatProc.MemUsage.MonitorThresholdWithDuration()
				if notify {
					psm.NotifyTenantProcessUsage(runningTenant, monTenatProc.MemUsage, msg)
				}
				monTenatProc.ConnectionsCount.Value = runningTenant.ConnectionsCount
				notify, msg = monTenatProc.ConnectionsCount.MonitorThresholdWithDuration()
				if notify {
					psm.NotifyTenantConnCount(runningTenant, monTenatProc.ConnectionsCount, msg)
				}

				//psm.checkS3stats(runningTenant, tenant)

			} else {
				log.Printf("configuration not available for %s", runningTenant.DNS)
			}

			log.Printf("Tenant: %s, PID: %d, CPU Usage: %.2f%%, Memory Usage: %.2f%%", tenantProcessInfo.DNS, runningTenant.PID, runningTenant.CPUUsage, runningTenant.MemUsage)
		}
		time.Sleep(30 * time.Second) // Adjust interval as needed
	}

}

//func (psm *PrcessStatsMonitor) MonitorTenantsS3Stats() {

// 	for {
// 		tenantsFromApiServer, err := psm.apiServerClient.GetTenatsListFromApiServer()
// 		if err != nil {
// 			//notify watchdog not able to fetch tenantlist from api server
// 			log.Printf("Failed to fetch tenant list from API server: %v", err)
// 		}
// 		for _, tenant := range tenantsFromApiServer {
// 			psm.checkS3stats(, tenant)
// 		}
// 		time.Sleep(15 * time.Second) // Adjust interval as needed

// 	}

// }

func (psm *PrcessStatsMonitor) checkS3stats(runningTenant collector.TenantProcessMetrics, tenant dto.Tenant) {
	s3stats, err := psm.s3MetricsCollector.CollectS3Metrics(tenant)

	if err != nil {
		//Notify it why it is not able to get the s3 metics
		log.Printf("Failed to collect S3 metrics for tenant %s: %v", tenant.DNS, err)
		return
	}
	exist, tenatS3Config := psm.config.GetS3MetricCinfig(tenant.DNS)
	if exist {
		// analyze s3stats and notify if any metric reach the threshold
		tenatS3Config.BucketListing.Value = s3stats.BucketListingDuration
		notify, msg := tenatS3Config.BucketListing.MonitorThresholdWithDuration()
		if notify {
			psm.NotifyS3Stats(runningTenant, tenatS3Config.BucketListing, msg)
		}
		log.Printf("Tenant: %s, BucketCount: %d, Time taken in bucket listing: %v", s3stats.DNS, s3stats.BucketsCount, s3stats.BucketListingDuration)
		for bucket, objMetric := range s3stats.ObjectMetricsMap {
			a, b := tenatS3Config.BucketObjectListMap[bucket]
			if !b {
				tenatS3Config.BucketObjectListMap[bucket] = &dto.Metric[time.Duration]{
					Name:             tenatS3Config.ObjectListing.Name,
					Value:            objMetric.ObjecttListingDuration,
					Threshold:        tenatS3Config.BucketListing.Threshold,
					HighLoadDuration: tenatS3Config.BucketListing.HighLoadDuration,
				}
				a = tenatS3Config.BucketObjectListMap[bucket]
			} else {
				a.Value = objMetric.ObjecttListingDuration
			}
			a.MonitorThresholdWithDuration()

			log.Printf("Tenant: %s, Bucket: %s, ObjectCount %d, Time taken in object listing: %v", s3stats.DNS, bucket, objMetric.ObjectsCount, objMetric.ObjecttListingDuration.Seconds())
		}
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

func (psm *PrcessStatsMonitor) notifyProcessUsage(id int32, name, msg string, metric *dto.Metric[float64]) {
	systeNot := actions.SystemNotification[float64]{
		Type:      actions.ProcessMetric,
		NodeId:    psm.config.ApiServerConfig.NodeId,
		TimeStamp: time.Now().Format(time.RFC3339),
		Metric:    metric,
		Actions:   []actions.Action{actions.Notify},
		Message:   msg,
	}
	procNot := actions.ProcessNotification[float64]{
		SystemNotification: systeNot,
		ProcessName:        name,
		ProcessId:          id,
	}

	payload, err := json.Marshal(procNot)

	if err != nil {
		log.Printf("Error marshalling system notification: %v", err)
		return
	}

	psm.apiServerClient.Notify(payload)
}

func (psm *PrcessStatsMonitor) notifyConnectionsCount(id int32, name, msg string, metric *dto.Metric[int]) {
	systeNot := actions.SystemNotification[int]{
		Type:      actions.ProcessMetric,
		NodeId:    psm.config.ApiServerConfig.NodeId,
		TimeStamp: time.Now().Format(time.RFC3339),
		Metric:    metric,
		Actions:   []actions.Action{actions.Notify},
		Message:   msg,
	}

	procNot := actions.ProcessNotification[int]{
		SystemNotification: systeNot,
		ProcessName:        name,
		ProcessId:          id,
	}

	payload, err := json.Marshal(procNot)
	if err != nil {
		log.Printf("Error marshalling system notification: %v", err)
		return
	}

	psm.apiServerClient.Notify(payload)
}

func (psm *PrcessStatsMonitor) NotifyTenantProcessUsage(tenantProInfo collector.TenantProcessMetrics, metric *dto.Metric[float64], msg string) {
	systeNot := actions.SystemNotification[float64]{
		Type:      actions.ProcessMetric,
		NodeId:    psm.config.ApiServerConfig.NodeId,
		TimeStamp: time.Now().Format(time.RFC3339),
		Metric:    metric,
		Actions:   []actions.Action{actions.Notify},
		Message:   msg,
	}
	procNot := actions.ProcessNotification[float64]{
		SystemNotification: systeNot,
		ProcessName:        tenantProInfo.Name,
		ProcessId:          tenantProInfo.PID,
	}

	s3Not := actions.S3Notification[float64]{
		ProcessNotification: procNot,
		S3Dns:               tenantProInfo.DNS,
	}

	payload, err := json.Marshal(s3Not)
	if err != nil {
		log.Printf("Error marshalling system notification: %v", err)
		return
	}

	psm.apiServerClient.Notify(payload)
}

func (psm *PrcessStatsMonitor) NotifyTenantConnCount(tenantProInfo collector.TenantProcessMetrics, metric *dto.Metric[int], msg string) {
	systeNot := actions.SystemNotification[int]{
		Type:      actions.ProcessMetric,
		NodeId:    psm.config.ApiServerConfig.NodeId,
		TimeStamp: time.Now().Format(time.RFC3339),
		Metric:    metric,
		Actions:   []actions.Action{actions.Notify},
		Message:   msg,
	}
	procNot := actions.ProcessNotification[int]{
		SystemNotification: systeNot,
		ProcessName:        tenantProInfo.Name,
		ProcessId:          tenantProInfo.PID,
	}

	s3Not := actions.S3Notification[int]{
		ProcessNotification: procNot,
		S3Dns:               tenantProInfo.DNS,
	}

	payload, err := json.Marshal(s3Not)
	if err != nil {
		log.Printf("Error marshalling system notification: %v", err)
		return
	}

	psm.apiServerClient.Notify(payload)
}

func (psm *PrcessStatsMonitor) NotifyS3Stats(tenantProInfo collector.TenantProcessMetrics, metric *dto.Metric[time.Duration], msg string) {
	systeNot := actions.SystemNotification[time.Duration]{
		Type:      actions.S3Metric,
		NodeId:    psm.config.ApiServerConfig.NodeId,
		TimeStamp: time.Now().Format(time.RFC3339),
		Metric:    metric,
		Actions:   []actions.Action{actions.Notify},
		Message:   msg,
	}
	procNot := actions.ProcessNotification[time.Duration]{
		SystemNotification: systeNot,
		ProcessName:        tenantProInfo.Name,
		ProcessId:          tenantProInfo.PID,
	}

	s3Not := actions.S3Notification[time.Duration]{
		ProcessNotification: procNot,
		S3Dns:               tenantProInfo.DNS,
	}

	payload, err := json.Marshal(s3Not)
	if err != nil {
		log.Printf("Error marshalling system notification: %v", err)
		return
	}

	psm.apiServerClient.Notify(payload)
}
