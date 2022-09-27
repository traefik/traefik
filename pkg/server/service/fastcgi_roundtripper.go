package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/yookoala/gofast"
)

type FastCgiRoundTripper struct {
	client  gofast.Client
	handler gofast.SessionHandler
}

func NewFastCgiRoundTripper(addr, filename string) (*FastCgiRoundTripper, error) {
	connFactory := gofast.SimpleConnFactory("tcp", addr)

	client, err := gofast.SimpleClientFactory(connFactory)()
	if err != nil {
		return nil, fmt.Errorf("FastCGI connection failed: %w", err)
	}

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
		client:  client,
		handler: chain(gofast.BasicSession),
	}, nil
}

func (rt *FastCgiRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() {
		if rt.client == nil {
			return
		}

		// TODO: when close ?
		//if err := rt.client.Close(); err != nil {
		//		log.WithoutContext().Errorf("gofast: error closing client: %s", err.Error())
		//}
	}()

	resp, err := rt.handler(rt.client, gofast.NewRequest(req))
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
