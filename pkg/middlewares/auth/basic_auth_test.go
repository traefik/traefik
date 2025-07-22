package auth

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestBasicAuthFail(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.BasicAuth{
		Users: []string{"test"},
	}
	_, err := NewBasic(t.Context(), next, auth, "authName")
	require.Error(t, err)

	auth2 := dynamic.BasicAuth{
		Users: []string{"test:test"},
	}
	authMiddleware, err := NewBasic(t.Context(), next, auth2, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "they should be equal")
}

func TestBasicAuthSuccess(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.BasicAuth{
		Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
	}
	authMiddleware, err := NewBasic(t.Context(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}

func TestBasicAuthUserHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test", r.Header["X-Webauth-User"][0], "auth user should be set")
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.BasicAuth{
		Users:       []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
		HeaderField: "X-Webauth-User",
	}
	middleware, err := NewBasic(t.Context(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, "traefik\n", string(body))
}

func TestBasicAuthHeaderRemoved(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get(authorizationHeader))
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.BasicAuth{
		RemoveHeader: true,
		Users:        []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
	}
	middleware, err := NewBasic(t.Context(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, "traefik\n", string(body))
}

func TestBasicAuthHeaderPresent(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get(authorizationHeader))
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.BasicAuth{
		Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
	}
	middleware, err := NewBasic(t.Context(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, "traefik\n", string(body))
}

func TestBasicAuthConcurrentHashOnce(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	auth := dynamic.BasicAuth{
		Users: []string{"test:$2a$04$.8sTYfcxbSplCtoxt5TdJOgpBYkarKtZYsYfYxQ1edbYRuO1DNi0e"},
	}

	authMiddleware, err := NewBasic(t.Context(), next, auth, "authName")
	require.NoError(t, err)

	hashCount := 0
	ba := authMiddleware.(*basicAuth)
	ba.checkSecret = func(password, secret string) bool {
		hashCount++
		// delay to ensure the second request arrives
		time.Sleep(time.Millisecond)
		return true
	}

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	for range 2 {
		go func() {
			defer wg.Done()
			req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
			req.SetBasicAuth("test", "test")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")
		}()
	}

	wg.Wait()
	assert.Equal(t, 1, hashCount)
}

func TestBasicAuthUsersFromFile(t *testing.T) {
	testCases := []struct {
		desc            string
		userFileContent string
		expectedUsers   map[string]string
		givenUsers      []string
		realm           string
	}{
		{
			desc:            "Finds the users in the file",
			userFileContent: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/\ntest2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0\n",
			givenUsers:      []string{},
			expectedUsers:   map[string]string{"test": "test", "test2": "test2"},
		},
		{
			desc:            "Merges given users with users from the file",
			userFileContent: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/\n",
			givenUsers:      []string{"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "test3:$apr1$3rJbDP0q$RfzJiorTk78jQ1EcKqWso0"},
			expectedUsers:   map[string]string{"test": "test", "test2": "test2", "test3": "test3"},
		},
		{
			desc:            "Given users have priority over users in the file",
			userFileContent: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/\ntest2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0\n",
			givenUsers:      []string{"test2:$apr1$mK.GtItK$ncnLYvNLek0weXdxo68690"},
			expectedUsers:   map[string]string{"test": "test", "test2": "overridden"},
		},
		{
			desc:            "Should authenticate the correct user based on the realm",
			userFileContent: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/\ntest2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0\n",
			givenUsers:      []string{"test2:$apr1$mK.GtItK$ncnLYvNLek0weXdxo68690"},
			expectedUsers:   map[string]string{"test": "test", "test2": "overridden"},
			realm:           "traefik",
		},
		{
			desc:            "Should skip comments",
			userFileContent: "#Comment\ntest:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/\ntest2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0\n",
			givenUsers:      []string{},
			expectedUsers:   map[string]string{"test": "test", "test2": "test2"},
			realm:           "traefiker",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// Creates the temporary configuration file with the users
			usersFile, err := os.CreateTemp(t.TempDir(), "auth-users")
			require.NoError(t, err)

			_, err = usersFile.WriteString(test.userFileContent)
			require.NoError(t, err)

			// Creates the configuration for our Authenticator
			authenticatorConfiguration := dynamic.BasicAuth{
				Users:     test.givenUsers,
				UsersFile: usersFile.Name(),
				Realm:     test.realm,
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "traefik")
			})

			authenticator, err := NewBasic(t.Context(), next, authenticatorConfiguration, "authName")
			require.NoError(t, err)

			ts := httptest.NewServer(authenticator)
			defer ts.Close()

			for userName, userPwd := range test.expectedUsers {
				req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
				req.SetBasicAuth(userName, userPwd)

				var res *http.Response
				res, err = http.DefaultClient.Do(req)
				require.NoError(t, err)

				require.Equal(t, http.StatusOK, res.StatusCode, "Cannot authenticate user "+userName)

				var body []byte
				body, err = io.ReadAll(res.Body)
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)

				require.Equal(t, "traefik\n", string(body))
			}

			// Checks that user foo doesn't work
			req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
			req.SetBasicAuth("foo", "foo")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusUnauthorized, res.StatusCode)
			if len(test.realm) > 0 {
				require.Equal(t, `Basic realm="`+test.realm+`"`, res.Header.Get("WWW-Authenticate"))
			}

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			require.NotContains(t, "traefik", string(body))
		})
	}
}
