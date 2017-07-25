package route53

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// MockResponse represents a predefined response used by a mock server
type MockResponse struct {
	StatusCode int
	Body       string
}

// MockResponseMap maps request paths to responses
type MockResponseMap map[string]MockResponse

func newMockServer(t *testing.T, responses MockResponseMap) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		resp, ok := responses[path]
		if !ok {
			msg := fmt.Sprintf("Requested path not found in response map: %s", path)
			require.FailNow(t, msg)
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(resp.StatusCode)
		w.Write([]byte(resp.Body))
	}))

	time.Sleep(100 * time.Millisecond)
	return ts
}
