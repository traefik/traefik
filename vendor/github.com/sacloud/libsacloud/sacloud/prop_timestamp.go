package sacloud

import "time"

// propCreatedAt 作成日時内包型
type propCreatedAt struct {
	CreatedAt *time.Time `json:",omitempty"` // 作成日時
}

// GetCreatedAt 作成日時 取得
func (p *propCreatedAt) GetCreatedAt() *time.Time {
	return p.CreatedAt
}

// propModifiedAt 変更日時内包型
type propModifiedAt struct {
	// ModifiedAt 変更日時
	ModifiedAt *time.Time `json:",omitempty"`
}

// GetModifiedAt 変更日時 取得
func (p *propModifiedAt) GetModifiedAt() *time.Time {
	return p.ModifiedAt
}

// propUpdatedAt 変更日時内包型
type propUpdatedAt struct {
	// UpdatedAt 変更日時
	UpdatedAt *time.Time `json:",omitempty"`
}

// GetModifiedAt 変更日時 取得
func (p *propUpdatedAt) GetModifiedAt() *time.Time {
	return p.UpdatedAt
}
