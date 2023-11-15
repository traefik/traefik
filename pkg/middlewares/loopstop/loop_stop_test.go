package loopstop

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHTTP(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "", nil)
	require.NoError(t, err)

	m, err := NewLoopStop("myRouter", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	require.Equal(t, m.routerNameHash, req.Header.Get(xTraefikRouter))

	recorder = httptest.NewRecorder()
	m.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
