package fast

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	proxyhttputil "github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/http/httpguts"
)

const (
	bufferSize = 32 * 1024
	bufioSize  = 64 * 1024
)

var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

type pool[T any] struct {
	pool sync.Pool
}

func (p *pool[T]) Get() T {
	if tmp := p.pool.Get(); tmp != nil {
		return tmp.(T)
	}

	var res T
	return res
}

func (p *pool[T]) Put(x T) {
	p.pool.Put(x)
}

type writeDetector struct {
	net.Conn

	written bool
}

func (w *writeDetector) Write(p []byte) (int, error) {
	n, err := w.Conn.Write(p)
	if n > 0 {
		w.written = true
	}

	return n, err
}

type writeFlusher struct {
	io.Writer
}

func (w *writeFlusher) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	if f, ok := w.Writer.(http.Flusher); ok {
		f.Flush()
	}

	return n, err
}

type timeoutError struct {
	error
}

func (t timeoutError) Timeout() bool {
	return true
}

func (t timeoutError) Temporary() bool {
	return false
}

// ReverseProxy is the FastProxy reverse proxy implementation.
type ReverseProxy struct {
	debug bool

	connPool *connPool

	writerPool pool[*bufio.Writer]

	proxyAuth string

	targetURL      *url.URL
	passHostHeader bool
	preservePath   bool
}

// NewReverseProxy creates a new ReverseProxy.
func NewReverseProxy(targetURL, proxyURL *url.URL, debug, passHostHeader, preservePath bool, connPool *connPool) (*ReverseProxy, error) {
	var proxyAuth string
	if proxyURL != nil && proxyURL.User != nil && targetURL.Scheme == "http" {
		username := proxyURL.User.Username()
		password, _ := proxyURL.User.Password()
		proxyAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	}

	return &ReverseProxy{
		debug:          debug,
		passHostHeader: passHostHeader,
		preservePath:   preservePath,
		targetURL:      targetURL,
		proxyAuth:      proxyAuth,
		connPool:       connPool,
	}, nil
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Body != nil {
		defer req.Body.Close()
	}

	outReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(outReq)

	// This is not required as the headers are already normalized by net/http.
	outReq.Header.DisableNormalizing()

	for k, v := range req.Header {
		for _, s := range v {
			outReq.Header.Add(k, s)
		}
	}

	removeConnectionHeaders(&outReq.Header)

	for _, header := range hopHeaders {
		outReq.Header.Del(header)
	}

	if p.proxyAuth != "" {
		outReq.Header.Set("Proxy-Authorization", p.proxyAuth)
	}

	if httpguts.HeaderValuesContainsToken(req.Header["Te"], "trailers") {
		outReq.Header.Set("Te", "trailers")
	}

	if p.debug {
		outReq.Header.Set("X-Traefik-Fast-Proxy", "enabled")
	}

	reqUpType := upgradeType(req.Header)
	if !isGraphic(reqUpType) {
		proxyhttputil.ErrorHandler(rw, req, fmt.Errorf("client tried to switch to invalid protocol %q", reqUpType))
		return
	}

	if reqUpType != "" {
		outReq.Header.Set("Connection", "Upgrade")
		outReq.Header.Set("Upgrade", reqUpType)

		if strings.EqualFold(reqUpType, "websocket") {
			cleanWebSocketHeaders(&outReq.Header)
		}
	}

	u2 := new(url.URL)
	*u2 = *req.URL
	u2.Scheme = p.targetURL.Scheme
	u2.Host = p.targetURL.Host

	u := req.URL
	if req.RequestURI != "" {
		parsedURL, err := url.ParseRequestURI(req.RequestURI)
		if err == nil {
			u = parsedURL
		}
	}

	u2.Path = u.Path
	u2.RawPath = u.RawPath

	if p.preservePath {
		u2.Path, u2.RawPath = proxyhttputil.JoinURLPath(p.targetURL, u)
	}

	u2.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")

	outReq.SetHost(u2.Host)
	outReq.Header.SetHost(u2.Host)

	if p.passHostHeader {
		outReq.Header.SetHost(req.Host)
	}

	outReq.SetRequestURI(u2.RequestURI())

	outReq.SetBodyStream(req.Body, int(req.ContentLength))

	outReq.Header.SetMethod(req.Method)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		prior, ok := req.Header["X-Forwarded-For"]
		if len(prior) > 0 {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}

		omit := ok && prior == nil // Go Issue 38079: nil now means don't populate the header
		if !omit {
			outReq.Header.Set("X-Forwarded-For", clientIP)
		}
	}

	if err := p.roundTrip(rw, req, outReq, reqUpType); err != nil {
		proxyhttputil.ErrorHandler(rw, req, err)
	}
}

