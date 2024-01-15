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

	next := 0
	m, err := New("myRouter", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		next++
	}))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "995d26092d19a224", m.routerNameHash)
	assert.Equal(t, m.routerNameHash, req.Header.Get(xTraefikRouter))
	assert.Equal(t, 1, next)

	recorder = httptest.NewRecorder()
	m.ServeHTTP(recorder, req)

	assert.Equal(t, 1, next)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
