package headers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

// Middleware tests based on https://github.com/unrolled/secure

func Test_newSecure_modifyResponse(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      dynamic.Headers
		expected http.Header
	}{
		{
			desc: "PermissionsPolicy",
			cfg: dynamic.Headers{
				PermissionsPolicy: "microphone=(),",
			},
			expected: http.Header{"Permissions-Policy": []string{"microphone=(),"}},
		},
		{
			desc: "STSSeconds",
			cfg: dynamic.Headers{
				STSSeconds:     1,
				ForceSTSHeader: true,
			},
			expected: http.Header{"Strict-Transport-Security": []string{"max-age=1"}},
		},
		{
			desc: "STSSeconds and STSPreload",
			cfg: dynamic.Headers{
				STSSeconds:     1,
				ForceSTSHeader: true,
				STSPreload:     true,
			},
			expected: http.Header{"Strict-Transport-Security": []string{"max-age=1; preload"}},
		},
		{
			desc: "CustomFrameOptionsValue",
			cfg: dynamic.Headers{
				CustomFrameOptionsValue: "foo",
			},
			expected: http.Header{"X-Frame-Options": []string{"foo"}},
		},
		{
			desc: "FrameDeny",
			cfg: dynamic.Headers{
				FrameDeny: true,
			},
			expected: http.Header{"X-Frame-Options": []string{"DENY"}},
		},
		{
			desc: "ContentTypeNosniff",
			cfg: dynamic.Headers{
				ContentTypeNosniff: true,
			},
			expected: http.Header{"X-Content-Type-Options": []string{"nosniff"}},
		},
	}

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			secure := newSecure(emptyHandler, test.cfg, "mymiddleware")

			req := httptest.NewRequest(http.MethodGet, "/foo", nil)

			rw := httptest.NewRecorder()

			secure.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Result().Header)
		})
	}
}
