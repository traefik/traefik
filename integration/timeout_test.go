package integration

import (
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

func (s *TimeoutSuite) TestRouterRespondingTimeouts() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/router_responding_timeouts.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/roundTrip`)"))
	require.NoError(s.T(), err)

	// Check that timeout service is available.
	statusURL := fmt.Sprintf("http://%s/statusTest?status=200",
		net.JoinHostPort(timeoutEndpointIP, "9000"))
	require.NoError(s.T(), try.GetRequest(statusURL, 60*time.Second, try.StatusCodeIs(http.StatusOK)))

	// A single idle connection per host forces the requests below to reuse the same keep-alive connection,
	// so the "no leaked deadline" assertion actually exercises connection reuse.
	client := &http.Client{Transport: &http.Transport{MaxIdleConns: 1, MaxIdleConnsPerHost: 1}}

	// A backend answering after the router roundTrip timeout yields a 504.
	response, err := client.Get("http://127.0.0.1:8000/roundTrip?sleep=2000")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)

	// Drain the response to allow the keep-alive connection reuse.
	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())

	// A backend answering below the router roundTrip timeout is not affected,
	// even on the same keep-alive connection (no leaked connection deadline).
	response, err = client.Get("http://127.0.0.1:8000/roundTrip?sleep=100")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)
}

// TestRouterRespondingTimeoutsSupersedeEntryPointWriteTimeout covers a router roundTrip on an entrypoint that
// has its own writeTimeout: the router deadline supersedes it for the request, and the entrypoint one is
// restored before the response is flushed, which happens after the router handler has returned.
func (s *TimeoutSuite) TestRouterRespondingTimeoutsSupersedeEntryPointWriteTimeout() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/router_responding_timeouts_write_timeout.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/roundTrip`)", "Path(`/noRoundTrip`)"))
	require.NoError(s.T(), err)

	// Check that timeout service is available.
	statusURL := fmt.Sprintf("http://%s/statusTest?status=200",
		net.JoinHostPort(timeoutEndpointIP, "9000"))
	require.NoError(s.T(), try.GetRequest(statusURL, 60*time.Second, try.StatusCodeIs(http.StatusOK)))

	// A single idle connection per host forces the requests below to reuse the same keep-alive connection,
	// so the write-deadline restore is exercised across a real connection reuse.
	client := &http.Client{Transport: &http.Transport{MaxIdleConns: 1, MaxIdleConnsPerHost: 1}}

	// The entrypoint writeTimeout is in force: without a router roundTrip, a backend answering after it
	// leaves the response unwritable and the connection is closed with no response at all.
	_, err = client.Get("http://127.0.0.1:8000/noRoundTrip?sleep=2000")
	require.Error(s.T(), err)

	// The router roundTrip supersedes it: the very same backend now answers.
	response, err := client.Get("http://127.0.0.1:8000/roundTrip?sleep=2000")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	// Drain the response to allow the keep-alive connection reuse.
	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())

	// Past the roundTrip the 504 still reaches the client: the response is flushed once the router handler
	// has returned, under a write deadline restored from the entrypoint writeTimeout rather than an expired one.
	response, err = client.Get("http://127.0.0.1:8000/roundTrip?sleep=4000")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)

	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())

	// The next request on the same keep-alive connection is unaffected: the server re-arms the entrypoint
	// write deadline per request.
	response, err = client.Get("http://127.0.0.1:8000/roundTrip?sleep=100")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)
}

// TestRouterRespondingTimeoutsWebSocket covers the upgrade path of the router roundTrip: an established
// WebSocket tunnel is never severed at the deadline (the handshake timer is disarmed at the protocol switch),
// while a hung upgrade handshake still yields a 504 (pre-101 the gateway has not responded).
func (s *TimeoutSuite) TestRouterRespondingTimeoutsWebSocket() {
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

	// A backend that never answers the handshake: it holds the request until the proxy tears the leg down.
	hang := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	s.T().Cleanup(hang.Close)

	file := s.adaptFile("fixtures/timeout/router_responding_timeouts_websocket.toml", struct {
		WebsocketServer string
		HangServer      string
	}{
		WebsocketServer: echo.URL,
		HangServer:      hang.URL,
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/ws`)", "Path(`/ws-hang`)"))
	require.NoError(s.T(), err)

	// An established tunnel survives well past the 1s roundTrip: the timer is disarmed at the protocol switch.
	conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", nil)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { _ = conn.Close() })

	time.Sleep(2 * time.Second)

	require.NoError(s.T(), conn.WriteMessage(gorillawebsocket.TextMessage, []byte("after-deadline")))

	_, msg, err := conn.ReadMessage()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "after-deadline", string(msg))

	// A hung upgrade handshake yields a 504 once the roundTrip elapses: pre-101 the gateway has not responded.
	_, resp, err := gorillawebsocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws-hang", nil)
	require.Error(s.T(), err)
	require.NotNil(s.T(), resp)
	assert.Equal(s.T(), http.StatusGatewayTimeout, resp.StatusCode)
}

func (s *TimeoutSuite) TestRouterRespondingTimeoutsHTTP3() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/router_responding_timeouts_http3.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/roundTrip`)"))
	require.NoError(s.T(), err)

	transport := &http3.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: "snitest.com"}}
	defer func() { _ = transport.Close() }()

	// Wait for the HTTP/3 entry point to be ready with a below-timeout request.
	fastReq, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/roundTrip?sleep=100", nil)
	require.NoError(s.T(), err)
	fastReq.Host = "snitest.com"
	require.NoError(s.T(), try.RequestWithTransport(fastReq, 60*time.Second, transport, try.StatusCodeIs(http.StatusOK)))

	client := &http.Client{Transport: transport}

	// A backend answering after the router roundTrip timeout yields a 504 over HTTP/3.
	slowReq, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/roundTrip?sleep=2000", nil)
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
	fastReq2, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/roundTrip?sleep=100", nil)
	require.NoError(s.T(), err)
	fastReq2.Host = "snitest.com"

	response, err = client.Do(fastReq2)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())
}
