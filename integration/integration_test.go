// This is the main file that sets up integration tests using go-check.
package integration

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/compose/v2/cmd/formatter"
	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/fatih/structs"
	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/pkg/log"
	checker "github.com/vdemeester/shakers"
)

var (
	integration = flag.Bool("integration", false, "run integration tests")
	showLog     = flag.Bool("tlog", false, "always show Traefik logs")
)

func Test(t *testing.T) {
	if !*integration {
		log.WithoutContext().Info("Integration tests disabled.")
		return
	}

	check.Suite(&AccessLogSuite{})
	check.Suite(&AcmeSuite{})
	check.Suite(&ConsulCatalogSuite{})
	check.Suite(&ConsulSuite{})
	check.Suite(&DockerComposeSuite{})
	check.Suite(&DockerSuite{})
	check.Suite(&ErrorPagesSuite{})
	check.Suite(&EtcdSuite{})
	check.Suite(&FileSuite{})
	check.Suite(&GRPCSuite{})
	check.Suite(&HeadersSuite{})
	check.Suite(&HealthCheckSuite{})
	check.Suite(&HostResolverSuite{})
	check.Suite(&HTTPSSuite{})
	check.Suite(&HTTPSuite{})
	check.Suite(&K8sSuite{})
	check.Suite(&KeepAliveSuite{})
	check.Suite(&LogRotationSuite{})
	check.Suite(&MarathonSuite15{})
	check.Suite(&MarathonSuite{})
	check.Suite(&ProxyProtocolSuite{})
	check.Suite(&RateLimitSuite{})
	check.Suite(&RedisSuite{})
	check.Suite(&RestSuite{})
	check.Suite(&RetrySuite{})
	check.Suite(&SimpleSuite{})
	check.Suite(&TCPSuite{})
	check.Suite(&TimeoutSuite{})
	check.Suite(&ThrottlingSuite{})
	check.Suite(&TLSClientHeadersSuite{})
	check.Suite(&TracingSuite{})
	check.Suite(&UDPSuite{})
	check.Suite(&WebsocketSuite{})
	check.Suite(&ZookeeperSuite{})

	check.TestingT(t)
}

var traefikBinary = "../dist/traefik"

type BaseSuite struct {
	composeProject       *types.Project
	dockerComposeService composeapi.Service
	dockerClient         *client.Client
}

func (s *BaseSuite) TearDownSuite(c *check.C) {
	if s.composeProject != nil && s.dockerComposeService != nil {
		s.composeDown(c)
	}
}

// createComposeProject creates the docker compose project stored as a field in the BaseSuite.
// This method should be called before starting and/or stopping compose services.
func (s *BaseSuite) createComposeProject(c *check.C, name string) {
	projectName := fmt.Sprintf("traefik-integration-test-%s", name)
	composeFile := fmt.Sprintf("resources/compose/%s.yml", name)

	var err error
	s.dockerClient, err = client.NewClientWithOpts()
	c.Assert(err, checker.IsNil)

	s.dockerComposeService = compose.NewComposeService(s.dockerClient, &configfile.ConfigFile{})
	ops, err := cli.NewProjectOptions([]string{composeFile}, cli.WithName(projectName))
	c.Assert(err, checker.IsNil)

	s.composeProject, err = cli.ProjectFromOptions(ops)
	c.Assert(err, checker.IsNil)
}

// composeUp starts the given services of the current docker compose project, if they are not already started.
// Already running services are not affected (i.e. not stopped).
func (s *BaseSuite) composeUp(c *check.C, services ...string) {
	c.Assert(s.composeProject, check.NotNil)
	c.Assert(s.dockerComposeService, check.NotNil)

	// We use Create and Restart instead of Up, because the only option that actually works to control which containers
	// are started is within the RestartOptions.
	err := s.dockerComposeService.Create(context.Background(), s.composeProject, composeapi.CreateOptions{})
	c.Assert(err, checker.IsNil)

	err = s.dockerComposeService.Restart(context.Background(), s.composeProject, composeapi.RestartOptions{Services: services})
	c.Assert(err, checker.IsNil)
}

// composeStop stops the given services of the current docker compose project and removes the corresponding containers.
func (s *BaseSuite) composeStop(c *check.C, services ...string) {
	c.Assert(s.dockerComposeService, check.NotNil)
	c.Assert(s.composeProject, check.NotNil)

	err := s.dockerComposeService.Stop(context.Background(), s.composeProject, composeapi.StopOptions{Services: services})
	c.Assert(err, checker.IsNil)

	err = s.dockerComposeService.Remove(context.Background(), s.composeProject, composeapi.RemoveOptions{
		Services: services,
		Force:    true,
	})
	c.Assert(err, checker.IsNil)
}

// composeDown stops all compose project services and removes the corresponding containers.
func (s *BaseSuite) composeDown(c *check.C) {
	c.Assert(s.dockerComposeService, check.NotNil)
	c.Assert(s.composeProject, check.NotNil)

	err := s.dockerComposeService.Down(context.Background(), s.composeProject.Name, composeapi.DownOptions{})
	c.Assert(err, checker.IsNil)
}

