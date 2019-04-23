package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sacloud/libsacloud/sacloud"
)

// SearchMobileGatewayResponse モバイルゲートウェイ検索レスポンス
type SearchMobileGatewayResponse struct {
	// Total 総件数
	Total int `json:",omitempty"`
	// From ページング開始位置
	From int `json:",omitempty"`
	// Count 件数
	Count int `json:",omitempty"`
	// MobileGateways モバイルゲートウェイ リスト
	MobileGateways []sacloud.MobileGateway `json:"Appliances,omitempty"`
}

// MobileGatewaySIMRequest SIM一覧取得リクエスト
type MobileGatewaySIMRequest struct {
	From    int                    `json:",omitempty"`
	Count   int                    `json:",omitempty"`
	Sort    []string               `json:",omitempty"`
	Filter  map[string]interface{} `json:",omitempty"`
	Exclude []string               `json:",omitempty"`
	Include []string               `json:",omitempty"`
}

type mobileGatewayResponse struct {
	*sacloud.ResultFlagValue
	*sacloud.MobileGateway `json:"Appliance,omitempty"`
	Success                interface{} `json:",omitempty"` //HACK: さくらのAPI側仕様: 戻り値:Successがbool値へ変換できないためinterface{}
}

type mobileGatewaySIMResponse struct {
	*sacloud.ResultFlagValue
	SIM     []sacloud.SIMInfo `json:"sim,omitempty"`
	Success interface{}       `json:",omitempty"` //HACK: さくらのAPI側仕様: 戻り値:Successがbool値へ変換できないためinterface{}
}

type trafficMonitoringBody struct {
	TrafficMonitoring *sacloud.TrafficMonitoringConfig `json:"traffic_monitoring_config"`
}

type trafficStatusBody struct {
	TrafficStatus *sacloud.TrafficStatus `json:"traffic_status"`
}

// MobileGatewayAPI モバイルゲートウェイAPI
type MobileGatewayAPI struct {
	*baseAPI
}

// NewMobileGatewayAPI モバイルゲートウェイAPI作成
func NewMobileGatewayAPI(client *Client) *MobileGatewayAPI {
	return &MobileGatewayAPI{
		&baseAPI{
			client: client,
			FuncGetResourceURL: func() string {
				return "appliance"
			},
			FuncBaseSearchCondition: func() *sacloud.Request {
				res := &sacloud.Request{}
				res.AddFilter("Class", "mobilegateway")
				return res
			},
		},
	}
}

// Find 検索
func (api *MobileGatewayAPI) Find() (*SearchMobileGatewayResponse, error) {
	data, err := api.client.newRequest("GET", api.getResourceURL(), api.getSearchState())
	if err != nil {
		return nil, err
	}
	var res SearchMobileGatewayResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (api *MobileGatewayAPI) request(f func(*mobileGatewayResponse) error) (*sacloud.MobileGateway, error) {
	res := &mobileGatewayResponse{}
	err := f(res)
	if err != nil {
		return nil, err
	}
	return res.MobileGateway, nil
}

func (api *MobileGatewayAPI) createRequest(value *sacloud.MobileGateway) *mobileGatewayResponse {
	return &mobileGatewayResponse{MobileGateway: value}
}

// Create 新規作成
func (api *MobileGatewayAPI) Create(value *sacloud.MobileGateway) (*sacloud.MobileGateway, error) {
	return api.request(func(res *mobileGatewayResponse) error {
		return api.create(api.createRequest(value), res)
	})
}

// Read 読み取り
func (api *MobileGatewayAPI) Read(id int64) (*sacloud.MobileGateway, error) {
	return api.request(func(res *mobileGatewayResponse) error {
		return api.read(id, nil, res)
	})
}

// Update 更新
func (api *MobileGatewayAPI) Update(id int64, value *sacloud.MobileGateway) (*sacloud.MobileGateway, error) {
	return api.request(func(res *mobileGatewayResponse) error {
		return api.update(id, api.createRequest(value), res)
	})
}

// UpdateSetting 設定更新
func (api *MobileGatewayAPI) UpdateSetting(id int64, value *sacloud.MobileGateway) (*sacloud.MobileGateway, error) {
	req := &sacloud.MobileGateway{
		// Settings
		Settings: value.Settings,
	}
	return api.request(func(res *mobileGatewayResponse) error {
		return api.update(id, api.createRequest(req), res)
	})
}

// Delete 削除
func (api *MobileGatewayAPI) Delete(id int64) (*sacloud.MobileGateway, error) {
	return api.request(func(res *mobileGatewayResponse) error {
		return api.delete(id, nil, res)
	})
}

// Config 設定変更の反映
func (api *MobileGatewayAPI) Config(id int64) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/config", api.getResourceURL(), id)
	)
	return api.modify(method, uri, nil)
}

// IsUp 起動しているか判定
func (api *MobileGatewayAPI) IsUp(id int64) (bool, error) {
	lb, err := api.Read(id)
	if err != nil {
		return false, err
	}
	return lb.Instance.IsUp(), nil
}

