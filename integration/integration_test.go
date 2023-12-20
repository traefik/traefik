// This is the main file that sets up integration tests using go-check.
package integration

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"k8s.io/utils/strings/slices"
	stdlog "log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/docker/docker/api/types/container"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/fatih/structs"
	"github.com/go-check/check"
	notaryclient "github.com/theupdateframework/notary/client"
	"github.com/testcontainers/testcontainers-go"
	"github.com/traefik/traefik/v2/pkg/log"
	checker "github.com/vdemeester/shakers"
	"gopkg.in/yaml.v3"
)

var (
	integration = flag.Bool("integration", true, "run integration tests")
	showLog     = flag.Bool("tlog", false, "always show Traefik logs")
)

type composeConfig struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Image       string            `yaml:"image"`
	Labels      map[string]string `yaml:"labels"`
	Hostname    string            `yaml:"hostname"`
	Volumes     []string          `yaml:"volumes"`
	CapAdd      []string          `yaml:"cap_add"`
	Command     []string          `yaml:"command"`
	Environment map[string]string `yaml:"environment"`
	Privileged  bool              `yaml:"privileged"`
}

func Test(t *testing.T) {
	if !*integration {
		log.WithoutContext().Info("Integration tests disabled.")
		// return
	}

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Global logrus replacement
	logrus.StandardLogger().Out = logs.NoLevel(log.Logger, zerolog.DebugLevel)

	// configure default standard log.
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)
	stdlog.SetOutput(logs.NoLevel(log.Logger, zerolog.DebugLevel))

	check.Suite(&AccessLogSuite{})
	//if !useVPN {
	// check.Suite(&AcmeSuite{})
	//}
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
	//if !useVPN {
	check.Suite(&K8sSuite{})
	//}
	check.Suite(&KeepAliveSuite{})
	check.Suite(&LogRotationSuite{})
	check.Suite(&MarathonSuite{})
	check.Suite(&MarathonSuite15{})
	//if !useVPN {
	check.Suite(&ProxyProtocolSuite{})
	//}
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

var networkName = "traefik-test-network"

type BaseSuite struct {
	containers map[string]testcontainers.Container
	network    testcontainers.Network
}

func (s *BaseSuite) SetUpSuite(c *check.C) {
	_ = os.Setenv("TESTCONTAINERS_HOST_OVERRIDE", "host.docker.internal")
	// Create docker network
	// docker network create traefik-test-network --driver bridge --subnet 172.31.42.0/24
	req := testcontainers.GenericNetworkRequest{
		ProviderType: testcontainers.ProviderDocker,
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
			Driver:         "bridge",
			IPAM: &dockernetwork.IPAM{
				Driver:  "default",
				Options: nil,
				Config: []dockernetwork.IPAMConfig{
					{
						Subnet: "172.31.42.0/24",
					},
				},
			},
		},
	}
	network, err := testcontainers.GenericNetwork(context.Background(), req)
	c.Assert(err, checker.IsNil)

	if os.Getenv("IN_DOCKER") != "true" {
		setupVPN(nil, "tailscale.secret")
	}

	s.network = network
}

func (s *BaseSuite) TearDownSuite(c *check.C) {
	s.composeDown(c)

	if s.network != nil {
		err := s.network.Remove(context.Background())
		c.Assert(err, checker.IsNil)
	}
}

// createComposeProject creates the docker compose project stored as a field in the BaseSuite.
// This method should be called before starting and/or stopping compose services.
func (s *BaseSuite) createComposeProject(c *check.C, name string) {
	composeFile := fmt.Sprintf("resources/compose/%s.yml", name)

	file, err := os.ReadFile(composeFile)
	c.Assert(err, checker.IsNil)

	var composeConfigData composeConfig
	err = yaml.Unmarshal(file, &composeConfigData)
	c.Assert(err, checker.IsNil)

	if s.containers == nil {
		s.containers = map[string]testcontainers.Container{}
	}

	ctx := context.Background()

	for id, containerConfig := range composeConfigData.Services {
		var mounts testcontainers.ContainerMounts
		for _, volume := range containerConfig.Volumes {
			split := strings.Split(volume, ":")
			if len(split) != 2 {
				continue
			}

			if strings.HasPrefix(split[0], "./") {
				path, err := os.Getwd()
				if err != nil {
					log.Err(err).Msg("can't determine current directory")
					continue
				}
				split[0] = strings.Replace(split[0], "./", path+"/", 1)
			}

			mounts = append(mounts, testcontainers.BindMount(split[0], testcontainers.ContainerMountTarget(split[1])))
		}

		req := testcontainers.ContainerRequest{
			Image:    containerConfig.Image,
			Name:     id,
			Hostname: containerConfig.Hostname,
			HostConfigModifier: func(config *container.HostConfig) {
				if containerConfig.CapAdd != nil {
					config.CapAdd = containerConfig.CapAdd
				}
			},
			Cmd:        containerConfig.Command,
			Mounts:     testcontainers.Mounts(mounts...),
			Labels:     containerConfig.Labels,
			Networks:   []string{networkName},
			Env:        containerConfig.Environment,
			Privileged: containerConfig.Privileged,
		}
		con, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          false,
		})

		s.containers[id] = con

		c.Assert(err, checker.IsNil)
	}
}

// composeUp starts the given services of the current docker compose project, if they are not already started.
// Already running services are not affected (i.e. not stopped).
func (s *BaseSuite) composeUp(c *check.C, services ...string) {
	for name, con := range s.containers {
		if len(services) == 0 || slices.Contains(services, name) {
			err := con.Start(context.Background())
			c.Assert(err, checker.IsNil)
		}
	}
}

