package clients

import (
	"ChintuIdrive/storage-node-watchdog/dto"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type APIserverClient struct {
	apiserverConfig *dto.ApiServerConfig
}

func NewApiServerClient(apiserverConfig *dto.ApiServerConfig) *APIserverClient {
	return &APIserverClient{
		apiserverConfig: apiserverConfig,
	}
}

func (asc *APIserverClient) Notify(payload []byte) {
	log.Println("Notifying to API server")
	log.Printf("Notification Payload: %s", string(payload))
	// url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, asc.apiserverConfig.Notify)
	// method := "POST"

	// res, err := FireRequest(method, url, payload)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer res.Body.Close()

}

func (asc *APIserverClient) GetTenatsListFromApiServer() ([]dto.Tenant, error) {

	var tenatList []dto.Tenant

	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, asc.apiserverConfig.TenantListApi)
	method := "POST"

	payload := []byte(fmt.Sprintf(`{"NodeId":"%s"}`, asc.apiserverConfig.NodeId))

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
