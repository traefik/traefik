package sacloud

// propWaitDiskMigration ディスク作成待ちフラグ内包型
type propWaitDiskMigration struct {
	WaitDiskMigration bool `json:",omitempty"`
}

// GetWaitDiskMigration ディスク作成待ちフラグ 取得
func (p *propWaitDiskMigration) GetWaitDiskMigration() bool {
	return p.WaitDiskMigration
}

// SetWaitDiskMigration ディスク作成待ちフラグ 設定
func (p *propWaitDiskMigration) SetWaitDiskMigration(f bool) {
	p.WaitDiskMigration = f
}
