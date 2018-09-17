package gandi

import (
	"encoding/xml"
	"fmt"
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

// POSTing/Marshalling/Unmarshalling

type rpcError struct {
	faultCode   int
	faultString string
}

func (e rpcError) Error() string {
	return fmt.Sprintf("Gandi DNS: RPC Error: (%d) %s", e.faultCode, e.faultString)
}
