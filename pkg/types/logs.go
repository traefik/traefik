package types

import "github.com/traefik/paerser/types"

const (
	// AccessLogKeep is the keep string value.
	AccessLogKeep = "keep"
	// AccessLogDrop is the drop string value.
	AccessLogDrop = "drop"
	// AccessLogRedact is the redact string value.
	AccessLogRedact = "redact"
)

const (
	// CommonFormat is the common logging format (CLF).
	CommonFormat string = "common"

	// JSONFormat is the JSON logging format.
	JSONFormat string = "json"
)

// TraefikLog holds the configuration settings for the traefik logger.
type TraefikLog struct {
	Level    string `description:"Log level set to traefik logs." json:"level,omitempty" toml:"level,omitempty" yaml:"level,omitempty" export:"true"`
	FilePath string `description:"Traefik log file path. Stdout is used when omitted or empty." json:"filePath,omitempty" toml:"filePath,omitempty" yaml:"filePath,omitempty"`
	Format   string `description:"Traefik log format: json | common" json:"format,omitempty" toml:"format,omitempty" yaml:"format,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (l *TraefikLog) SetDefaults() {
	l.Format = CommonFormat
	l.Level = "ERROR"
}

// AccessLog holds the configuration settings for the access logger (middlewares/accesslog).
type AccessLog struct {
	FilePath      string            `description:"Access log file path. Stdout is used when omitted or empty." json:"filePath,omitempty" toml:"filePath,omitempty" yaml:"filePath,omitempty"`
	Format        string            `description:"Access log format: json | common" json:"format,omitempty" toml:"format,omitempty" yaml:"format,omitempty" export:"true"`
	Filters       *AccessLogFilters `description:"Access log filters, used to keep only specific access logs." json:"filters,omitempty" toml:"filters,omitempty" yaml:"filters,omitempty" export:"true"`
	Fields        *AccessLogFields  `description:"AccessLogFields." json:"fields,omitempty" toml:"fields,omitempty" yaml:"fields,omitempty" export:"true"`
	BufferingSize int64             `description:"Number of access log lines to process in a buffered way." json:"bufferingSize,omitempty" toml:"bufferingSize,omitempty" yaml:"bufferingSize,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (l *AccessLog) SetDefaults() {
	l.Format = CommonFormat
	l.FilePath = ""
	l.Filters = &AccessLogFilters{}
	l.Fields = &AccessLogFields{}
	l.Fields.SetDefaults()
}

// AccessLogFilters holds filters configuration.
type AccessLogFilters struct {
	StatusCodes   []string       `description:"Keep access logs with status codes in the specified range." json:"statusCodes,omitempty" toml:"statusCodes,omitempty" yaml:"statusCodes,omitempty" export:"true"`
	RetryAttempts bool           `description:"Keep access logs when at least one retry happened." json:"retryAttempts,omitempty" toml:"retryAttempts,omitempty" yaml:"retryAttempts,omitempty" export:"true"`
	MinDuration   types.Duration `description:"Keep access logs when request took longer than the specified duration." json:"minDuration,omitempty" toml:"minDuration,omitempty" yaml:"minDuration,omitempty" export:"true"`
}

// FieldHeaders holds configuration for access log headers.
type FieldHeaders struct {
	DefaultMode string            `description:"Default mode for fields: keep | drop | redact" json:"defaultMode,omitempty" toml:"defaultMode,omitempty" yaml:"defaultMode,omitempty" export:"true"`
	Names       map[string]string `description:"Override mode for headers" json:"names,omitempty" toml:"names,omitempty" yaml:"names,omitempty" export:"true"`
}

// AccessLogFields holds configuration for access log fields.
type AccessLogFields struct {
	DefaultMode string            `description:"Default mode for fields: keep | drop" json:"defaultMode,omitempty" toml:"defaultMode,omitempty" yaml:"defaultMode,omitempty"  export:"true"`
	Names       map[string]string `description:"Override mode for fields" json:"names,omitempty" toml:"names,omitempty" yaml:"names,omitempty" export:"true"`
	Headers     *FieldHeaders     `description:"Headers to keep, drop or redact" json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (f *AccessLogFields) SetDefaults() {
	f.DefaultMode = AccessLogKeep
	f.Headers = &FieldHeaders{
		DefaultMode: AccessLogDrop,
	}
}

// Keep check if the field need to be kept or dropped.
func (f *AccessLogFields) Keep(field string) bool {
	defaultKeep := true
	if f != nil {
		defaultKeep = checkFieldValue(f.DefaultMode, defaultKeep)

		if v, ok := f.Names[field]; ok {
			return checkFieldValue(v, defaultKeep)
		}
	}
	return defaultKeep
}

// KeepHeader checks if the headers need to be kept, dropped or redacted and returns the status.
func (f *AccessLogFields) KeepHeader(header string) string {
	defaultValue := AccessLogKeep
	if f != nil && f.Headers != nil {
		defaultValue = checkFieldHeaderValue(f.Headers.DefaultMode, defaultValue)

		if v, ok := f.Headers.Names[header]; ok {
			return checkFieldHeaderValue(v, defaultValue)
		}
	}
	return defaultValue
}

func checkFieldValue(value string, defaultKeep bool) bool {
	switch value {
	case AccessLogKeep:
		return true
	case AccessLogDrop:
		return false
	default:
		return defaultKeep
	}
}

func checkFieldHeaderValue(value, defaultValue string) string {
	if value == AccessLogKeep || value == AccessLogDrop || value == AccessLogRedact {
		return value
	}
	return defaultValue
}
