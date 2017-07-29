package integration

import (
	"net/http"
	"time"

	"bytes"
	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
	"io/ioutil"
	"os"
	"os/exec"
)

// Plugin test suites
type PluginSuite struct{ BaseSuite }

func (s *PluginSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "plugin")
	s.composeProject.Start(c)
}

func (s *PluginSuite) TestMiddleware(c *check.C) {
	// check that the plugin has been compiled
	plugin := "resources/plugin/middleware/middleware.so"
	err := try.FileExists(plugin, 30*time.Second)
	c.Assert(err, checker.IsNil)

	// check that traefik binary has been compiled
	traefikBinary := "../dist/traefik-musl"
	err = try.FileExists(traefikBinary, 60*time.Second)
	c.Assert(err, checker.IsNil)

	// load the plugin
	whoamiHost := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/plugin/simple.toml", struct {
		PluginPath string
		Server1    string
	}{plugin, whoamiHost})
	defer os.Remove(file)

	cmd := exec.Command(traefikBinary, withConfigFile(file))
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	req.Host = "test.localhost"

	resp, err := try.ResponseUntilStatusCode(req, 30*time.Second, http.StatusOK)
	c.Assert(err, checker.IsNil)

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)
	c.Assert(string(body), checker.Contains, "plugin.middleware.reponse.body")
}

func (s *PluginSuite) TestNoLoad(c *check.C) {
	// check that the plugin has been compiled
	plugin := "resources/plugin/no-load/no-load.so"
	err := try.FileExists(plugin, 30*time.Second)
	c.Assert(err, checker.IsNil)

	// check that traefik binary has been compiled
	traefikBinary := "../dist/traefik-musl"
	err = try.FileExists(traefikBinary, 60*time.Second)
	c.Assert(err, checker.IsNil)

	// load the plugin
	whoamiHost := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/plugin/simple.toml", struct {
		PluginPath string
		Server1    string
	}{plugin, whoamiHost})
	defer os.Remove(file)

	cmd := exec.Command(traefikBinary, withConfigFile(file))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8081/api/providers", 60*time.Second, try.BodyContains("Host:test.localhost"))
	c.Assert(err, checker.IsNil)

	c.Assert(out.String(), checker.Contains, "Error loading plugin: error in plugin Lookup: plugin: symbol Load not found in plugin")
}

func (s *PluginSuite) TestBadLoad(c *check.C) {
	// check that the plugin has been compiled
	plugin := "resources/plugin/bad-load/bad-load.so"
	err := try.FileExists(plugin, 30*time.Second)
	c.Assert(err, checker.IsNil)

	// check that traefik binary has been compiled
	traefikBinary := "../dist/traefik-musl"
	err = try.FileExists(traefikBinary, 60*time.Second)
	c.Assert(err, checker.IsNil)

	// load the plugin
	whoamiHost := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/plugin/simple.toml", struct {
		PluginPath string
		Server1    string
	}{plugin, whoamiHost})
	defer os.Remove(file)

	cmd := exec.Command(traefikBinary, withConfigFile(file))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8081/api/providers", 60*time.Second, try.BodyContains("Host:test.localhost"))
	c.Assert(err, checker.IsNil)

	c.Assert(out.String(), checker.Contains, "bad-load.so} does not implement Load() interface{} function")
}

func (s *PluginSuite) TestNoInterface(c *check.C) {
	// check that the plugin has been compiled
	plugin := "resources/plugin/no-interface/no-interface.so"
	err := try.FileExists(plugin, 30*time.Second)
	c.Assert(err, checker.IsNil)

	// check that traefik binary has been compiled
	traefikBinary := "../dist/traefik-musl"
	err = try.FileExists(traefikBinary, 60*time.Second)
	c.Assert(err, checker.IsNil)

	// load the plugin
	whoamiHost := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/plugin/simple.toml", struct {
		PluginPath string
		Server1    string
	}{plugin, whoamiHost})
	defer os.Remove(file)

	cmd := exec.Command(traefikBinary, withConfigFile(file))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8081/api/providers", 60*time.Second, try.BodyContains("Host:test.localhost"))
	c.Assert(err, checker.IsNil)

	c.Assert(out.String(), checker.Contains, "no-interface.so} does not implement any plugin interface")
}
