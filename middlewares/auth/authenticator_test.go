package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
)

func TestAuthUsersFromFile(t *testing.T) {
	tests := []struct {
		authType   string
		usersStr   string
		userKeys   []string
		parserFunc func(fileName string) (map[string]string, error)
	}{
		{
			authType: "basic",
			usersStr: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/\ntest2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0\n",
			userKeys: []string{"test", "test2"},
			parserFunc: func(fileName string) (map[string]string, error) {
				basic := &types.Basic{
					UsersFile: fileName,
				}
				return parserBasicUsers(basic)
			},
		},
		{
			authType: "digest",
			usersStr: "test:traefik:a2688e031edb4be6a3797f3882655c05 \ntest2:traefik:518845800f9e2bfb1f1f740ec24f074e\n",
			userKeys: []string{"test:traefik", "test2:traefik"},
			parserFunc: func(fileName string) (map[string]string, error) {
				digest := &types.Digest{
					UsersFile: fileName,
				}
				return parserDigestUsers(digest)
			},
		},
		{
			authType: "basic",
			usersStr: "#Comment\ntest:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/\ntest2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0\n",
			userKeys: []string{"test", "test2"},
			parserFunc: func(fileName string) (map[string]string, error) {
				basic := &types.Basic{
					UsersFile: fileName,
				}
				return parserBasicUsers(basic)
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.authType, func(t *testing.T) {
			t.Parallel()
			usersFile, err := ioutil.TempFile("", "auth-users")
			require.NoError(t, err)
			defer os.Remove(usersFile.Name())

			_, err = usersFile.Write([]byte(test.usersStr))
			require.NoError(t, err)

			users, err := test.parserFunc(usersFile.Name())
			require.NoError(t, err)
			assert.Equal(t, 2, len(users), "they should be equal")

			_, ok := users[test.userKeys[0]]
			assert.True(t, ok, "user test should be found")
			_, ok = users[test.userKeys[1]]
			assert.True(t, ok, "user test2 should be found")
		})
	}
}

func TestBasicAuthFail(t *testing.T) {
	_, err := NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			Users: []string{"test"},
		},
	}, &tracing.Tracing{})
	assert.Contains(t, err.Error(), "error parsing Authenticator user", "should contains")

	authMiddleware, err := NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			Users: []string{"test:test"},
		},
	}, &tracing.Tracing{})
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "they should be equal")
}

func TestBasicAuthSuccess(t *testing.T) {
	authMiddleware, err := NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
		},
	}, &tracing.Tracing{})
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}

func TestDigestAuthFail(t *testing.T) {
	_, err := NewAuthenticator(&types.Auth{
		Digest: &types.Digest{
			Users: []string{"test"},
		},
	}, &tracing.Tracing{})
	assert.Contains(t, err.Error(), "error parsing Authenticator user", "should contains")

	authMiddleware, err := NewAuthenticator(&types.Auth{
		Digest: &types.Digest{
			Users: []string{"test:traefik:test"},
		},
	}, &tracing.Tracing{})
	require.NoError(t, err)
	assert.NotNil(t, authMiddleware, "this should not be nil")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(authMiddleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "they should be equal")
}

func TestBasicAuthUserHeader(t *testing.T) {
	middleware, err := NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
		},
		HeaderField: "X-Webauth-User",
	}, &tracing.Tracing{})
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test", r.Header["X-Webauth-User"][0], "auth user should be set")
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(middleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}

func TestBasicAuthHeaderRemoved(t *testing.T) {
	middleware, err := NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			RemoveHeader: true,
			Users:        []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
		},
	}, &tracing.Tracing{})
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get(authorizationHeader))
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(middleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}

func TestBasicAuthHeaderPresent(t *testing.T) {
	middleware, err := NewAuthenticator(&types.Auth{
		Basic: &types.Basic{
			Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
		},
	}, &tracing.Tracing{})
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get(authorizationHeader))
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New(middleware)
	n.UseHandler(handler)
	ts := httptest.NewServer(n)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")
	res, err := client.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}
