package conf

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/dto"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/shirou/gopsutil/process"
)

var MonitoredDisks = []string{
	"/",
	"/data1",
	"/data2",
	"/data3",
	"/data4",
}

type Config struct {
	LogFilePath       string                `json:"log-file-path"`
	TenantProcessName string                `json:"tenant-process-name"`
	MonitoredDisks    []string              `json:"monitored-disks"`
	ApiServerConfig   *dto.ApiServerConfig  `json:"api-server-config"`
	ControllerConfig  *dto.ControllerConfig `json:"controller-config"`
	//SystemMetrics         dto.SystemMetrics             `json:"system-metrics"`
	MonitoredProcesses    []*dto.MonitoredProcess       `json:"monitored-processes"`
	TenantProcessList     []*dto.MonitoredTenantProcess `json:"monitored-tenant-processes"`
	TenantS3MetricsConfig []*dto.TenantS3Metrics
	//ResourceMetrics    []*MetricConfig   `json:"metric-threshold-map"`
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func GetDefaultConfig() *Config {
	asc := &dto.ApiServerConfig{
		NodeId:        "nc1",
		APIPort:       ":8080",
		APIServerKey:  "E8AA3FBB0F512B32",
		APIServerDNS:  "e2-api.edgedrive.com",
		TenantListApi: "api/tenant/list",
	}
	apiClinet := clients.NewApiServerClient(asc)
	tenants, err := apiClinet.GetTenatsListFromApiServer()
	if err != nil {
		log.Printf("error while listing tenants fro api server %v", err)
	}

	return &Config{
		LogFilePath: "watchdog.log",

		ApiServerConfig: asc,
		ControllerConfig: &dto.ControllerConfig{
			AccessKeyDir:         "access-keys",
			ControllerDNS:        "localhost:44344",
			AddServiceAccountApi: "admin/v1/add_service_account",
			GetTenantInfoApi:     "admin/v1/get_tenant_info",
		},

		TenantProcessName:     "minio",
		MonitoredProcesses:    GetMonitoredProcessList(),
		TenantProcessList:     loadTenantProcess(tenants),
		TenantS3MetricsConfig: loadTenantS3Config(tenants),
		MonitoredDisks:        MonitoredDisks,
		// SystemMetrics: dto.SystemMetrics{
		// 	ResourceMetrics: GetResourceMetrics(),
		// 	DiskMetrics:     GetDiskMetrics(MonitoredDisks),
		// },
		//ResourceMetrics: GetResorceMetrics(),
		//TenatS3ConfigMap: make(map[string]*S3Config),
	}
}

func (config *Config) GetProcessToMonitor() []*dto.MonitoredProcess {
	return config.MonitoredProcesses
}

func (config *Config) AddProcessToMonitor(processName string) bool {
	//TODO: validate process name
	processList, _ := process.Processes()
	existingProcess := slices.ContainsFunc(processList, func(proc *process.Process) bool {
		name, _ := proc.Name()
		return name == processName
	})

	if !existingProcess {
		process := &dto.MonitoredProcess{
			Name: processName,
			CPUusage: &dto.Metric[float64]{
				Name:             "cpu_usage",
				Threshold:        90,
				HighLoadDuration: 1 * time.Minute,
			},
			MemUsage: &dto.Metric[float64]{
				Name:             "mem_usage",
				Threshold:        20,
				HighLoadDuration: 1 * time.Minute,
			},
			ConnectionsCount: &dto.Metric[int]{
				Name:             "conn_count",
				Threshold:        15,
				HighLoadDuration: 1 * time.Minute,
			},
		}
		config.MonitoredProcesses = append(config.MonitoredProcesses, process)
		log.Printf("%s process found in the system and added to monitored process list", processName)
		return true
	}

	log.Printf("%s process not found in the system", processName)
	return false
}

func (config *Config) GetDisksToMonitor() []string {
	return config.MonitoredDisks
}

func (config *Config) AddDiskToMonitor(diskName string) {
	//TODO: validate disk name
	config.MonitoredDisks = append(config.MonitoredDisks, diskName)
}

func (config *Config) AddDefaultS3Config(tenant dto.Tenant) (*dto.S3Config, error) {
	s3config := &dto.S3Config{
		DNS:            tenant.DNS,
		BucketSelector: 1,
		PageSelector:   1,
	}
	s3configDir := filepath.Join(config.ControllerConfig.AccessKeyDir, tenant.DNS)
	if _, err := os.Stat(s3configDir); os.IsNotExist(err) {
		err := os.MkdirAll(s3configDir, os.ModePerm)
		if err != nil {
			//log.Fatalf("Failed to create access key directory: %v", err)
			return nil, err
		}
	}
	s3configPath := filepath.Join(s3configDir, "s3-config.json")
	s3configFile, err := os.Create(s3configPath)
	if err != nil {
		//log.Fatalf("Failed to create s3-config file for tenant %s: %v", tenant.DNS, err)
		return nil, err
	}
	defer s3configFile.Close()

	s3configData, err := json.MarshalIndent(s3config, "", "  ")
	if err != nil {
		//log.Fatalf("Failed to marshal s3-config data for tenant %s: %v", tenant.DNS, err)
		return nil, err
	}

	s3configFile.Write(s3configData)

	//config.TenatS3ConfigMap[tenant.DNS] = s3config

	return s3config, nil
}

func (config *Config) GetS3Config(tenant dto.Tenant) (*dto.S3Config, error) {
	s3configPath := filepath.Join(config.ControllerConfig.AccessKeyDir, tenant.DNS, "s3-config.json")
	data, err := os.ReadFile(s3configPath)
	if err != nil {
		// If the file does not exist, create a default S3Config
		log.Printf("S3 configuration not available for tenant %s, adding default configuration", tenant.DNS)

		return config.AddDefaultS3Config(tenant)
	}

	var s3config dto.S3Config
	err = json.Unmarshal(data, &s3config)
	if err != nil {
		// If there is an error in unmarshalling, return a default S3Config
		return config.AddDefaultS3Config(tenant)
	}

	//config.TenatS3ConfigMap[tenant.DNS] = &s3config
	return &s3config, nil
}

func (config *Config) LoadS3Config(tenantsFromApiServer []dto.Tenant) {
	for _, tenant := range tenantsFromApiServer {
		config.AddDefaultS3Config(tenant)
	}
}

func GetResourceMetrics() []*dto.Metric[float64] {

	avgLoad1 := &dto.Metric[float64]{
		Name:             "avg_load1",
		Threshold:        2.0,
		HighLoadDuration: 1 * time.Minute,
	}
	CPUusage := &dto.Metric[float64]{
		Name:             "cpu_usage",
		Threshold:        90,
		HighLoadDuration: 1 * time.Minute,
	}

	MemoryUsage := &dto.Metric[float64]{
		Name:             "mem_usage",
		Threshold:        20,
		HighLoadDuration: 1 * time.Minute,
	}

	// DiskUsage := dto.Metric[float64]{
	// 	Name: "disk_usage",
	// 	Threshold: dto.Threshold[float64]{
	// 		Limit:    50,
	// 		Duration: 1 * time.Minute,
	// 	},
	// }
	cpuMetrics := []*dto.Metric[float64]{avgLoad1, CPUusage, MemoryUsage}
	//cpuMetrics = append(cpuMetrics, avgLoad1)

	return cpuMetrics
}

func GetDiskMetrics(disks []string) []*dto.DiskMetrics {
	diskMetrics := make([]*dto.DiskMetrics, 0)
	//iometricsMap := make(map[string][]dto.Metric[uint64], 0)
	for _, disk := range disks {

		diskMetric := &dto.DiskMetrics{
			Name: disk,
			DiskUsage: &dto.Metric[float64]{
				Name:             "disk_usage",
				Threshold:        50,
				HighLoadDuration: 5 * time.Minute,
			},
		}

		readBytes := &dto.Metric[uint64]{
			Name:             "read_bytes",
			Threshold:        100000000000000,
			HighLoadDuration: 5 * time.Minute,
		}

		writeBytes := &dto.Metric[uint64]{
			Name:             "write_bytes",
			Threshold:        100000000000000,
			HighLoadDuration: 5 * time.Minute,
		}

		readCount := &dto.Metric[uint64]{
			Name:             "read_count",
			Threshold:        100000000000,
			HighLoadDuration: 5 * time.Minute,
		}

		writeCount := &dto.Metric[uint64]{
			Name:             "write_count",
			Threshold:        10000000000,
			HighLoadDuration: 5 * time.Minute,
		}

		iometrics := []*dto.Metric[uint64]{readBytes, writeBytes, readCount, writeCount}
		//iometricsMap[disk] = ipmetrics
		diskMetric.IoMetrics = iometrics
		diskMetrics = append(diskMetrics, diskMetric)
	}
	return diskMetrics

}

func GetResorceMetrics() []*dto.MetricConfig {

	//resourceMetric := []string{"avg_load1", "cpu_usage", "mem_usage"}
	avgLoad1 := &dto.MetricConfig{
		Name:     "avg_load1",
		Limit:    3,
		Duration: 5,
	}

	cpuUsage := &dto.MetricConfig{
		Name:     "cpu_usage",
		Limit:    50,
		Duration: 5,
	}

	memUsage := &dto.MetricConfig{
		Name:     "mem_usage",
		Limit:    50,
		Duration: 5,
	}

	return []*dto.MetricConfig{avgLoad1, cpuUsage, memUsage}
}

// func (config *Config) AddReSourceMetric(metricname string, limit float64, duration int) {

// 	// if config.SystemMetrics.ResourceMetrics contains metric which name is name if it is not there then add otherwise update the values
// 	// Check if the metric exists using slices.ContainsFunc
// 	var metric *dto.Metric[float64]
// 	existingMetric := slices.ContainsFunc(config.SystemMetrics.ResourceMetrics, func(m *dto.Metric[float64]) bool {
// 		if m.Name == metricname {
// 			metric = m
// 			return true
// 		}
// 		return false
// 	})
// 	if existingMetric {
// 		metric.Threshold = limit
// 		metric.HighLoadDuration = time.Duration(duration) * time.Minute

// 	} else {
// 		metric = &dto.Metric[float64]{
// 			Name:             metricname,
// 			Threshold:        limit,
// 			HighLoadDuration: time.Duration(duration) * time.Minute, // like 5 minute
// 		}
// 		config.SystemMetrics.ResourceMetrics = append(config.SystemMetrics.ResourceMetrics, metric)
// 	}
// }

func GetMonitoredProcessList() []*dto.MonitoredProcess {
	processNames := []string{
		"e2_node_controller_service",
		"trash-cleaner-service",
		"rclone",
		"kes",
		"vault",
		"load-simulator",
	}
	processList := make([]*dto.MonitoredProcess, 0)
	for _, processName := range processNames {
		process := &dto.MonitoredProcess{
			Name: processName,
			CPUusage: &dto.Metric[float64]{
				Name:             "cpu_usage",
				Threshold:        90,
				HighLoadDuration: 1 * time.Minute,
			},
			MemUsage: &dto.Metric[float64]{
				Name:             "mem_usage",
				Threshold:        20,
				HighLoadDuration: 1 * time.Minute,
			},
			ConnectionsCount: &dto.Metric[int]{
				Name:             "conn_count",
				Threshold:        15,
				HighLoadDuration: 1 * time.Minute,
			},
		}
		processList = append(processList, process)

	}
	return processList
}

func loadTenantProcess(tenants []dto.Tenant) []*dto.MonitoredTenantProcess {
	var tenantProcList []*dto.MonitoredTenantProcess

	for _, tenant := range tenants {
		tenantProcess := &dto.MonitoredTenantProcess{
			DNS: tenant.DNS,
			MonitoredProcess: &dto.MonitoredProcess{
				CPUusage: &dto.Metric[float64]{
					Name:             "cpu_usage",
					Threshold:        90,
					HighLoadDuration: 1 * time.Minute,
				},
				MemUsage: &dto.Metric[float64]{
					Name:             "mem_usage",
					Threshold:        20,
					HighLoadDuration: 1 * time.Minute,
				},
				ConnectionsCount: &dto.Metric[int]{
					Name:             "conn_count",
					Threshold:        15,
					HighLoadDuration: 1 * time.Minute,
				},
			},
		}
		tenantProcList = append(tenantProcList, tenantProcess)
	}
	//config.TenantProcessList = tenantProcList
	return tenantProcList
}

func loadTenantS3Config(tenants []dto.Tenant) []*dto.TenantS3Metrics {

	var tenantsS3Metrics []*dto.TenantS3Metrics
	for _, tenant := range tenants {
		tenantS3Metric := &dto.TenantS3Metrics{
			//DNS: tenant.DNS,
			MonitoredTenantProcess: &dto.MonitoredTenantProcess{
				DNS: tenant.DNS,
			},
			BucketListing: &dto.Metric[time.Duration]{
				Name:             "bucket_listing",
				Threshold:        3,
				HighLoadDuration: 5 * time.Minute,
			},
			ObjectListing: &dto.Metric[time.Duration]{
				Name:             "object_listing",
				Threshold:        3,
				HighLoadDuration: 5 * time.Minute,
			},
			BucketObjectListMap: map[string]*dto.Metric[time.Duration]{},
		}
		tenantsS3Metrics = append(tenantsS3Metrics, tenantS3Metric)
	}
	return tenantsS3Metrics
}

func (config *Config) GetS3MetricCinfig(dns string) (bool, *dto.TenantS3Metrics) {
	var tenantS3Config *dto.TenantS3Metrics
	available := slices.ContainsFunc(config.TenantS3MetricsConfig, func(ts3Confif *dto.TenantS3Metrics) bool {
		if ts3Confif.DNS == dns {
			tenantS3Config = ts3Confif
			return true
		}
		return false
	})
	return available, tenantS3Config
}
