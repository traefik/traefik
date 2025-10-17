package integration

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
)

const traefikTestOTLPLogFile = "traefik_otlp.log"

// DualLoggingSuite tests that both OTLP and stdout logging can work together.
type DualLoggingSuite struct {
	BaseSuite
	otlpLogs  []string
	collector *httptest.Server
}

func TestDualLoggingSuite(t *testing.T) {
	suite.Run(t, new(DualLoggingSuite))
}

func (s *DualLoggingSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	// Clean up any existing log files
	os.Remove(traefikTestLogFile)
	os.Remove(traefikTestOTLPLogFile)
}

func (s *DualLoggingSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()

	// Clean up log files
	generatedFiles := []string{
		traefikTestLogFile,
		traefikTestOTLPLogFile,
	}

	for _, filename := range generatedFiles {
		if err := os.Remove(filename); err != nil {
			s.T().Logf("Failed to remove %s: %v", filename, err)
		}
	}
}

func (s *DualLoggingSuite) SetupTest() {
	s.otlpLogs = []string{}

	// Create mock OTLP collector
	s.collector = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		gzr, err := gzip.NewReader(r.Body)
		if err != nil {
			s.T().Logf("Error creating gzip reader: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer gzr.Close()

		body, err := io.ReadAll(gzr)
		if err != nil {
			s.T().Logf("Error reading gzipped body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req := plogotlp.NewExportRequest()
		err = req.UnmarshalProto(body)
		if err != nil {
			s.T().Logf("Error unmarshaling protobuf: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		marshalledReq, err := json.Marshal(req)
		if err != nil {
			s.T().Logf("Error marshaling to JSON: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.otlpLogs = append(s.otlpLogs, string(marshalledReq))

		w.WriteHeader(http.StatusOK)
	}))
}

func (s *DualLoggingSuite) TearDownTest() {
	if s.collector != nil {
		s.collector.Close()
		s.collector = nil
	}
}

func (s *DualLoggingSuite) TestOTLPAndStdoutLogging() {
	tempObjects := struct {
		CollectorURL string
	}{
		CollectorURL: s.collector.URL + "/v1/logs",
	}

	file := s.adaptFile("fixtures/dual_logging/otlp_and_stdout.toml", tempObjects)

	cmd, display := s.cmdTraefik(withConfigFile(file))
	defer s.displayTraefikLogFile(traefikTestLogFile)

	s.waitForTraefik("dashboard")

	time.Sleep(3 * time.Second)

	s.killCmd(cmd)
	time.Sleep(1 * time.Second)

	assert.NotEmpty(s.T(), s.otlpLogs)

	output := display.String()
	otlpOutput := strings.Join(s.otlpLogs, "\n")

	foundStdoutLog := strings.Contains(output, "Starting provider")
	assert.True(s.T(), foundStdoutLog)
	foundOTLPLog := strings.Contains(otlpOutput, "Starting provider")
	assert.True(s.T(), foundOTLPLog)
}
