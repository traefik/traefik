package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/baqupio/baqup/v3/integration/try"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// File tests suite.
type FileSuite struct{ BaseSuite }

func TestFileSuite(t *testing.T) {
	suite.Run(t, new(FileSuite))
}

func (s *FileSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("file")
	s.composeUp()
}

func (s *FileSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *FileSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/file/simple.toml", struct{}{})
	s.baqupCmd(withConfigFile(file))

	// Expected a 404 as we did not configure anything
	err := try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

// #56 regression test, make sure it does not fail?
func (s *FileSuite) TestSimpleConfigurationNoPanic() {
	s.baqupCmd(withConfigFile("fixtures/file/56-simple-panic.toml"))

	// Expected a 404 as we did not configure anything
	err := try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *FileSuite) TestDirectoryConfiguration() {
	s.baqupCmd(withConfigFile("fixtures/file/directory.toml"))

	// Expected a 404 as we did not configure anything at /test
	err := try.GetRequest("http://127.0.0.1:8000/test", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	// Expected a 502 as there is no backend server
	err = try.GetRequest("http://127.0.0.1:8000/test2", 1000*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	require.NoError(s.T(), err)
}
