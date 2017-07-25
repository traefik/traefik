package trace

import (
	"bytes"
	"fmt"
	"io"
	"log/syslog"
	"net/http"
	"net/url"

	"github.com/codegangsta/cli"
	oxytrace "github.com/vulcand/oxy/trace"
	"github.com/vulcand/vulcand/plugin"
)

const Type = "trace"

// Trace plugin emits structured logs to syslog facility
type Trace struct {
	// ReqHeaders - request headers to capture
	ReqHeaders []string
	// RespHeaders - response headers to capture
	RespHeaders []string
	// Address in format syslog://host:port or syslog:///path/socket.sock
	Addr string
}

// New returns a new Trace plugin
func New(addr string, reqHeaders, respHeaders []string) (*Trace, error) {
	if _, err := newWriter(addr); err != nil {
		return nil, err
	}
	return &Trace{
		ReqHeaders:  reqHeaders,
		RespHeaders: respHeaders,
		Addr:        addr,
	}, nil
}

// NewHandler creates a new http.Handler middleware
func (t *Trace) NewHandler(next http.Handler) (http.Handler, error) {
	return newTraceHandler(next, t)
}

// String is a user-friendly representation of the handler
func (t *Trace) String() string {
	return fmt.Sprintf("addr=%v, reqHeaders=%v, respHeaders=%v", t.Addr, t.ReqHeaders, t.RespHeaders)
}

func newTraceHandler(next http.Handler, t *Trace) (*oxytrace.Tracer, error) {
	writer, err := newWriter(t.Addr)
	if err != nil {
		return nil, err
	}
	return oxytrace.New(next, writer, oxytrace.RequestHeaders(t.ReqHeaders...), oxytrace.ResponseHeaders(t.RespHeaders...))
}

// FromOther creates and validates Trace plugin instance from serialized format
func FromOther(t Trace) (plugin.Middleware, error) {
	return New(t.Addr, t.ReqHeaders, t.RespHeaders)
}

// FromCli creates a Trace plugin object from command line
func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return New(c.String("addr"), c.StringSlice("reqHeader"), c.StringSlice("respHeader"))
}

// GetSpec returns all information neccessary for Vulcand to plugin this extension
func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      Type,
		FromOther: FromOther,
		FromCli:   FromCli,
		CliFlags:  CliFlags(),
	}
}

// CliFlags is used to add command-line arguments to the CLI tool - vctl
func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Usage: "Address of the output, e.g. syslog:///tmp/out.sock",
		},
		cli.StringSliceFlag{
			Name:  "reqHeader",
			Usage: "if provided, captures headers from requests",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:  "respHeader",
			Usage: "if provided, captures headers from response",
			Value: &cli.StringSlice{},
		},
	}
}

func newWriter(addr string) (io.Writer, error) {
	u, err := url.Parse(addr)
	if u.Scheme != "syslog" {
		return nil, fmt.Errorf("unsupported scheme '%v' currently supported only 'syslog'", u.Scheme)
	}
	pr, err := parseSyslogPriority(u)
	if err != nil {
		return nil, err
	}

	var w io.Writer
	if u.Host != "" {
		w, err = syslog.Dial("udp", u.Host, pr, SyslogTag)
	} else if u.Path != "" {
		w, err = syslog.Dial("unixgram", u.Path, pr, SyslogTag)
	} else if u.Host == "" && u.Path == "" {
		w, err = syslog.Dial("", "", pr, SyslogTag)
	} else {
		return nil, fmt.Errorf("unsupported address format: %v", addr)
	}
	if err != nil {
		return nil, err
	}
	return &prefixWriter{p: []byte(parsePrefix(u)), w: w}, nil
}

func parsePrefix(u *url.URL) string {
	t := u.Query().Get("prefix")
	if t != "" {
		return t
	}
	return SyslogPrefix
}

func parseSyslogPriority(u *url.URL) (syslog.Priority, error) {
	vals := u.Query()
	pr, err := sevToString(vals.Get("sev"))
	if err != nil {
		return 0, err
	}
	f, err := fToString(vals.Get("f"))
	if err != nil {
		return 0, err
	}
	return pr | f, nil
}

func sevToString(sev string) (pr syslog.Priority, err error) {
	switch sev {
	case "ALERT":
		pr |= syslog.LOG_ALERT
	case "CRIT":
		pr |= syslog.LOG_CRIT
	case "ERR":
		pr |= syslog.LOG_ERR
	case "WARNING":
		pr |= syslog.LOG_WARNING
	case "NOTICE":
		pr |= syslog.LOG_NOTICE
	case "INFO":
		pr |= syslog.LOG_INFO
	case "DEBUG", "":
		pr |= syslog.LOG_DEBUG
	default:
		return 0, fmt.Errorf("uknown severity: %v", sev)
	}
	return pr, nil
}

func fToString(v string) (f syslog.Priority, err error) {
	switch v {
	case "USER":
		f |= syslog.LOG_USER
	case "MAIL":
		f |= syslog.LOG_MAIL
	case "DAEMON":
		f |= syslog.LOG_DAEMON
	case "AUTH":
		f |= syslog.LOG_AUTH
	case "SYSLOG":
		f |= syslog.LOG_SYSLOG
	case "LPR":
		f |= syslog.LOG_LPR
	case "NEWS":
		f |= syslog.LOG_NEWS
	case "UUCP":
		f |= syslog.LOG_UUCP
	case "CRON":
		f |= syslog.LOG_CRON
	case "AUTHPRIV":
		f |= syslog.LOG_AUTHPRIV
	case "FTP":
		f |= syslog.LOG_FTP
	case "LOG_LOCAL0", "":
		f |= syslog.LOG_LOCAL0
	case "LOG_LOCAL1":
		f |= syslog.LOG_LOCAL1
	case "LOG_LOCAL2":
		f |= syslog.LOG_LOCAL2
	case "LOG_LOCAL3":
		f |= syslog.LOG_LOCAL3
	case "LOG_LOCAL4":
		f |= syslog.LOG_LOCAL4
	case "LOG_LOCAL5":
		f |= syslog.LOG_LOCAL5
	case "LOG_LOCAL6":
		f |= syslog.LOG_LOCAL6
	case "LOG_LOCAL7":
		f |= syslog.LOG_LOCAL7
	default:
		return 0, fmt.Errorf("unsupported facility: %v", v)
	}
	return f, nil
}

const SyslogPrefix = "@cee: "
const SyslogTag = "pid"

type prefixWriter struct {
	w io.Writer
	p []byte
}

func (p *prefixWriter) Write(val []byte) (int, error) {
	b := bytes.Buffer{}
	b.Write(p.p)
	b.Write(val)
	return p.w.Write(b.Bytes())
}
