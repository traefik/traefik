package accesslog

import (
	"bytes"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// default format for time presentation
const (
	commonLogTimeFormat = "02/Jan/2006:15:04:05 -0700"
	defaultValue        = "-"
)

// CommonLogFormatter provides formatting in the Traefik common log format
type CommonLogFormatter struct{}

// Format formats the log entry in the Traefik common log format
func (f *CommonLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	var timestamp = defaultValue
	if v, ok := entry.Data[StartUTC]; ok {
		timestamp = v.(time.Time).Format(commonLogTimeFormat)
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
		toLog(entry.Data, OriginStatus, defaultValue, true),
		toLog(entry.Data, OriginContentSize, defaultValue, true),
		toLog(entry.Data, "request_Referer", `"-"`, true),
		toLog(entry.Data, "request_User-Agent", `"-"`, true),
		toLog(entry.Data, RequestCount, defaultValue, true),
		toLog(entry.Data, FrontendName, defaultValue, true),
		toLog(entry.Data, BackendURL, defaultValue, true),
		elapsedMillis)

	return b.Bytes(), err
}

func toLog(fields logrus.Fields, key string, defaultValue string, quoted bool) interface{} {
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
func toLogEntry(s string, defaultValue string, quote bool) string {
	if len(s) == 0 {
		return defaultValue
	}

	if quote {
		return `"` + s + `"`
	}
	return s
}
