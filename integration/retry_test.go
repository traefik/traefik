package integration

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type RetrySuite struct {
	BaseSuite
	whoamiIP string
}

func TestRetrySuite(t *testing.T) {
	suite.Run(t, new(RetrySuite))
}

func (s *RetrySuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("retry")
	s.composeUp()

	s.whoamiIP = s.getComposeServiceIP("whoami")
}

func (s *RetrySuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *RetrySuite) TestRetry() {
	file := s.adaptFile("fixtures/retry/simple.toml", struct{ WhoamiIP string }{s.whoamiIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("PathPrefix(`/`)"))
	require.NoError(s.T(), err)

	response, err := http.Get("http://127.0.0.1:8000/")
	require.NoError(s.T(), err)

	// The test only verifies that the retry middleware makes sure that the working service is eventually reached.
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func (s *RetrySuite) TestRetryBackoff() {
	file := s.adaptFile("fixtures/retry/backoff.toml", struct{ WhoamiIP string }{s.whoamiIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("PathPrefix(`/`)"))
	require.NoError(s.T(), err)

	response, err := http.Get("http://127.0.0.1:8000/")
	require.NoError(s.T(), err)

	// The test only verifies that the retry middleware allows finally to reach the working service.
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func (s *RetrySuite) TestRetryWebsocket() {
	file := s.adaptFile("fixtures/retry/simple.toml", struct{ WhoamiIP string }{s.whoamiIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("PathPrefix(`/`)"))
	require.NoError(s.T(), err)

	// The test only verifies that the retry middleware makes sure that the working service is eventually reached.
	_, response, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/echo", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusSwitchingProtocols, response.StatusCode)

	// The test verifies a second time that the working service is eventually reached.
	_, response, err = websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/echo", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusSwitchingProtocols, response.StatusCode)
}

func (s *RetrySuite) TestRetryWithStripPrefix() {
	file := s.adaptFile("fixtures/retry/strip_prefix.toml", struct{ WhoamiIP string }{s.whoamiIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("PathPrefix(`/`)"))
	require.NoError(s.T(), err)

	response, err := http.Get("http://127.0.0.1:8000/test")
	require.NoError(s.T(), err)

	body, err := io.ReadAll(response.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "GET / HTTP/1.1")
	assert.Contains(s.T(), string(body), "X-Forwarded-Prefix: /test")
}
