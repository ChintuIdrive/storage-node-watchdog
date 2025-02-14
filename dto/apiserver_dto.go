package dto

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

type Password struct {
	CString string `json:"CString"`
}
