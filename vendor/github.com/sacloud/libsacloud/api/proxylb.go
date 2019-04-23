package api

import (
	"encoding/json" //	"strings"
	"fmt"

	"github.com/sacloud/libsacloud/sacloud"
)

//HACK: さくらのAPI側仕様: CommonServiceItemsの内容によってJSONフォーマットが異なるため
//      DNS/ProxyLB/シンプル監視それぞれでリクエスト/レスポンスデータ型を定義する。

// SearchProxyLBResponse ProxyLB検索レスポンス
type SearchProxyLBResponse struct {
	// Total 総件数
	Total int `json:",omitempty"`
	// From ページング開始位置
	From int `json:",omitempty"`
	// Count 件数
	Count int `json:",omitempty"`
	// CommonServiceProxyLBItems ProxyLBリスト
	CommonServiceProxyLBItems []sacloud.ProxyLB `json:"CommonServiceItems,omitempty"`
}

type proxyLBRequest struct {
	CommonServiceProxyLBItem *sacloud.ProxyLB       `json:"CommonServiceItem,omitempty"`
	From                     int                    `json:",omitempty"`
	Count                    int                    `json:",omitempty"`
	Sort                     []string               `json:",omitempty"`
	Filter                   map[string]interface{} `json:",omitempty"`
	Exclude                  []string               `json:",omitempty"`
	Include                  []string               `json:",omitempty"`
}

type proxyLBResponse struct {
	*sacloud.ResultFlagValue
	*sacloud.ProxyLB `json:"CommonServiceItem,omitempty"`
}

// ProxyLBAPI ProxyLB API
type ProxyLBAPI struct {
	*baseAPI
}

// NewProxyLBAPI ProxyLB API作成
func NewProxyLBAPI(client *Client) *ProxyLBAPI {
	return &ProxyLBAPI{
		&baseAPI{
			client: client,
			FuncGetResourceURL: func() string {
				return "commonserviceitem"
			},
			FuncBaseSearchCondition: func() *sacloud.Request {
				res := &sacloud.Request{}
				res.AddFilter("Provider.Class", "proxylb")
				return res
			},
		},
	}
}

// Find 検索
func (api *ProxyLBAPI) Find() (*SearchProxyLBResponse, error) {

	data, err := api.client.newRequest("GET", api.getResourceURL(), api.getSearchState())
	if err != nil {
		return nil, err
	}
	var res SearchProxyLBResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (api *ProxyLBAPI) request(f func(*proxyLBResponse) error) (*sacloud.ProxyLB, error) {
	res := &proxyLBResponse{}
	err := f(res)
	if err != nil {
		return nil, err
	}
	return res.ProxyLB, nil
}

func (api *ProxyLBAPI) createRequest(value *sacloud.ProxyLB) *proxyLBResponse {
	return &proxyLBResponse{ProxyLB: value}
}

// New 新規作成用パラメーター作成
func (api *ProxyLBAPI) New(name string) *sacloud.ProxyLB {
	return sacloud.CreateNewProxyLB(name)
}

// Create 新規作成
func (api *ProxyLBAPI) Create(value *sacloud.ProxyLB) (*sacloud.ProxyLB, error) {
	return api.request(func(res *proxyLBResponse) error {
		return api.create(api.createRequest(value), res)
	})
}

// Read 読み取り
func (api *ProxyLBAPI) Read(id int64) (*sacloud.ProxyLB, error) {
	return api.request(func(res *proxyLBResponse) error {
		return api.read(id, nil, res)
	})
}

// Update 更新
func (api *ProxyLBAPI) Update(id int64, value *sacloud.ProxyLB) (*sacloud.ProxyLB, error) {
	return api.request(func(res *proxyLBResponse) error {
		return api.update(id, api.createRequest(value), res)
	})
}

// UpdateSetting 設定更新
func (api *ProxyLBAPI) UpdateSetting(id int64, value *sacloud.ProxyLB) (*sacloud.ProxyLB, error) {
	req := &sacloud.ProxyLB{
		// Settings
		Settings: value.Settings,
	}
	return api.request(func(res *proxyLBResponse) error {
		return api.update(id, api.createRequest(req), res)
	})
}

// Delete 削除
func (api *ProxyLBAPI) Delete(id int64) (*sacloud.ProxyLB, error) {
	return api.request(func(res *proxyLBResponse) error {
		return api.delete(id, nil, res)
	})
}

// ChangePlan プラン変更
func (api *ProxyLBAPI) ChangePlan(id int64, newPlan sacloud.ProxyLBPlan) (*sacloud.ProxyLB, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/plan", api.getResourceURL(), id)
	)
	body := &sacloud.ProxyLB{}
	body.SetPlan(newPlan)
	realBody := map[string]interface{}{
		"CommonServiceItem": map[string]interface{}{
			"ServiceClass": body.ServiceClass,
		},
	}

	return api.request(func(res *proxyLBResponse) error {
		return api.baseAPI.request(method, uri, realBody, res)
	})
}

type proxyLBCertificateResponse struct {
	*sacloud.ResultFlagValue
	ProxyLB *sacloud.ProxyLBCertificates `json:",omitempty"`
}

// GetCertificates 証明書取得
func (api *ProxyLBAPI) GetCertificates(id int64) (*sacloud.ProxyLBCertificates, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%d/proxylb/sslcertificate", api.getResourceURL(), id)
		res    = &proxyLBCertificateResponse{}
	)
	err := api.baseAPI.request(method, uri, nil, res)
	if err != nil {
		return nil, err
	}
	if res.ProxyLB == nil {
		return nil, nil
	}
	return res.ProxyLB, nil
}

// SetCertificates 証明書設定
func (api *ProxyLBAPI) SetCertificates(id int64, certs *sacloud.ProxyLBCertificates) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/proxylb/sslcertificate", api.getResourceURL(), id)
		res    = &proxyLBCertificateResponse{}
	)
	err := api.baseAPI.request(method, uri, map[string]interface{}{
		"ProxyLB": certs,
	}, res)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteCertificates 証明書削除
func (api *ProxyLBAPI) DeleteCertificates(id int64) (bool, error) {
	var (
		method = "DELETE"
		uri    = fmt.Sprintf("%s/%d/proxylb/sslcertificate", api.getResourceURL(), id)
	)
	return api.baseAPI.modify(method, uri, nil)
}

type proxyLBHealthResponse struct {
	*sacloud.ResultFlagValue
	ProxyLB *sacloud.ProxyLBStatus `json:",omitempty"`
}

// Health ヘルスチェックステータス取得
func (api *ProxyLBAPI) Health(id int64) (*sacloud.ProxyLBStatus, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%d/health", api.getResourceURL(), id)
		res    = &proxyLBHealthResponse{}
	)
	err := api.baseAPI.request(method, uri, nil, res)
	if err != nil {
		return nil, err
	}
	if res.ProxyLB == nil {
		return nil, nil
	}
	return res.ProxyLB, nil
}

// Monitor アクティビティーモニター取得
func (api *ProxyLBAPI) Monitor(id int64, body *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
	return api.baseAPI.applianceMonitorBy(id, "activity/proxylb", 0, body)
}
