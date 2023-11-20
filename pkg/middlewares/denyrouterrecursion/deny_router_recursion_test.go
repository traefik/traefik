package denyrouterrecursion

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

	_, err = New("", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	require.Error(t, err)

	next := false
	m, err := New("myRouter", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		next = true
	}))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	assert.Equal(t, m.routerNameHash, req.Header.Get(xTraefikRouter))

	assert.True(t, next)

	recorder = httptest.NewRecorder()
	m.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
