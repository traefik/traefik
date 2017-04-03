// package forwarder implements http handler that forwards requests to remote server
// and serves back the response
// websocket proxying support based on https://github.com/yhat/wsutil
package forward

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/vulcand/oxy/utils"
)

// ReqRewriter can alter request headers and body
type ReqRewriter interface {
	Rewrite(r *http.Request)
}

type optSetter func(f *Forwarder) error

// PassHostHeader specifies if a client's Host header field should
// be delegated
func PassHostHeader(b bool) optSetter {
	return func(f *Forwarder) error {
		f.passHost = b
		return nil
	}
}

// StreamResponse forces streaming body (flushes response directly to client)
func StreamResponse(b bool) optSetter {
	return func(f *Forwarder) error {
		f.httpForwarder.streamResponse = b
		return nil
	}
}

// RoundTripper sets a new http.RoundTripper
// Forwarder will use http.DefaultTransport as a default round tripper
func RoundTripper(r http.RoundTripper) optSetter {
	return func(f *Forwarder) error {
		f.roundTripper = r
		return nil
	}
}

// Rewriter defines a request rewriter for the HTTP forwarder
func Rewriter(r ReqRewriter) optSetter {
	return func(f *Forwarder) error {
		f.httpForwarder.rewriter = r
		return nil
	}
}

// WebsocketRewriter defines a request rewriter for the websocket forwarder
func WebsocketRewriter(r ReqRewriter) optSetter {
	return func(f *Forwarder) error {
		f.websocketForwarder.rewriter = r
		return nil
	}
}

// ErrorHandler is a functional argument that sets error handler of the server
func ErrorHandler(h utils.ErrorHandler) optSetter {
	return func(f *Forwarder) error {
		f.errHandler = h
		return nil
	}
}

// Logger specifies the logger to use.
// Forwarder will default to oxyutils.NullLogger if no logger has been specified
func Logger(l utils.Logger) optSetter {
	return func(f *Forwarder) error {
		f.log = l
		return nil
	}
}

// Forwarder wraps two traffic forwarding implementations: HTTP and websockets.
// It decides based on the specified request which implementation to use
type Forwarder struct {
	*httpForwarder
	*websocketForwarder
	*handlerContext
}

// handlerContext defines a handler context for error reporting and logging
type handlerContext struct {
	errHandler utils.ErrorHandler
	log        utils.Logger
}

// httpForwarder is a handler that can reverse proxy
// HTTP traffic
type httpForwarder struct {
	roundTripper   http.RoundTripper
	rewriter       ReqRewriter
	passHost       bool
	streamResponse bool
}

// websocketForwarder is a handler that can reverse proxy
// websocket traffic
type websocketForwarder struct {
	rewriter        ReqRewriter
	TLSClientConfig *tls.Config
}

// New creates an instance of Forwarder based on the provided list of configuration options
func New(setters ...optSetter) (*Forwarder, error) {
	f := &Forwarder{
		httpForwarder:      &httpForwarder{},
		websocketForwarder: &websocketForwarder{},
		handlerContext:     &handlerContext{},
	}
	for _, s := range setters {
		if err := s(f); err != nil {
			return nil, err
		}
	}
	if f.httpForwarder.roundTripper == nil {
		f.httpForwarder.roundTripper = http.DefaultTransport
	}
	if f.httpForwarder.rewriter == nil {
		h, err := os.Hostname()
		if err != nil {
			h = "localhost"
		}
		f.httpForwarder.rewriter = &HeaderRewriter{TrustForwardHeader: true, Hostname: h}
	}
	if f.log == nil {
		f.log = utils.NullLogger
	}
	if f.errHandler == nil {
		f.errHandler = utils.DefaultHandler
	}
	return f, nil
}

// ServeHTTP decides which forwarder to use based on the specified
// request and delegates to the proper implementation
func (f *Forwarder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if isWebsocketRequest(req) {
		f.websocketForwarder.serveHTTP(w, req, f.handlerContext)
	} else {
		f.httpForwarder.serveHTTP(w, req, f.handlerContext)
	}
}

