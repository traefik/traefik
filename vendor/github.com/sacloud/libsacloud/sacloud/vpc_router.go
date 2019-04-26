package sacloud

import (
	"fmt"
	"net"
)

// VPCRouter VPCルーター
type VPCRouter struct {
	*Appliance // アプライアンス共通属性

	Remark   *VPCRouterRemark   `json:",omitempty"` // リマーク
	Settings *VPCRouterSettings `json:",omitempty"` // VPCルーター設定リスト
}

// VPCRouterRemark リマーク
type VPCRouterRemark struct {
	*ApplianceRemarkBase
	// TODO Zone
	//Zone *Resource
}

// VPCRouterSettings VPCルーター設定リスト
type VPCRouterSettings struct {
	Router *VPCRouterSetting `json:",omitempty"` // VPCルーター設定
}

// CreateNewVPCRouter VPCルーター作成
func CreateNewVPCRouter() *VPCRouter {
	return &VPCRouter{
		Appliance: &Appliance{
			Class:      "vpcrouter",
			propPlanID: propPlanID{Plan: &Resource{}},
		},
		Remark: &VPCRouterRemark{
			ApplianceRemarkBase: &ApplianceRemarkBase{
				Servers: []interface{}{""},
				Switch:  &ApplianceRemarkSwitch{},
			},
		},
		Settings: &VPCRouterSettings{
			Router: &VPCRouterSetting{},
		},
	}
}

// InitVPCRouterSetting VPCルーター設定初期化
func (v *VPCRouter) InitVPCRouterSetting() {
	settings := &VPCRouterSettings{
		Router: &VPCRouterSetting{},
	}

	if v.Settings != nil && v.Settings.Router != nil && v.Settings.Router.Interfaces != nil {
		settings.Router.Interfaces = v.Settings.Router.Interfaces
	}
	if v.Settings != nil && v.Settings.Router != nil && v.Settings.Router.VRID != nil {
		settings.Router.VRID = v.Settings.Router.VRID
	}

	v.Settings = settings
}

// IsStandardPlan スタンダードプランか判定
func (v *VPCRouter) IsStandardPlan() bool {
	return v.Plan.ID == 1
}

// IsPremiumPlan プレミアムプランか判定
func (v *VPCRouter) IsPremiumPlan() bool {
	return v.Plan.ID == 2
}

// IsHighSpecPlan ハイスペックプランか判定
func (v *VPCRouter) IsHighSpecPlan() bool {
	return v.Plan.ID == 3
}

// SetStandardPlan スタンダードプランへ設定
func (v *VPCRouter) SetStandardPlan() {
	v.Plan.SetID(1)
	v.Remark.Switch = &ApplianceRemarkSwitch{
		// Scope
		propScope: propScope{Scope: "shared"},
	}
	v.Settings = nil
}

// SetPremiumPlan プレミアムプランへ設定
func (v *VPCRouter) SetPremiumPlan(switchID string, virtualIPAddress string, ipAddress1 string, ipAddress2 string, vrid int, ipAliases []string) {
	v.Plan.SetID(2)
	v.setPremiumServices(switchID, virtualIPAddress, ipAddress1, ipAddress2, vrid, ipAliases)
}

// SetHighSpecPlan ハイスペックプランへ設定
func (v *VPCRouter) SetHighSpecPlan(switchID string, virtualIPAddress string, ipAddress1 string, ipAddress2 string, vrid int, ipAliases []string) {
	v.Plan.SetID(3)
	v.setPremiumServices(switchID, virtualIPAddress, ipAddress1, ipAddress2, vrid, ipAliases)
}

func (v *VPCRouter) setPremiumServices(switchID string, virtualIPAddress string, ipAddress1 string, ipAddress2 string, vrid int, ipAliases []string) {
	v.Remark.Switch = &ApplianceRemarkSwitch{
		ID: switchID,
	}
	v.Remark.Servers = []interface{}{
		map[string]string{"IPAddress": ipAddress1},
		map[string]string{"IPAddress": ipAddress2},
	}

	v.Settings = &VPCRouterSettings{
		Router: &VPCRouterSetting{
			Interfaces: []*VPCRouterInterface{
				{
					IPAddress: []string{
						ipAddress1,
						ipAddress2,
					},
					VirtualIPAddress: virtualIPAddress,
					IPAliases:        ipAliases,
				},
			},
			VRID: &vrid,
		},
	}

}

