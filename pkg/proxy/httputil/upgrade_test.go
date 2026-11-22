package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http/httpguts"
)

const http2Settings = "AAMAAABkAAQAoAAAAAIAAAAA"

func TestH2CUpgradeNotForwarded(t *testing.T) {
	testCases := []struct {
		desc            string
		requestHeaders  http.Header
		expectedHeaders http.Header
	}{
		{
			desc: "h2c upgrade with HTTP2-Settings listed in Connection",
			requestHeaders: http.Header{
				"Connection":     {"Upgrade, HTTP2-Settings"},
				"Upgrade":        {"h2c"},
				"Http2-Settings": {http2Settings},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c upgrade with HTTP2-Settings smuggled out of Connection",
			requestHeaders: http.Header{
				"Connection":     {"Upgrade"},
				"Upgrade":        {"h2c"},
				"Http2-Settings": {http2Settings},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c upgrade without HTTP2-Settings",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"h2c"},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c upgrade with an uppercase token",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"H2C"},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c upgrade combined with a websocket upgrade",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket, h2c"},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "HTTP2-Settings without an upgrade",
			requestHeaders: http.Header{
				"Http2-Settings": {http2Settings},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "websocket upgrade is preserved",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket"},
			},
			expectedHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket"},
			},
		},
		// Only the first Upgrade value is inspected, as the upstream fix does, hence the two outcomes below.
		// Both are equivalent in the end: the reverse proxy replaces the Upgrade values it forwards with the first
		// one, so a later h2c value never reaches the backend either.
		{
			desc: "h2c as the first Upgrade value drops every upgrade",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"h2c", "websocket"},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c as a later Upgrade value keeps the first one",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket", "h2c"},
			},
			expectedHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				forwarded := http.Header{}
				for _, name := range []string{"Connection", "Upgrade", "Http2-Settings"} {
					if values := req.Header.Values(name); len(values) > 0 {
						forwarded[name] = values
					}
				}
				assert.Equal(t, test.expectedHeaders, forwarded)

				// The backend switches to h2c on any Upgrade: h2c, without the validations RFC 7540 section 3.2.1 requires.
				if !httpguts.HeaderValuesContainsToken([]string{req.Header.Get("Upgrade")}, "h2c") {
					return
				}

				rw.Header().Set("Connection", "Upgrade")
				rw.Header().Set("Upgrade", "h2c")
				rw.WriteHeader(http.StatusSwitchingProtocols)
			}))
			t.Cleanup(backend.Close)

			proxy := createProxyWithForwarder(t, backend.URL, http.DefaultTransport)

			req, err := http.NewRequest(http.MethodGet, proxy.URL+"/public", http.NoBody)
			require.NoError(t, err)

			req.Header = test.requestHeaders

			resp, err := proxy.Client().Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
