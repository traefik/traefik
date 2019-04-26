package sacloud

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AllowDatabaseBackupWeekdays データベースバックアップ実行曜日リスト
func AllowDatabaseBackupWeekdays() []string {
	return []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
}

// Database データベース(appliance)
type Database struct {
	*Appliance // アプライアンス共通属性

	Remark   *DatabaseRemark   `json:",omitempty"` // リマーク
	Settings *DatabaseSettings `json:",omitempty"` // データベース設定
}

// DatabaseRemark データベースリマーク
type DatabaseRemark struct {
	*ApplianceRemarkBase
	propPlanID                             // プランID
	DBConf          *DatabaseCommonRemarks // コンフィグ
	Network         *DatabaseRemarkNetwork // ネットワーク
	SourceAppliance *Resource              // クローン元DB
	Zone            struct {               // ゾーン
		ID json.Number `json:",omitempty"` // ゾーンID
	}
}

// DatabaseRemarkNetwork ネットワーク
type DatabaseRemarkNetwork struct {
	NetworkMaskLen int    `json:",omitempty"` // ネットワークマスク長
	DefaultRoute   string `json:",omitempty"` // デフォルトルート
}

// UnmarshalJSON JSONアンマーシャル(配列、オブジェクトが混在するためここで対応)
func (s *DatabaseRemarkNetwork) UnmarshalJSON(data []byte) error {
	targetData := strings.Replace(strings.Replace(string(data), " ", "", -1), "\n", "", -1)
	if targetData == `[]` {
		return nil
	}

	tmp := &struct {
		// NetworkMaskLen
		NetworkMaskLen int `json:",omitempty"`
		// DefaultRoute
		DefaultRoute string `json:",omitempty"`
	}{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	s.NetworkMaskLen = tmp.NetworkMaskLen
	s.DefaultRoute = tmp.DefaultRoute
	return nil
}

// DatabaseCommonRemarks リマークリスト
type DatabaseCommonRemarks struct {
	Common *DatabaseCommonRemark // Common
}

// DatabaseCommonRemark リマーク
type DatabaseCommonRemark struct {
	DatabaseName     string `json:",omitempty"` // 名称
	DatabaseRevision string `json:",omitempty"` // リビジョン
	DatabaseTitle    string `json:",omitempty"` // タイトル
	DatabaseVersion  string `json:",omitempty"` // バージョン
}

// DatabaseSettings データベース設定リスト
type DatabaseSettings struct {
	DBConf *DatabaseSetting `json:",omitempty"` // コンフィグ
}

// DatabaseSetting データベース設定
type DatabaseSetting struct {
	Backup      *DatabaseBackupSetting      `json:",omitempty"` // バックアップ設定
	Common      *DatabaseCommonSetting      `json:",oitempty"`  // 共通設定
	Replication *DatabaseReplicationSetting `json:",omitempty"` // レプリケーション設定
}

// DatabaseServer データベースサーバー情報
type DatabaseServer struct {
	IPAddress  string `json:",omitempty"` // IPアドレス
	Port       string `json:",omitempty"` // ポート
	Enabled    string `json:",omitempty"` // 有効/無効
	Status     string `json:",omitempty"` // ステータス
	ActiveConn string `json:",omitempty"` // アクティブコネクション
}

// DatabasePlan プラン
type DatabasePlan int

var (
	// DatabasePlanMini ミニプラン(後方互換用)
	DatabasePlanMini = DatabasePlan(10)
	// DatabasePlan10G 10Gプラン
	DatabasePlan10G = DatabasePlan(10)
	// DatabasePlan30G 30Gプラン
	DatabasePlan30G = DatabasePlan(30)
	// DatabasePlan90G 90Gプラン
	DatabasePlan90G = DatabasePlan(90)
	// DatabasePlan240G 240Gプラン
	DatabasePlan240G = DatabasePlan(240)
	// DatabasePlan500G 500Gプラン
	DatabasePlan500G = DatabasePlan(500)
	// DatabasePlan1T 1Tプラン
	DatabasePlan1T = DatabasePlan(1000)
)

// AllowDatabasePlans 指定可能なデータベースプラン
func AllowDatabasePlans() []int {
	return []int{
		int(DatabasePlan10G),
		int(DatabasePlan30G),
		int(DatabasePlan90G),
		int(DatabasePlan240G),
		int(DatabasePlan500G),
		int(DatabasePlan1T),
	}
}

// DatabaseBackupSetting バックアップ設定
type DatabaseBackupSetting struct {
	Rotate    int      `json:",omitempty"` // ローテーション世代数
	Time      string   `json:",omitempty"` // 開始時刻
	DayOfWeek []string `json:",omitempty"` // 取得曜日
}

// DatabaseCommonSetting 共通設定
type DatabaseCommonSetting struct {
	DefaultUser     string        `json:",omitempty"` // ユーザー名
	UserPassword    string        `json:",omitempty"` // ユーザーパスワード
	WebUI           interface{}   `json:",omitempty"` // WebUIのIPアドレス or FQDN
	ReplicaPassword string        `json:",omitempty"` // レプリケーションパスワード
	ReplicaUser     string        `json:",omitempty"` // レプリケーションユーザー
	ServicePort     json.Number   `json:",omitempty"` // ポート番号
	SourceNetwork   SourceNetwork // 接続許可ネットワーク
}

// SourceNetwork 接続許可ネットワーク
type SourceNetwork []string

// UnmarshalJSON JSONアンマーシャル(配列と文字列が混在するためここで対応)
func (s *SourceNetwork) UnmarshalJSON(data []byte) error {
	// SourceNetworkが未設定の場合、APIレスポンスが""となるため回避する
	if string(data) == `""` {
		return nil
	}

	tmp := []string{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	source := SourceNetwork(tmp)
	*s = source
	return nil
}

// MarshalJSON JSONマーシャル(配列と文字列が混在するためここで対応)
func (s *SourceNetwork) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte(""), nil
	}

	list := []string(*s)
	if len(list) == 0 || (len(list) == 1 && list[0] == "") {
		return []byte(`""`), nil
	}

	return json.Marshal(list)
}

