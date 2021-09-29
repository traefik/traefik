package accesslog

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
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

type PatternLogFormatter struct {
	pattern string
}

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
		toLog(entry.Data, OriginStatus, defaultValue, true),
		toLog(entry.Data, OriginContentSize, defaultValue, true),
		toLog(entry.Data, "request_Referer", `"-"`, true),
		toLog(entry.Data, "request_User-Agent", `"-"`, true),
		toLog(entry.Data, RequestCount, defaultValue, true),
		toLog(entry.Data, RouterName, `"-"`, true),
		toLog(entry.Data, ServiceURL, `"-"`, true),
		elapsedMillis)

	return b.Bytes(), err
}

func toLog(fields logrus.Fields, key, defaultValue string, quoted bool) interface{} {
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

func toLogEntry(s, defaultValue string, quote bool) string {
	if len(s) == 0 {
		return defaultValue
	}

	if quote {
		return `"` + s + `"`
	}
	return s
}

func (f *PatternLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	re, regexerr := regexp.Compile(`([^$][\w]+)*`)

	if regexerr != nil {
		return nil, regexerr
	}

	formattedPatterns := f.pattern
	var logEntryFields []string

	for _, match := range re.FindAllStringSubmatch(f.pattern, -1) {
		if len(match[1]) > 0 {
			matchfield := match[1]
			logEntryFields = append(logEntryFields, matchfield)
			replaceValue := "%v"
			if matchfield == Duration {
				replaceValue = "%vms"
			}
			formattedPatterns = strings.Replace(formattedPatterns, fmt.Sprintf("$%s", match[1]), replaceValue, -1)
		}
	}

	var toLogEntryies []interface{}
	quoted := false

	for _, logEntryField := range logEntryFields {
		switch logEntryField {
		case OriginStatus, OriginContentSize, RouterName, ServiceURL:
			quoted = true
			toLogEntryies = append(toLogEntryies, toLog(entry.Data, logEntryField, defaultValue, quoted))
		case StartUTC:
			if v, ok := entry.Data[StartUTC]; ok {
				timestamp := v.(time.Time).Format(commonLogTimeFormat)
				toLogEntryies = append(toLogEntryies, timestamp)
			}
		case StartLocal:
			if v, ok := entry.Data[StartLocal]; ok {
				timestamp := v.(time.Time).Local().Format(commonLogTimeFormat)
				toLogEntryies = append(toLogEntryies, timestamp)
			}
		case Duration:
			var elapsedMillis int64
			if v, ok := entry.Data[Duration]; ok {
				elapsedMillis = v.(time.Duration).Nanoseconds() / 1000000
			}
			toLogEntryies = append(toLogEntryies, elapsedMillis)
		case RequestRefererHeader:
			toLogEntryies = append(toLogEntryies, toLog(entry.Data, "request_Referer", "-", true))
		case RequestUserAgentHeader:
			toLogEntryies = append(toLogEntryies, toLog(entry.Data, "request_User-Agent", "-", true))
		default:
			quoted = false
			toLogEntryies = append(toLogEntryies, toLog(entry.Data, logEntryField, defaultValue, quoted))
		}
	}

	_, err := fmt.Fprintf(b, formattedPatterns, toLogEntryies...)

	return b.Bytes(), err
}
