package tap

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/middlewares"
)

var _ middlewares.Stateful = &statusRecorder{}

// statusRecorder is the http.ResponseWriter given to the tap services, discarding their response body.
type statusRecorder struct {
	status int
	header http.Header
}

func (s *statusRecorder) Header() http.Header {
	return s.header
}

func (s *statusRecorder) WriteHeader(status int) {
	// An informational status commits the tap service to nothing:
	// the one that follows tells whether the record is accepted.
	if status >= 100 && status <= 199 {
		return
	}

	if s.status == 0 {
		s.status = status
	}
}

func (s *statusRecorder) Write(p []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}

	return len(p), nil
}

func (s *statusRecorder) Flush() {}

func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("connection on %T cannot be hijacked", s)
}