// DatabaseReplicationSetting レプリケーション設定
type DatabaseReplicationSetting struct {
	// Model レプリケーションモデル
	Model DatabaseReplicationModels `json:",omitempty"`
	// Appliance マスター側アプライアンス
	Appliance *struct {
		ID string
	} `json:",omitempty"`
	// IPAddress IPアドレス
	IPAddress string `json:",omitempty"`
	// Port ポート
	Port int `json:",omitempty"`
	// User ユーザー
	User string `json:",omitempty"`
	// Password パスワード
	Password string `json:",omitempty"`
}

// DatabaseReplicationModels データベースのレプリケーションモデル
type DatabaseReplicationModels string

const (
	// DatabaseReplicationModelMasterSlave レプリケーションモデル: Master-Slave(マスター側)
	DatabaseReplicationModelMasterSlave = "Master-Slave"
	// DatabaseReplicationModelAsyncReplica レプリケーションモデル: Async-Replica(スレーブ側)
	DatabaseReplicationModelAsyncReplica = "Async-Replica"
)

// CreateDatabaseValue データベース作成用パラメータ
type CreateDatabaseValue struct {
	Plan             DatabasePlan // プラン
	AdminPassword    string       // 管理者パスワード
	DefaultUser      string       // ユーザー名
	UserPassword     string       // パスワード
	SourceNetwork    []string     // 接続許可ネットワーク
	ServicePort      int          // ポート
	EnableBackup     bool         // バックアップ有効化
	BackupRotate     int          // バックアップ世代数
	BackupTime       string       // バックアップ開始時間
	BackupDayOfWeek  []string     // バックアップ取得曜日
	SwitchID         string       // 接続先スイッチ
	IPAddress1       string       // IPアドレス1
	MaskLen          int          // ネットワークマスク長
	DefaultRoute     string       // デフォルトルート
	Name             string       // 名称
	Description      string       // 説明
	Tags             []string     // タグ
	Icon             *Resource    // アイコン
	WebUI            bool         // WebUI有効
	DatabaseName     string       // データベース名
	DatabaseRevision string       // リビジョン
	DatabaseTitle    string       // データベースタイトル
	DatabaseVersion  string       // データベースバージョン
	// ReplicaUser      string    // レプリケーションユーザー 現在はreplica固定
	ReplicaPassword string    // レプリケーションパスワード
	SourceAppliance *Resource // クローン元DB
}

