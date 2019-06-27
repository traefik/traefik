package types

const (
	// AccessLogKeep is the keep string value
	AccessLogKeep = "keep"
	// AccessLogDrop is the drop string value
	AccessLogDrop = "drop"
	// AccessLogRedact is the redact string value
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
	Level    string `description:"Log level set to traefik logs." export:"true"`
	FilePath string `json:"file,omitempty" description:"Traefik log file path. Stdout is used when omitted or empty."`
	Format   string `json:"format,omitempty" description:"Traefik log format: json | common"`
}

// SetDefaults sets the default values.
func (l *TraefikLog) SetDefaults() {
	l.Format = CommonFormat
	l.Level = "ERROR"
}

// AccessLog holds the configuration settings for the access logger (middlewares/accesslog).
type AccessLog struct {
	FilePath      string            `json:"file,omitempty" description:"Access log file path. Stdout is used when omitted or empty." export:"true"`
	Format        string            `json:"format,omitempty" description:"Access log format: json | common" export:"true"`
	Filters       *AccessLogFilters `json:"filters,omitempty" description:"Access log filters, used to keep only specific access logs." export:"true"`
	Fields        *AccessLogFields  `json:"fields,omitempty" description:"AccessLogFields." export:"true"`
	BufferingSize int64             `json:"bufferingSize,omitempty" description:"Number of access log lines to process in a buffered way." export:"true"`
}

// SetDefaults sets the default values.
func (l *AccessLog) SetDefaults() {
	l.Format = CommonFormat
	l.FilePath = ""
	l.Filters = &AccessLogFilters{}
	l.Fields = &AccessLogFields{}
	l.Fields.SetDefaults()
}

// AccessLogFilters holds filters configuration
type AccessLogFilters struct {
	StatusCodes   []string `json:"statusCodes,omitempty" description:"Keep access logs with status codes in the specified range." export:"true"`
	RetryAttempts bool     `json:"retryAttempts,omitempty" description:"Keep access logs when at least one retry happened." export:"true"`
	MinDuration   Duration `json:"duration,omitempty" description:"Keep access logs when request took longer than the specified duration." export:"true"`
}

// FieldHeaders holds configuration for access log headers
type FieldHeaders struct {
	DefaultMode string            `json:"defaultMode,omitempty" description:"Default mode for fields: keep | drop | redact" export:"true"`
	Names       map[string]string `json:"names,omitempty" description:"Override mode for headers" export:"true"`
}

// AccessLogFields holds configuration for access log fields
type AccessLogFields struct {
	DefaultMode string            `json:"defaultMode,omitempty" description:"Default mode for fields: keep | drop" export:"true"`
	Names       map[string]string `json:"names,omitempty" description:"Override mode for fields" export:"true"`
	Headers     *FieldHeaders     `json:"headers,omitempty" description:"Headers to keep, drop or redact" export:"true"`
}

// SetDefaults sets the default values.
func (f *AccessLogFields) SetDefaults() {
	f.DefaultMode = AccessLogKeep
	f.Headers = &FieldHeaders{
		DefaultMode: AccessLogDrop,
	}
}

// Keep check if the field need to be kept or dropped
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

// KeepHeader checks if the headers need to be kept, dropped or redacted and returns the status
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

func checkFieldHeaderValue(value string, defaultValue string) string {
	if value == AccessLogKeep || value == AccessLogDrop || value == AccessLogRedact {
		return value
	}
	return defaultValue
}
