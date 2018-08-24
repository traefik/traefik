// This is the main file that sets up integration tests using go-check.
package integration

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/go-check/check"
	compose "github.com/libkermit/compose/check"
	checker "github.com/vdemeester/shakers"
)

var integration = flag.Bool("integration", false, "run integration tests")
var container = flag.Bool("container", false, "run container integration tests")
var host = flag.Bool("host", false, "run host integration tests")
var showLog = flag.Bool("tlog", false, "always show Traefik logs")

func Test(t *testing.T) {
	check.TestingT(t)
}

func init() {
	flag.Parse()
	if !*integration {
		log.Info("Integration tests disabled.")
		return
	}

	if *container {
		// tests launched from a container
		check.Suite(&AccessLogSuite{})
		check.Suite(&AcmeSuite{})
		check.Suite(&ConstraintSuite{})
		check.Suite(&ConsulCatalogSuite{})
		check.Suite(&ConsulSuite{})
		check.Suite(&DockerComposeSuite{})
		check.Suite(&DockerSuite{})
		check.Suite(&DynamoDBSuite{})
		check.Suite(&EtcdSuite{})
		check.Suite(&ErrorPagesSuite{})
		check.Suite(&EurekaSuite{})
		check.Suite(&FileSuite{})
		check.Suite(&GRPCSuite{})
		check.Suite(&HealthCheckSuite{})
		check.Suite(&HostResolverSuite{})
		check.Suite(&HTTPSSuite{})
		check.Suite(&LogRotationSuite{})
		check.Suite(&MarathonSuite{})
		check.Suite(&MarathonSuite15{})
		check.Suite(&MesosSuite{})
		check.Suite(&RateLimitSuite{})
		check.Suite(&RetrySuite{})
		check.Suite(&SimpleSuite{})
		check.Suite(&TLSClientHeadersSuite{})
		check.Suite(&TimeoutSuite{})
		check.Suite(&TracingSuite{})
		check.Suite(&WebsocketSuite{})
	}
	if *host {
		// tests launched from the host
		check.Suite(&ProxyProtocolSuite{})
		check.Suite(&Etcd3Suite{})
	}
}

var traefikBinary = "../dist/traefik"

type BaseSuite struct {
	composeProject *compose.Project
}

func (s *BaseSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *BaseSuite) createComposeProject(c *check.C, name string) {
	projectName := fmt.Sprintf("integration-test-%s", name)
	composeFile := fmt.Sprintf("resources/compose/%s.yml", name)

	addrs, err := net.InterfaceAddrs()
	c.Assert(err, checker.IsNil)
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		c.Assert(err, checker.IsNil)
		if !ip.IsLoopback() && ip.To4() != nil {
			os.Setenv("DOCKER_HOST_IP", ip.String())
			break
		}
	}

	s.composeProject = compose.CreateProject(c, projectName, composeFile)
}

func withConfigFile(file string) string {
	return "--configFile=" + file
}

func (s *BaseSuite) cmdTraefik(args ...string) (*exec.Cmd, *bytes.Buffer) {
	cmd := exec.Command(traefikBinary, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	return cmd, &out
}

func (s *BaseSuite) traefikCmd(args ...string) (*exec.Cmd, func(*check.C)) {
	cmd, out := s.cmdTraefik(args...)
	return cmd, func(c *check.C) {
		if c.Failed() || *showLog {
			s.displayTraefikLog(c, out)
		}
	}
}

func (s *BaseSuite) displayTraefikLog(c *check.C, output *bytes.Buffer) {
	if output == nil || output.Len() == 0 {
		log.Printf("%s: No Traefik logs.", c.TestName())
	} else {
		log.Printf("%s: Traefik logs: ", c.TestName())
		log.Println(output.String())
	}
}

func (s *BaseSuite) adaptFileForHost(c *check.C, path string) string {
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		// Default docker socket
		dockerHost = "unix:///var/run/docker.sock"
	}
	tempObjects := struct{ DockerHost string }{dockerHost}
	return s.adaptFile(c, path, tempObjects)
}

func (s *BaseSuite) adaptFile(c *check.C, path string, tempObjects interface{}) string {
	// Load file
	tmpl, err := template.ParseFiles(path)
	c.Assert(err, checker.IsNil)

	folder, prefix := filepath.Split(path)
	tmpFile, err := ioutil.TempFile(folder, prefix)
	c.Assert(err, checker.IsNil)
	defer tmpFile.Close()

	err = tmpl.ExecuteTemplate(tmpFile, prefix, tempObjects)
	c.Assert(err, checker.IsNil)
	err = tmpFile.Sync()

	c.Assert(err, checker.IsNil)

	return tmpFile.Name()
}