// HasSetting VPCルータ設定を保持しているか
func (v *VPCRouter) HasSetting() bool {
	return v.Settings != nil && v.Settings.Router != nil
}

// HasInterfaces NIC設定を保持しているか
func (v *VPCRouter) HasInterfaces() bool {
	return v.HasSetting() && v.Settings.Router.HasInterfaces()
}

// HasStaticNAT スタティックNAT設定を保持しているか
func (v *VPCRouter) HasStaticNAT() bool {
	return v.HasSetting() && v.Settings.Router.HasStaticNAT()
}

// HasPortForwarding ポートフォワーディング設定を保持しているか
func (v *VPCRouter) HasPortForwarding() bool {
	return v.HasSetting() && v.Settings.Router.HasPortForwarding()
}

// HasFirewall ファイアウォール設定を保持しているか
func (v *VPCRouter) HasFirewall() bool {
	return v.HasSetting() && v.Settings.Router.HasFirewall()
}

// HasDHCPServer DHCPサーバー設定を保持しているか
func (v *VPCRouter) HasDHCPServer() bool {
	return v.HasSetting() && v.Settings.Router.HasDHCPServer()
}

// HasDHCPStaticMapping DHCPスタティックマッピング設定を保持しているか
func (v *VPCRouter) HasDHCPStaticMapping() bool {
	return v.HasSetting() && v.Settings.Router.HasDHCPStaticMapping()
}

// HasL2TPIPsecServer L2TP/IPSecサーバを保持しているか
func (v *VPCRouter) HasL2TPIPsecServer() bool {
	return v.HasSetting() && v.Settings.Router.HasL2TPIPsecServer()
}

// HasPPTPServer PPTPサーバを保持しているか
func (v *VPCRouter) HasPPTPServer() bool {
	return v.HasSetting() && v.Settings.Router.HasPPTPServer()
}

// HasRemoteAccessUsers リモートアクセスユーザー設定を保持しているか
func (v *VPCRouter) HasRemoteAccessUsers() bool {
	return v.HasSetting() && v.Settings.Router.HasRemoteAccessUsers()
}

// HasSiteToSiteIPsecVPN サイト間VPN設定を保持しているか
func (v *VPCRouter) HasSiteToSiteIPsecVPN() bool {
	return v.HasSetting() && v.Settings.Router.HasSiteToSiteIPsecVPN()
}

// HasStaticRoutes スタティックルートを保持しているか
func (v *VPCRouter) HasStaticRoutes() bool {
	return v.HasSetting() && v.Settings.Router.HasStaticRoutes()
}

// RealIPAddress プランに応じて外部向けIPアドレスを返す
//
// Standard: IPAddress1
// Other: VirtualIPAddress
func (v *VPCRouter) RealIPAddress(index int) (string, int) {
	if !v.HasInterfaces() {
		return "", -1
	}
	for i, nic := range v.Settings.Router.Interfaces {
		if i == index {
			if index > 0 && nic == nil {
				return "", -1
			}

			if index == 0 && v.IsStandardPlan() {
				return v.Interfaces[0].IPAddress, v.Interfaces[0].Switch.Subnet.NetworkMaskLen
			}

			nwMask := nic.NetworkMaskLen
			if index == 0 {
				nwMask = v.Interfaces[0].Switch.Subnet.NetworkMaskLen
			}

			if v.IsStandardPlan() {
				return nic.IPAddress[0], nwMask
			}
			return nic.VirtualIPAddress, nwMask
		}
	}
	return "", -1
}

