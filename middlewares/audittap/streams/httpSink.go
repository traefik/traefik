package streams

import (
	"bytes"
	"fmt"
	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"net/http"
	"net/url"
)

type httpSink struct {
	method, endpoint string
	render           Renderer
}

// NewHTTPSink creates a new HTTP sink
func NewHTTPSink(method, endpoint string, renderer Renderer) (AuditSink, error) {
	if method == "" {
		method = http.MethodPost
	}
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("Cannot access endpoint '%s': %v", endpoint, err)
	}
	return &httpSink{method, endpoint, renderer}, nil
}

func (has *httpSink) Audit(encoded audittypes.Encoded) error {
	request, err := http.NewRequest(has.method, has.endpoint, bytes.NewBuffer(encoded.Bytes))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Length", fmt.Sprintf("%d", encoded.Length()))

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	return res.Body.Close()
}

func (has *httpSink) Close() error {
	return nil
}
