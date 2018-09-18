package protocol

import (
	"reflect"
)

// RecordAdd POST record (同期)
//  http://manual.iij.jp/dns/doapi/754517.html
type RecordAdd struct {
	DoServiceCode string `json:"-"` // DO契約のサービスコード(do########)
	ZoneName      string `json:"-"` // Zone Name
	Owner         string // owner of record
	TTL           string // TTL of record
	RecordType    string // type of record
	RData         string // data of record
}

// URI /:GisServiceCode/fw-lbs/:IflServiceCode/filters/:IpVersion/:Direction.json
func (t RecordAdd) URI() string {
	return "/{{.DoServiceCode}}/{{.ZoneName}}/record.json"
}

// APIName RecordAdd
func (t RecordAdd) APIName() string {
	return "RecordAdd"
}

// Method POST
func (t RecordAdd) Method() string {
	return "POST"
}

// http://manual.iij.jp/dns/doapi/754517.html
func (t RecordAdd) Document() string {
	return "http://manual.iij.jp/dns/doapi/754517.html"
}

// JPName POST record
func (t RecordAdd) JPName() string {
	return "POST record"
}
func init() {
	APIlist = append(APIlist, RecordAdd{})
	TypeMap["RecordAdd"] = reflect.TypeOf(RecordAdd{})
}

// RecordAddResponse POST recordのレスポンス
type RecordAddResponse struct {
	*CommonResponse
	Record ResourceRecord
}