// serveHTTP forwards HTTP traffic using the configured transport
func (f *httpForwarder) serveHTTP(w http.ResponseWriter, req *http.Request, ctx *handlerContext) {
	start := time.Now().UTC()
	response, err := f.roundTripper.RoundTrip(f.copyRequest(req, req.URL))
	if err != nil {
		ctx.log.Errorf("Error forwarding to %v, err: %v", req.URL, err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}

	utils.CopyHeaders(w.Header(), response.Header)
	// Remove hop-by-hop headers.
	utils.RemoveHeaders(w.Header(), HopHeaders...)
	w.WriteHeader(response.StatusCode)

	stream := f.streamResponse
	if !stream {
		contentType, err := utils.GetHeaderMediaType(response.Header, ContentType)
		if err == nil {
			stream = contentType == "text/event-stream"
		}
	}
	written, err := io.Copy(newResponseFlusher(w, stream), response.Body)

	if req.TLS != nil {
		ctx.log.Infof("Round trip: %v, code: %v, duration: %v tls:version: %x, tls:resume:%t, tls:csuite:%x, tls:server:%v",
			req.URL, response.StatusCode, time.Now().UTC().Sub(start),
			req.TLS.Version,
			req.TLS.DidResume,
			req.TLS.CipherSuite,
			req.TLS.ServerName)
	} else {
		ctx.log.Infof("Round trip: %v, code: %v, duration: %v",
			req.URL, response.StatusCode, time.Now().UTC().Sub(start))
	}

	defer response.Body.Close()

	if err != nil {
		ctx.log.Errorf("Error copying upstream response Body: %v", err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}

	if written != 0 {
		w.Header().Set(ContentLength, strconv.FormatInt(written, 10))
	}
}

// copyRequest makes a copy of the specified request to be sent using the configured
// transport
func (f *httpForwarder) copyRequest(req *http.Request, u *url.URL) *http.Request {
	outReq := new(http.Request)
	*outReq = *req // includes shallow copies of maps, but we handle this below

	outReq.URL = utils.CopyURL(req.URL)
	outReq.URL.Scheme = u.Scheme
	outReq.URL.Host = u.Host
	outReq.URL.Opaque = req.RequestURI
	// raw query is already included in RequestURI, so ignore it to avoid dupes
	outReq.URL.RawQuery = ""
	// Do not pass client Host header unless optsetter PassHostHeader is set.
	if !f.passHost {
		outReq.Host = u.Host
	}
	outReq.Proto = "HTTP/1.1"
	outReq.ProtoMajor = 1
	outReq.ProtoMinor = 1

	// Overwrite close flag so we can keep persistent connection for the backend servers
	outReq.Close = false

	outReq.Header = make(http.Header)
	utils.CopyHeaders(outReq.Header, req.Header)

	if f.rewriter != nil {
		f.rewriter.Rewrite(outReq)
	}
	return outReq
}

// serveHTTP forwards websocket traffic
func (f *websocketForwarder) serveHTTP(w http.ResponseWriter, req *http.Request, ctx *handlerContext) {
	outReq := f.copyRequest(req, req.URL)
	host := outReq.URL.Host
	dial := net.Dial

	// if host does not specify a port, use the default http port
	if !strings.Contains(host, ":") {
		if outReq.URL.Scheme == "wss" {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}

	if outReq.URL.Scheme == "wss" {
		if f.TLSClientConfig == nil {
			f.TLSClientConfig = http.DefaultTransport.(*http.Transport).TLSClientConfig
		}
		dial = func(network, address string) (net.Conn, error) {
			return tls.Dial("tcp", host, f.TLSClientConfig)
		}
	}

	targetConn, err := dial("tcp", host)
	if err != nil {
		ctx.log.Errorf("Error dialing `%v`: %v", host, err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		ctx.log.Errorf("Unable to hijack the connection: %v", reflect.TypeOf(w))
		ctx.errHandler.ServeHTTP(w, req, nil)
		return
	}
	underlyingConn, _, err := hijacker.Hijack()
	if err != nil {
		ctx.log.Errorf("Unable to hijack the connection: %v %v", reflect.TypeOf(w), err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}
	// it is now caller's responsibility to Close the underlying connection
	defer underlyingConn.Close()
	defer targetConn.Close()

	ctx.log.Infof("Writing outgoing Websocket request to target connection: %+v", outReq)

	// write the modified incoming request to the dialed connection
	if err = outReq.Write(targetConn); err != nil {
		ctx.log.Errorf("Unable to copy request to target: %v", err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}
	errc := make(chan error, 2)
	replicate := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}
	go replicate(targetConn, underlyingConn)
	go replicate(underlyingConn, targetConn)
	<-errc
}

// copyRequest makes a copy of the specified request.
func (f *websocketForwarder) copyRequest(req *http.Request, u *url.URL) (outReq *http.Request) {
	outReq = new(http.Request)
	*outReq = *req // includes shallow copies of maps, but we handle this below

	outReq.URL = utils.CopyURL(req.URL)
	outReq.URL.Scheme = u.Scheme

	//sometimes backends might be registered as HTTP/HTTPS servers so translate URLs to websocket URLs.
	switch u.Scheme {
	case "https":
		outReq.URL.Scheme = "wss"
	case "http":
		outReq.URL.Scheme = "ws"
	}

	outReq.URL.Host = u.Host
	outReq.URL.Opaque = req.RequestURI
	// raw query is already included in RequestURI, so ignore it to avoid dupes
	outReq.URL.RawQuery = ""

	outReq.Proto = "HTTP/1.1"
	outReq.ProtoMajor = 1
	outReq.ProtoMinor = 1

	// Overwrite close flag so we can keep persistent connection for the backend servers
	outReq.Close = false

	outReq.Header = make(http.Header)
	utils.CopyHeaders(outReq.Header, req.Header)

	if f.rewriter != nil {
		f.rewriter.Rewrite(outReq)
	}
	return outReq
}

// isWebsocketRequest determines if the specified HTTP request is a
// websocket handshake request
func isWebsocketRequest(req *http.Request) bool {
	containsHeader := func(name, value string) bool {
		items := strings.Split(req.Header.Get(name), ",")
		for _, item := range items {
			if value == strings.ToLower(strings.TrimSpace(item)) {
				return true
			}
		}
		return false
	}
	return containsHeader(Connection, "upgrade") && containsHeader(Upgrade, "websocket")
}
