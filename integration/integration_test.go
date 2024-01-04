// This is the main file that sets up integration tests using go-check.
package integration

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	stdlog "log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/docker/docker/api/types/container"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/fatih/structs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/integration/try"
	"gopkg.in/yaml.v3"
)

var (
	showLog = flag.Bool("tlog", false, "always show Traefik logs")
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
	Deploy      composeDeploy     `yaml:"deploy"`
}

type composeDeploy struct {
	Replicas int `yaml:"replicas"`
}

var traefikBinary = "../dist/traefik"

var networkName = "traefik-test-network"

type BaseSuite struct {
	suite.Suite
	containers map[string]testcontainers.Container
	network    testcontainers.Network
	hostIP     string
}

func (s *BaseSuite) waitForTraefik(containerName string) {
	time.Sleep(1 * time.Second)

	// Wait for Traefik to turn ready.
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/rawdata", nil)
	require.NoError(s.T(), err)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains(containerName))
	require.NoError(s.T(), err)
}

func (s *BaseSuite) displayTraefikLogFile(path string) {
	if s.T().Failed() {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			content, errRead := os.ReadFile(path)
			// TODO TestName
			// fmt.Printf("%s: Traefik logs: \n", c.TestName())
			fmt.Print("Traefik logs: \n")
			if errRead == nil {
				fmt.Println(content)
			} else {
				fmt.Println(errRead)
			}
		} else {
			// fmt.Printf("%s: No Traefik logs.\n", c.TestName())
			fmt.Print("No Traefik logs.\n")
		}
		errRemove := os.Remove(path)
		if errRemove != nil {
			fmt.Println(errRemove)
		}
	}
}

func (s *BaseSuite) SetupSuite() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Global logrus replacement
	logrus.StandardLogger().Out = logs.NoLevel(log.Logger, zerolog.DebugLevel)

	// configure default standard log.
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)
	stdlog.SetOutput(logs.NoLevel(log.Logger, zerolog.DebugLevel))

	// Create docker network
	// docker network create traefik-test-network --driver bridge --subnet 172.31.42.0/24
	req := testcontainers.GenericNetworkRequest{
		ProviderType: testcontainers.ProviderDocker,
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
			Driver:         "bridge",
			IPAM: &dockernetwork.IPAM{
				Driver: "default",
				Config: []dockernetwork.IPAMConfig{
					{
						Subnet: "172.31.42.0/24",
					},
				},
			},
		},
	}
	ctx := context.Background()
	network, err := testcontainers.GenericNetwork(ctx, req)
	require.NoError(s.T(), err)

	s.network = network
	s.hostIP = "172.31.42.1"
	if isDockerDesktop(ctx, s.T()) {
		s.hostIP = getDockerDesktopHostIP(ctx, s.T())
		s.setupVPN("tailscale.secret")
	}
}

func getDockerDesktopHostIP(ctx context.Context, t *testing.T) string {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image: "alpine",
		HostConfigModifier: func(config *container.HostConfig) {
			config.AutoRemove = true
		},
		Cmd: []string{"getent", "hosts", "host.docker.internal"},
	}

	con, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	closer, err := con.Logs(ctx)
	require.NoError(t, err)

	all, err := io.ReadAll(closer)
	require.NoError(t, err)

	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	matches := ipRegex.FindAllString(string(all), -1)
	require.Len(t, matches, 1)

	return matches[0]
}

func isDockerDesktop(ctx context.Context, t *testing.T) bool {
	t.Helper()

	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		t.Fatalf("failed to create docker client: %s", err)
	}

	info, err := cli.Info(ctx)
	if err != nil {
		t.Fatalf("failed to get docker info: %s", err)
	}

	return info.OperatingSystem == "Docker Desktop"
}

func (s *BaseSuite) TearDownSuite() {
	s.composeDown()

	err := try.Do(5*time.Second, func() error {
		if s.network != nil {
			err := s.network.Remove(context.Background())
			if err != nil {
				return err
			}
		}

		return nil
	})
	require.NoError(s.T(), err)
}

// createComposeProject creates the docker compose project stored as a field in the BaseSuite.
// This method should be called before starting and/or stopping compose services.
func (s *BaseSuite) createComposeProject(name string) {
	composeFile := fmt.Sprintf("resources/compose/%s.yml", name)

	file, err := os.ReadFile(composeFile)
	require.NoError(s.T(), err)

	var composeConfigData composeConfig
	err = yaml.Unmarshal(file, &composeConfigData)
	require.NoError(s.T(), err)

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

		if containerConfig.Deploy.Replicas > 0 {
			for i := 0; i < containerConfig.Deploy.Replicas; i++ {
				id = fmt.Sprintf("%s-%d", id, i+1)
				con, err := s.createContainer(ctx, containerConfig, id, mounts)
				require.NoError(s.T(), err)
				s.containers[id] = con
			}
			continue
		}

		con, err := s.createContainer(ctx, containerConfig, id, mounts)
		require.NoError(s.T(), err)
		s.containers[id] = con
	}
}

