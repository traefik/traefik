package ready

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_NotReady(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusServiceUnavailable, rw.Code)
	assert.Equal(t, "Service Unavailable", rw.Body.String())
}

func TestHandler_Ready(t *testing.T) {
	handler := NewHandler()
	handler.SetReady()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "OK", rw.Body.String())
}
