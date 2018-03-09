package types

import (
	"fmt"
	"strings"
)

// TraefikLog holds the configuration settings for the traefik logger.
type TraefikLog struct {
	FilePath string `json:"file,omitempty" description:"Traefik log file path. Stdout is used when omitted or empty"`
	Format   string `json:"format,omitempty" description:"Traefik log format: json | common"`
}

// AccessLog holds the configuration settings for the access logger (middlewares/accesslog).
type AccessLog struct {
	FilePath string   `json:"file,omitempty" description:"Access log file path. Stdout is used when omitted or empty" export:"true"`
	Format   string   `json:"format,omitempty" description:"Access log format: json | common" export:"true"`
	Filters  *Filters `json:"filters,omitempty" description:"Access log filters, used to keep only specific access log" export:"true"`
	Fields   *Fields  `json:"fields,omitempty" description:"Fields" export:"true"`
}

// StatusCodes holds status codes ranges to filter access log
type StatusCodes []string

// Filters holds filters configuration
type Filters struct {
	StatusCodes StatusCodes `json:"statusCodes,omitempty" description:"Keep only specific ranges of HTTP Status codes" export:"true"`
}

// FieldsNames holds maps of fields with specific mode
type FieldsNames map[string]string

// Fields holds configuration for access log fields
type Fields struct {
	DefaultMode string         `json:"defaultMode,omitempty" description:"Default mode for fields: keep | drop" export:"true"`
	Names       FieldsNames    `json:"names,omitempty" description:"Override mode for fields" export:"true"`
	Headers     *FieldsHeaders `json:"headers,omitempty" description:"Headers to keep, drop or redact" export:"true"`
}

// FieldsHeadersNames holds maps of fields with specific mode
type FieldsHeadersNames map[string]string

// FieldsHeaders holds configuration for access log headers
type FieldsHeaders struct {
	DefaultMode string             `json:"defaultMode,omitempty" description:"Default mode for fields: keep | drop | redact" export:"true"`
	Names       FieldsHeadersNames `json:"names,omitempty" description:"Override mode for headers" export:"true"`
}

// Set adds strings elem into the the parser
// it splits str on , and ;
func (s *StatusCodes) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	*s = append(*s, slice...)
	return nil
}

// Get StatusCodes
func (s *StatusCodes) Get() interface{} { return *s }

// String return slice in a string
func (s *StatusCodes) String() string { return fmt.Sprintf("%v", *s) }

// SetValue sets StatusCodes into the parser
func (s *StatusCodes) SetValue(val interface{}) {
	*s = val.(StatusCodes)
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (f *FieldsNames) String() string {
	return fmt.Sprintf("%+v", *f)
}

// Get return the FieldsNames map
func (f *FieldsNames) Get() interface{} {
	return *f
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (f *FieldsNames) Set(value string) error {
	fields := strings.Fields(value)

	//(*f) make(FieldsNames)
	for _, field := range fields {
		n := strings.SplitN(field, "=", 2)
		if len(n) == 2 {
			(*f)[n[0]] = n[1]
		}
	}

	return nil
}

// SetValue sets the FieldsNames map with val
func (f *FieldsNames) SetValue(val interface{}) {
	*f = val.(FieldsNames)
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (f *FieldsHeadersNames) String() string {
	return fmt.Sprintf("%+v", *f)
}

// Get return the FieldsHeadersNames map
func (f *FieldsHeadersNames) Get() interface{} {
	return *f
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (f *FieldsHeadersNames) Set(value string) error {
	fields := strings.Fields(value)

	//f = make(FieldsHeadersNames)
	for _, field := range fields {
		n := strings.SplitN(field, "=", 2)
		(*f)[n[0]] = n[1]
	}

	return nil
}

// SetValue sets the FieldsHeadersNames map with val
func (f *FieldsHeadersNames) SetValue(val interface{}) {
	*f = val.(FieldsHeadersNames)
}
