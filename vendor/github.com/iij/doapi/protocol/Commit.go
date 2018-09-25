package protocol

import (
	"reflect"
)

// Commit
type Commit struct {
	DoServiceCode string `json:"-"` // DO契約のサービスコード(do########)
}

// URI /{{.DoServiceCode}}/commit.json
func (t Commit) URI() string {
	return "/{{.DoServiceCode}}/commit.json"
}

// APIName Commit
func (t Commit) APIName() string {
	return "Commit"
}

// Method PUT
func (t Commit) Method() string {
	return "PUT"
}

// http://manual.iij.jp/dns/doapi/754632.html
func (t Commit) Document() string {
	return "http://manual.iij.jp/dns/doapi/754632.html"
}

// JPName PUT Commit
func (t Commit) JPName() string {
	return "PUT commit"
}
func init() {
	APIlist = append(APIlist, Commit{})
	TypeMap["Commit"] = reflect.TypeOf(Commit{})
}

// CommitResponse PUT Commitのレスポンス
type CommitResponse struct {
	*CommonResponse
}
