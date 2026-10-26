package integration

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type TimeoutSuite struct{ BaseSuite }

func TestTimeoutSuite(t *testing.T) {
	suite.Run(t, new(TimeoutSuite))
}

func (s *TimeoutSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("timeout")
	s.composeUp()
}

func (s *TimeoutSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TimeoutSuite) TestForwardingTimeouts() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/forwarding_timeouts.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/dialTimeout`)"))
	require.NoError(s.T(), err)

	// This simulates a DialTimeout when connecting to the backend server.
	response, err := http.Get("http://127.0.0.1:8000/dialTimeout")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)

	// Check that timeout service is available
	statusURL := fmt.Sprintf("http://%s/statusTest?status=200",
		net.JoinHostPort(timeoutEndpointIP, "9000"))
	assert.NoError(s.T(), try.GetRequest(statusURL, 60*time.Second, try.StatusCodeIs(http.StatusOK)))

	// This simulates a ResponseHeaderTimeout.
	response, err = http.Get("http://127.0.0.1:8000/responseHeaderTimeout?sleep=1000")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)
}

// TestRouterRespondingTimeoutsWriteTimeout covers a router writeTimeout on an entrypoint that
// has its own writeTimeout: the router deadline supersedes it for the request.
func (s *TimeoutSuite) TestRouterRespondingTimeoutsWriteTimeout() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/router_responding_timeouts_write_timeout.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/writeTimeout`)", "Path(`/noWriteTimeout`)"))
	require.NoError(s.T(), err)

	// Check that timeout service is available.
	statusURL := fmt.Sprintf("http://%s/statusTest?status=200",
		net.JoinHostPort(timeoutEndpointIP, "9000"))
	require.NoError(s.T(), try.GetRequest(statusURL, 60*time.Second, try.StatusCodeIs(http.StatusOK)))

	// A single idle connection per host forces the requests below to reuse the same keep-alive connection,
	// so the write-deadline restore is exercised across a real connection reuse.
	client := &http.Client{Transport: &http.Transport{MaxIdleConns: 1, MaxIdleConnsPerHost: 1}}

	// The entrypoint writeTimeout is in force: without a router writeTimeout, a backend answering after it
	// leaves the response unwritable and the connection is closed with no response at all.
	_, err = client.Get("http://127.0.0.1:8000/noWriteTimeout?sleep=2000")
	require.Error(s.T(), err)

	// The router writeTimeout supersedes it: the very same backend now answers.
	response, err := client.Get("http://127.0.0.1:8000/writeTimeout?sleep=2000")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	// Drain the response to allow the keep-alive connection reuse.
	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())

	// Past the writeTimeout the 504 still reaches the client: it is written under a grace period
	// rather than the expired deadline.
	response, err = client.Get("http://127.0.0.1:8000/writeTimeout?sleep=4000")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)

	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())

	// The next request on the same keep-alive connection is unaffected: the server re-arms the entrypoint
	// write deadline per request.
	response, err = client.Get("http://127.0.0.1:8000/writeTimeout?sleep=100")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)
}

