package middlewares

import (
	"context"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestSSLForceHost(t *testing.T) {
	testCases := []struct {
		desc             string
		host             string
		secureMiddleware *SecureConfig
		expected         int
	}{
		{
			desc: "http should return a 301",
			host: "http://powpow.example.com",
			secureMiddleware: NewSecure(&types.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "http sub domain should return a 301",
			host: "http://www.powpow.example.com",
			secureMiddleware: NewSecure(&types.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "https should return a 200",
			host: "https://powpow.example.com",
			secureMiddleware: NewSecure(&types.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusOK,
		},
		{
			desc: "https sub domain should return a 301",
			host: "https://www.powpow.example.com",
			secureMiddleware: NewSecure(&types.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "http without force host and sub domain should return a 301",
			host: "http://www.powpow.example.com",
			secureMiddleware: NewSecure(&types.Headers{
				SSLRedirect:  true,
				SSLForceHost: false,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusMovedPermanently,
		},
		{
			desc: "https without force host and sub domain should return a 301",
			host: "https://www.powpow.example.com",
			secureMiddleware: NewSecure(&types.Headers{
				SSLRedirect:  true,
				SSLForceHost: false,
				SSLHost:      "powpow.example.com",
			}),
			expected: http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := testhelpers.MustNewRequest(http.MethodGet, test.host, nil)
			next := func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("OK"))
			}

			rw := httptest.NewRecorder()
			test.secureMiddleware.HandlerFuncWithNextForRequestOnlyWithContextCheck(rw, req, next)

			assert.Equal(t, test.expected, rw.Result().StatusCode)
		})
	}
}

func TestSSLRedirectWithModifiedRequest(t *testing.T) {
	testCases := []struct {
		desc          string
		addPrefix     bool
		replacePrefix bool
		stripPrefix   bool
		url           string
		key           string
		expected      string
	}{
		{
			desc:        "StripPrefix",
			stripPrefix: true,
			url:         "http://powpow.example.com/foo",
			key:         "/bacon/foo",
			expected:    "/bacon/foo",
		},
		{
			desc:      "AddPrefix",
			addPrefix: true,
			url:       "http://powpow.example.com/bacon/foo",
			key:       "/bacon",
			expected:  "/foo",
		},
		{
			desc:          "ReplacePrefix",
			replacePrefix: true,
			url:           "http://powpow.example.com/foo",
			key:           "/bacon/foo",
			expected:      "/bacon/foo",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			secureMiddleware := NewSecure(&types.Headers{
				SSLRedirect:  true,
				SSLForceHost: true,
				SSLHost:      "powpow.example.com",
			})

			next := func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("OK"))
			}
			req := testhelpers.MustNewRequest(http.MethodGet, test.url, nil)
			switch {
			case test.stripPrefix:
				req = req.WithContext(context.WithValue(req.Context(), StripPrefixKey, test.key))
			case test.addPrefix:
				req = req.WithContext(context.WithValue(req.Context(), AddPrefixKey, test.key))
			case test.replacePrefix:
				req = req.WithContext(context.WithValue(req.Context(), ReplacePathKey, test.key))
			}
			req.RequestURI = req.URL.RequestURI()
			rw := httptest.NewRecorder()
			secureMiddleware.HandlerFuncWithNextForRequestOnlyWithContextCheck(rw, req, next)
			returnedLocation, err := rw.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, test.expected, returnedLocation.Path)
		})
	}
}
