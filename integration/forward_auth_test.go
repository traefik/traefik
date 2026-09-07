package integration

import (
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

// ForwardAuthSuite tests the ForwardAuth service option, which routes the authentication
// request through a Traefik service instead of dialing a hardcoded address.
type ForwardAuthSuite struct{ BaseSuite }

func TestForwardAuthSuite(t *testing.T) {
	suite.Run(t, new(ForwardAuthSuite))
}

func (s *ForwardAuthSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("forward_auth")
	s.composeUp()
}

func (s *ForwardAuthSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *ForwardAuthSuite) TestForwardAuthWithService() {
	auth1Port, auth1Count := s.authServer("auth-1")
	auth2Port, auth2Count := s.authServer("auth-2")

	whoamiIP := s.getComposeServiceIP("whoami")
	require.NotEmpty(s.T(), whoamiIP)

	file := s.adaptFile("fixtures/forward_auth/service.toml", struct {
		Auth1Port string
		Auth2Port string
		WhoamiIP  string
	}{
		Auth1Port: auth1Port,
		Auth2Port: auth2Port,
		WhoamiIP:  whoamiIP,
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.BodyContains("protected-router"))
	require.NoError(s.T(), err)

	s.Run("allows a valid token and forwards the auth response headers", func() {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/protected", nil)
		require.NoError(s.T(), err)
		req.Header.Set("Authorization", "Bearer valid-token")

		err = try.Request(req, 5*time.Second,
			try.StatusCodeIs(http.StatusOK),
			try.BodyContains("X-Auth-User: bob"))
		require.NoError(s.T(), err)
	})

	s.Run("denies a request without a token", func() {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/protected", nil)
		require.NoError(s.T(), err)

		err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusUnauthorized))
		require.NoError(s.T(), err)
	})

	s.Run("forwards a relative Location header unchanged", func() {
		// The default client follows redirects, which would hide the Location header under way.
		client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}}

		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/protected", nil)
		require.NoError(s.T(), err)
		req.Header.Set("Authorization", "Bearer redirect-token")

		res, err := client.Do(req)
		require.NoError(s.T(), err)
		require.NoError(s.T(), res.Body.Close())

		assert.Equal(s.T(), http.StatusFound, res.StatusCode)

		// A relative Location must not be resolved against the http placeholder URL,
		// which would downgrade an HTTPS request to plaintext.
		assert.Equal(s.T(), "/login?rd=%2Fprotected", res.Header.Get("Location"))
	})

	s.Run("load balances the authentication request across the service servers", func() {
		before1, before2 := auth1Count.Load(), auth2Count.Load()

		for range 6 {
			req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/protected", nil)
			require.NoError(s.T(), err)
			req.Header.Set("Authorization", "Bearer valid-token")

			err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK))
			require.NoError(s.T(), err)
		}

		got1, got2 := auth1Count.Load()-before1, auth2Count.Load()-before2

		assert.Positive(s.T(), got1, "auth-1 served no authentication request")
		assert.Positive(s.T(), got2, "auth-2 served no authentication request")
		assert.Equal(s.T(), int64(6), got1+got2)
	})
}

// authServer starts an authentication server on a random port, and reports the port it
// listens on along with a counter of the requests it served.
func (s *ForwardAuthSuite) authServer(name string) (string, *atomic.Int64) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { _ = listener.Close() })

	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(s.T(), err)

	var count atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)

		// The path option must be honored.
		if r.URL.Path != "/verify" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("X-Auth-Backend", name)

		switch strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") {
		case "valid-token":
			w.Header().Set("X-Auth-User", "bob")
			w.WriteHeader(http.StatusOK)
		case "redirect-token":
			w.Header().Set("Location", "/login?rd=%2Fprotected")
			w.WriteHeader(http.StatusFound)
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	go func() {
		_ = http.Serve(listener, handler)
	}()

	return port, &count
}
