package middlewares

import (
	"net/http"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
)

func TestNewAddRequestIdWithRequestID(t *testing.T) {
	expected := uuid.NewV4().String()

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Set(RequestID, expected)
	rw := httptest.NewRecorder()

	handler := &AddRequestID{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}

	handler.ServeHTTP(rw, req, nil)

	assert.Equal(t, expected, req.Header.Get(RequestID), "Unexpected X-Request-Id.")
	assert.Equal(t, expected, rw.HeaderMap.Get(RequestID), "Unexpected X-Request-Id.")
}

func TestNewAddRequestIdWithoutRequestID(t *testing.T) {
	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	rw := httptest.NewRecorder()

	handler := &AddRequestID{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}

	handler.ServeHTTP(rw, req, nil)

	assert.NotEmpty(t, req.Header.Get(RequestID), "Unexpected X-Request-Id.")
	assert.NotEmpty(t, rw.HeaderMap.Get(RequestID), "Unexpected X-Request-Id.")
}
