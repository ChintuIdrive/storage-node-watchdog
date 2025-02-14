package conf

import (
	"encoding/json"
	"os"
)

type Config struct {
	LogFilePath          string   `json:"log-file-path"`
	APIPort              string   `json:"api-port"`
	APIServerKey         string   `json:"api-server-key"`
	APIServerDNS         string   `json:"api-server-dns"`
	TenantListApi        string   `json:"tenant-list-api"`
	NodeId               string   `json:"node-id"`
	ControllerDNS        string   `json:"controller-dns"`
	AddServiceAccountApi string   `json:"add-service-account-api"`
	GetTenantInfoApi     string   `json:"get-tenant-info-api"`
	TenantProcessName    string   `json:"tenant-process-name"`
	MonitoredProcesses   []string `json:"monitored-processes"`
	MonitoredDisks       []string `json:"monitored-disks"`
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
		LogFilePath:          "watchdog.log",
		APIPort:              ":8080",
		APIServerKey:         "E8AA3FBB0F512B32",
		APIServerDNS:         "e2-api.edgedrive.com",
		TenantListApi:        "api/tenant/list",
		NodeId:               "nc1",
		ControllerDNS:        "localhost:44344",
		AddServiceAccountApi: "admin/v1/add_service_account",
		GetTenantInfoApi:     "admin/v1/get_tenant_info",
		TenantProcessName:    "minio",
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
	}
}
