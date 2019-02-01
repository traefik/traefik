package middlewares

// Middleware tests based on https://github.com/unrolled/secure

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
	"gitlab.com/JanMa/correlation"
)

// newCorrelation constructs a new correlation instance with supplied options.
func newCorrelation(options ...correlation.Options) *correlation.Correlation {
	var opt correlation.Options
	if len(options) == 0 {
		opt = correlation.Options{}
	} else {
		opt = options[0]
	}

	return correlation.New(opt)
}

func TestCorrelationNoConfig(t *testing.T) {
	cor := newCorrelation()

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "http://example.com/foo", nil)

	cor.Handler(myHandler).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "bar", res.Body.String(), "Body not the expected")
	assert.NotEqual(t, res.Header().Get("X-Correlation-ID"), "", "Correlation ID Header not present")
}

func TestCorrelationHeaderAlreadyPresent(t *testing.T) {
	cor := newCorrelation(correlation.Options{
		CorrelationHeaderName: "foo",
	})
	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "http://example.com/foo", nil)
	req.Header.Add("foo", "baz")

	cor.Handler(myHandler).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "bar", res.Body.String(), "Body not the expected")
	assert.Equal(t, res.Header().Get("foo"), "baz", "Wrong Correlation ID header")
}
