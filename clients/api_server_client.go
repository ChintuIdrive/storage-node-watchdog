package clients

import (
	"ChintuIdrive/storage-node-watchdog/dto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type APIserverClient struct {
	apiServerDNS string
	tenantList   string
}

func NewApiServerClient() *APIserverClient {
	return &APIserverClient{}
}

func (asc *APIserverClient) GetTenatsListFromApiServer(nodeId string) ([]dto.Tenant, error) {

	var tenatList []dto.Tenant

	url := fmt.Sprintf("https://%s/%s", asc.apiServerDNS, asc.tenantList)
	method := "POST"

	payload := []byte(`{"NodeId":"nc1"}`)

	res, err := FireRequest(method, url, payload)
	if err != nil {
		return tenatList, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return tenatList, err
	}
	var nodeInfo dto.TenantList

	err = json.Unmarshal(body, &nodeInfo)
	if err != nil {
		return tenatList, err
	} else {
		tenatList = nodeInfo.TenantList
	}
	return tenatList, err

}

func FireRequest(method, url string, payload []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return res, nil
}
