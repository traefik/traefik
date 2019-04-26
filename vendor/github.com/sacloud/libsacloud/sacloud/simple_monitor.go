package sacloud

import "time"

// SimpleMonitor シンプル監視
type SimpleMonitor struct {
	*Resource        // ID
	propName         // 名称
	propDescription  // 説明
	propServiceClass // サービスクラス
	propIcon         // アイコン
	propTags         // タグ
	propCreatedAt    // 作成日時
	propModifiedAt   // 変更日時

	Settings *SimpleMonitorSettings `json:",omitempty"` // 設定
	Status   *SimpleMonitorStatus   `json:",omitempty"` // ステータス
	Provider *SimpleMonitorProvider `json:",omitempty"` // プロバイダ

}

// SimpleMonitorSettings シンプル監視設定 リスト
type SimpleMonitorSettings struct {
	SimpleMonitor *SimpleMonitorSetting `json:",omitempty"` // シンプル監視設定値
}

// SimpleMonitorSetting シンプル監視設定
type SimpleMonitorSetting struct {
	DelayLoop   int                       `json:",omitempty"` // 監視間隔
	HealthCheck *SimpleMonitorHealthCheck `json:",omitempty"` // ヘルスチェック
	Enabled     string                    `json:",omitempty"` // 有効/無効
	NotifyEmail *SimpleMonitorNotify      `json:",omitempty"` // Email通知
	NotifySlack *SimpleMonitorNotify      `json:",omitempty"` // Slack通知
}

// SimpleMonitorStatus シンプル監視ステータス
type SimpleMonitorStatus struct {
	Target string `json:",omitempty"` // 対象(IP or FQDN)
}

// SimpleMonitorProvider プロバイダ
type SimpleMonitorProvider struct {
	*Resource        // ID
	propName         // 名称
	propServiceClass // サービスクラス

	Class string `json:",omitempty"` // クラス
}

// SimpleMonitorHealthCheck ヘルスチェック
type SimpleMonitorHealthCheck struct {
	Protocol          string `json:",omitempty"` // プロトコル
	Port              string `json:",omitempty"` // ポート
	Path              string `json:",omitempty"` // HTTP/HTTPS監視の場合のリクエストパス
	Status            string `json:",omitempty"` // HTTP/HTTPS監視の場合の期待ステータスコード
	SNI               string `json:",omitempty"` // HTTPS監視時のSNI有効/無効
	Host              string `json:",omitempty"` // 対象ホスト(IP or FQDN)
	BasicAuthUsername string `json:",omitempty"` // HTTP/HTTPS監視の場合のBASIC認証 ユーザー名
	BasicAuthPassword string `json:",omitempty"` // HTTP/HTTPS監視の場合のBASIC認証 パスワード
	QName             string `json:",omitempty"` // DNS監視の場合の問い合わせFQDN
	ExpectedData      string `json:",omitempty"` // 期待値
	Community         string `json:",omitempty"` // SNMP監視の場合のコミュニティ名
	SNMPVersion       string `json:",omitempty"` // SNMP監視 SNMPバージョン
	OID               string `json:",omitempty"` // SNMP監視 OID
	RemainingDays     int    `json:",omitempty"` // SSL証明書 有効残日数
}

// SimpleMonitorNotify シンプル監視通知
type SimpleMonitorNotify struct {
	Enabled             string `json:",omitempty"` // 有効/無効
	HTML                string `json:",omitempty"` // メール通知の場合のHTMLメール有効フラグ
	IncomingWebhooksURL string `json:",omitempty"` // Slack通知の場合のWebhook URL
}

// ESimpleMonitorHealth シンプル監視ステータス
type ESimpleMonitorHealth string

var (
	// EHealthUp Up
	EHealthUp = ESimpleMonitorHealth("UP")
	// EHealthDown Down
	EHealthDown = ESimpleMonitorHealth("DOWN")
)

// IsUp アップ
func (e ESimpleMonitorHealth) IsUp() bool {
	return e == EHealthUp
}

// IsDown ダウン
func (e ESimpleMonitorHealth) IsDown() bool {
	return e == EHealthDown
}

// SimpleMonitorHealthCheckStatus シンプル監視ステータス
type SimpleMonitorHealthCheckStatus struct {
	LastCheckedAt       time.Time
	LastHealthChangedAt time.Time
	Health              ESimpleMonitorHealth
}

// CreateNewSimpleMonitor シンプル監視作成
func CreateNewSimpleMonitor(target string) *SimpleMonitor {
	return &SimpleMonitor{
		propName: propName{Name: target},
		Provider: &SimpleMonitorProvider{
			Class: "simplemon",
		},
		Status: &SimpleMonitorStatus{
			Target: target,
		},
		Settings: &SimpleMonitorSettings{
			SimpleMonitor: &SimpleMonitorSetting{
				HealthCheck: &SimpleMonitorHealthCheck{},
				Enabled:     "True",
				NotifyEmail: &SimpleMonitorNotify{
					Enabled: "False",
				},
				NotifySlack: &SimpleMonitorNotify{
					Enabled: "False",
				},
			},
		},
	}

}

