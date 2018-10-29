package protocol

import (
	"reflect"
)

// Reset PUT reset (同期)
type Reset struct {
	DoServiceCode string `json:"-"` // DO契約のサービスコード(do########)
	ZoneName      string `json:"-"` // Zone name
}

// URI /{{.DoServiceCode}}/{{.ZoneName}}/reset.json
func (t Reset) URI() string {
	return "/{{.DoServiceCode}}/{{.ZoneName}}/reset.json"
}

// APIName Reset
func (t Reset) APIName() string {
	return "Reset"
}

// Method PUT
func (t Reset) Method() string {
	return "PUT"
}

// http://manual.iij.jp/dns/doapi/754610.html
func (t Reset) Document() string {
	return "http://manual.iij.jp/dns/doapi/754610.html"
}

// JPName PUT reset
func (t Reset) JPName() string {
	return "PUT Reset"
}
func init() {
	APIlist = append(APIlist, Reset{})
	TypeMap["Reset"] = reflect.TypeOf(Reset{})
}

// ResetResponse PUT resetのレスポンス
type ResetResponse struct {
	*CommonResponse
}
