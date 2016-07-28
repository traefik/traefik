package middlewares

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuthFail(t *testing.T) {
	authMiddleware, err := NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			Users: []string{"test"},
		},
	})
	assert.Contains(t, err.Error(), "Error parsing Authenticator user", "should contains")

	authMiddleware, err = NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			Users: []string{"test:test"},
		},
	})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "they should be equal")
}

func TestBasicAuthSuccess(t *testing.T) {
	authMiddleware, err := NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
		},
	})
	assert.NoError(t, err, "there should be no error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}

func TestDigestAuthFail(t *testing.T) {
	authMiddleware, err := NewAuthenticator(&types.Auth{
		Digest: &types.Digest{
			Users: []string{"test"},
		},
	})
	assert.Contains(t, err.Error(), "Error parsing Authenticator user", "should contains")

	authMiddleware, err = NewAuthenticator(&types.Auth{
		Digest: &types.Digest{
			Users: []string{"test:traefik:test"},
		},
	})
	assert.NoError(t, err, "there should be no error")
	assert.NotNil(t, authMiddleware, "this should not be nil")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "they should be equal")
}
