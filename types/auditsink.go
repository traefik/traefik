package types

import (
	"fmt"
	"strings"
)

// Exclusion excludes a request from auditing if the http header contains any of the specified values
type Exclusion struct {
	HeaderName string   `json:"headerName,omitempty" description:"Request header name to evaluate"`
	Contains   []string `json:"contains,omitempty" description:"Substring values to exclude"`
	EndsWith   []string `json:"endsWith,omitempty" description:"End of string values to exclude"`
	StartsWith []string `json:"startsWith,omitempty" description:"Start of string values to exclude"`
	Matches    []string `json:"matches,omitempty" description:"Regex matches to exclude"`
}

// Enabled states whether any exclusion filters are specified
func (e *Exclusion) Enabled() bool {
	return len(e.Contains) > 0 || len(e.EndsWith) > 0 || len(e.StartsWith) > 0
}

// Exclusions is a container type for Exclusion
type Exclusions map[string]*Exclusion

type MaskFields []string

// HeaderMapping maps a audit event field to a header value
//type HeaderMapping map[string]string

// HeaderMappings allows dynamic definition of audit fields whose values are sourced
// from request/response headers. The key defines the section of the audit structure
// where the fields will be defined.
//type HeaderMappings map[string]*HeaderMapping

// AuditSink holds AuditSink configuration
type AuditSink struct {
	Exclusions               Exclusions `json:"exclusions,omitempty"`
	Type                     string     `json:"type,omitempty" description:"The type of sink: File/HTTP/Kafka/AMQP/Blackhole"`
	ClientID                 string     `json:"clientId,omitempty" description:"Identifier to be used for the sink client"`
	ClientVersion            string     `json:"clientVersion,omitempty" description:"Version info to identify the sink client"`
	Endpoint                 string     `json:"endpoint,omitempty" description:"Endpoint for audit tap. e.g. url for HTTP/Kafka/AMQP or filename for File"`
	Destination              string     `json:"destination,omitempty" description:"For Kafka the topic, AMQP the exchange etc."`
	MaxEntityLength          string     `json:"maxEntityLength,omitempty" description:"MaxEntityLength truncates entities (bodies) longer than this (units are allowed, eg. 32KiB)"`
	NumProducers             int        `json:"numProducers,omitempty" description:"The number of concurrent producers which can send messages to the endpoint"`
	ChannelLength            int        `json:"channelLength,omitempty" description:"Size of the in-memory message channel.  Used as a buffer in case of Producer failure"`
	DiskStorePath            string     `json:"diskStorePath,omitempty" description:"Directory path for disk-backed persistent audit message queue"`
	ProxyingFor              string     `json:"proxyingFor,omitempty" description:"Defines the style of auditing event required. e.g API, RATE"`
	AuditSource              string     `json:"auditSource,omitempty" description:"Value to use for auditSource in audit message"`
	AuditType                string     `json:"auditType,omitempty" description:"Value to use for auditType in audit message"`
	EncryptSecret            string     `json:"encryptSecret,omitempty" description:"Key for encrypting failed events. If present events will be AES encrypted"`
	MaxAuditLength           string     `json:"maxAuditLength,omitempty" description:"The allowed maximum size of an audit event (units are allowed, eg. 32K)"`
	MaxPayloadContentsLength string     `json:"maxPayloadContentsLength,omitempty" description:"The allowed maximum combined size of audit requestPayload.contents and responsePayload.contents (units are allowed, eg. 32K)"`
	MaskValue                string     `json:"maskValue,omitempty" description:"The value to be used when obfuscating fields. Default is #########"`
	MaskFields               MaskFields `json:"maskFields,omitempty" description:"Names of payload fields whose values should be obfuscated"`
	//HeaderMappings           HeaderMappings `json:"headerMappings,omitempty" description:"Configuration of dynamic audit fields whose value is sourced form a header"`
}

// Set adds strings elem into the the parser
// it splits str on , and ;
func (s *MaskFields) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	*s = append(*s, slice...)
	return nil
}

// Get MaskFields
func (s *MaskFields) Get() interface{} { return *s }

// String return slice in a string
func (s *MaskFields) String() string { return fmt.Sprintf("%v", *s) }

// SetValue sets MaskFields into the parser
func (s *MaskFields) SetValue(val interface{}) {
	*s = val.(MaskFields)
}

/*
// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (f *HeaderMapping) String() string {
	return fmt.Sprintf("%+v", *f)
}

// Get return the HeaderMapping map
func (f *HeaderMapping) Get() interface{} {
	return *f
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a space-separated list, so we split it.
func (f *HeaderMapping) Set(value string) error {
	fields := strings.Fields(value)

	for _, field := range fields {
		n := strings.SplitN(field, "=", 2)
		(*f)[n[0]] = n[1]
	}

	return nil
}

// SetValue sets the HeaderMapping map with val
func (f *HeaderMapping) SetValue(val interface{}) {
	*f = val.(HeaderMapping)
}
*/
