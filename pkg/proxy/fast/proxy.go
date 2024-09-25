package fast

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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

type buffConn struct {
	*bufio.Reader
	net.Conn
}

func (b buffConn) Read(p []byte) (int, error) {
	return b.Reader.Read(p)
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

	bufferPool      pool[[]byte]
	readerPool      pool[*bufio.Reader]
	writerPool      pool[*bufio.Writer]
	limitReaderPool pool[*io.LimitedReader]

	proxyAuth string

	targetURL             *url.URL
	passHostHeader        bool
	responseHeaderTimeout time.Duration
}

// NewReverseProxy creates a new ReverseProxy.
func NewReverseProxy(targetURL *url.URL, proxyURL *url.URL, debug, passHostHeader bool, responseHeaderTimeout time.Duration, connPool *connPool) (*ReverseProxy, error) {
	var proxyAuth string
	if proxyURL != nil && proxyURL.User != nil && targetURL.Scheme == "http" {
		username := proxyURL.User.Username()
		password, _ := proxyURL.User.Password()
		proxyAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	}

	return &ReverseProxy{
		debug:                 debug,
		passHostHeader:        passHostHeader,
		targetURL:             targetURL,
		proxyAuth:             proxyAuth,
		connPool:              connPool,
		responseHeaderTimeout: responseHeaderTimeout,
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
		if reqUpType == "websocket" {
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

		wd := &writeDetector{Conn: co}

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

	br := p.readerPool.Get()
	if br == nil {
		br = bufio.NewReaderSize(co, bufioSize)
	}
	defer p.readerPool.Put(br)

	br.Reset(co)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	res.Header.SetNoDefaultContentType(true)

	for {
		var timer *time.Timer
		errTimeout := atomic.Pointer[timeoutError]{}
		if p.responseHeaderTimeout > 0 {
			timer = time.AfterFunc(p.responseHeaderTimeout, func() {
				errTimeout.Store(&timeoutError{errors.New("timeout awaiting response headers")})
				co.Close()
			})
		}

		res.Header.SetNoDefaultContentType(true)
		if err := res.Header.Read(br); err != nil {
			if p.responseHeaderTimeout > 0 {
				if errT := errTimeout.Load(); errT != nil {
					return errT
				}
			}
			co.Close()
			return err
		}

		if timer != nil {
			timer.Stop()
		}

		fixPragmaCacheControl(&res.Header)

		resCode := res.StatusCode()
		is1xx := 100 <= resCode && resCode <= 199
		// treat 101 as a terminal status, see issue 26161
		is1xxNonTerminal := is1xx && resCode != http.StatusSwitchingProtocols
		if is1xxNonTerminal {
			removeConnectionHeaders(&res.Header)
			h := rw.Header()

			for _, header := range hopHeaders {
				res.Header.Del(header)
			}

			res.Header.VisitAll(func(key, value []byte) {
				rw.Header().Add(string(key), string(value))
			})

			rw.WriteHeader(res.StatusCode())
			// Clear headers, it's not automatically done by ResponseWriter.WriteHeader() for 1xx responses
			for k := range h {
				delete(h, k)
			}

			res.Reset()
			res.Header.Reset()
			res.Header.SetNoDefaultContentType(true)

			continue
		}
		break
	}

	announcedTrailers := res.Header.Peek("Trailer")

	// Deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if res.StatusCode() == http.StatusSwitchingProtocols {
		// As the connection has been hijacked, it cannot be added back to the pool.
		handleUpgradeResponse(rw, req, reqUpType, res, buffConn{Conn: co, Reader: br})
		return nil
	}

	removeConnectionHeaders(&res.Header)

	for _, header := range hopHeaders {
		res.Header.Del(header)
	}

	if len(announcedTrailers) > 0 {
		res.Header.Add("Trailer", string(announcedTrailers))
	}

	res.Header.VisitAll(func(key, value []byte) {
		rw.Header().Add(string(key), string(value))
	})

	rw.WriteHeader(res.StatusCode())

	// Chunked response, Content-Length is set to -1 by FastProxy when "Transfer-Encoding: chunked" header is received.
	if res.Header.ContentLength() == -1 {
		cbr := httputil.NewChunkedReader(br)

		b := p.bufferPool.Get()
		if b == nil {
			b = make([]byte, bufferSize)
		}
		defer p.bufferPool.Put(b)

		if _, err := io.CopyBuffer(&writeFlusher{rw}, cbr, b); err != nil {
			co.Close()
			return err
		}

		res.Header.Reset()
		res.Header.SetNoDefaultContentType(true)
		if err := res.Header.ReadTrailer(br); err != nil {
			co.Close()
			return err
		}

		if res.Header.Len() > 0 {
			var announcedTrailersKey []string
			if len(announcedTrailers) > 0 {
				announcedTrailersKey = strings.Split(string(announcedTrailers), ",")
			}

			res.Header.VisitAll(func(key, value []byte) {
				for _, s := range announcedTrailersKey {
					if strings.EqualFold(s, strings.TrimSpace(string(key))) {
						rw.Header().Add(string(key), string(value))
						return
					}
				}

				rw.Header().Add(http.TrailerPrefix+string(key), string(value))
			})
		}

		p.connPool.ReleaseConn(co)

		return nil
	}

	brl := p.limitReaderPool.Get()
	if brl == nil {
		brl = &io.LimitedReader{}
	}
	defer p.limitReaderPool.Put(brl)

	brl.R = br
	brl.N = int64(res.Header.ContentLength())

	b := p.bufferPool.Get()
	if b == nil {
		b = make([]byte, bufferSize)
	}
	defer p.bufferPool.Put(b)

	if _, err := io.CopyBuffer(rw, brl, b); err != nil {
		co.Close()
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
	SetBytesV(key string, value []byte)
	DelBytes(key []byte)
	Del(key string)
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
	headers.SetBytesV("Sec-WebSocket-Key", headers.Peek("Sec-Websocket-Key"))
	headers.Del("Sec-Websocket-Key")

	headers.SetBytesV("Sec-WebSocket-Extensions", headers.Peek("Sec-Websocket-Extensions"))
	headers.Del("Sec-Websocket-Extensions")

	headers.SetBytesV("Sec-WebSocket-Accept", headers.Peek("Sec-Websocket-Accept"))
	headers.Del("Sec-Websocket-Accept")

	headers.SetBytesV("Sec-WebSocket-Protocol", headers.Peek("Sec-Websocket-Protocol"))
	headers.Del("Sec-Websocket-Protocol")

	headers.SetBytesV("Sec-WebSocket-Version", headers.Peek("Sec-Websocket-Version"))
	headers.DelBytes([]byte("Sec-Websocket-Version"))
}
