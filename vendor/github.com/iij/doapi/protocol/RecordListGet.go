package protocol

import (
	"reflect"
)

// GET records
type RecordListGet struct {
	DoServiceCode string `json:"-"` // DO契約のサービスコード(do########)
	ZoneName      string `json:"-"` // ゾーン名
}

// URI /{{.DoServiceCode}}/{{.ZoneName}}/records/DETAIL.json
func (t RecordListGet) URI() string {
	return "/{{.DoServiceCode}}/{{.ZoneName}}/records/DETAIL.json"
}

// APIName RecordListGet
func (t RecordListGet) APIName() string {
	return "RecordListGet"
}

// Method GET
func (t RecordListGet) Method() string {
	return "GET"
}

// http://manual.iij.jp/dns/doapi/754619.html
func (t RecordListGet) Document() string {
	return "http://manual.iij.jp/dns/doapi/754619.html"
}

// JPName GET records
func (t RecordListGet) JPName() string {
	return "GET records"
}
func init() {
	APIlist = append(APIlist, RecordListGet{})
	TypeMap["RecordListGet"] = reflect.TypeOf(RecordListGet{})
}

// RecordListGetResponse GET recordsのレスポンス
type RecordListGetResponse struct {
	*CommonResponse
	RecordList       []ResourceRecord
	StaticRecordList []ResourceRecord
}
