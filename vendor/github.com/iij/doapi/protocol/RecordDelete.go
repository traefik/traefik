package protocol

import (
	"reflect"
)

// RecordDelete DELETE record
//  http://manual.iij.jp/dns/doapi/754525.html
type RecordDelete struct {
	DoServiceCode string `json:"-"` // DO契約のサービスコード(do########)
	ZoneName      string `json:"-"` // Zone Name
	RecordID      string `json:"-"` // Record ID
}

// URI /{{.DoServiceCode}}/{{.ZoneName}}/record/{{.RecordID}}.json
func (t RecordDelete) URI() string {
	return "/{{.DoServiceCode}}/{{.ZoneName}}/record/{{.RecordID}}.json"
}

// APIName RecordDelete
func (t RecordDelete) APIName() string {
	return "RecordDelete"
}

// Method DELETE
func (t RecordDelete) Method() string {
	return "DELETE"
}

// http://manual.iij.jp/dns/doapi/754525.html
func (t RecordDelete) Document() string {
	return "http://manual.iij.jp/dns/doapi/754525.html"
}

// JPName DELETE record
func (t RecordDelete) JPName() string {
	return "DELETE record"
}
func init() {
	APIlist = append(APIlist, RecordDelete{})
	TypeMap["RecordDelete"] = reflect.TypeOf(RecordDelete{})
}

// RecordDeleteResponse DELETE recordのレスポンス
type RecordDeleteResponse struct {
	*CommonResponse
}
