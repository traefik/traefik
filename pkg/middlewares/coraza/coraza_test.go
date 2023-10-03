package coraza

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestNewCoraza(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	directives := `
	SecDebugLogLevel 9
	SecRuleEngine On
	SecRule REQUEST_URI "@streq /admin" "id:101,phase:1,t:lowercase,log,deny"
`

	conf := dynamic.Coraza{
		Directives: directives,
	}
	handler, err := NewCoraza(context.Background(), next, conf, "coraza")
	require.NoError(t, err)
	assert.NotNil(t, handler, "this should not be nil")

	type testCase struct {
		name       string
		statusCode int
		path       string
	}
	testCases := []testCase{
		{
			name:       "request should pass because / is allowed",
			statusCode: 200,
			path:       "/",
		},
		{
			name:       "request should not pass because its using /admin which is denied",
			statusCode: 403,
			path:       "/admin",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			handler.ServeHTTP(w, req)
			body := w.Body.String()
			require.Equal(t, tc.statusCode, w.Code, body)
		})
	}
}

func TestCorazaCRS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	directives := `
	SecDefaultAction "phase:1,log,auditlog,deny,status:403"
	SecDefaultAction "phase:2,log,auditlog,deny,status:403"

	SecAction \
    	"id:900110,\
    	phase:1,\
    	pass,\
    	t:none,\
    	nolog,\
    	setvar:tx.inbound_anomaly_score_threshold=5,\
    	setvar:tx.outbound_anomaly_score_threshold=4"

	SecAction \
	    "id:900200,\
	    phase:1,\
	    pass,\
	    t:none,\
	    nolog,\
	    setvar:'tx.allowed_methods=GET HEAD POST OPTIONS'"

	Include @owasp_crs/REQUEST-911-METHOD-ENFORCEMENT.conf
	Include @owasp_crs/REQUEST-949-BLOCKING-EVALUATION.conf
`

	conf := dynamic.Coraza{
		Directives: directives,
		CRSEnabled: true,
	}
	handler, err := NewCoraza(context.Background(), next, conf, "coraza")
	require.NoError(t, err)
	assert.NotNil(t, handler, "this should not be nil")

	type testCase struct {
		name       string
		statusCode int
		path       string
		method     string
	}
	testCases := []testCase{
		{
			name:       "request should pass because its using GET which is in allowed_methods",
			statusCode: 200,
			path:       "/",
			method:     http.MethodGet,
		},
		{
			name:       "request should not pass because its using PUT which is not in allowed_methods",
			statusCode: 403,
			path:       "/",
			method:     http.MethodPut,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			handler.ServeHTTP(w, req)
			body := w.Body.String()
			require.Equal(t, tc.statusCode, w.Code, body)
		})
	}
}