// IsDown ダウンしているか判定
func (api *MobileGatewayAPI) IsDown(id int64) (bool, error) {
	lb, err := api.Read(id)
	if err != nil {
		return false, err
	}
	return lb.Instance.IsDown(), nil
}

// Boot 起動
func (api *MobileGatewayAPI) Boot(id int64) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/power", api.getResourceURL(), id)
	)
	return api.modify(method, uri, nil)
}

// Shutdown シャットダウン(graceful)
func (api *MobileGatewayAPI) Shutdown(id int64) (bool, error) {
	var (
		method = "DELETE"
		uri    = fmt.Sprintf("%s/%d/power", api.getResourceURL(), id)
	)

	return api.modify(method, uri, nil)
}

// Stop シャットダウン(force)
func (api *MobileGatewayAPI) Stop(id int64) (bool, error) {
	var (
		method = "DELETE"
		uri    = fmt.Sprintf("%s/%d/power", api.getResourceURL(), id)
	)

	return api.modify(method, uri, map[string]bool{"Force": true})
}

// RebootForce 再起動
func (api *MobileGatewayAPI) RebootForce(id int64) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/reset", api.getResourceURL(), id)
	)

	return api.modify(method, uri, nil)
}

// ResetForce リセット
func (api *MobileGatewayAPI) ResetForce(id int64, recycleProcess bool) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/reset", api.getResourceURL(), id)
	)

	return api.modify(method, uri, map[string]bool{"RecycleProcess": recycleProcess})
}

// SleepUntilUp 起動するまで待機
func (api *MobileGatewayAPI) SleepUntilUp(id int64, timeout time.Duration) error {
	handler := waitingForUpFunc(func() (hasUpDown, error) {
		return api.Read(id)
	}, 0)
	return blockingPoll(handler, timeout)
}

// SleepUntilDown ダウンするまで待機
func (api *MobileGatewayAPI) SleepUntilDown(id int64, timeout time.Duration) error {
	handler := waitingForDownFunc(func() (hasUpDown, error) {
		return api.Read(id)
	}, 0)
	return blockingPoll(handler, timeout)
}

// SleepWhileCopying コピー終了まで待機
func (api *MobileGatewayAPI) SleepWhileCopying(id int64, timeout time.Duration, maxRetry int) error {
	handler := waitingForAvailableFunc(func() (hasAvailable, error) {
		return api.Read(id)
	}, maxRetry)
	return blockingPoll(handler, timeout)
}

// AsyncSleepWhileCopying コピー終了まで待機(非同期)
func (api *MobileGatewayAPI) AsyncSleepWhileCopying(id int64, timeout time.Duration, maxRetry int) (chan (interface{}), chan (interface{}), chan (error)) {
	handler := waitingForAvailableFunc(func() (hasAvailable, error) {
		return api.Read(id)
	}, maxRetry)
	return poll(handler, timeout)
}

// ConnectToSwitch 指定のインデックス位置のNICをスイッチへ接続
func (api *MobileGatewayAPI) ConnectToSwitch(id int64, switchID int64) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/interface/%d/to/switch/%d", api.getResourceURL(), id, 1, switchID)
	)
	return api.modify(method, uri, nil)
}

// DisconnectFromSwitch 指定のインデックス位置のNICをスイッチから切断
func (api *MobileGatewayAPI) DisconnectFromSwitch(id int64) (bool, error) {
	var (
		method = "DELETE"
		uri    = fmt.Sprintf("%s/%d/interface/%d/to/switch", api.getResourceURL(), id, 1)
	)
	return api.modify(method, uri, nil)
}

// GetDNS DNSサーバ設定 取得
func (api *MobileGatewayAPI) GetDNS(id int64) (*sacloud.MobileGatewayResolver, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/dnsresolver", api.getResourceURL(), id)
	)

	data, err := api.client.newRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}
	var res sacloud.MobileGatewayResolver
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, err
}

// SetDNS DNSサーバ設定
func (api *MobileGatewayAPI) SetDNS(id int64, dns *sacloud.MobileGatewayResolver) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/dnsresolver", api.getResourceURL(), id)
	)

	return api.modify(method, uri, dns)
}

// GetSIMRoutes SIMルート 取得
func (api *MobileGatewayAPI) GetSIMRoutes(id int64) ([]*sacloud.MobileGatewaySIMRoute, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/simroutes", api.getResourceURL(), id)
	)

	data, err := api.client.newRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}
	var res sacloud.MobileGatewaySIMRoutes

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res.SIMRoutes, err
}

// SetSIMRoutes SIMルート 設定
func (api *MobileGatewayAPI) SetSIMRoutes(id int64, simRoutes *sacloud.MobileGatewaySIMRoutes) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/simroutes", api.getResourceURL(), id)
	)

	return api.modify(method, uri, simRoutes)
}

