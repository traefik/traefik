package integration

import (
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type KeepAliveSuite struct {
	BaseSuite
}

func TestKeepAliveSuite(t *testing.T) {
	suite.Run(t, new(KeepAliveSuite))
}

type KeepAliveConfig struct {
	KeepAliveServer string
	IdleConnTimeout string
}

type connStateChangeEvent struct {
	key   string
	state http.ConnState
}

func (s *KeepAliveSuite) TestShouldRespectConfiguredBackendHttpKeepAliveTime() {
	idleTimeout := time.Duration(75) * time.Millisecond

	connStateChanges := make(chan connStateChangeEvent)
	noMoreRequests := make(chan bool, 1)
	completed := make(chan bool, 1)

	// keep track of HTTP connections and their status changes and measure their idle period
	go func() {
		connCount := 0
		idlePeriodStartMap := make(map[string]time.Time)
		idlePeriodLengthMap := make(map[string]time.Duration)

		maxWaitDuration := 5 * time.Second
		maxWaitTimeExceeded := time.After(maxWaitDuration)
		moreRequestsExpected := true

		// Ensure that all idle HTTP connections are closed before verification phase
		for moreRequestsExpected || len(idlePeriodLengthMap) < connCount {
			select {
			case event := <-connStateChanges:
				switch event.state {
				case http.StateNew:
					connCount++
				case http.StateIdle:
					idlePeriodStartMap[event.key] = time.Now()
				case http.StateClosed:
					idlePeriodLengthMap[event.key] = time.Since(idlePeriodStartMap[event.key])
				}
			case <-noMoreRequests:
				moreRequestsExpected = false
			case <-maxWaitTimeExceeded:
				log.Info().Msgf("timeout waiting for all connections to close, waited for %v, configured idle timeout was %v", maxWaitDuration, idleTimeout)
				s.T().Fail()
				close(completed)
				return
			}
		}

		require.Equal(s.T(), 1, connCount)

		for _, idlePeriod := range idlePeriodLengthMap {
			// Our method of measuring the actual idle period is not precise, so allow some sub-ms deviation
			require.LessOrEqual(s.T(), math.Round(idlePeriod.Seconds()), idleTimeout.Seconds())
		}

		close(completed)
	}()

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	server.Config.ConnState = func(conn net.Conn, state http.ConnState) {
		connStateChanges <- connStateChangeEvent{key: conn.RemoteAddr().String(), state: state}
	}
	server.Start()
	defer server.Close()

	config := KeepAliveConfig{KeepAliveServer: server.URL, IdleConnTimeout: idleTimeout.String()}
	file := s.adaptFile("fixtures/timeout/keepalive.toml", config)

	s.traefikCmd(withConfigFile(file))

	// Wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Duration(1)*time.Second, try.StatusCodeIs(200), try.BodyContains("PathPrefix(`/keepalive`)"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/keepalive", time.Duration(1)*time.Second, try.StatusCodeIs(200))
	require.NoError(s.T(), err)

	close(noMoreRequests)
	<-completed
}
