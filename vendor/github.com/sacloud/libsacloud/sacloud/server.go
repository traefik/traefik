package sacloud

import "fmt"

// Server サーバー
type Server struct {
	*Resource             // ID
	propName              // 名称
	propDescription       // 説明
	propHostName          // ホスト名
	propInterfaceDriver   // インターフェースドライバ
	propAvailability      // 有功状態
	propServerPlan        // サーバープラン
	propZone              // ゾーン
	propServiceClass      // サービスクラス
	propConnectedSwitches // 接続スイッチ
	propDisks             // ディスク配列
	propInstance          // インスタンス
	propInterfaces        // インターフェース配列
	propPrivateHost       // 専有ホスト
	propIcon              // アイコン
	propTags              // タグ
	propCreatedAt         // 作成日時
	propWaitDiskMigration // サーバ作成時のディスク作成待ち
}

// DNSServers サーバの所属するリージョンの推奨ネームサーバリスト
func (s *Server) DNSServers() []string {
	return s.Zone.Region.NameServers
}

// IPAddress サーバの1番目のNIC(eth0)のIPアドレス
func (s *Server) IPAddress() string {
	// has NIC?
	if len(s.Interfaces) == 0 {
		return ""
	}
	ip := s.Interfaces[0].IPAddress
	if ip == "" {
		ip = s.Interfaces[0].UserIPAddress
	}
	return ip
}

// Gateway デフォルトゲートウェイアドレス
func (s *Server) Gateway() string {
	if len(s.Interfaces) == 0 || s.Interfaces[0].Switch == nil {
		return ""
	}
	if s.Interfaces[0].Switch.UserSubnet != nil {
		return s.Interfaces[0].Switch.UserSubnet.DefaultRoute
	}
	return s.Interfaces[0].Switch.Subnet.DefaultRoute
}

// DefaultRoute デフォルトゲートウェイアドレス(Gatewayのエイリアス)
func (s *Server) DefaultRoute() string {
	return s.Gateway()
}

// NetworkMaskLen サーバの1番目のNIC(eth0)のネットワークマスク長
func (s *Server) NetworkMaskLen() int {
	if len(s.Interfaces) == 0 || s.Interfaces[0].Switch == nil {
		return 0
	}
	if s.Interfaces[0].Switch.UserSubnet != nil {
		return s.Interfaces[0].Switch.UserSubnet.NetworkMaskLen
	}
	return s.Interfaces[0].Switch.Subnet.NetworkMaskLen
}

// NetworkAddress サーバの1番目のNIC(eth0)のネットワークアドレス
func (s *Server) NetworkAddress() string {
	if len(s.Interfaces) == 0 || s.Interfaces[0].Switch == nil || s.Interfaces[0].Switch.Subnet == nil {
		return ""
	}
	return s.Interfaces[0].Switch.Subnet.NetworkAddress
}

// CIDRIPAddress サーバの1番目のNIC(eth0)のIPアドレス+ネットワークマスク長
func (s *Server) CIDRIPAddress() string {
	ip, maskLen := s.IPAddress(), s.NetworkMaskLen()
	if ip != "" && maskLen > 0 {
		return fmt.Sprintf("%s/%d", ip, maskLen)
	}
	return ""
}

// UpstreamType 1番目(0番目)のNICの上流ネットワーク種別
func (s *Server) UpstreamType() EUpstreamNetworkType {
	return s.UpstreamTypeAt(0)
}

// UpstreamTypeAt 指定インデックスのNICの上流ネットワーク種別
func (s *Server) UpstreamTypeAt(index int) EUpstreamNetworkType {
	if len(s.Interfaces) <= index {
		return EUpstreamNetworkUnknown
	}
	return s.Interfaces[index].UpstreamType()
}

// SwitchID 上流のスイッチのID
//
// NICがない、上流スイッチが見つからない、上流が共有セグメントの場合は-1を返す
func (s *Server) SwitchID() int64 {
	return s.SwitchIDAt(0)
}

// SwitchIDAt 上流ネットワークのスイッチのID
//
// NICがない、上流スイッチが見つからない、上流が共有セグメントの場合は-1を返す
func (s *Server) SwitchIDAt(index int) int64 {
	if len(s.Interfaces) <= index {
		return -1
	}

	nic := s.Interfaces[index]
	if nic.Switch == nil || nic.Switch.Scope == ESCopeShared {
		return -1
	}
	return nic.Switch.ID
}

// SwitchName 上流のスイッチのID
//
// NICがない、上流スイッチが見つからない、上流が共有セグメントの場合は空文字を返す
func (s *Server) SwitchName() string {
	return s.SwitchNameAt(0)
}