// AddSIMRoute SIMルート 個別追加
func (api *MobileGatewayAPI) AddSIMRoute(id int64, simID int64, prefix string) (bool, error) {

	routes, err := api.GetSIMRoutes(id)
	if err != nil {
		return false, err
	}

	param := &sacloud.MobileGatewaySIMRoutes{
		SIMRoutes: routes,
	}
	index, added := param.AddSIMRoute(simID, prefix)
	if index < 0 || added == nil {
		return false, nil
	}

	return api.SetSIMRoutes(id, param)
}

// DeleteSIMRoute SIMルート 個別削除
func (api *MobileGatewayAPI) DeleteSIMRoute(id int64, simID int64, prefix string) (bool, error) {

	routes, err := api.GetSIMRoutes(id)
	if err != nil {
		return false, err
	}

	param := &sacloud.MobileGatewaySIMRoutes{
		SIMRoutes: routes,
	}
	deleted := param.DeleteSIMRoute(simID, prefix)
	if !deleted {
		return false, nil
	}

	_, err = api.SetSIMRoutes(id, param)
	return deleted, err
}

// DeleteSIMRoutes SIMルート 全件削除
func (api *MobileGatewayAPI) DeleteSIMRoutes(id int64) (bool, error) {
	return api.SetSIMRoutes(id, &sacloud.MobileGatewaySIMRoutes{
		SIMRoutes: []*sacloud.MobileGatewaySIMRoute{},
	})
}

// ListSIM SIM一覧取得
func (api *MobileGatewayAPI) ListSIM(id int64, req *MobileGatewaySIMRequest) ([]sacloud.SIMInfo, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/sims", api.getResourceURL(), id)
	)

	data, err := api.client.newRequest(method, uri, req)
	if err != nil {
		return nil, err
	}
	var res mobileGatewaySIMResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res.SIM, nil
}

// AddSIM SIM登録
func (api *MobileGatewayAPI) AddSIM(id int64, simID int64) (bool, error) {
	var (
		method = "POST"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/sims", api.getResourceURL(), id)
	)

	return api.modify(method, uri, map[string]interface{}{
		"sim": map[string]interface{}{
			"resource_id": fmt.Sprintf("%d", simID),
		},
	})
}

// DeleteSIM SIM登録
func (api *MobileGatewayAPI) DeleteSIM(id int64, simID int64) (bool, error) {
	var (
		method = "DELETE"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/sims/%d", api.getResourceURL(), id, simID)
	)
	return api.modify(method, uri, nil)
}

// Logs セッションログ取得(複数SIM)
func (api *MobileGatewayAPI) Logs(id int64, body interface{}) ([]sacloud.SIMLog, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/sessionlog", api.getResourceURL(), id)
	)

	res := &simLogResponse{}
	err := api.baseAPI.request(method, uri, body, res)
	if err != nil {
		return nil, err
	}
	return res.Logs, nil
}

// GetTrafficMonitoringConfig トラフィックコントロール 取得
func (api *MobileGatewayAPI) GetTrafficMonitoringConfig(id int64) (*sacloud.TrafficMonitoringConfig, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/traffic_monitoring", api.getResourceURL(), id)
	)

	res := &trafficMonitoringBody{}
	err := api.baseAPI.request(method, uri, nil, res)
	if err != nil {
		return nil, err
	}
	return res.TrafficMonitoring, nil
}

// SetTrafficMonitoringConfig トラフィックコントロール 設定
func (api *MobileGatewayAPI) SetTrafficMonitoringConfig(id int64, trafficMonConfig *sacloud.TrafficMonitoringConfig) (bool, error) {
	var (
		method = "PUT"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/traffic_monitoring", api.getResourceURL(), id)
	)

	req := &trafficMonitoringBody{
		TrafficMonitoring: trafficMonConfig,
	}
	return api.modify(method, uri, req)
}

// DisableTrafficMonitoringConfig トラフィックコントロール 解除
func (api *MobileGatewayAPI) DisableTrafficMonitoringConfig(id int64) (bool, error) {
	var (
		method = "DELETE"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/traffic_monitoring", api.getResourceURL(), id)
	)
	return api.modify(method, uri, nil)
}

// GetTrafficStatus 当月通信量 取得
func (api *MobileGatewayAPI) GetTrafficStatus(id int64) (*sacloud.TrafficStatus, error) {
	var (
		method = "GET"
		uri    = fmt.Sprintf("%s/%d/mobilegateway/traffic_status", api.getResourceURL(), id)
	)

	res := &trafficStatusBody{}
	err := api.baseAPI.request(method, uri, nil, res)
	if err != nil {
		return nil, err
	}
	return res.TrafficStatus, nil
}

// MonitorBy 指定位置のインターフェースのアクティビティーモニター取得
func (api *MobileGatewayAPI) MonitorBy(id int64, nicIndex int, body *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
	return api.baseAPI.applianceMonitorBy(id, "interface", nicIndex, body)
}
