// +build !go1.11

package forward

import (
	"context"
	"net/http"
)

type key string

const (
	teHeader key = "TeHeader"
)

type TeTrailerRoundTripper struct {
	http.RoundTripper
}

func (t *TeTrailerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	teHeader := req.Context().Value(teHeader)
	if teHeader != nil {
		req.Header.Set("Te", teHeader.(string))
	}
	return t.RoundTripper.RoundTrip(req)
}

type TeTrailerRewriter struct {
	ReqRewriter
}

func (t *TeTrailerRewriter) Rewrite(req *http.Request) {
	if req.Header.Get("Te") == "trailers" {
		*req = *req.WithContext(context.WithValue(req.Context(), teHeader, req.Header.Get("Te")))
	}
	t.ReqRewriter.Rewrite(req)
}

func (f *Forwarder) postConfig() {
	f.roundTripper = &TeTrailerRoundTripper{RoundTripper: f.roundTripper}
	f.rewriter = &TeTrailerRewriter{ReqRewriter: f.rewriter}
}