// SwitchNameAt 上流ネットワークのスイッチのID
//
// NICがない、上流スイッチが見つからない、上流が共有セグメントの場合は空文字を返す
func (s *Server) SwitchNameAt(index int) string {
	if len(s.Interfaces) <= index {
		return ""
	}

	nic := s.Interfaces[index]
	if nic.Switch == nil || nic.Switch.Scope == ESCopeShared {
		return ""
	}
	return nic.Switch.Name
}

// Bandwidth 上流ネットワークの帯域幅(単位:Mbps)
//
// -1: 1番目(0番目)のNICが存在しない場合 or 切断されている場合
// 0 : 制限なしの場合
// 以外: 帯域幅(Mbps)
func (s *Server) Bandwidth() int {
	return s.BandwidthAt(0)
}

// BandwidthAt 上流ネットワークの帯域幅(単位:Mbps)
//
// -1: 存在しないインデックスを取得した場合 or 切断されている場合
// 0 : 制限なしの場合
// 以外: 帯域幅(Mbps)
func (s *Server) BandwidthAt(index int) int {
	if len(s.Interfaces) <= index {
		return -1
	}

	nic := s.Interfaces[index]

	switch nic.UpstreamType() {
	case EUpstreamNetworkNone:
		return -1
	case EUpstreamNetworkShared:
		return 100
	case EUpstreamNetworkSwitch, EUpstreamNetworkRouter:
		//
		// 上流ネットワークがスイッチだった場合の帯域制限
		// https://manual.sakura.ad.jp/cloud/support/technical/network.html#support-network-03
		//

		// 専有ホストの場合は制限なし
		if s.PrivateHost != nil {
			return 0
		}

		// メモリに応じた制限
		memory := s.GetMemoryGB()
		switch {
		case memory < 32:
			return 1000
		case 32 <= memory && memory < 128:
			return 2000
		case 128 <= memory && memory < 224:
			return 5000
		case 224 <= memory:
			return 10000
		default:
			return -1
		}
	default:
		return -1
	}
}

const (
	// ServerMaxInterfaceLen サーバーに接続できるNICの最大数
	ServerMaxInterfaceLen = 10
	// ServerMaxDiskLen サーバーに接続できるディスクの最大数
	ServerMaxDiskLen = 4
)

// KeyboardRequest キーボード送信リクエスト
type KeyboardRequest struct {
	Keys []string `json:",omitempty"` // キー(複数)
	Key  string   `json:",omitempty"` // キー(単体)
}

// MouseRequest マウス送信リクエスト
type MouseRequest struct {
	X       *int                 `json:",omitempty"` // X
	Y       *int                 `json:",omitempty"` // Y
	Z       *int                 `json:",omitempty"` // Z
	Buttons *MouseRequestButtons `json:",omitempty"` // マウスボタン

}

// VNCSnapshotRequest VNCスナップショット取得リクエスト
type VNCSnapshotRequest struct {
	ScreenSaverExitTimeMS int `json:",omitempty"` // スクリーンセーバーからの復帰待ち時間
}

// MouseRequestButtons マウスボタン
type MouseRequestButtons struct {
	L bool `json:",omitempty"` // 左ボタン
	R bool `json:",omitempty"` // 右ボタン
	M bool `json:",omitempty"` // 中ボタン
}

// VNCProxyResponse VNCプロキシ取得レスポンス
type VNCProxyResponse struct {
	*ResultFlagValue
	Status       string `json:",omitempty"` // ステータス
	Host         string `json:",omitempty"` // プロキシホスト
	IOServerHost string `json:",omitempty"` // 新プロキシホスト(Hostがlocalhostの場合にこちらを利用する)
	Port         string `json:",omitempty"` // ポート番号
	Password     string `json:",omitempty"` // VNCパスワード
	VNCFile      string `json:",omitempty"` // VNC接続情報ファイル(VNCビューア用)
}

// ActualHost プロキシホスト名(Host or IOServerHost)を返す
func (r *VNCProxyResponse) ActualHost() string {
	host := r.Host
	if host == "localhost" {
		host = r.IOServerHost
	}
	return host
}

// VNCSizeResponse VNC画面サイズレスポンス
type VNCSizeResponse struct {
	Width  int `json:",string,omitempty"` // 幅
	Height int `json:",string,omitempty"` // 高さ
}

// VNCSnapshotResponse VPCスナップショットレスポンス
type VNCSnapshotResponse struct {
	Image string `json:",omitempty"` // スナップショット画像データ
}