// Note that unlike the net/http RoundTrip:
//   - we are not supporting "100 Continue" response to forward them as-is to the client.
//   - we are not asking for compressed response automatically. That is because this will add an extra cost when the
//     client is asking for an uncompressed response, as we will have to un-compress it, and nowadays most clients are
//     already asking for compressed response (allowing "passthrough" compression).
func (p *ReverseProxy) roundTrip(rw http.ResponseWriter, req *http.Request, outReq *fasthttp.Request, reqUpType string) error {
	ctx := req.Context()
	trace := httptrace.ContextClientTrace(ctx)

	var co *conn
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
		}

		var err error
		co, err = p.connPool.AcquireConn()
		if err != nil {
			return fmt.Errorf("acquire connection: %w", err)
		}

		// Before writing the request,
		// we mark the conn as expecting to handle a response.
		co.expectedResponse.Store(true)

		wd := &writeDetector{Conn: co}

		// TODO: do not wait to write the full request before reading the response (to handle "100 Continue").
		// TODO: this is currently impossible with fasthttp to write the request partially (headers only).
		// Currently, writing the request fully is a mandatory step before handling the response.
		err = p.writeRequest(wd, outReq)
		if wd.written && trace != nil && trace.WroteRequest != nil {
			// WroteRequest hook is used by the tracing middleware to detect if the request has been written.
			trace.WroteRequest(httptrace.WroteRequestInfo{})
		}
		if err == nil {
			break
		}

		log.Ctx(ctx).Debug().Err(err).Msg("Error while writing request")

		co.Close()

		if wd.written && !isReplayable(req) {
			return err
		}
	}

	// Sending the responseWriter unlocks the connection readLoop, to handle the response.
	co.RWCh <- rwWithUpgrade{
		ReqMethod: req.Method,
		RW:        rw,
		Upgrade:   upgradeResponseHandler(req.Context(), reqUpType),
	}

	if err := <-co.ErrCh; err != nil {
		return err
	}

	p.connPool.ReleaseConn(co)
	return nil
}

func (p *ReverseProxy) writeRequest(co net.Conn, outReq *fasthttp.Request) error {
	bw := p.writerPool.Get()
	if bw == nil {
		bw = bufio.NewWriterSize(co, bufioSize)
	}
	defer p.writerPool.Put(bw)

	bw.Reset(co)

	if err := outReq.Write(bw); err != nil {
		return err
	}

	return bw.Flush()
}

// isReplayable returns whether the request is replayable.
func isReplayable(req *http.Request) bool {
	if req.Body == nil || req.Body == http.NoBody {
		switch req.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			return true
		}

		// The Idempotency-Key, while non-standard, is widely used to
		// mean a POST or other request is idempotent. See
		// https://golang.org/issue/19943#issuecomment-421092421
		if _, ok := req.Header["Idempotency-Key"]; ok {
			return true
		}

		if _, ok := req.Header["X-Idempotency-Key"]; ok {
			return true
		}
	}

	return false
}

// isGraphic returns whether s is ASCII and printable according to
// https://tools.ietf.org/html/rfc20#section-4.2.
func isGraphic(s string) bool {
	for i := range len(s) {
		if s[i] < ' ' || s[i] > '~' {
			return false
		}
	}

	return true
}

type fasthttpHeader interface {
	Peek(key string) []byte
	Set(key string, value string)
	SetCanonical(key []byte, value []byte)
	DelBytes(key []byte)
	Del(key string)
	ConnectionUpgrade() bool
}

// removeConnectionHeaders removes hop-by-hop headers listed in the "Connection" header of h.
// See RFC 7230, section 6.1.
func removeConnectionHeaders(h fasthttpHeader) {
	f := h.Peek(fasthttp.HeaderConnection)
	for _, sf := range bytes.Split(f, []byte{','}) {
		if sf = bytes.TrimSpace(sf); len(sf) > 0 {
			h.DelBytes(sf)
		}
	}
}

// RFC 7234, section 5.4: Should treat Pragma: no-cache like Cache-Control: no-cache.
func fixPragmaCacheControl(header fasthttpHeader) {
	if pragma := header.Peek("Pragma"); bytes.Equal(pragma, []byte("no-cache")) {
		if len(header.Peek("Cache-Control")) == 0 {
			header.Set("Cache-Control", "no-cache")
		}
	}
}

// cleanWebSocketHeaders Even if the websocket RFC says that headers should be case-insensitive,
// some servers need Sec-WebSocket-Key, Sec-WebSocket-Extensions, Sec-WebSocket-Accept,
// Sec-WebSocket-Protocol and Sec-WebSocket-Version to be case-sensitive.
// https://tools.ietf.org/html/rfc6455#page-20
func cleanWebSocketHeaders(headers fasthttpHeader) {
	secWebsocketKey := headers.Peek("Sec-Websocket-Key")
	if len(secWebsocketKey) > 0 {
		headers.SetCanonical([]byte("Sec-WebSocket-Key"), secWebsocketKey)
		headers.Del("Sec-Websocket-Key")
	}

	secWebsocketExtensions := headers.Peek("Sec-Websocket-Extensions")
	if len(secWebsocketExtensions) > 0 {
		headers.SetCanonical([]byte("Sec-WebSocket-Extensions"), secWebsocketExtensions)
		headers.Del("Sec-Websocket-Extensions")
	}

	secWebsocketAccept := headers.Peek("Sec-Websocket-Accept")
	if len(secWebsocketAccept) > 0 {
		headers.SetCanonical([]byte("Sec-WebSocket-Accept"), secWebsocketAccept)
		headers.Del("Sec-Websocket-Accept")
	}

	secWebsocketProtocol := headers.Peek("Sec-Websocket-Protocol")
	if len(secWebsocketProtocol) > 0 {
		headers.SetCanonical([]byte("Sec-WebSocket-Protocol"), secWebsocketProtocol)
		headers.Del("Sec-Websocket-Protocol")
	}

	secWebsocketVersion := headers.Peek("Sec-Websocket-Version")
	if len(secWebsocketVersion) > 0 {
		headers.SetCanonical([]byte("Sec-WebSocket-Version"), secWebsocketVersion)
		headers.Del("Sec-Websocket-Version")
	}
}
