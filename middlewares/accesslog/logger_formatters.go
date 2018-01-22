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

//Format formats the log entry in the Traefik common log format
func (f *CommonLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	timestamp := entry.Data[StartUTC].(time.Time).Format(commonLogTimeFormat)
	elapsedMillis := entry.Data[Duration].(time.Duration).Nanoseconds() / 1000000

	_, err := fmt.Fprintf(b, "%s - %s [%s] \"%s %s %s\" %v %v %s %s %v %s %s %dms\n",
		entry.Data[ClientHost],
		entry.Data[ClientUsername],
		timestamp,
		entry.Data[RequestMethod],
		entry.Data[RequestPath],
		entry.Data[RequestProtocol],
		toLog(entry.Data[OriginStatus], defaultValue),
		toLog(entry.Data[OriginContentSize], defaultValue),
		toLog(entry.Data["request_Referer"], `"-"`),
		toLog(entry.Data["request_User-Agent"], `"-"`),
		toLog(entry.Data[RequestCount], defaultValue),
		toLog(entry.Data[FrontendName], defaultValue),
		toLog(entry.Data[BackendURL], defaultValue),
		elapsedMillis)

	return b.Bytes(), err
}

func toLog(v interface{}, defaultValue string) interface{} {
	if v == nil {
		return defaultValue
	}

	switch s := v.(type) {
	case string:
		return quoted(s, defaultValue)

	case fmt.Stringer:
		return quoted(s.String(), defaultValue)

	default:
		return v
	}

}

func quoted(s string, defaultValue string) string {
	if len(s) == 0 {
		return defaultValue
	}
	return `"` + s + `"`
}