// composeStop stops the given services of the current docker compose project and removes the corresponding containers.
func (s *BaseSuite) composeStop(c *check.C, services ...string) {
	for name, con := range s.containers {
		if len(services) == 0 || slices.Contains(services, name) {
			timeout := 10 * time.Second
			err := con.Stop(context.Background(), &timeout)
			c.Assert(err, checker.IsNil)
		}
	}
}

// composeDown stops all compose project services and removes the corresponding containers.
func (s *BaseSuite) composeDown(c *check.C) {
	for _, container := range s.containers {
		err := container.Terminate(context.Background())
		c.Assert(err, checker.IsNil)
	}
	s.containers = map[string]testcontainers.Container{}
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
			s.displayLogK3S()
			s.displayLogCompose(c)
			s.displayTraefikLog(c, out)
		}
	}
}

func (s *BaseSuite) displayLogK3S() {
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
	// if s.dockerComposeService == nil || s.composeProject == nil {
	// 	log.Info().Str("testName", c.TestName()).Msg("No docker compose logs.")
	// 	return
	// }
	//
	// log.Info().Str("testName", c.TestName()).Msg("docker compose logs")
	//
	// logConsumer := formatter.NewLogConsumer(context.Background(), logs.NoLevel(log.Logger, zerolog.InfoLevel), false, true)
	//
	// err := s.dockerComposeService.Logs(context.Background(), s.composeProject.Name, logConsumer, composeapi.LogOptions{})
	// c.Assert(err, checker.IsNil)
	//
	// log.Print()
	// log.Print("################################")
	// log.Print()
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
	container, ok := s.containers[name]
	if !ok {
		return ""
	}
	ip, err := container.ContainerIP(context.Background())
	if err != nil {
		return ""
	}
	return ip
}

func (s *BaseSuite) getContainerIP(c *check.C, name string) string {
	// container, err := s.dockerClient.ContainerInspect(context.Background(), name)
	// c.Assert(err, checker.IsNil)
	// c.Assert(container.NetworkSettings.Networks, check.NotNil)
	//
	// for _, network := range container.NetworkSettings.Networks {
	// 	return network.IPAddress
	// }
	//
	// // Should never happen.
	// c.Error("No network found")
	return ""
}

func withConfigFile(file string) string {
	return "--configFile=" + file
}

// tailscaleNotSuite includes a BaseSuite out of convenience, so we can benefit
// from composeUp et co., but it is not meant to function as a TestSuite per se.
type tailscaleNotSuite struct{ BaseSuite }

// setupVPN starts Tailscale on the corresponding container, and makes it a subnet
// router, for all the other containers (whoamis, etc) subsequently started for the
// integration tests.
// It only does so if the file provided as argument exists, and contains a
// Tailscale auth key (an ephemeral, but reusable, one is recommended).
//
// Add this section to your tailscale ACLs to auto-approve the routes for the
// containers in the docker subnet:
//
//	"autoApprovers": {
//	  // Allow myself to automatically advertize routes for docker networks
//	  "routes": {
//	    "172.0.0.0/8": ["your_tailscale_identity"],
//	  },
//	},
func setupVPN(c *check.C, keyFile string) {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Fatal().Err(err).Send()
		}

		return
	}
	authKey := strings.TrimSpace(string(data))
	// // TODO: copy and create versions that don't need a check.C?
	vpn := &tailscaleNotSuite{}
	vpn.createComposeProject(c, "tailscale")
	vpn.composeUp(c)
	time.Sleep(5 * time.Second)
	// If we ever change the docker subnet in the Makefile,
	// we need to change this one below correspondingly.
	vpn.composeExec(c, "tailscaled", "tailscale", "up", "--authkey="+authKey, "--advertise-routes=172.31.42.0/24")
}

// composeExec runs the command in the given args in the given compose service container.
// Already running services are not affected (i.e. not stopped).
func (s *BaseSuite) composeExec(c *check.C, service string, args ...string) {
	c.Assert(s.containers[service], check.NotNil)

	_, _, err := s.containers[service].Exec(context.Background(), args)
	c.Assert(err, checker.IsNil)
}

type FakeDockerCLI struct {
	client client.APIClient
}

func (f FakeDockerCLI) Client() client.APIClient {
	return f.client
}

func (f FakeDockerCLI) In() *streams.In {
	return streams.NewIn(os.Stdin)
}

func (f FakeDockerCLI) Out() *streams.Out {
	return streams.NewOut(os.Stdout)
}

func (f FakeDockerCLI) Err() io.Writer {
	return streams.NewOut(os.Stderr)
}

func (f FakeDockerCLI) SetIn(in *streams.In) {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) Apply(ops ...command.DockerCliOption) error {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) ConfigFile() *configfile.ConfigFile {
	return &configfile.ConfigFile{}
}

func (f FakeDockerCLI) ServerInfo() command.ServerInfo {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) NotaryClient(imgRefAndAuth trust.ImageRefAndAuth, actions []string) (notaryclient.Repository, error) {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) DefaultVersion() string {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) CurrentVersion() string {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) ManifestStore() manifeststore.Store {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) RegistryClient(b bool) registryclient.RegistryClient {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) ContentTrustEnabled() bool {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) BuildKitEnabled() (bool, error) {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) ContextStore() store.Store {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) CurrentContext() string {
	panic("implement me if you need me")
}

func (f FakeDockerCLI) DockerEndpoint() docker.Endpoint {
	panic("implement me if you need me")
}
