package headers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestHeaders_requestIdAdded(t *testing.T) {
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	var requestID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID = r.Header.Get("X-Request-ID")
	})
	h := NewHeaders()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req, next)
	assert.True(t, strings.HasPrefix(requestID, "s"))
	id, _ := uuid.FromString(strings.TrimPrefix(requestID, "s"))
	assert.Equal(t, uint(0x4), id.Version())
}
