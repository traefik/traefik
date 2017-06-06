package accesslog

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
)

// default format for time presentation
const commonLogTimeFormat = "02/Jan/2006:15:04:05 -0700"

// CommonLogFormatter provides formatting in the Traefik common log format
type CommonLogFormatter struct{}

//Format formats the log entry in the Traefik common log format
func (f *CommonLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	timestamp := entry.Data[StartUTC].(time.Time).Format(commonLogTimeFormat)
	elapsedMillis := entry.Data[Duration].(time.Duration).Nanoseconds() / 1000000

	_, err := fmt.Fprintf(b, "%s - %s [%s] \"%s %s %s\" %d %d %s %s %d %s %s %dms\n",
		entry.Data[ClientHost],
		entry.Data[ClientUsername],
		timestamp,
		entry.Data[RequestMethod],
		entry.Data[RequestPath],
		entry.Data[RequestProtocol],
		entry.Data[OriginStatus],
		entry.Data[OriginContentSize],
		toLogString(entry.Data["request_Referer"]),
		toLogString(entry.Data["request_User-Agent"]),
		entry.Data[RequestCount],
		toLogString(entry.Data[FrontendName]),
		toLogString(entry.Data[BackendURL]),
		elapsedMillis)

	return b.Bytes(), err
}

func toLogString(v interface{}) string {
	defaultValue := "-"
	if v == nil {
		return defaultValue
	}

	switch s := v.(type) {
	case string:
		return quoted(s, defaultValue)

	case fmt.Stringer:
		return quoted(s.String(), defaultValue)

	default:
		return defaultValue
	}

}

func quoted(s string, defaultValue string) string {
	if len(s) == 0 {
		return defaultValue
	}
	return `"` + s + `"`
}