// FindBelongsInterface 指定のIPアドレスが所属するIPレンジを持つインターフェースを取得
func (v *VPCRouter) FindBelongsInterface(ip net.IP) (int, *VPCRouterInterface) {
	if !v.HasInterfaces() {
		return -1, nil
	}

	for i, nic := range v.Settings.Router.Interfaces {
		nicIP, maskLen := v.RealIPAddress(i)
		if nicIP != "" {
			_, ipv4Net, err := net.ParseCIDR(fmt.Sprintf("%s/%d", nicIP, maskLen))
			if err != nil {
				return -1, nil
			}
			if ipv4Net.Contains(ip) {
				return i, nic
			}
		}
	}
	return -1, nil
}

// IPAddress1 1番目(0番目)のNICのIPアドレス1
func (v *VPCRouter) IPAddress1() string {
	return v.IPAddress1At(0)
}

// IPAddress1At 指定インデックスのNICのIPアドレス1
func (v *VPCRouter) IPAddress1At(index int) string {
	if len(v.Interfaces) <= index {
		return ""
	}

	if index == 0 {
		if v.IsStandardPlan() {
			return v.Interfaces[0].IPAddress
		}

		if !v.HasInterfaces() {
			return ""
		}
		if len(v.Settings.Router.Interfaces[0].IPAddress) < 1 {
			return ""
		}
		return v.Settings.Router.Interfaces[0].IPAddress[0]
	}

	nic := v.Settings.Router.Interfaces[index]
	if len(nic.IPAddress) < 1 {
		return ""
	}
	return nic.IPAddress[0]
}

// IPAddress2 1番目(0番目)のNICのIPアドレス2
func (v *VPCRouter) IPAddress2() string {
	return v.IPAddress2At(0)
}

// IPAddress2At 指定インデックスのNICのIPアドレス2
func (v *VPCRouter) IPAddress2At(index int) string {
	if v.IsStandardPlan() {
		return ""
	}
	if len(v.Interfaces) <= index {
		return ""
	}

	if index == 0 {
		if !v.HasInterfaces() {
			return ""
		}
		if len(v.Settings.Router.Interfaces[0].IPAddress) < 2 {
			return ""
		}
		return v.Settings.Router.Interfaces[0].IPAddress[1]
	}

	nic := v.Settings.Router.Interfaces[index]
	if len(nic.IPAddress) < 2 {
		return ""
	}
	return nic.IPAddress[1]
}

// VirtualIPAddress 1番目(0番目)のNICのVIP
func (v *VPCRouter) VirtualIPAddress() string {
	return v.VirtualIPAddressAt(0)
}

// VirtualIPAddressAt 指定インデックスのNICのVIP
func (v *VPCRouter) VirtualIPAddressAt(index int) string {
	if v.IsStandardPlan() {
		return ""
	}
	if len(v.Interfaces) <= index {
		return ""
	}

	return v.Settings.Router.Interfaces[0].VirtualIPAddress
}

// NetworkMaskLen 1番目(0番目)のNICのネットワークマスク長
func (v *VPCRouter) NetworkMaskLen() int {
	return v.NetworkMaskLenAt(0)
}

// NetworkMaskLenAt 指定インデックスのNICのネットワークマスク長
func (v *VPCRouter) NetworkMaskLenAt(index int) int {
	if !v.HasInterfaces() {
		return -1
	}
	if len(v.Interfaces) <= index {
		return -1
	}

	if index == 0 {
		return v.Interfaces[0].Switch.Subnet.NetworkMaskLen
	}

	return v.Settings.Router.Interfaces[index].NetworkMaskLen
}

// Zone スイッチから現在のゾーン名を取得
//
// Note: 共有セグメント接続時は取得不能
func (v *VPCRouter) Zone() string {
	if v.Switch != nil {
		return v.Switch.GetZoneName()
	}

	if len(v.Interfaces) > 0 && v.Interfaces[0].Switch != nil {
		return v.Interfaces[0].Switch.GetZoneName()
	}

	return ""
}

// VRID VRIDを取得
//
// スタンダードプラン、またはVRIDの参照に失敗した場合は-1を返す
func (v *VPCRouter) VRID() int {
	if v.IsStandardPlan() {
		return -1
	}

	if !v.HasSetting() || v.Settings.Router.VRID == nil {
		return -1
	}

	return *v.Settings.Router.VRID
}
