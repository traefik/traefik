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
	"runtime"
	"slices"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/fatih/structs"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/traefik/traefik/v3/integration/try"
	"gopkg.in/yaml.v3"
)

var (
	showLog               = flag.Bool("tlog", false, "always show Traefik logs")
	k8sConformance        = flag.Bool("k8sConformance", false, "run K8s Gateway API conformance test")
	k8sConformanceRunTest = flag.String("k8sConformanceRunTest", "", "run a specific K8s Gateway API conformance test")
)

const tailscaleSecretFilePath = "tailscale.secret"

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

type BaseSuite struct {
	suite.Suite
	containers map[string]testcontainers.Container
	network    *testcontainers.DockerNetwork
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
	if isDockerDesktop(context.Background(), s.T()) {
		_, err := os.Stat(tailscaleSecretFilePath)
		require.NoError(s.T(), err, "Tailscale need to be configured when running integration tests with Docker Desktop: (https://doc.traefik.io/traefik/v2.11/contributing/building-testing/#testing)")
	}

	// configure default standard log.
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)
	// TODO
	// stdlog.SetOutput(log.Logger)

	ctx := context.Background()
	// Create docker network
	// docker network create traefik-test-network --driver bridge --subnet 172.31.42.0/24
	var opts []network.NetworkCustomizer
	opts = append(opts, network.WithDriver("bridge"))
	opts = append(opts, network.WithIPAM(&dockernetwork.IPAM{
		Driver: "default",
		Config: []dockernetwork.IPAMConfig{
			{
				Subnet: "172.31.42.0/24",
			},
		},
	}))
	dockerNetwork, err := network.New(ctx, opts...)
	require.NoError(s.T(), err)

	s.network = dockerNetwork
	s.hostIP = "172.31.42.1"
	if isDockerDesktop(ctx, s.T()) {
		s.hostIP = getDockerDesktopHostIP(ctx, s.T())
		s.setupVPN(tailscaleSecretFilePath)
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
		var mounts []mount.Mount
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

			abs, err := filepath.Abs(split[0])
			require.NoError(s.T(), err)

			mounts = append(mounts, mount.Mount{Source: abs, Target: split[1], Type: mount.TypeBind})
		}

		if containerConfig.Deploy.Replicas > 0 {
			for i := range containerConfig.Deploy.Replicas {
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

func (s *BaseSuite) createContainer(ctx context.Context, containerConfig composeService, id string, mounts []mount.Mount) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:      containerConfig.Image,
		Env:        containerConfig.Environment,
		Cmd:        containerConfig.Command,
		Labels:     containerConfig.Labels,
		Name:       id,
		Hostname:   containerConfig.Hostname,
		Privileged: containerConfig.Privileged,
		Networks:   []string{s.network.Name},
		HostConfigModifier: func(config *container.HostConfig) {
			if containerConfig.CapAdd != nil {
				config.CapAdd = containerConfig.CapAdd
			}
			if !isDockerDesktop(ctx, s.T()) {
				config.ExtraHosts = append(config.ExtraHosts, "host.docker.internal:"+s.hostIP)
			}
			config.Mounts = mounts
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
	for _, c := range s.containers {
		err := c.Terminate(context.Background())
		require.NoError(s.T(), err)
	}
	s.containers = map[string]testcontainers.Container{}
}

func (s *BaseSuite) cmdTraefik(args ...string) (*exec.Cmd, *bytes.Buffer) {
	binName := "traefik"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	traefikBinPath := filepath.Join("..", "dist", runtime.GOOS, runtime.GOARCH, binName)
	cmd := exec.Command(traefikBinPath, args...)

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
		log.Error().Err(err).Msg("Kill")
	}

	time.Sleep(100 * time.Millisecond)
}

func (s *BaseSuite) traefikCmd(args ...string) *exec.Cmd {
	cmd, out := s.cmdTraefik(args...)

	s.T().Cleanup(func() {
		if s.T().Failed() || *showLog {
			s.displayLogK3S()
			s.displayLogCompose()
			s.displayTraefikLog(out)
		}
	})

	return cmd
}

func (s *BaseSuite) displayLogK3S() {
	filePath := "./fixtures/k8s/config.skip/k3s.log"
	if _, err := os.Stat(filePath); err == nil {
		content, errR := os.ReadFile(filePath)
		if errR != nil {
			log.Error().Err(errR).Send()
		}
		log.Print(string(content))
	}
	log.Print()
	log.Print("################################")
	log.Print()
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
		log.Info().Msg("No Traefik logs.")
	} else {
		for _, line := range strings.Split(output.String(), "\n") {
			log.Info().Msg(line)
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

	s.T().Cleanup(func() {
		os.Remove(tmpFile.Name())
	})
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
			log.Error().Err(err).Send()
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
