package fast

import (
	"context"
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

func (c switchProtocolCopier) copyFromBackend(errCh chan<- error) {
	_, err := io.Copy(c.user, c.backend)
	errCh <- err
}

func (c switchProtocolCopier) copyToBackend(errCh chan<- error) {
	_, err := io.Copy(c.backend, c.user)
	errCh <- err
}

type upgradeHandler func(rw http.ResponseWriter, res *fasthttp.Response, backConn net.Conn)

func upgradeResponseHandler(ctx context.Context, reqUpType string) upgradeHandler {
	return func(rw http.ResponseWriter, res *fasthttp.Response, backConn net.Conn) {
		resUpType := upgradeTypeFastHTTP(&res.Header)

		if !strings.EqualFold(reqUpType, resUpType) {
			httputil.ErrorHandlerWithContext(ctx, rw, fmt.Errorf("backend tried to switch protocol %q when %q was requested", resUpType, reqUpType))
			backConn.Close()
			return
		}

		hj, ok := rw.(http.Hijacker)
		if !ok {
			httputil.ErrorHandlerWithContext(ctx, rw, fmt.Errorf("can't switch protocols using non-Hijacker ResponseWriter type %T", rw))
			backConn.Close()
			return
		}
		backConnCloseCh := make(chan bool)
		go func() {
			// Ensure that the cancellation of a request closes the backend.
			// See issue https://golang.org/issue/35559.
			select {
			case <-ctx.Done():
			case <-backConnCloseCh:
			}
			_ = backConn.Close()
		}()
		defer close(backConnCloseCh)

		conn, brw, err := hj.Hijack()
		if err != nil {
			httputil.ErrorHandlerWithContext(ctx, rw, fmt.Errorf("hijack failed on protocol switch: %w", err))
			return
		}
		defer conn.Close()

		for k, values := range rw.Header() {
			for _, v := range values {
				res.Header.Add(k, v)
			}
		}

		if err := res.Header.Write(brw.Writer); err != nil {
			httputil.ErrorHandlerWithContext(ctx, rw, fmt.Errorf("response write: %w", err))
			return
		}

		if err := brw.Flush(); err != nil {
			httputil.ErrorHandlerWithContext(ctx, rw, fmt.Errorf("response flush: %w", err))
			return
		}

		errCh := make(chan error, 1)
		spc := switchProtocolCopier{user: conn, backend: backConn}
		go spc.copyToBackend(errCh)
		go spc.copyFromBackend(errCh)
		<-errCh
	}
}

func upgradeType(h http.Header) string {
	if !httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade") {
		return ""
	}

	return h.Get("Upgrade")
}

func upgradeTypeFastHTTP(h fasthttpHeader) string {
	if !h.ConnectionUpgrade() {
		return ""
	}

	return string(h.Peek("Upgrade"))
}
