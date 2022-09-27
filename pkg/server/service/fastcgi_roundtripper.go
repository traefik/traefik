package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/yookoala/gofast"
)

type FastCgiRoundTripper struct {
	client  *gofast.Client
	handler gofast.SessionHandler
}

func makeClient(proto, addr string) (gofast.Client, error) {
	connFactory := gofast.SimpleConnFactory(proto, addr)
	return gofast.SimpleClientFactory(connFactory)()
}

func NewFastCgiRoundTripper(filename string) (*FastCgiRoundTripper, error) {
	chain := gofast.Chain(
		gofast.BasicParamsMap,
		gofast.MapHeader,
		gofast.MapEndpoint(filename),
		func(handler gofast.SessionHandler) gofast.SessionHandler {
			return func(client gofast.Client, req *gofast.Request) (*gofast.ResponsePipe, error) {
				req.Params["HTTP_HOST"] = req.Params["SERVER_NAME"]
				req.Params["SERVER_SOFTWARE"] = "Traefik"

				// Gofast sets this param to `fastcgi` which is not what the backend will expect.
				delete(req.Params, "REQUEST_SCHEME")

				return handler(client, req)
			}
		},
	)

	return &FastCgiRoundTripper{
		client:  nil,
		handler: chain(gofast.BasicSession),
	}, nil
}

func (rt *FastCgiRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() {
		// TODO: Close the client only when the server shuts down.
		if rt.client != nil {
			if err := (*rt.client).Close(); err != nil {
				log.WithoutContext().Errorf("gofast: error closing client: %s", err.Error())
			}
			rt.client = nil
		}
	}()

	if rt.client == nil {
		client, err := makeClient("tcp", req.URL.Host)
		if err != nil {
			return nil, fmt.Errorf("FastCGI connection failed: %w", err)
		}

		rt.client = &client
	}

	resp, err := rt.handler(*rt.client, gofast.NewRequest(req))
	if err != nil {
		return nil, fmt.Errorf("failed to process request: %w", err)
	}

	rr := httptest.NewRecorder()

	errBuffer := new(bytes.Buffer)
	resp.WriteTo(rr, errBuffer)

	if errBuffer.Len() > 0 {
		if strings.Contains(errBuffer.String(), "Primary script unknown") {
			body := http.StatusText(http.StatusNotFound)
			return &http.Response{
				Status:        body,
				StatusCode:    http.StatusNotFound,
				Body:          io.NopCloser(bytes.NewBufferString(body)),
				ContentLength: int64(len(body)),
				Request:       req,
				Header:        make(http.Header),
			}, nil
		} else {
			return nil, fmt.Errorf("FastCGI application error %s", errBuffer.String())
		}
	}

	return rr.Result(), nil
}
