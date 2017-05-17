package accesslog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

type tuple struct {
	sourceKey, translatedKey string
}

type jsonLogFormatter struct {
	timeFormat        string
	coreMapping       []tuple
	requestMapping    []tuple
	originMapping     []tuple
	downstreamMapping []tuple
}

// convertFieldsToMappings splits all the input strings on the colon character
// and returns the {sourceKey, translatedKey} tuples. Any input string without a
// colon will produce a tuple in which the sourceKey and translatedKey are the same.
func convertFieldsToMappings(fields []string) []tuple {
	var mapping []tuple
	for _, s := range fields {
		fields := strings.SplitN(s, ":", 2)
		t := tuple{sourceKey: strings.TrimSpace(fields[0]), translatedKey: strings.TrimSpace(fields[0])}
		if len(fields) == 2 {
			t.translatedKey = strings.TrimSpace(fields[1])
		}

		mapping = append(mapping, t)
	}

	return mapping
}

// newJSONLogFormatter constructs a jsonLogFormatter with its four categories
// of configuration (i.e. four sets of tuples). The core fields will be given
// defaults if the config has not specified any.
func newJSONLogFormatter(settings *types.AccessLog) (jsonLogFormatter, []error) {
	jlf := jsonLogFormatter{timeFormat: settings.TimeFormat}

	var errors []error

	if len(settings.CoreFields) == 0 {
		// default is to propagate all fields
		jlf.coreMapping = convertFieldsToMappings(defaultCoreKeys)
	} else {
		jlf.coreMapping = convertFieldsToMappings(settings.CoreFields)
		if err := requireNoDuplicates(jlf.coreMapping); err != nil {
			errors = append(errors, err)
		}

		if err := validateCoreFields(jlf.coreMapping); err != nil {
			errors = append(errors, err)
		}
	}

	if len(settings.RequestHeaders) > 0 {
		jlf.requestMapping = convertFieldsToMappings(settings.RequestHeaders)
		if err := requireNoDuplicates(jlf.requestMapping); err != nil {
			errors = append(errors, err)
		}
	}

	if len(settings.OriginResponseHeaders) > 0 {
		jlf.originMapping = convertFieldsToMappings(settings.OriginResponseHeaders)
		if err := requireNoDuplicates(jlf.originMapping); err != nil {
			errors = append(errors, err)
		}
	}

	if len(settings.DownstreamResponseHeaders) > 0 {
		jlf.downstreamMapping = convertFieldsToMappings(settings.DownstreamResponseHeaders)
		if err := requireNoDuplicates(jlf.downstreamMapping); err != nil {
			errors = append(errors, err)
		}
	}

	return jlf, errors
}

func requireNoDuplicates(mappings []tuple) error {
	sourceKeys := make(map[string]bool)
	transKeys := make(map[string]bool)
	var duplicates []string

	for _, m := range mappings {
		if _, exists := sourceKeys[m.sourceKey]; exists {
			duplicates = append(duplicates, m.sourceKey)
		}
		if _, exists := transKeys[m.translatedKey]; exists {
			duplicates = append(duplicates, m.translatedKey)
		}
		sourceKeys[m.sourceKey] = true
		transKeys[m.translatedKey] = true
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("Duplicate access log fields: [%s]", strings.Join(duplicates, ","))
	}
	return nil
}

func validateCoreFields(mappings []tuple) error {
	var invalidFields []string

	for _, m := range mappings {
		key := m.sourceKey
		if _, exists := allCoreKeys[key]; !exists {
			invalidFields = append(invalidFields, key)
		}
	}

	if len(invalidFields) > 0 {
		return fmt.Errorf("Unsupported access log fields: [%v]", strings.Join(invalidFields, ","))
	}
	return nil
}

func (l jsonLogFormatter) Write(w io.Writer, logDataTable *LogData) error {

	buf := &bytes.Buffer{}
	buf.WriteString("{")

	for _, tup := range l.coreMapping {
		v, exists := logDataTable.Core[tup.sourceKey]
		if exists {
			l.writeField(buf, tup.translatedKey, v)
		}
	}

	l.writeHeaders(buf, l.requestMapping, logDataTable.Request)
	l.writeHeaders(buf, l.originMapping, logDataTable.OriginResponse)
	l.writeHeaders(buf, l.downstreamMapping, logDataTable.DownstreamResponse)

	buf.WriteString("}\n")

	_, err := w.Write(buf.Bytes())
	return err
}

func (l jsonLogFormatter) writeHeaders(buf *bytes.Buffer, mapping []tuple, hdr http.Header) {
	for _, tup := range mapping {
		v := hdr[tup.sourceKey]
		if len(v) > 0 {
			l.writeField(buf, tup.translatedKey, v[0])
			// note that multi-valued headers are not fully reported
		}
	}
}

func writeFieldKey(buf *bytes.Buffer, key string) {
	if buf.Len() > 1 {
		buf.WriteByte(',')
	}
	buf.WriteByte('"')
	buf.WriteString(key)
	buf.WriteByte('"')
	buf.WriteByte(':')
}

func (l jsonLogFormatter) writeField(buf *bytes.Buffer, key string, value interface{}) {

	switch v := value.(type) {
	case time.Duration:
		writeFieldKey(buf, key)
		buf.WriteString(strconv.FormatFloat(v.Seconds(), 'f', 3, 64))

	case float64:
		if !math.IsInf(v, 0) && !math.IsNaN(v) {
			writeFieldKey(buf, key)
			buf.WriteString(strconv.FormatFloat(v, 'f', 4, 64))
		}

	case float32:
		l.writeField(buf, key, float64(v))

	case time.Time:
		l.writeField(buf, key, v.Format(l.timeFormat))

	case fmt.Stringer:
		l.writeField(buf, key, v.String())

	default:
		enc, err := json.Marshal(value)

		if err != nil {
			log.Errorf("Unable to encode value '%v' to JSON: %v", value, err)
			return
		}

		writeFieldKey(buf, key)
		buf.Write(enc)
	}
}
