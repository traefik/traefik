package responsemodifiers

import (
	"net/http"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares/headers"
)

func buildHeaders(hdrs *config.Headers) func(*http.Response) error {
	return func(resp *http.Response) error {
		if hdrs.HasCustomHeadersDefined() || hdrs.HasCorsHeadersDefined() {
			err := headers.NewHeader(nil, *hdrs).ModifyResponseHeaders(resp)
			if err != nil {
				return err
			}
		}

		if hdrs.HasSecureHeadersDefined() {
			err := headers.NewSecure(nil, *hdrs).ModifyResponseHeaders(resp)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
