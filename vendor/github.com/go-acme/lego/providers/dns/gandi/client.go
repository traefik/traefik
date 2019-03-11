package gandi

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
)

// types for XML-RPC method calls and parameters

type param interface {
	param()
}

type paramString struct {
	XMLName xml.Name `xml:"param"`
	Value   string   `xml:"value>string"`
}

type paramInt struct {
	XMLName xml.Name `xml:"param"`
	Value   int      `xml:"value>int"`
}

type structMember interface {
	structMember()
}
type structMemberString struct {
	Name  string `xml:"name"`
	Value string `xml:"value>string"`
}
type structMemberInt struct {
	Name  string `xml:"name"`
	Value int    `xml:"value>int"`
}
type paramStruct struct {
	XMLName       xml.Name       `xml:"param"`
	StructMembers []structMember `xml:"value>struct>member"`
}

func (p paramString) param()               {}
func (p paramInt) param()                  {}
func (m structMemberString) structMember() {}
func (m structMemberInt) structMember()    {}
func (p paramStruct) param()               {}

type methodCall struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
	Params     []param  `xml:"params"`
}

// types for XML-RPC responses

type response interface {
	faultCode() int
	faultString() string
}

type responseFault struct {
	FaultCode   int    `xml:"fault>value>struct>member>value>int"`
	FaultString string `xml:"fault>value>struct>member>value>string"`
}

func (r responseFault) faultCode() int      { return r.FaultCode }
func (r responseFault) faultString() string { return r.FaultString }

type responseStruct struct {
	responseFault
	StructMembers []struct {
		Name     string `xml:"name"`
		ValueInt int    `xml:"value>int"`
	} `xml:"params>param>value>struct>member"`
}

type responseInt struct {
	responseFault
	Value int `xml:"params>param>value>int"`
}

type responseBool struct {
	responseFault
	Value bool `xml:"params>param>value>boolean"`
}

type rpcError struct {
	faultCode   int
	faultString string
}

func (e rpcError) Error() string {
	return fmt.Sprintf("Gandi DNS: RPC Error: (%d) %s", e.faultCode, e.faultString)
}

// rpcCall makes an XML-RPC call to Gandi's RPC endpoint by
// marshaling the data given in the call argument to XML and sending
// that via HTTP Post to Gandi.
// The response is then unmarshalled into the resp argument.
func (d *DNSProvider) rpcCall(call *methodCall, resp response) error {
	// marshal
	b, err := xml.MarshalIndent(call, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}

	// post
	b = append([]byte(`<?xml version="1.0"?>`+"\n"), b...)
	respBody, err := d.httpPost(d.config.BaseURL, "text/xml", bytes.NewReader(b))
	if err != nil {
		return err
	}

	// unmarshal
	err = xml.Unmarshal(respBody, resp)
	if err != nil {
		return fmt.Errorf("unmarshal error: %v", err)
	}
	if resp.faultCode() != 0 {
		return rpcError{
			faultCode: resp.faultCode(), faultString: resp.faultString()}
	}
	return nil
}

// functions to perform API actions

func (d *DNSProvider) getZoneID(domain string) (int, error) {
	resp := &responseStruct{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.info",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramString{Value: domain},
		},
	}, resp)
	if err != nil {
		return 0, err
	}

	var zoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "zone_id" {
			zoneID = member.ValueInt
		}
	}

	if zoneID == 0 {
		return 0, fmt.Errorf("could not determine zone_id for %s", domain)
	}
	return zoneID, nil
}

func (d *DNSProvider) cloneZone(zoneID int, name string) (int, error) {
	resp := &responseStruct{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.clone",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
			paramInt{Value: 0},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "name",
						Value: name,
					}},
			},
		},
	}, resp)
	if err != nil {
		return 0, err
	}

	var newZoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "id" {
			newZoneID = member.ValueInt
		}
	}

	if newZoneID == 0 {
		return 0, fmt.Errorf("could not determine cloned zone_id")
	}
	return newZoneID, nil
}

func (d *DNSProvider) newZoneVersion(zoneID int) (int, error) {
	resp := &responseInt{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.version.new",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return 0, err
	}

	if resp.Value == 0 {
		return 0, fmt.Errorf("could not create new zone version")
	}
	return resp.Value, nil
}

func (d *DNSProvider) addTXTRecord(zoneID int, version int, name string, value string, ttl int) error {
	resp := &responseStruct{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.record.add",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "type",
						Value: "TXT",
					}, structMemberString{
						Name:  "name",
						Value: name,
					}, structMemberString{
						Name:  "value",
						Value: value,
					}, structMemberInt{
						Name:  "ttl",
						Value: ttl,
					}},
			},
		},
	}, resp)
	return err
}

func (d *DNSProvider) setZoneVersion(zoneID int, version int) error {
	resp := &responseBool{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.version.set",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
		},
	}, resp)
	if err != nil {
		return err
	}

	if !resp.Value {
		return fmt.Errorf("could not set zone version")
	}
	return nil
}

func (d *DNSProvider) setZone(domain string, zoneID int) error {
	resp := &responseStruct{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.set",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramString{Value: domain},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return err
	}

	var respZoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "zone_id" {
			respZoneID = member.ValueInt
		}
	}

	if respZoneID != zoneID {
		return fmt.Errorf("could not set new zone_id for %s", domain)
	}
	return nil
}

func (d *DNSProvider) deleteZone(zoneID int) error {
	resp := &responseBool{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.delete",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return err
	}

	if !resp.Value {
		return fmt.Errorf("could not delete zone_id")
	}
	return nil
}

func (d *DNSProvider) httpPost(url string, bodyType string, body io.Reader) ([]byte, error) {
	resp, err := d.config.HTTPClient.Post(url, bodyType, body)
	if err != nil {
		return nil, fmt.Errorf("HTTP Post Error: %v", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP Post Error: %v", err)
	}

	return b, nil
}
