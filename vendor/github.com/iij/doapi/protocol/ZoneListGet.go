package protocol

import (
	"reflect"
)

// GET zones
type ZoneListGet struct {
	DoServiceCode string `json:"-"` // DO契約のサービスコード(do########)
}

// URI /{{.DoServiceCode}}/zones.json
func (t ZoneListGet) URI() string {
	return "/{{.DoServiceCode}}/zones.json"
}

// APIName ZoneListGet
func (t ZoneListGet) APIName() string {
	return "ZoneListGet"
}

// Method GET
func (t ZoneListGet) Method() string {
	return "GET"
}

// http://manual.iij.jp/dns/doapi/754466.html
func (t ZoneListGet) Document() string {
	return "http://manual.iij.jp/dns/doapi/754466.html"
}

// JPName GET zones
func (t ZoneListGet) JPName() string {
	return "GET zones"
}
func init() {
	APIlist = append(APIlist, ZoneListGet{})
	TypeMap["ZoneListGet"] = reflect.TypeOf(ZoneListGet{})
}

// ZoneListGetResponse GET zonesのレスポンス
type ZoneListGetResponse struct {
	*CommonResponse
	ZoneList []string
}
