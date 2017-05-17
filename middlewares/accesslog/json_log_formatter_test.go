package accesslog

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestJSONLogFormatter(t *testing.T) {
	jlf, errs := newJSONLogFormatter(&types.AccessLog{
		TimeFormat: commonLogTimeFormat,
		CoreFields: []string{
			"StartUTC:StartUTC",
			"Duration:Duration",
			"FrontendName:thefrontend",
			"BackendName:BackendName",
			"BackendURL:BackendURL",
			"OriginDuration:OriginDuration",
			"OriginContentSize:OriginContentSize",
			"RequestAddr:RequestAddr",
			"RequestMethod:RequestMethod",
			"RequestPath:RequestPath",
			"RequestProtocol:RequestProtocol",
			"OriginStatus:OriginStatus",
			"DownstreamContentSize:DownstreamContentSize",
			"RequestCount:RequestCount",
			"ClientHost:ClHost",
			"ClientPort:ClPort",
			"ClientUsername:ClUsername",
		},
		RequestHeaders:            []string{"User-Agent: user_agent", "Referer: referrer"},
		OriginResponseHeaders:     []string{"Server: upstream_http_server"},
		DownstreamResponseHeaders: []string{"Location: sent_http_location"},
	})

	assert.Len(t, errs, 0)

	buf := &bytes.Buffer{}
	jlf.Write(buf, fixtureLogDataTable(12345))
	s := buf.String()
	assert.Equal(t,
		`{"StartUTC":"`,
		s[:13])
	_, err := time.Parse(commonLogTimeFormat, s[13:39])
	assert.NoError(t, err)
	assert.Equal(t,
		`","Duration":0.002,"thefrontend":"frontend","BackendName":"backend","BackendURL":"http://test.host.name:8181/a/b/c?q=1#z1","OriginDuration":0.001,`+
			`"OriginContentSize":102,`+
			`"RequestAddr":"test.host.name:8181","RequestMethod":"GET","RequestPath":"/y/xy/z","RequestProtocol":"HTTP/1.1",`+
			`"OriginStatus":200,"DownstreamContentSize":82,"RequestCount":12345,`+
			`"ClHost":"190.190.190.190","ClPort":"20121","ClUsername":"-",`+
			`"user_agent":"user-agent-very-very-long-string","referrer":"http://example.com/x/y/z",`+
			`"upstream_http_server":"foobar v1",`+
			`"sent_http_location":"http://somewhere.else/a/b"`+"}\n",
		s[39:])
}

func TestGzipRatioDivideByZero(t *testing.T) {
	jlf, errs := newJSONLogFormatter(&types.AccessLog{
		TimeFormat: commonLogTimeFormat,
		CoreFields: []string{"FrontendName", "GzipRatio", "BackendName"},
	})

	if errs != nil {
		panic("Unexpected error")
	}

	buf := &bytes.Buffer{}
	data := fixtureLogDataTable(12345)
	data.Core[DownstreamContentSize] = 0 // cause the calculation to reach infinity
	jlf.Write(buf, data)
	s := buf.String()
	assert.Equal(t, `{"FrontendName":"frontend","BackendName":"backend"`+"}\n", s)
}

func TestNewJSONLogFormatterValidation(t *testing.T) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	_, errs := newJSONLogFormatter(&types.AccessLog{
		CoreFields:                []string{"Duration", "StartUTC", "FrontendName:frontend", "NonExistent"},
		RequestHeaders:            []string{"Host: http_host"},
		OriginResponseHeaders:     []string{"Server: upstream_http_server"},
		DownstreamResponseHeaders: []string{"Location: sent_http_location"},
	})

	assert.Len(t, errs, 1)
	assert.Equal(t, []error{errors.New("Unsupported access log fields: [NonExistent]")}, errs)
}

func TestNewJSONLogFormatterValidationNonBlank(t *testing.T) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	_, errs := newJSONLogFormatter(&types.AccessLog{
		CoreFields:                []string{" ", " : "},
		RequestHeaders:            []string{" ", " : "},
		OriginResponseHeaders:     []string{" ", " : "},
		DownstreamResponseHeaders: []string{" ", " : "},
	})

	assert.Len(t, errs, 5)
	assert.Equal(t, []error{
		errors.New("Duplicate access log fields: [,]"),
		errors.New("Unsupported access log fields: [,]"),
		errors.New("Duplicate access log fields: [,]"),
		errors.New("Duplicate access log fields: [,]"),
		errors.New("Duplicate access log fields: [,]"),
	}, errs)
}

func TestNewJSONLogFormatterValidationNoDuplicates(t *testing.T) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	_, errs := newJSONLogFormatter(&types.AccessLog{
		CoreFields:                []string{"Duration:x", "Duration:y", "FrontendName:end", "BackendName:end"},
		RequestHeaders:            []string{"Host: http_host", "Host: http_host"},
		OriginResponseHeaders:     []string{"Server: server", "Server: server"},
		DownstreamResponseHeaders: []string{"Location: location", "Location: location"},
	})

	assert.Len(t, errs, 4)
	assert.Equal(t, []error{
		errors.New("Duplicate access log fields: [Duration,end]"),
		errors.New("Duplicate access log fields: [Host,http_host]"),
		errors.New("Duplicate access log fields: [Server,server]"),
		errors.New("Duplicate access log fields: [Location,location]"),
	}, errs)
}
