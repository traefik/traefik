package accesslog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
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
		colon := strings.IndexByte(s, ':')

		var t tuple

		if colon > 0 {
			t = tuple{
				strings.TrimSpace(s[:colon]),
				strings.TrimSpace(s[colon+1:]),
			}
		} else {
			ws := strings.TrimSpace(s)
			t = tuple{ws, ws}
		}

		mapping = append(mapping, t)
	}
	return mapping
}

// newJSONLogFormatter constructs a jsonLogFormatter with its four categories
// of configuration (i.e. four sets of tuples). The core fields will be given
// defaults if the config has not specified any.
func newJSONLogFormatter(settings *types.AccessLog) jsonLogFormatter {
	jlf := jsonLogFormatter{timeFormat: settings.TimeFormat}
	ok := true

	if len(settings.CoreFields) == 0 {
		// default is to propagate all fields
		for _, name := range defaultCoreKeys {
			t := tuple{name, name}
			jlf.coreMapping = append(jlf.coreMapping, t)
		}
	} else {
		jlf.coreMapping = convertFieldsToMappings(settings.CoreFields)
		ok = requireNoDuplicates(jlf.coreMapping) && ok
		ok = validateCoreFields(jlf.coreMapping) && ok
	}

	if len(settings.RequestHeaders) > 0 {
		jlf.requestMapping = convertFieldsToMappings(settings.RequestHeaders)
		ok = requireNoDuplicates(jlf.requestMapping) && ok
	}

	if len(settings.OriginResponseHeaders) > 0 {
		jlf.originMapping = convertFieldsToMappings(settings.OriginResponseHeaders)
		ok = requireNoDuplicates(jlf.originMapping) && ok
	}

	if len(settings.DownstreamResponseHeaders) > 0 {
		jlf.downstreamMapping = convertFieldsToMappings(settings.DownstreamResponseHeaders)
		ok = requireNoDuplicates(jlf.downstreamMapping) && ok
	}

	if !ok {
		exiter.Exit(1)
	}

	return jlf
}

func requireNoDuplicates(mappings []tuple) bool {
	sourceKeys := make(map[string]struct{})
	transKeys := make(map[string]struct{})
	duplicates := ""

	for _, m := range mappings {
		if _, exists := sourceKeys[m.sourceKey]; exists {
			duplicates = duplicates + " " + m.sourceKey
		}
		if _, exists := transKeys[m.translatedKey]; exists {
			duplicates = duplicates + " " + m.translatedKey
		}
		sourceKeys[m.sourceKey] = struct{}{}
		transKeys[m.translatedKey] = struct{}{}
	}

	if len(duplicates) > 0 {
		log.Errorf("Duplicate access log fields:%s", duplicates)
		return false
	}
	return true
}

func validateCoreFields(mappings []tuple) bool {
	var invalidFields []string

	for _, m := range mappings {
		key := m.sourceKey
		if _, exists := allCoreKeys[key]; !exists {
			invalidFields = append(invalidFields, key)
		}
	}

	if len(invalidFields) > 0 {
		log.Errorf("Unsupported access log fields: %v", invalidFields)
		return false
	}
	return true
}

//-------------------------------------------------------------------------------------------------

// Exiter provides a facade for os.Exit.
type Exiter interface {
	Exit(code int)
}

type stdExiter struct{}

func (stdExiter) Exit(code int) {
	os.Exit(code)
}

// exiter provides a seam for testing code containing os.Exit.
var exiter Exiter = stdExiter{}

//-------------------------------------------------------------------------------------------------

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

func asSeconds(d time.Duration) float64 {
	// half-rounded up to the nearest millisecond, then converted to seconds
	return float64((d.Nanoseconds()+500000)/1000000) / 1000
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
		buf.WriteString(strconv.FormatFloat(asSeconds(v), 'f', 3, 64))

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
