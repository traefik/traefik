package protocol

//go:generate python doc2struct.py

import (
	"fmt"
	"reflect"
)

type CommonArg interface {
	APIName() string
	Method() string
	URI() string
	Document() string
	JPName() string
}

type CommonResponse struct {
	RequestId     string
	ErrorResponse struct {
		RequestId    string
		ErrorType    string
		ErrorMessage string
	}
}

var APIlist []CommonArg

var TypeMap = map[string]reflect.Type{}

type ResourceRecord struct {
	Id         string `json:",omitempty"`
	Status     string
	Owner      string
	TTL        string
	RecordType string
	RData      string
}

func (r *ResourceRecord) String() string {
	return fmt.Sprintf("%s %s IN %s %s", r.Owner, r.TTL, r.RecordType, r.RData)
}

func (r *ResourceRecord) FQDN(zone string) string {
	return fmt.Sprintf("%s.%s %s IN %s %s", r.Owner, zone, r.TTL, r.RecordType, r.RData)
}

const (
	UNCAHNGED = "UNCHANGED"
	ADDING    = "ADDING"
	DELETING  = "DELETING"
	DELETED   = "DELETED"
)