// AllowSimpleMonitorHealthCheckProtocol シンプル監視対応プロトコルリスト
func AllowSimpleMonitorHealthCheckProtocol() []string {
	return []string{"http", "https", "ping", "tcp", "dns", "ssh", "smtp", "pop3", "snmp", "sslcertificate"}
}

func createSimpleMonitorNotifyEmail(withHTML bool) *SimpleMonitorNotify {
	n := &SimpleMonitorNotify{
		Enabled: "True",
		HTML:    "False",
	}

	if withHTML {
		n.HTML = "True"
	}

	return n
}

func createSimpleMonitorNotifySlack(incomingWebhooksURL string) *SimpleMonitorNotify {
	return &SimpleMonitorNotify{
		Enabled:             "True",
		IncomingWebhooksURL: incomingWebhooksURL,
	}

}

// SetTarget 対象ホスト(IP or FQDN)の設定
func (s *SimpleMonitor) SetTarget(target string) {
	s.Name = target
	s.Status.Target = target
}

// SetHealthCheckPing pingでのヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckPing() {
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol: "ping",
	}
}

// SetHealthCheckTCP TCPでのヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckTCP(port string) {
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol: "tcp",
		Port:     port,
	}
}

// SetHealthCheckHTTP HTTPでのヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckHTTP(port string, path string, status string, host string, user, pass string) {
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol:          "http",
		Port:              port,
		Path:              path,
		Status:            status,
		Host:              host,
		BasicAuthUsername: user,
		BasicAuthPassword: pass,
	}
}

// SetHealthCheckHTTPS HTTPSでのヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckHTTPS(port string, path string, status string, host string, sni bool, user, pass string) {
	strSNI := "False"
	if sni {
		strSNI = "True"
	}
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol:          "https",
		Port:              port,
		Path:              path,
		Status:            status,
		Host:              host,
		SNI:               strSNI,
		BasicAuthUsername: user,
		BasicAuthPassword: pass,
	}
}

// SetHealthCheckDNS DNSクエリでのヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckDNS(qname string, expectedData string) {
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol:     "dns",
		QName:        qname,
		ExpectedData: expectedData,
	}
}

// SetHealthCheckSSH SSHヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckSSH(port string) {
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol: "ssh",
		Port:     port,
	}
}

// SetHealthCheckSMTP SMTPヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckSMTP(port string) {
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol: "smtp",
		Port:     port,
	}
}

// SetHealthCheckPOP3 POP3ヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckPOP3(port string) {
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol: "pop3",
		Port:     port,
	}
}

// SetHealthCheckSNMP SNMPヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckSNMP(community string, version string, oid string, expectedData string) {
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol:     "snmp",
		Community:    community,
		SNMPVersion:  version,
		OID:          oid,
		ExpectedData: expectedData,
	}
}

// SetHealthCheckSSLCertificate SSLサーバ証明書有効期限ヘルスチェック設定
func (s *SimpleMonitor) SetHealthCheckSSLCertificate(remainingDays int) {
	// set default
	if remainingDays < 0 {
		remainingDays = 30
	}
	s.Settings.SimpleMonitor.HealthCheck = &SimpleMonitorHealthCheck{
		Protocol:      "sslcertificate",
		RemainingDays: remainingDays,
	}
}

// EnableNotifyEmail Email通知の有効か
func (s *SimpleMonitor) EnableNotifyEmail(withHTML bool) {
	s.Settings.SimpleMonitor.NotifyEmail = createSimpleMonitorNotifyEmail(withHTML)
}

// DisableNotifyEmail Email通知の無効化
func (s *SimpleMonitor) DisableNotifyEmail() {
	s.Settings.SimpleMonitor.NotifyEmail = &SimpleMonitorNotify{
		Enabled: "False",
	}
}

// EnableNofitySlack Slack通知の有効化
func (s *SimpleMonitor) EnableNofitySlack(incomingWebhooksURL string) {
	s.Settings.SimpleMonitor.NotifySlack = createSimpleMonitorNotifySlack(incomingWebhooksURL)
}

// DisableNotifySlack Slack通知の無効化
func (s *SimpleMonitor) DisableNotifySlack() {
	s.Settings.SimpleMonitor.NotifySlack = &SimpleMonitorNotify{
		Enabled: "False",
	}
}

// SetDelayLoop 監視間隔の設定
func (s *SimpleMonitor) SetDelayLoop(loop int) {
	s.Settings.SimpleMonitor.DelayLoop = loop
}
