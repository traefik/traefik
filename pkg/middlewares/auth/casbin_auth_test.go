package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

func TestCasbinAuthFail(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.CasbinAuth{
		ModelPath:  "./rbac_with_pattern_model.conf",
		PolicyPath: "./examples/rbac_with_pattern_policy.csv",
	}

	_, err := NewCasbin(context.Background(), next, auth, "authName")

	require.Error(t, err)

	auth2 := dynamic.CasbinAuth{
		ModelPath:  "./examples/rbac_with_pattern_model.conf",
		PolicyPath: "./examples/rbac_with_pattern_policy.csv",
	}
	authMiddleware, err := NewCasbin(context.Background(), next, auth2, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, fmt.Sprintf("%s/pen/2", ts.URL), nil)
	req.Header.Add(CasbinAuthHeader, "alice")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "they should be equal")
}

func TestCasbinAuthSuccess(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.CasbinAuth{
		ModelPath:  "./examples/rbac_with_pattern_model.conf",
		PolicyPath: "./examples/rbac_with_pattern_policy.csv",
	}
	authMiddleware, err := NewCasbin(context.Background(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, fmt.Sprintf("%s/pen/1", ts.URL), nil)
	req.Header.Add(CasbinAuthHeader, "alice")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}
