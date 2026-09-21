package auth

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestNewBasicEmpty(t *testing.T) {
	auth := dynamic.BasicAuth{
		Users: []string{},
	}

	_, err := NewBasic(t.Context(), nil, auth, "authName")
	require.Error(t, err)
}

func TestNewBasicNotFoundSecretIsSet(t *testing.T) {
	auth := dynamic.BasicAuth{
		Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
	}
	middleware, err := NewBasic(t.Context(), nil, auth, "authName")
	require.NoError(t, err)

	ba := middleware.(*basicAuth)
	assert.Equal(t, "$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", ba.notFoundSecret)
}

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

func TestBasicAuthUserHeaderCanonical(t *testing.T) {
	var nextCalled bool
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		nextCalled = true
		assert.Empty(t, req.Header.Get("X-User"))
		assert.Equal(t, []string{"test"}, req.Header["x-user"])
	})
	auth := dynamic.BasicAuth{
		Users:       []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"},
		HeaderField: "x-user",
	}
	m, err := NewBasic(t.Context(), next, auth, "test")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.SetBasicAuth("test", "test")
	req.Header.Set("X-User", "admin")
	rw := httptest.NewRecorder()
	m.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Result().StatusCode)
	assert.True(t, nextCalled)
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

	var hashCount atomic.Int64
	ba := authMiddleware.(*basicAuth)
	ba.checkSecret = func(password, secret string) bool {
		hashCount.Add(1)
		// delay to ensure the second request arrives
		time.Sleep(50 * time.Millisecond)
		return true
	}

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	var wg sync.WaitGroup

	for range 2 {
		wg.Go(func() {
			req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
			req.SetBasicAuth("test", "test")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode, "they should be equal")
		})
	}

	wg.Wait()
	assert.Equal(t, int64(1), hashCount.Load())
}

func TestBasicAuthConcurrentNoUserEnumeration(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	auth := dynamic.BasicAuth{
		Users: []string{"test:$2a$04$.8sTYfcxbSplCtoxt5TdJOgpBYkarKtZYsYfYxQ1edbYRuO1DNi0e"},
	}

	authMiddleware, err := NewBasic(t.Context(), next, auth, "authName")
	require.NoError(t, err)

	var hashCount atomic.Int64
	proceed := make(chan struct{})

	// Hold the computations in flight, to give the other request the opportunity to join the singleflight call.
	ba := authMiddleware.(*basicAuth)
	ba.checkSecret = func(password, secret string) bool {
		hashCount.Add(1)
		<-proceed

		return false
	}

	ts := httptest.NewServer(authMiddleware)
	t.Cleanup(ts.Close)

	// The release must also run on an early exit, or the handlers held in flight would deadlock ts.Close.
	release := sync.OnceFunc(func() { close(proceed) })
	t.Cleanup(release)

	var wg sync.WaitGroup

	for _, user := range []string{"unknown1", "unknown2"} {
		wg.Go(func() {
			req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
			req.SetBasicAuth(user, "test")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		})
	}

	// A deduplicated request never reaches checkSecret, leaving the condition unsatisfied.
	assert.Eventually(t, func() bool { return hashCount.Load() == 2 }, 5*time.Second, 10*time.Millisecond)

	release()
	wg.Wait()
}

// TestBasicAuthConcurrentNoKeyCollision ensures an unconfigured user cannot inherit
// a configured user's successful authentication through singleflight deduplication
// (GHSA-6765-c87h-8mrf).
func TestBasicAuthConcurrentNoKeyCollision(t *testing.T) {
	var forwardedUser string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardedUser = r.Header.Get("X-WebAuth-User")
		fmt.Fprintln(w, "traefik")
	})

	const hash = "$2a$04$.8sTYfcxbSplCtoxt5TdJOgpBYkarKtZYsYfYxQ1edbYRuO1DNi0e"
	auth := dynamic.BasicAuth{
		Users:       []string{"viewer:" + hash},
		HeaderField: "X-WebAuth-User",
	}

	authMiddleware, err := NewBasic(t.Context(), next, auth, "authName")
	require.NoError(t, err)

	// Validate only the configured pair, and stay in flight long enough for the
	// attacker request to join a colliding singleflight call.
	ba := authMiddleware.(*basicAuth)
	ba.checkSecret = func(password, secret string) bool {
		time.Sleep(50 * time.Millisecond)
		return secret == hash && password == "test"
	}

	ts := httptest.NewServer(authMiddleware)
	defer ts.Close()

	// The attacker's crafted request must fail on its own.
	attackerReq := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	attackerReq.SetBasicAuth("admin", "test"+hash)
	baseline, err := http.DefaultClient.Do(attackerReq)
	require.NoError(t, err)
	baseline.Body.Close()
	require.Equal(t, http.StatusUnauthorized, baseline.StatusCode)

	// Fire a valid request, then let the attacker join it concurrently.
	var wg sync.WaitGroup
	wg.Go(func() {
		req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
		req.SetBasicAuth("viewer", "test")
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	time.Sleep(10 * time.Millisecond)

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.SetBasicAuth("admin", "test"+hash)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	res.Body.Close()

	wg.Wait()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.NotEqual(t, "admin", forwardedUser)
}

// Test_singleflightKey pins the two properties the key must hold: identical credentials collapse to
// one call, and distinct pairs never collide. The colon cases cover boundaries no client can submit,
// since http.Request.BasicAuth cuts on the first colon.
func Test_singleflightKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc          string
		user          string
		password      string
		otherUser     string
		otherPassword string
		equal         bool
	}{
		{
			desc:          "identical credentials share a call",
			user:          "viewer",
			password:      "test",
			otherUser:     "viewer",
			otherPassword: "test",
			equal:         true,
		},
		{
			desc:          "colon moved across the user boundary",
			user:          "viewer",
			password:      "admin:test",
			otherUser:     "viewer:admin",
			otherPassword: "test",
		},
		{
			desc:          "user boundary shifted into the password",
			user:          "ab",
			password:      "cd",
			otherUser:     "a",
			otherPassword: "bcd",
		},
		{
			desc:          "empty user and empty password swapped",
			user:          "",
			password:      "viewer",
			otherUser:     "viewer",
			otherPassword: "",
		},
		{
			desc:          "password embedding the stored hash (GHSA-6765-c87h-8mrf)",
			user:          "admin",
			password:      "test$2a$04$.8sTYfcxbSplCtoxt5TdJOgpBYkarKtZYsYfYxQ1edbYRuO1DNi0e",
			otherUser:     "viewer",
			otherPassword: "test",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			key := singleflightKey(test.user, test.password)
			otherKey := singleflightKey(test.otherUser, test.otherPassword)

			if test.equal {
				assert.Equal(t, key, otherKey)
				return
			}

			assert.NotEqual(t, key, otherKey)
		})
	}
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
