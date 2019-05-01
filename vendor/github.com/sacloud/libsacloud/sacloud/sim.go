package sacloud

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

const (
	// SIMOperatorsKDDI KDDI
	SIMOperatorsKDDI = "KDDI"
	// SIMOperatorsDOCOMO Docomo
	SIMOperatorsDOCOMO = "NTT DOCOMO"
	// SIMOperatorsSoftBank SoftBank
	SIMOperatorsSoftBank = "SoftBank"
)

// SIM SIM(CommonServiceItem)
type SIM struct {
	*Resource        // ID
	propName         // 名称
	propDescription  // 説明
	propServiceClass // サービスクラス
	propIcon         // アイコン
	propTags         // タグ
	propCreatedAt    // 作成日時
	propModifiedAt   // 変更日時
	propAvailability // 有効状態

	Status   SIMStatus   `json:",omitempty"` // ステータス
	Provider SIMProvider `json:",omitempty"` // プロバイダ
	Remark   *SIMRemark  `json:",omitempty"` // Remark
}

// SIMStatus SIMステータス
type SIMStatus struct {
	ICCID   string   `json:",omitempty"`    // ICCID
	SIMInfo *SIMInfo `json:"sim,omitempty"` // SIM詳細情報
}

// SIMInfo SIM詳細情報
type SIMInfo struct {
	ICCID                      string           `json:"iccid,omitempty"`
	IMSI                       []string         `json:"imsi,omitempty"`
	IP                         string           `json:"ip,omitempty"`
	SessionStatus              string           `json:"session_status,omitempty"`
	IMEILock                   bool             `json:"imei_lock,omitempty"`
	Registered                 bool             `json:"registered,omitempty"`
	Activated                  bool             `json:"activated,omitempty"`
	ResourceID                 string           `json:"resource_id,omitempty"`
	RegisteredDate             *time.Time       `json:"registered_date,omitempty"`
	ActivatedDate              *time.Time       `json:"activated_date,omitempty"`
	DeactivatedDate            *time.Time       `json:"deactivated_date,omitempty"`
	SIMGroupID                 string           `json:"simgroiup_id,omitempty"`
	TrafficBytesOfCurrentMonth *SIMTrafficBytes `json:"traffic_bytes_of_current_month,omitempty"`
	ConnectedIMEI              string           `json:"connected_imei,omitempty"`
}

// SIMTrafficBytes 当月通信量
type SIMTrafficBytes struct {
	UplinkBytes   uint64 `json:"uplink_bytes,omitempty"`
	DownlinkBytes uint64 `json:"downlink_bytes,omitempty"`
}

// UnmarshalJSON JSONアンマーシャル(配列、オブジェクトが混在するためここで対応)
func (s *SIMTrafficBytes) UnmarshalJSON(data []byte) error {
	targetData := strings.Replace(strings.Replace(string(data), " ", "", -1), "\n", "", -1)
	if targetData == `[]` {
		return nil
	}
	tmp := &struct {
		UplinkBytes   string `json:"uplink_bytes,omitempty"`
		DownlinkBytes string `json:"downlink_bytes,omitempty"`
	}{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	var err error
	s.UplinkBytes, err = strconv.ParseUint(tmp.UplinkBytes, 10, 64)
	if err != nil {
		return err
	}
	s.DownlinkBytes, err = strconv.ParseUint(tmp.DownlinkBytes, 10, 64)
	if err != nil {
		return err
	}
	return nil
}

// SIMRemark remark
type SIMRemark struct {
	PassCode string `json:",omitempty"`
}

// SIMProvider プロバイダ
type SIMProvider struct {
	Class        string `json:",omitempty"` // クラス
	Name         string `json:",omitempty"`
	ServiceClass string `json:",omitempty"`
}

// SIMLog SIMログ
type SIMLog struct {
	Date          *time.Time `json:"date,omitempty"`
	SessionStatus string     `json:"session_status,omitempty"`
	ResourceID    string     `json:"resource_id,omitempty"`
	IMEI          string     `json:"imei,omitempty"`
	IMSI          string     `json:"imsi,omitempty"`
}

// SIMNetworkOperatorConfig SIM通信キャリア設定
type SIMNetworkOperatorConfig struct {
	Allow       bool   `json:"allow,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	Name        string `json:"name,omitempty"`
}

// SIMNetworkOperatorConfigs SIM通信キャリア設定 リクエストパラメータ
type SIMNetworkOperatorConfigs struct {
	NetworkOperatorConfigs []*SIMNetworkOperatorConfig `json:"network_operator_config,omitempty"`
}

// CreateNewSIM SIM作成
func CreateNewSIM(name string, iccID string, passcode string) *SIM {
	return &SIM{
		Resource: &Resource{},
		propName: propName{Name: name},
		Provider: SIMProvider{
			Class: "sim",
		},
		Status: SIMStatus{
			ICCID: iccID,
		},
		Remark: &SIMRemark{
			PassCode: passcode,
		},
	}
}
