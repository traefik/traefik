package accesslog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
)

// default format for time presentation.
const (
	commonLogTimeFormat = "02/Jan/2006:15:04:05 -0700"
	defaultValue        = "-"
)

// CommonLogFormatter provides formatting in the Traefik common log format.
type CommonLogFormatter struct{}

// Format formats the log entry in the Traefik common log format.
func (f *CommonLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	timestamp := defaultValue
	if v, ok := entry.Data[StartUTC]; ok {
		timestamp = v.(time.Time).Format(commonLogTimeFormat)
	} else if v, ok := entry.Data[StartLocal]; ok {
		timestamp = v.(time.Time).Local().Format(commonLogTimeFormat)
	}

	var elapsedMillis int64
	if v, ok := entry.Data[Duration]; ok {
		elapsedMillis = v.(time.Duration).Nanoseconds() / 1000000
	}

	_, err := fmt.Fprintf(b, "%s - %s [%s] \"%s %s %s\" %v %v %s %s %v %s %s %dms\n",
		toLog(entry.Data, ClientHost, defaultValue, false),
		toLog(entry.Data, ClientUsername, defaultValue, false),
		timestamp,
		toLog(entry.Data, RequestMethod, defaultValue, false),
		toLog(entry.Data, RequestPath, defaultValue, false),
		toLog(entry.Data, RequestProtocol, defaultValue, false),
		toLog(entry.Data, DownstreamStatus, defaultValue, true),
		toLog(entry.Data, DownstreamContentSize, defaultValue, true),
		toLog(entry.Data, "request_Referer", `"-"`, true),
		toLog(entry.Data, "request_User-Agent", `"-"`, true),
		toLog(entry.Data, RequestCount, defaultValue, true),
		toLog(entry.Data, RouterName, `"-"`, true),
		toLog(entry.Data, ServiceURL, `"-"`, true),
		elapsedMillis)

	return b.Bytes(), err
}

// GenericCLFLogFormatter provides formatting in the generic CLF log format.
type GenericCLFLogFormatter struct{}

// Format formats the log entry in the generic CLF log format.
func (f *GenericCLFLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	timestamp := defaultValue
	if v, ok := entry.Data[StartUTC]; ok {
		timestamp = v.(time.Time).Format(commonLogTimeFormat)
	} else if v, ok := entry.Data[StartLocal]; ok {
		timestamp = v.(time.Time).Local().Format(commonLogTimeFormat)
	}

	_, err := fmt.Fprintf(b, "%s - %s [%s] \"%s %s %s\" %v %v %s %s\n",
		toLog(entry.Data, ClientHost, defaultValue, false),
		toLog(entry.Data, ClientUsername, defaultValue, false),
		timestamp,
		toLog(entry.Data, RequestMethod, defaultValue, false),
		toLog(entry.Data, RequestPath, defaultValue, false),
		toLog(entry.Data, RequestProtocol, defaultValue, false),
		toLog(entry.Data, DownstreamStatus, defaultValue, true),
		toLog(entry.Data, DownstreamContentSize, defaultValue, true),
		toLog(entry.Data, "request_Referer", `"-"`, true),
		toLog(entry.Data, "request_User-Agent", `"-"`, true))

	return b.Bytes(), err
}

func toLog(fields logrus.Fields, key, defaultValue string, quoted bool) any {
	if v, ok := fields[key]; ok {
		if v == nil {
			return defaultValue
		}

		switch s := v.(type) {
		case string:
			return toLogEntry(s, defaultValue, quoted)

		case fmt.Stringer:
			return toLogEntry(s.String(), defaultValue, quoted)

		default:
			return v
		}
	}

	return defaultValue
}

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// jsonString auto-escapes its content when rendered by text/template, so
// "{{ index . "RequestPath" }}" is safe even when the value contains quotes or backslashes.
// MarshalJSON delegates to the raw string so {{ json (index . "field") }} also works.
type jsonString struct{ v string }

func (s jsonString) String() string {
	b, err := json.Marshal(s.v)
	if err != nil {
		return ""
	}
	return string(b[1 : len(b)-1]) // strip surrounding quotes; content is embedded in user-provided quotes
}

func (s jsonString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

// TemplateJSONFormatter formats access log entries using a user-supplied Go text/template.
// String fields auto-escape; use {{ json (index . "field") }} for a self-quoted JSON value.
// The template must produce valid JSON; invalid output is rejected at runtime.
type TemplateJSONFormatter struct {
	tmpl *template.Template
}

// NewTemplateJSONFormatter compiles the template string. Returns an error if it is invalid.
func NewTemplateJSONFormatter(tmpl string) (*TemplateJSONFormatter, error) {
	t, err := template.New("accesslog").Funcs(template.FuncMap{
		"json": jsonMarshal,
	}).Parse(tmpl)
	if err != nil {
		return nil, err
	}
	return &TemplateJSONFormatter{tmpl: t}, nil
}

// jsonMarshal is the "json" template function: encodes v as a complete JSON value.
func jsonMarshal(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (f *TemplateJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(map[string]any, len(entry.Data)+3)
	for k, v := range entry.Data {
		if s, ok := v.(string); ok {
			data[k] = jsonString{v: s}
		} else {
			data[k] = v
		}
	}
	data["time"] = entry.Time
	data["level"] = jsonString{v: entry.Level.String()}
	data["msg"] = jsonString{v: entry.Message}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	if err := f.tmpl.Execute(buf, data); err != nil {
		return nil, err
	}

	if !json.Valid(buf.Bytes()) {
		return nil, fmt.Errorf("accessLog jsonTemplate produced invalid JSON: %.200s", buf.String())
	}

	out := make([]byte, buf.Len()+1)
	copy(out, buf.Bytes())
	out[buf.Len()] = '\n'
	return out, nil
}

func toLogEntry(s, defaultValue string, quote bool) string {
	if len(s) == 0 {
		return defaultValue
	}

	if quote {
		return `"` + s + `"`
	}
	return s
}
