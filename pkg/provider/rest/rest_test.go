package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestServeHTTPRejectsOversizedBody(t *testing.T) {
	provider := &Provider{configurationChan: make(chan dynamic.Message, 1)}
	body := strings.NewReader(`{"x":"` + strings.Repeat("a", maxBodySize+1) + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/providers/rest", body)
	req = mux.SetURLVars(req, map[string]string{"provider": "rest"})
	rw := httptest.NewRecorder()

	provider.ServeHTTP(rw, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, rw.Code)
	require.Empty(t, provider.configurationChan)
}

func TestServeHTTPRejectsOversizedTrailingBody(t *testing.T) {
	provider := &Provider{configurationChan: make(chan dynamic.Message, 1)}
	body := strings.NewReader(`{}` + strings.Repeat(" ", maxBodySize+1))
	req := httptest.NewRequest(http.MethodPut, "/api/providers/rest", body)
	req = mux.SetURLVars(req, map[string]string{"provider": "rest"})
	rw := httptest.NewRecorder()

	provider.ServeHTTP(rw, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, rw.Code)
	require.Empty(t, provider.configurationChan)
}