// TestRouterRespondingTimeoutsReadTimeout covers a router readTimeout on an entrypoint that has its own
// readTimeout: the router one supersedes it for the request body, in both directions.
func (s *TimeoutSuite) TestRouterRespondingTimeoutsReadTimeout() {
	// The backend answers once it has read the whole body, so that only the upload is timed.
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if _, err := io.Copy(io.Discard, req.Body); err != nil {
			return
		}

		rw.WriteHeader(http.StatusOK)
	}))
	s.T().Cleanup(backend.Close)

	file := s.adaptFile("fixtures/timeout/router_responding_timeouts_read_timeout.toml", struct{ BackendURL string }{backend.URL})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/readTimeout`)", "Path(`/noReadTimeout`)"))
	require.NoError(s.T(), err)

	// uploadSlowly sends the first half of the body, then the second half after the pause, unless it is negative.
	uploadSlowly := func(path string, pause time.Duration) (*http.Response, error) {
		conn, err := net.Dial("tcp", "127.0.0.1:8000")
		require.NoError(s.T(), err)
		s.T().Cleanup(func() { _ = conn.Close() })

		require.NoError(s.T(), conn.SetDeadline(time.Now().Add(10*time.Second)))

		_, err = fmt.Fprintf(conn, "POST %s HTTP/1.1\r\nHost: 127.0.0.1\r\nContent-Length: 4\r\n\r\nab", path)
		require.NoError(s.T(), err)

		if pause >= 0 {
			time.Sleep(pause)

			// The server may already have given up on the request.
			_, _ = fmt.Fprint(conn, "cd")
		}

		return http.ReadResponse(bufio.NewReader(conn), nil)
	}

	// The entrypoint readTimeout is in force: without a router readTimeout, an upload slower than it fails.
	response, err := uploadSlowly("/noReadTimeout", 2*time.Second)
	if err == nil {
		assert.NotEqual(s.T(), http.StatusOK, response.StatusCode)
	}

	// The router readTimeout supersedes it: the very same upload now succeeds.
	response, err = uploadSlowly("/readTimeout", 2*time.Second)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	// An upload stalling for good is still bounded, by the router readTimeout.
	start := time.Now()
	response, err = uploadSlowly("/readTimeout", -1)
	require.NoError(s.T(), err)
	assert.NotEqual(s.T(), http.StatusOK, response.StatusCode)
	assert.Less(s.T(), time.Since(start), 6*time.Second)
}

// TestRouterRespondingTimeoutsWriteTimeoutWebSocket covers the upgrade requests, which the router writeTimeout leaves alone:
// an established WebSocket tunnel is never severed at the deadline.
func (s *TimeoutSuite) TestRouterRespondingTimeoutsWriteTimeoutWebSocket() {
	upgrader := gorillawebsocket.Upgrader{}

	// A healthy echo backend: it completes the handshake and relays messages both ways.
	echo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				return
			}
			if err = c.WriteMessage(mt, message); err != nil {
				return
			}
		}
	}))
	s.T().Cleanup(echo.Close)

	file := s.adaptFile("fixtures/timeout/router_responding_timeouts_write_timeout_websocket.toml", struct{ WebsocketServer string }{echo.URL})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/ws`)"))
	require.NoError(s.T(), err)

	// An established tunnel survives well past the 1s writeTimeout.
	conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", nil)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { _ = conn.Close() })

	time.Sleep(2 * time.Second)

	require.NoError(s.T(), conn.WriteMessage(gorillawebsocket.TextMessage, []byte("after-deadline")))

	_, msg, err := conn.ReadMessage()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "after-deadline", string(msg))
}

func (s *TimeoutSuite) TestRouterRespondingTimeoutsWriteTimeoutHTTP3() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/router_responding_timeouts_write_timeout_http3.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/writeTimeout`)"))
	require.NoError(s.T(), err)

	transport := &http3.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: "snitest.com"}}
	defer func() { _ = transport.Close() }()

	// Wait for the HTTP/3 entry point to be ready with a below-timeout request.
	fastReq, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/writeTimeout?sleep=100", nil)
	require.NoError(s.T(), err)
	fastReq.Host = "snitest.com"
	require.NoError(s.T(), try.RequestWithTransport(fastReq, 60*time.Second, transport, try.StatusCodeIs(http.StatusOK)))

	client := &http.Client{Transport: transport}

	// A backend answering after the router writeTimeout yields a 504 over HTTP/3.
	slowReq, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/writeTimeout?sleep=2000", nil)
	require.NoError(s.T(), err)
	slowReq.Host = "snitest.com"

	response, err := client.Do(slowReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)

	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())

	// A subsequent below-timeout request on the same HTTP/3 connection is unaffected,
	// as the deadline is per-stream and cannot leak to another stream.
	fastReq2, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/writeTimeout?sleep=100", nil)
	require.NoError(s.T(), err)
	fastReq2.Host = "snitest.com"

	response, err = client.Do(fastReq2)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())
}
