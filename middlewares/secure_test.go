package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/unrolled/secure"
)

func TestSSLForceHost(t *testing.T) {
	testCases := []struct {
		desc             string
		host             string
		secureMiddleware *secure.Secure
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
			test.secureMiddleware.HandlerFuncWithNextForRequestOnly(rw, req, next)

			assert.Equal(t, test.expected, rw.Result().StatusCode)
		})
	}
}
