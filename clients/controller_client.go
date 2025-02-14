package clients

import (
	"ChintuIdrive/storage-node-watchdog/cryption"
	"ChintuIdrive/storage-node-watchdog/dto"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ControllerClient struct {
	controllerDNS     string
	addServiceAccount string
	getTenantInfo     string
}

func NewControllerClientt(controllerDNS, addServiceAccount, getTenantInfo string) *ControllerClient {
	return &ControllerClient{
		controllerDNS:     controllerDNS,
		addServiceAccount: addServiceAccount,
		getTenantInfo:     getTenantInfo,
	}
}
func (cc *ControllerClient) GetTenantListFromController() ([]dto.TenatWithProcessInfo, error) {
	var tenants []dto.TenatWithProcessInfo
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

				var serverInfo dto.TenatWithProcessInfo
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

func (cc *ControllerClient) GetAccessKeys(tenat dto.Tenant) (*cryption.SecretData, error) {

	url := fmt.Sprintf("https://%s/%s", cc.controllerDNS, cc.addServiceAccount)
	method := "POST"
	// url := "https://localhost:44344/admin/v1/add_service_account"
	// method := "POST"
	addSrvAcctReq := dto.ServiceAccountReq{
		Name: "wtachdog",
		BaseReq: dto.BaseReq{
			DNS: tenat.DNS,
			SID: tenat.UserID,
			SK: dto.SecretKey{
				CString: tenat.Password.CString,
			},
		},
		IsInternal:    false,
		IsTestAccount: false,
		Permissions:   2,
	}

	payload, err := json.Marshal(addSrvAcctReq)
	if err != nil {
		return nil, err
	}

	res, err := FireRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var accesskey *cryption.SecretData

	err = json.Unmarshal(body, &accesskey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		return accesskey, nil
	}

}

func (cc *ControllerClient) GetTenantWithProcessInfo(tenat dto.Tenant) (*dto.TenatWithProcessInfo, error) {

	url := fmt.Sprintf("https://%s/%s", cc.controllerDNS, cc.getTenantInfo)
	method := "POST"

	clientreq := dto.TenantProcessInfoReq{
		BaseReq: dto.BaseReq{
			DNS: tenat.DNS,
			SID: tenat.UserID,
			SK:  dto.SecretKey{CString: tenat.Password.CString},
		},
	}
	payload, err := json.Marshal(clientreq)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	res, err := FireRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var resp dto.TenantProcessInfoResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.TenatWithProcessInfo, err
}
