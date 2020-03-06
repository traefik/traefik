package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnonAuthFail(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.AnonAuth{
		Users: []string{`\i`},
	}
	_, err := NewAnon(context.Background(), next, auth, "authName")
	require.Error(t, err)

	auth2 := dynamic.AnonAuth{
		Users: []string{"testtest"},
	}
	authMiddleware, err := NewAnon(context.Background(), next, auth2, "authTest")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "they should be equal")
}

func TestAnonAuthSuccess(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.AnonAuth{
		Users: []string{"test.*"},
	}
	authMiddleware, err := NewAnon(context.Background(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("testtest", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, "traefik\n", string(body), "they should be equal")
}

func TestAnonAuthUserHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test", r.Header["X-Webauth-User"][0], "auth user should be set")
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.AnonAuth{
		Users:       []string{"test"},
		HeaderField: "X-Webauth-User",
	}
	middleware, err := NewAnon(context.Background(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, "traefik\n", string(body))
}

func TestAnonAuthHeaderRemoved(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get(authorizationHeader))
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.AnonAuth{
		RemoveHeader: true,
	}
	middleware, err := NewAnon(context.Background(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, "traefik\n", string(body))
}

func TestAnonAuthHeaderPresent(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get(authorizationHeader))
		fmt.Fprintln(w, "traefik")
	})

	auth := dynamic.AnonAuth{}

	middleware, err := NewAnon(context.Background(), next, auth, "authName")
	require.NoError(t, err)

	ts := httptest.NewServer(middleware)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("test", "test")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, "traefik\n", string(body))
}

func TestAnonAuthUsersFromFile(t *testing.T) {
	testCases := []struct {
		desc            string
		userFileContent string
		expectedUsers   map[string]string
		givenUsers      []string
		realm           string
	}{
		{
			desc:            "Finds the users in the file",
			userFileContent: "test\ntest2\n",
			givenUsers:      []string{},
			expectedUsers:   map[string]string{"test": "test", "test2": "test2"},
		},
		{
			desc:            "Merges given users with users from the file",
			userFileContent: "test\n",
			givenUsers:      []string{"test2", "test3"},
			expectedUsers:   map[string]string{"test": "test", "test2": "test2", "test3": "test3"},
		},
		{
			desc:            "Should skip comments",
			userFileContent: "#foo\ntest\ntest2",
			givenUsers:      []string{},
			expectedUsers:   map[string]string{"test": "test", "test2": "test2"},
			realm:           "traefiker",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// Creates the temporary configuration file with the users
			usersFile, err := ioutil.TempFile("", "auth-users")
			require.NoError(t, err)
			defer os.Remove(usersFile.Name())

			_, err = usersFile.Write([]byte(test.userFileContent))
			require.NoError(t, err)

			// Creates the configuration for our Authenticator
			authenticatorConfiguration := dynamic.AnonAuth{
				Users:     test.givenUsers,
				UsersFile: usersFile.Name(),
				Realm:     test.realm,
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "traefik")
			})

			authenticator, err := NewAnon(context.Background(), next, authenticatorConfiguration, "authName")
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
				body, err = ioutil.ReadAll(res.Body)
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

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			require.NotContains(t, "traefik", string(body))
		})
	}
}