func (s *BaseSuite) cmdTraefik(args ...string) (*exec.Cmd, *bytes.Buffer) {
	cmd := exec.Command(traefikBinary, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	return cmd, &out
}

func (s *BaseSuite) killCmd(cmd *exec.Cmd) {
	err := cmd.Process.Kill()
	if err != nil {
		log.WithoutContext().Errorf("Kill: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

func (s *BaseSuite) traefikCmd(args ...string) (*exec.Cmd, func(*check.C)) {
	cmd, out := s.cmdTraefik(args...)
	return cmd, func(c *check.C) {
		if c.Failed() || *showLog {
			s.displayLogK3S(c)
			s.displayLogCompose(c)
			s.displayTraefikLog(c, out)
		}
	}
}

func (s *BaseSuite) displayLogK3S(c *check.C) {
	filePath := "./fixtures/k8s/config.skip/k3s.log"
	if _, err := os.Stat(filePath); err == nil {
		content, errR := os.ReadFile(filePath)
		if errR != nil {
			log.WithoutContext().Error(errR)
		}
		log.WithoutContext().Println(string(content))
	}
	log.WithoutContext().Println()
	log.WithoutContext().Println("################################")
	log.WithoutContext().Println()
}

func (s *BaseSuite) displayLogCompose(c *check.C) {
	if s.dockerComposeService == nil || s.composeProject == nil {
		log.WithoutContext().Infof("%s: No docker compose logs.", c.TestName())
		return
	}

	log.WithoutContext().Infof("%s: docker compose logs: ", c.TestName())

	logWriter := log.WithoutContext().WriterLevel(log.GetLevel())
	logConsumer := formatter.NewLogConsumer(context.Background(), logWriter, false, true)

	err := s.dockerComposeService.Logs(context.Background(), s.composeProject.Name, logConsumer, composeapi.LogOptions{})
	c.Assert(err, checker.IsNil)

	log.WithoutContext().Println()
	log.WithoutContext().Println("################################")
	log.WithoutContext().Println()
}

func (s *BaseSuite) displayTraefikLog(c *check.C, output *bytes.Buffer) {
	if output == nil || output.Len() == 0 {
		log.WithoutContext().Infof("%s: No Traefik logs.", c.TestName())
	} else {
		log.WithoutContext().Infof("%s: Traefik logs: ", c.TestName())
		log.WithoutContext().Infof(output.String())
	}
}

func (s *BaseSuite) getDockerHost() string {
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		// Default docker socket
		dockerHost = "unix:///var/run/docker.sock"
	}

	return dockerHost
}

func (s *BaseSuite) adaptFile(c *check.C, path string, tempObjects interface{}) string {
	// Load file
	tmpl, err := template.ParseFiles(path)
	c.Assert(err, checker.IsNil)

	folder, prefix := filepath.Split(path)
	tmpFile, err := os.CreateTemp(folder, strings.TrimSuffix(prefix, filepath.Ext(prefix))+"_*"+filepath.Ext(prefix))
	c.Assert(err, checker.IsNil)
	defer tmpFile.Close()

	model := structs.Map(tempObjects)
	model["SelfFilename"] = tmpFile.Name()

	err = tmpl.ExecuteTemplate(tmpFile, prefix, model)
	c.Assert(err, checker.IsNil)
	err = tmpFile.Sync()

	c.Assert(err, checker.IsNil)

	return tmpFile.Name()
}

func (s *BaseSuite) getComposeServiceIP(c *check.C, name string) string {
	filter := filters.NewArgs(
		filters.Arg("label", fmt.Sprintf("%s=%s", composeapi.ProjectLabel, s.composeProject.Name)),
		filters.Arg("label", fmt.Sprintf("%s=%s", composeapi.ServiceLabel, name)),
	)

	containers, err := s.dockerClient.ContainerList(context.Background(), dockertypes.ContainerListOptions{Filters: filter})
	c.Assert(err, checker.IsNil)
	c.Assert(containers, checker.HasLen, 1)

	networkNames := s.composeProject.NetworkNames()
	c.Assert(networkNames, checker.HasLen, 1)

	network := s.composeProject.Networks[networkNames[0]]
	return containers[0].NetworkSettings.Networks[network.Name].IPAddress
}

func (s *BaseSuite) getContainerIP(c *check.C, name string) string {
	container, err := s.dockerClient.ContainerInspect(context.Background(), name)
	c.Assert(err, checker.IsNil)
	c.Assert(container.NetworkSettings.Networks, check.NotNil)

	for _, network := range container.NetworkSettings.Networks {
		return network.IPAddress
	}

	// Should never happen.
	c.Error("No network found")
	return ""
}

func withConfigFile(file string) string {
	return "--configFile=" + file
}
