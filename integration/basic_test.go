package integration

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// SimpleSuite
type SimpleSuite struct{ BaseSuite }

func (s *SimpleSuite) TestInvalidConfigShouldFail(c *check.C) {
	cmd, output := s.cmdTraefik(withConfigFile("fixtures/invalid_configuration.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.Do(500*time.Millisecond, func() error {
		expected := "Near line 0 (last key parsed ''): bare keys cannot contain '{'"
		actual := output.String()

		if !strings.Contains(actual, expected) {
			return fmt.Errorf("Got %s, wanted %s", actual, expected)
		}

		return nil
	})
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestSimpleDefaultConfig(c *check.C) {
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/simple_default.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestWithWebConfig(c *check.C) {
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/simple_web.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestDefaultEntryPoints(c *check.C) {
	cmd, output := s.cmdTraefik("--debug")

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.Do(500*time.Millisecond, func() error {
		expected := "\"DefaultEntryPoints\":[\"http\"]"
		actual := output.String()

		if !strings.Contains(actual, expected) {
			return fmt.Errorf("Got %s, wanted %s", actual, expected)
		}

		return nil
	})
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestPrintHelp(c *check.C) {
	cmd, output := s.cmdTraefik("--help")

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.Do(500*time.Millisecond, func() error {
		expected := "Usage:"
		notExpected := "panic:"
		actual := output.String()

		if strings.Contains(actual, notExpected) {
			return fmt.Errorf("Got %s", actual)
		}
		if !strings.Contains(actual, expected) {
			return fmt.Errorf("Got %s, wanted %s", actual, expected)
		}

		return nil
	})
	c.Assert(err, checker.IsNil)
}
