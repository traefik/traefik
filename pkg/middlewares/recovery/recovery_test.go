package recovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverHandler(t *testing.T) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		panic("I love panicing!")
	}
	recovery, err := New(context.Background(), http.HandlerFunc(fn), "foo-recovery")
	require.NoError(t, err)

	server := httptest.NewServer(recovery)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