// SlaveDatabaseValue スレーブデータベース作成用パラメータ
type SlaveDatabaseValue struct {
	Plan            DatabasePlan // プラン
	DefaultUser     string       // ユーザー名
	UserPassword    string       // パスワード
	SwitchID        string       // 接続先スイッチ
	IPAddress1      string       // IPアドレス1
	MaskLen         int          // ネットワークマスク長
	DefaultRoute    string       // デフォルトルート
	Name            string       // 名称
	Description     string       // 説明
	Tags            []string     // タグ
	Icon            *Resource    // アイコン
	DatabaseName    string       // データベース名
	DatabaseVersion string       // データベースバージョン
	// ReplicaUser      string    // レプリケーションユーザー 現在はreplica固定
	ReplicaPassword   string // レプリケーションパスワード
	MasterApplianceID int64  // クローン元DB
	MasterIPAddress   string // マスターIPアドレス
	MasterPort        int    // マスターポート
}

// NewCreatePostgreSQLDatabaseValue PostgreSQL作成用パラメーター
func NewCreatePostgreSQLDatabaseValue() *CreateDatabaseValue {
	return &CreateDatabaseValue{
		DatabaseName:    "postgres",
		DatabaseVersion: "10",
	}
}

// NewCreateMariaDBDatabaseValue MariaDB作成用パラメーター
func NewCreateMariaDBDatabaseValue() *CreateDatabaseValue {
	return &CreateDatabaseValue{
		DatabaseName:    "MariaDB",
		DatabaseVersion: "10.2",
	}
}

// NewCloneDatabaseValue クローンDB作成用パラメータ
func NewCloneDatabaseValue(db *Database) *CreateDatabaseValue {
	return &CreateDatabaseValue{
		DatabaseName:    db.Remark.DBConf.Common.DatabaseName,
		DatabaseVersion: db.Remark.DBConf.Common.DatabaseVersion,
		SourceAppliance: NewResource(db.ID),
	}
}

