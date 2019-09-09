package tracer

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	// TODO(gbbr): find a more effective way to keep this up to date,
	// e.g. via `go generate`
	tracerVersion = "v1.10.0"

	// We copy the transport to avoid using the default one, as it might be
	// augmented with tracing and we don't want these calls to be recorded.
	// See https://golang.org/pkg/net/http/#DefaultTransport .
	defaultRoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

const (
	defaultHostname    = "localhost"
	defaultPort        = "8126"
	defaultAddress     = defaultHostname + ":" + defaultPort
	defaultHTTPTimeout = time.Second             // defines the current timeout before giving up with the send process
	traceCountHeader   = "X-Datadog-Trace-Count" // header containing the number of traces in the payload
)

// transport is an interface for span submission to the agent.
type transport interface {
	// send sends the payload p to the agent using the transport set up.
	// It returns a non-nil response body when no error occurred.
	send(p *payload) (body io.ReadCloser, err error)
}

// newTransport returns a new Transport implementation that sends traces to a
// trace agent running on the given hostname and port, using a given
// http.RoundTripper. If the zero values for hostname and port are provided,
// the default values will be used ("localhost" for hostname, and "8126" for
// port). If roundTripper is nil, a default is used.
//
// In general, using this method is only necessary if you have a trace agent
// running on a non-default port, if it's located on another machine, or when
// otherwise needing to customize the transport layer, for instance when using
// a unix domain socket.
func newTransport(addr string, roundTripper http.RoundTripper) transport {
	if roundTripper == nil {
		roundTripper = defaultRoundTripper
	}
	return newHTTPTransport(addr, roundTripper)
}

// newDefaultTransport return a default transport for this tracing client
func newDefaultTransport() transport {
	return newHTTPTransport(defaultAddress, defaultRoundTripper)
}

type httpTransport struct {
	traceURL string            // the delivery URL for traces
	client   *http.Client      // the HTTP client used in the POST
	headers  map[string]string // the Transport headers
}

// newHTTPTransport returns an httpTransport for the given endpoint
func newHTTPTransport(addr string, roundTripper http.RoundTripper) *httpTransport {
	// initialize the default EncoderPool with Encoder headers
	defaultHeaders := map[string]string{
		"Datadog-Meta-Lang":             "go",
		"Datadog-Meta-Lang-Version":     strings.TrimPrefix(runtime.Version(), "go"),
		"Datadog-Meta-Lang-Interpreter": runtime.Compiler + "-" + runtime.GOARCH + "-" + runtime.GOOS,
		"Datadog-Meta-Tracer-Version":   tracerVersion,
		"Content-Type":                  "application/msgpack",
	}
	return &httpTransport{
		traceURL: fmt.Sprintf("http://%s/v0.4/traces", resolveAddr(addr)),
		client: &http.Client{
			Transport: roundTripper,
			Timeout:   defaultHTTPTimeout,
		},
		headers: defaultHeaders,
	}
}

func (t *httpTransport) send(p *payload) (body io.ReadCloser, err error) {
	// prepare the client and send the payload
	req, err := http.NewRequest("POST", t.traceURL, p)
	if err != nil {
		return nil, fmt.Errorf("cannot create http request: %v", err)
	}
	for header, value := range t.headers {
		req.Header.Set(header, value)
	}
	req.Header.Set(traceCountHeader, strconv.Itoa(p.itemCount()))
	req.Header.Set("Content-Length", strconv.Itoa(p.size()))
	response, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	if code := response.StatusCode; code >= 400 {
		// error, check the body for context information and
		// return a nice error.
		msg := make([]byte, 1000)
		n, _ := response.Body.Read(msg)
		response.Body.Close()
		txt := http.StatusText(code)
		if n > 0 {
			return nil, fmt.Errorf("%s (Status: %s)", msg[:n], txt)
		}
		return nil, fmt.Errorf("%s", txt)
	}
	return response.Body, nil
}

// resolveAddr resolves the given agent address and fills in any missing host
// and port using the defaults. Some environment variable settings will
// take precedence over configuration.
func resolveAddr(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// no port in addr
		host = addr
	}
	if host == "" {
		host = defaultHostname
	}
	if port == "" {
		port = defaultPort
	}
	if v := os.Getenv("DD_AGENT_HOST"); v != "" {
		host = v
	}
	if v := os.Getenv("DD_TRACE_AGENT_PORT"); v != "" {
		port = v
	}
	return fmt.Sprintf("%s:%s", host, port)
}
