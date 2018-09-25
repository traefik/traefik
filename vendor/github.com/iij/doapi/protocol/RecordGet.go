package protocol

import (
	"reflect"
)

// GET records
//  http://manual.iij.jp/dns/doapi/754619.html
type RecordGet struct {
	DoServiceCode string `json:"-"` // DO契約のサービスコード(do########)
	ZoneName      string `json:"-"` // ゾーン名
	RecordID      string `json:"-"` //
}

// URI /{{.DoServiceCode}}/{{.ZoneName}}/record/{{.RecordID}}.json
func (t RecordGet) URI() string {
	return "/{{.DoServiceCode}}/{{.ZoneName}}/record/{{.RecordID}}.json"
}

// APIName RecordGet
func (t RecordGet) APIName() string {
	return "RecordGet"
}

// Method GET
func (t RecordGet) Method() string {
	return "GET"
}

// http://manual.iij.jp/dns/doapi/754503.html
func (t RecordGet) Document() string {
	return "http://manual.iij.jp/dns/doapi/754503.html"
}

// JPName GET record
func (t RecordGet) JPName() string {
	return "GET record"
}
func init() {
	APIlist = append(APIlist, RecordGet{})
	TypeMap["RecordGet"] = reflect.TypeOf(RecordGet{})
}

// RecordGetResponse フィルタリングルール情報取得のレスポンス
type RecordGetResponse struct {
	*CommonResponse
	Record ResourceRecord
}