// CreateNewDatabase データベース作成
func CreateNewDatabase(values *CreateDatabaseValue) *Database {

	db := &Database{
		// Appliance
		Appliance: &Appliance{
			// Class
			Class: "database",
			// Name
			propName: propName{Name: values.Name},
			// Description
			propDescription: propDescription{Description: values.Description},
			// TagsType
			propTags: propTags{
				// Tags
				Tags: values.Tags,
			},
			// Icon
			propIcon: propIcon{
				&Icon{
					// Resource
					Resource: values.Icon,
				},
			},
			// Plan
			//propPlanID: propPlanID{Plan: &Resource{ID: int64(values.Plan)}},
		},
		// Remark
		Remark: &DatabaseRemark{
			// ApplianceRemarkBase
			ApplianceRemarkBase: &ApplianceRemarkBase{
				// Servers
				Servers: []interface{}{""},
			},
			// DBConf
			DBConf: &DatabaseCommonRemarks{
				// Common
				Common: &DatabaseCommonRemark{
					// DatabaseName
					DatabaseName: values.DatabaseName,
					// DatabaseRevision
					DatabaseRevision: values.DatabaseRevision,
					// DatabaseTitle
					DatabaseTitle: values.DatabaseTitle,
					// DatabaseVersion
					DatabaseVersion: values.DatabaseVersion,
				},
			},
			// Plan
			propPlanID:      propPlanID{Plan: &Resource{ID: int64(values.Plan)}},
			SourceAppliance: values.SourceAppliance,
		},
		// Settings
		Settings: &DatabaseSettings{
			// DBConf
			DBConf: &DatabaseSetting{
				// Backup
				Backup: &DatabaseBackupSetting{
					// Rotate
					// Rotate: values.BackupRotate,
					Rotate: 8,
					// Time
					Time: values.BackupTime,
					// DayOfWeek
					DayOfWeek: values.BackupDayOfWeek,
				},
				// Common
				Common: &DatabaseCommonSetting{
					// DefaultUser
					DefaultUser: values.DefaultUser,
					// UserPassword
					UserPassword: values.UserPassword,
					// SourceNetwork
					SourceNetwork: SourceNetwork(values.SourceNetwork),
				},
			},
		},
	}

	if values.ServicePort > 0 {
		db.Settings.DBConf.Common.ServicePort = json.Number(fmt.Sprintf("%d", values.ServicePort))
	}

	if !values.EnableBackup {
		db.Settings.DBConf.Backup = nil
	}

	db.Remark.Switch = &ApplianceRemarkSwitch{
		// ID
		ID: values.SwitchID,
	}
	db.Remark.Network = &DatabaseRemarkNetwork{
		// NetworkMaskLen
		NetworkMaskLen: values.MaskLen,
		// DefaultRoute
		DefaultRoute: values.DefaultRoute,
	}

	db.Remark.Servers = []interface{}{
		map[string]interface{}{"IPAddress": values.IPAddress1},
	}

	if values.WebUI {
		db.Settings.DBConf.Common.WebUI = values.WebUI
	}

	if values.ReplicaPassword != "" {
		db.Settings.DBConf.Common.ReplicaUser = "replica"
		db.Settings.DBConf.Common.ReplicaPassword = values.ReplicaPassword
		db.Settings.DBConf.Replication = &DatabaseReplicationSetting{
			Model: DatabaseReplicationModelMasterSlave,
		}
	}

	return db
}

// NewSlaveDatabaseValue スレーブ向けパラメータ作成
func NewSlaveDatabaseValue(values *SlaveDatabaseValue) *Database {
	db := &Database{
		// Appliance
		Appliance: &Appliance{
			// Class
			Class: "database",
			// Name
			propName: propName{Name: values.Name},
			// Description
			propDescription: propDescription{Description: values.Description},
			// TagsType
			propTags: propTags{
				// Tags
				Tags: values.Tags,
			},
			// Icon
			propIcon: propIcon{
				&Icon{
					// Resource
					Resource: values.Icon,
				},
			},
			// Plan
			//propPlanID: propPlanID{Plan: &Resource{ID: int64(values.Plan)}},
		},
		// Remark
		Remark: &DatabaseRemark{
			// ApplianceRemarkBase
			ApplianceRemarkBase: &ApplianceRemarkBase{
				// Servers
				Servers: []interface{}{""},
			},
			// DBConf
			DBConf: &DatabaseCommonRemarks{
				// Common
				Common: &DatabaseCommonRemark{
					// DatabaseName
					DatabaseName: values.DatabaseName,
					// DatabaseVersion
					DatabaseVersion: values.DatabaseVersion,
				},
			},
			// Plan
			propPlanID: propPlanID{Plan: &Resource{ID: int64(values.Plan)}},
		},
		// Settings
		Settings: &DatabaseSettings{
			// DBConf
			DBConf: &DatabaseSetting{
				// Common
				Common: &DatabaseCommonSetting{
					// DefaultUser
					DefaultUser: values.DefaultUser,
					// UserPassword
					UserPassword: values.UserPassword,
				},
				// Replication
				Replication: &DatabaseReplicationSetting{
					Model:     DatabaseReplicationModelAsyncReplica,
					Appliance: &struct{ ID string }{ID: fmt.Sprintf("%d", values.MasterApplianceID)},
					IPAddress: values.MasterIPAddress,
					Port:      values.MasterPort,
					User:      "replica",
					Password:  values.ReplicaPassword,
				},
			},
		},
	}

	db.Remark.Switch = &ApplianceRemarkSwitch{
		// ID
		ID: values.SwitchID,
	}
	db.Remark.Network = &DatabaseRemarkNetwork{
		// NetworkMaskLen
		NetworkMaskLen: values.MaskLen,
		// DefaultRoute
		DefaultRoute: values.DefaultRoute,
	}

	db.Remark.Servers = []interface{}{
		map[string]interface{}{"IPAddress": values.IPAddress1},
	}

	return db
}

