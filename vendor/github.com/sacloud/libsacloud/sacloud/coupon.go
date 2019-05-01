package sacloud

import "time"

// Coupon クーポン情報
type Coupon struct {
	CouponID       string    `json:",omitempty"` // クーポンID
	MemberID       string    `json:",omitempty"` // メンバーID
	ContractID     int64     `json:",omitempty"` // 契約ID
	ServiceClassID int64     `json:",omitempty"` // サービスクラスID
	Discount       int64     `json:",omitempty"` // クーポン残高
	AppliedAt      time.Time `json:",omitempty"` // 適用開始日
	UntilAt        time.Time `json:",omitempty"` // 有効期限
}