func (s *BaseSuite) createContainer(ctx context.Context, containerConfig composeService, id string, mounts testcontainers.ContainerMounts) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:      containerConfig.Image,
		Env:        containerConfig.Environment,
		Cmd:        containerConfig.Command,
		Labels:     containerConfig.Labels,
		Mounts:     testcontainers.Mounts(mounts...),
		Name:       id,
		Hostname:   containerConfig.Hostname,
		Privileged: containerConfig.Privileged,
		Networks:   []string{networkName},
		HostConfigModifier: func(config *container.HostConfig) {
			if containerConfig.CapAdd != nil {
				config.CapAdd = containerConfig.CapAdd
			}
			if !isDockerDesktop(ctx, s.T()) {
				config.ExtraHosts = append(config.ExtraHosts, "host.docker.internal:"+s.hostIP)
			}
		},
	}
	con, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false,
	})

	return con, err
}

// composeUp starts the given services of the current docker compose project, if they are not already started.
// Already running services are not affected (i.e. not stopped).
func (s *BaseSuite) composeUp(services ...string) {
	for name, con := range s.containers {
		if len(services) == 0 || slices.Contains(services, name) {
			err := con.Start(context.Background())
			require.NoError(s.T(), err)
		}
	}
}

// composeStop stops the given services of the current docker compose project and removes the corresponding containers.
func (s *BaseSuite) composeStop(services ...string) {
	for name, con := range s.containers {
		if len(services) == 0 || slices.Contains(services, name) {
			timeout := 10 * time.Second
			err := con.Stop(context.Background(), &timeout)
			require.NoError(s.T(), err)
		}
	}
}

// composeDown stops all compose project services and removes the corresponding containers.
func (s *BaseSuite) composeDown() {
	for _, container := range s.containers {
		err := container.Terminate(context.Background())
		require.NoError(s.T(), err)
	}
	s.containers = map[string]testcontainers.Container{}
}

func (s *BaseSuite) cmdTraefik(args ...string) (*exec.Cmd, *bytes.Buffer) {
	cmd := exec.Command(traefikBinary, args...)

	s.T().Cleanup(func() {
		s.killCmd(cmd)
	})
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Start()
	require.NoError(s.T(), err)

	return cmd, &out
}

func (s *BaseSuite) killCmd(cmd *exec.Cmd) {
	if cmd.Process == nil {
		log.Error().Msg("No process to kill")
		return
	}
	err := cmd.Process.Kill()
	if err != nil {
		log.WithoutContext().Errorf("Kill: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

func (s *BaseSuite) traefikCmd(args ...string) {
	_, out := s.cmdTraefik(args...)

	s.T().Cleanup(func() {
		if s.T().Failed() || *showLog {
			s.displayLogK3S()
			s.displayLogCompose()
			s.displayTraefikLog(out)
		}
	})
	return
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

func (s *BaseSuite) displayLogCompose() {
	for name, ctn := range s.containers {
		readCloser, err := ctn.Logs(context.Background())
		require.NoError(s.T(), err)
		for {
			b := make([]byte, 1024)
			_, err := readCloser.Read(b)
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(s.T(), err)

			trimLogs := bytes.Trim(bytes.TrimSpace(b), string([]byte{0}))
			if len(trimLogs) > 0 {
				log.Info().Str("container", name).Msg(string(trimLogs))
			}
		}
	}
}

func (s *BaseSuite) displayTraefikLog(output *bytes.Buffer) {
	if output == nil || output.Len() == 0 {
		log.WithoutContext().Info("No Traefik logs.")
	} else {
		for _, line := range strings.Split(output.String(), "\n") {
			log.WithoutContext().Info(line)
		}

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

func (s *BaseSuite) adaptFile(path string, tempObjects interface{}) string {
	// Load file
	tmpl, err := template.ParseFiles(path)
	require.NoError(s.T(), err)

	folder, prefix := filepath.Split(path)
	tmpFile, err := os.CreateTemp(folder, strings.TrimSuffix(prefix, filepath.Ext(prefix))+"_*"+filepath.Ext(prefix))
	require.NoError(s.T(), err)
	defer tmpFile.Close()

	model := structs.Map(tempObjects)
	model["SelfFilename"] = tmpFile.Name()

	err = tmpl.ExecuteTemplate(tmpFile, prefix, model)
	require.NoError(s.T(), err)
	err = tmpFile.Sync()

	require.NoError(s.T(), err)

	return tmpFile.Name()
}

func (s *BaseSuite) getComposeServiceIP(name string) string {
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

func withConfigFile(file string) string {
	return "--configFile=" + file
}

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
func (s *BaseSuite) setupVPN(keyFile string) {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Fatal().Err(err).Send()
		}

		return
	}
	authKey := strings.TrimSpace(string(data))
	// // TODO: copy and create versions that don't need a check.C?
	s.createComposeProject("tailscale")
	s.composeUp()
	time.Sleep(5 * time.Second)
	// If we ever change the docker subnet in the Makefile,
	// we need to change this one below correspondingly.
	s.composeExec("tailscaled", "tailscale", "up", "--authkey="+authKey, "--advertise-routes=172.31.42.0/24")
}

// composeExec runs the command in the given args in the given compose service container.
// Already running services are not affected (i.e. not stopped).
func (s *BaseSuite) composeExec(service string, args ...string) string {
	require.Contains(s.T(), s.containers, service)

	_, reader, err := s.containers[service].Exec(context.Background(), args)
	require.NoError(s.T(), err)

	content, err := io.ReadAll(reader)
	require.NoError(s.T(), err)

	return string(content)
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