// AddSourceNetwork 接続許可ネットワーク 追加
func (s *Database) AddSourceNetwork(nw string) {
	res := []string(s.Settings.DBConf.Common.SourceNetwork)
	res = append(res, nw)
	s.Settings.DBConf.Common.SourceNetwork = SourceNetwork(res)
}

// DeleteSourceNetwork 接続許可ネットワーク 削除
func (s *Database) DeleteSourceNetwork(nw string) {
	res := []string{}
	for _, s := range s.Settings.DBConf.Common.SourceNetwork {
		if s != nw {
			res = append(res, s)
		}
	}
	s.Settings.DBConf.Common.SourceNetwork = SourceNetwork(res)
}

// IsReplicationMaster レプリケーションが有効かつマスターとして構成されているか
func (s *Database) IsReplicationMaster() bool {
	return s.IsReplicationEnabled() && s.Settings.DBConf.Replication.Model == DatabaseReplicationModelMasterSlave
}

// IsReplicationEnabled レプリケーションが有効な場合はTrueを返す
func (s *Database) IsReplicationEnabled() bool {
	return s.Settings.DBConf.Replication != nil
}

// DatabaseName MariaDB or PostgreSQLの何れかを返す
func (s *Database) DatabaseName() string {
	return s.Remark.DBConf.Common.DatabaseName
}

// DatabaseRevision データベースのリビジョンを返す
//
// 例: MariaDBの場合 => 10.2.15 / PostgreSQLの場合 => 10.3
func (s *Database) DatabaseRevision() string {
	return s.Remark.DBConf.Common.DatabaseRevision
}

// DatabaseVersion データベースのバージョンを返す
//
// 例: MariaDBの場合 => 10.2 / PostgreSQLの場合 => 10
func (s *Database) DatabaseVersion() string {
	return s.Remark.DBConf.Common.DatabaseVersion
}

// WebUIAddress WebUIが有効な場合、IPアドレス or FQDNを返す、無効な場合は空文字を返す
func (s *Database) WebUIAddress() string {
	webUI := s.Settings.DBConf.Common.WebUI
	if webUI != nil {
		if v, ok := webUI.(string); ok {
			return v
		}
	}
	return ""
}

// IPAddress IPアドレスを取得
func (s *Database) IPAddress() string {
	if len(s.Remark.Servers) < 1 {
		return ""
	}
	v, ok := s.Remark.Servers[0].(map[string]string)
	if !ok {
		return ""
	}
	return v["IPAddress"]
}

// NetworkMaskLen ネットワークマスク長を取得
func (s *Database) NetworkMaskLen() int {
	if s.Remark.Network == nil {
		return -1
	}
	return s.Remark.Network.NetworkMaskLen
}

// DefaultRoute デフォルトゲートウェイアドレスを取得
func (s *Database) DefaultRoute() string {
	if s.Remark.Network == nil {
		return ""
	}
	return s.Remark.Network.DefaultRoute
}
