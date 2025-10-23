package integration

import (
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

// RoutingSuite tests multi-layer routing with authentication middleware.
type RoutingSuite struct{ BaseSuite }

func TestRoutingSuite(t *testing.T) {
	suite.Run(t, new(RoutingSuite))
}

func (s *RoutingSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("routing")
	s.composeUp()
}

func (s *RoutingSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

// authHandler implements the ForwardAuth protocol.
// It validates Bearer tokens and adds X-User-Role and X-User-Name headers.
func authHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	role, username, ok := getUserByToken(token)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Set headers that will be forwarded by Traefik
	w.Header().Set("X-User-Role", role)
	w.Header().Set("X-User-Name", username)
	w.WriteHeader(http.StatusOK)
}

// getUserByToken returns the role and username for a given token.
func getUserByToken(token string) (role, username string, ok bool) {
	users := map[string]struct {
		role     string
		username string
	}{
		"bob-token":   {role: "admin", username: "bob"},
		"jack-token":  {role: "developer", username: "jack"},
		"alice-token": {role: "guest", username: "alice"},
	}

	u, exists := users[token]
	return u.role, u.username, exists
}

// TestMultiLayerRoutingWithAuth tests the complete multi layer routing scenario:
// - Parent router matches path and applies authentication middleware
// - Auth middleware validates token and adds role header
// - Child routers route based on the role header added by the middleware
func (s *RoutingSuite) TestMultiLayerRoutingWithAuth() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err)
	defer listener.Close()

	_, authPort, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(s.T(), err)

	go func() {
		_ = http.Serve(listener, http.HandlerFunc(authHandler))
	}()

	adminIP := s.getComposeServiceIP("whoami-admin")
	require.NotEmpty(s.T(), adminIP)

	developerIP := s.getComposeServiceIP("whoami-developer")
	require.NotEmpty(s.T(), developerIP)

	file := s.adaptFile("fixtures/routing/multi_layer_auth.toml", struct {
		AuthPort    string
		AdminIP     string
		DeveloperIP string
	}{
		AuthPort:    authPort,
		AdminIP:     adminIP,
		DeveloperIP: developerIP,
	})

	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("parent-router"))
	require.NoError(s.T(), err)

	// Test 1: bob (admin role) routes to admin-service
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer bob-token")

	err = try.Request(req, 2*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("whoami-admin"))
	require.NoError(s.T(), err)

	// Test 2: jack (developer role) routes to developer-service
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer jack-token")

	err = try.Request(req, 2*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("whoami-developer"))
	require.NoError(s.T(), err)

	// Test 3: Invalid token returns 401
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer invalid-token")

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	// Test 4: Missing token returns 401
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	require.NoError(s.T(), err)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	// Test 5: Valid auth but role has no matching child router returns 404
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer alice-token")

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}
