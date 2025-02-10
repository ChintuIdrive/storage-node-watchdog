package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ServiceAccount struct {
	DNS                  string      `json:"dns"`
	SID                  string      `json:"sid"`
	SK                   SecretKey   `json:"sk"`
	IsInternal           bool        `json:"isInternal"`
	ValidTillUtc         time.Time   `json:"validTillUtc"`
	Name                 string      `json:"name"`
	Permissions          int         `json:"permissions"`
	Buckets              interface{} `json:"buckets"`
	SecreteId            string      `json:"secreteId"`
	GlobalKeySecret      *string     `json:"globalKeySecret"`
	GlobalKeyId          *string     `json:"globalKeyId"`
	IsTestAccount        bool        `json:"isTestAccount"`
	DisableDeleteVersion bool        `json:"disableDeleteVersion"`
	DisableDeleteBucket  bool        `json:"disableDeleteBucket"`
	DisableDeleteObject  bool        `json:"disableDeleteObject"`
}

type SecretKey struct {
	CString string `json:"CString"`
	DString string `json:"DString"`
}

type TenantList struct {
	RegionName               string        `json:"region_name"`
	RegionID                 int           `json:"region_id"`
	TenantList               []Tenant      `json:"TenantList"`
	Pool                     []interface{} `json:"pool"`
	SelfConfig               interface{}   `json:"self_config"`
	AvgLoad                  int           `json:"AvgLoad"`
	HealingAvgLoadLimit      int           `json:"healing_avg_load_limit"`
	HealingThreadPerTenant   int           `json:"healing_thread_per_tenant"`
	HealingConcurrentTenants int           `json:"healing_concurrent_tenants"`
}

type Tenant struct {
	DNS                       string      `json:"dns"`
	UserID                    string      `json:"userId"`
	Password                  Password    `json:"Password"`
	E2UserID                  string      `json:"e2_userid"`
	UserType                  int         `json:"user_type"`
	PublicBucketsEnabled      bool        `json:"public_buckets_enabled"`
	EnablePublicAccessOnE2URL bool        `json:"enable_public_access_on_e2_url"`
	UserStorageNodeID         int         `json:"user_storage_node_id"`
	CnameList                 []string    `json:"cname_list"`
	AllowedOrigin             string      `json:"AllowedOrigin"`
	Compression               bool        `json:"compression"`
	MaxAPIRequests            int         `json:"MaxApiRequests"`
	APIRequestsDeadline       int         `json:"ApiRequestsDeadline"`
	Whitelist                 []string    `json:"whitelist"`
	Blacklist                 interface{} `json:"blacklist"`
	UseDEC                    bool        `json:"UseDEC"`
	Restart                   bool        `json:"Restart"`
	ForceRestart              bool        `json:"ForceRestart"`
	DownloadLimit             interface{} `json:"DownloadLimit"`
	UploadLimit               interface{} `json:"UploadLimit"`
}

type StorageServerInfo struct {
	PlannedRestart                bool
	PublicBucketsEnabled          bool
	UserType                      string
	E2UserID                      string
	UserID                        string
	DNS                           string
	ProcessID                     int
	Password                      Password
	AdminPort                     int
	S3Port                        int
	FailedS3HealthChecks          int
	ProcessStartTime              string
	CNameList                     []string
	MarkedForForceRestart         bool
	RestartImmediately            bool
	AllowedOrigin                 string
	Compression                   bool
	RestartInProcess              bool
	MaxApiRequests                int
	ApiRequestsDeadline           int
	EnablePublicAccessOnE2URL     bool
	IsFServer                     bool
	TotalRequestsProcessed        int64
	RequestsProcessInLastInterval bool
	Whitelist                     []string
	Blacklist                     []string
	NewServerAdded                bool
	ErrorStat                     int
	UseDEC                        bool
	UseHttp                       bool
	UploadLimit                   int
	DownloadLimit                 int
}

type Password struct {
	CString string `json:"CString"`
}

type S3MetricCollector struct {
}

func (s3mc *S3MetricCollector) GetTenatsListFromApiServer(nodeId string) ([]Tenant, error) {

	var tenatList []Tenant

	url := "https://e2-api.edgedrive.com/api/tenant/list"
	method := "POST"

	payload := strings.NewReader(`{` + "" + `    "NodeId":"nc1"` + "" + `}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return tenatList, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return tenatList, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return tenatList, err
	}
	var nodeInfo TenantList

	err = json.Unmarshal(body, &nodeInfo)
	if err != nil {
		return tenatList, err
	} else {
		tenatList = nodeInfo.TenantList
	}
	return tenatList, err

}

func (s3mc *S3MetricCollector) GetTenantListFromController() ([]StorageServerInfo, error) {
	var tenants []StorageServerInfo
	minioProcessPath := "/opt/e2-node-controller-1/running_processes"
	dirEntries, err := os.ReadDir(minioProcessPath)
	if err != nil {
		fmt.Println(err)
		return tenants, err
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			if strings.HasSuffix(entry.Name(), ".info") {
				minioProcessFile := filepath.Join(minioProcessPath, entry.Name())
				file, err := os.Open(minioProcessFile)
				if err != nil {
					fmt.Println(err)
					continue
				}
				defer file.Close()

				var serverInfo StorageServerInfo
				decoder := json.NewDecoder(file)
				err = decoder.Decode(&serverInfo)
				if err != nil {
					fmt.Println(err)
					continue
				}

				tenants = append(tenants, serverInfo)
				fmt.Println(minioProcessFile)
			}
		}

	}

	return tenants, err

}
