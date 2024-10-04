package fast

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/http/httpguts"
)

// switchProtocolCopier exists so goroutines proxying data back and
// forth have nice names in stacks.
type switchProtocolCopier struct {
	user, backend io.ReadWriter
}

func (c switchProtocolCopier) copyFromBackend(errc chan<- error) {
	_, err := io.Copy(c.user, c.backend)
	errc <- err
}

func (c switchProtocolCopier) copyToBackend(errc chan<- error) {
	_, err := io.Copy(c.backend, c.user)
	errc <- err
}

func handleUpgradeResponse(rw http.ResponseWriter, req *http.Request, reqUpType string, res *fasthttp.Response, backConn net.Conn) {
	defer backConn.Close()

	resUpType := upgradeTypeFastHTTP(&res.Header)

	if !strings.EqualFold(reqUpType, resUpType) {
		httputil.ErrorHandler(rw, req, fmt.Errorf("backend tried to switch protocol %q when %q was requested", resUpType, reqUpType))
		return
	}

	hj, ok := rw.(http.Hijacker)
	if !ok {
		httputil.ErrorHandler(rw, req, fmt.Errorf("can't switch protocols using non-Hijacker ResponseWriter type %T", rw))
		return
	}
	backConnCloseCh := make(chan bool)
	go func() {
		// Ensure that the cancellation of a request closes the backend.
		// See issue https://golang.org/issue/35559.
		select {
		case <-req.Context().Done():
		case <-backConnCloseCh:
		}
		_ = backConn.Close()
	}()

	defer close(backConnCloseCh)

	conn, brw, err := hj.Hijack()
	if err != nil {
		httputil.ErrorHandler(rw, req, fmt.Errorf("hijack failed on protocol switch: %w", err))
		return
	}
	defer conn.Close()

	for k, values := range rw.Header() {
		for _, v := range values {
			res.Header.Add(k, v)
		}
	}

	if err := res.Header.Write(brw.Writer); err != nil {
		httputil.ErrorHandler(rw, req, fmt.Errorf("response write: %w", err))
		return
	}

	if err := brw.Flush(); err != nil {
		httputil.ErrorHandler(rw, req, fmt.Errorf("response flush: %w", err))
		return
	}

	errc := make(chan error, 1)
	spc := switchProtocolCopier{user: conn, backend: backConn}
	go spc.copyToBackend(errc)
	go spc.copyFromBackend(errc)
	<-errc
}

func upgradeType(h http.Header) string {
	if !httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade") {
		return ""
	}

	return h.Get("Upgrade")
}

func upgradeTypeFastHTTP(h fasthttpHeader) string {
	if !bytes.Contains(h.Peek("Connection"), []byte("Upgrade")) {
		return ""
	}

	return string(h.Peek("Upgrade"))
}
