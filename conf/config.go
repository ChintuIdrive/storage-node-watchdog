package conf

import (
	"ChintuIdrive/storage-node-watchdog/cryption"
	"ChintuIdrive/storage-node-watchdog/dto"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	LogFilePath        string               `json:"log-file-path"`
	TenantProcessName  string               `json:"tenant-process-name"`
	MonitoredProcesses []string             `json:"monitored-processes"`
	MonitoredDisks     []string             `json:"monitored-disks"`
	ApiServerConfig    *ApiServerConfig     `json:"api-server-config"`
	ControllerConfig   *ControllerConfig    `json:"controller-config"`
	TenatS3ConfigMap   map[string]*S3Config `json:"tenant-s3-config-map"`
}

type ApiServerConfig struct {
	NodeId        string `json:"node-id"`
	APIPort       string `json:"api-port"`
	APIServerKey  string `json:"api-server-key"`
	APIServerDNS  string `json:"api-server-dns"`
	TenantListApi string `json:"tenant-list-api"`
}

type ControllerConfig struct {
	AccessKeyDir         string `json:"access-keys-dir"`
	ControllerDNS        string `json:"controller-dns"`
	AddServiceAccountApi string `json:"add-service-account-api"`
	GetTenantInfoApi     string `json:"get-tenant-info-api"`
}

type S3Info struct {
	S3Credentials cryption.SecretData `json:"s3-credential"`
	S3Config      S3Config            `json:"s3-config"`
}
type S3Config struct {
	DNS            string `json:"dns"`
	BucketSelector int    `json:"bucket-selector"`
	PageSelector   int    `json:"page-selector"`
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
	return &Config{
		LogFilePath: "watchdog.log",

		ApiServerConfig: &ApiServerConfig{
			NodeId:        "nc1",
			APIPort:       ":8080",
			APIServerKey:  "E8AA3FBB0F512B32",
			APIServerDNS:  "e2-api.edgedrive.com",
			TenantListApi: "api/tenant/list",
		},
		ControllerConfig: &ControllerConfig{
			AccessKeyDir:         "access-keys",
			ControllerDNS:        "localhost:44344",
			AddServiceAccountApi: "admin/v1/add_service_account",
			GetTenantInfoApi:     "admin/v1/get_tenant_info",
		},

		TenantProcessName: "minio",
		MonitoredProcesses: []string{
			"e2_node_controller_service",
			"trash-cleaner-service",
			"rclone",
			"kes",
			"vault",
			"load-simulator",
		},

		MonitoredDisks: []string{
			"/",
			"/data1",
			"/data2",
			"/data3",
			"/data4",
		},
		TenatS3ConfigMap: make(map[string]*S3Config),
	}
}

func (config *Config) GetProcessToMonitor() []string {
	return config.MonitoredProcesses
}

func (config *Config) AddProcessToMonitor(processName string) {
	//TODO: validate process name
	config.MonitoredProcesses = append(config.MonitoredProcesses, processName)
}

func (config *Config) GetDisksToMonitor() []string {
	return config.MonitoredDisks
}

func (config *Config) AddDiskToMonitor(diskName string) {
	//TODO: validate disk name
	config.MonitoredDisks = append(config.MonitoredDisks, diskName)
}

func (config *Config) AddDefaultS3Config(tenant dto.Tenant) (*S3Config, error) {
	s3config := &S3Config{
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

	config.TenatS3ConfigMap[tenant.DNS] = s3config

	return s3config, nil
}

func (config *Config) GetS3Config(tenant dto.Tenant) (*S3Config, error) {
	s3configPath := filepath.Join(config.ControllerConfig.AccessKeyDir, tenant.DNS, "s3-config.json")
	data, err := os.ReadFile(s3configPath)
	if err != nil {
		// If the file does not exist, create a default S3Config
		log.Printf("S3 configuration not available for tenant %s, adding default configuration", tenant.DNS)

		return config.AddDefaultS3Config(tenant)
	}

	var s3config S3Config
	err = json.Unmarshal(data, &s3config)
	if err != nil {
		// If there is an error in unmarshalling, return a default S3Config
		return config.AddDefaultS3Config(tenant)
	}

	config.TenatS3ConfigMap[tenant.DNS] = &s3config
	return &s3config, nil
}

func (config *Config) LoadS3Config(tenantsFromApiServer []dto.Tenant) {
	for _, tenant := range tenantsFromApiServer {
		config.AddDefaultS3Config(tenant)
	}
}
