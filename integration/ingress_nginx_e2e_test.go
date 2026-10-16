// Use a build tag to include and run the ingress-nginx e2e tests,
// they live in a dedicated Go module that is built and run inside the k3s node.
//go:build ingressNGINXE2E

package integration

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
)

const ingressNGINXE2EDir = "./ingress-nginx-e2e"

// IngressNGINXE2ESuite runs the vendored ingress-nginx e2e annotation tests against the kubernetesIngressNGINX provider.
// The tests run on the k3s node because, like upstream, they reach the controller through its Service ClusterIP.
//
// Environment variables:
//   - FOCUS: ginkgo focus regex.
//   - SKIP: ginkgo skip regex, added to the known failures.
//   - E2E_NODES: number of parallel ginkgo processes (default 4).
//
// It does not extend BaseSuite, which requires the containers to be reachable from the host.
type IngressNGINXE2ESuite struct {
	suite.Suite

	k3sContainer *k3s.K3sContainer
	binDir       string
}

func TestIngressNGINXE2ESuite(t *testing.T) {
	suite.Run(t, new(IngressNGINXE2ESuite))
}

func (s *IngressNGINXE2ESuite) SetupSuite() {
	ctx := s.T().Context()

	provider, err := testcontainers.ProviderDocker.GetProvider()
	require.NoError(s.T(), err)

	images, err := provider.ListImages(ctx)
	require.NoError(s.T(), err)

	if !slices.ContainsFunc(images, func(img testcontainers.ImageInfo) bool {
		return img.Name == traefikImage
	}) {
		s.T().Fatal("Traefik image is not present")
	}

	s.binDir = s.T().TempDir()
	s.buildLinux("go", "test", "-c", "-o", filepath.Join(s.binDir, "e2e.test"), "./test/e2e")
	s.buildLinux("go", "build", "-o", filepath.Join(s.binDir, "ginkgo"), "github.com/onsi/ginkgo/v2/ginkgo")

	s.k3sContainer, err = k3s.Run(ctx, k3sImage)
	require.NoError(s.T(), err)

	require.NoError(s.T(), s.k3sContainer.LoadImages(ctx, traefikImage))

	for _, bin := range []string{"e2e.test", "ginkgo"} {
		require.NoError(s.T(), s.k3sContainer.CopyFileToContainer(ctx, filepath.Join(s.binDir, bin), "/e2e/"+bin, 0o755))
	}
}

func (s *IngressNGINXE2ESuite) TearDownSuite() {
	if s.k3sContainer != nil {
		if err := s.k3sContainer.Terminate(s.T().Context()); err != nil {
			s.T().Error(err)
		}
	}
}

func (s *IngressNGINXE2ESuite) TestAnnotations() {
	nodes := os.Getenv("E2E_NODES")
	if nodes == "" {
		nodes = "4"
	}

	skips := s.knownFailures()
	if skip := os.Getenv("SKIP"); skip != "" {
		skips = append(skips, skip)
	}

	args := []string{
		"exec",
		"-e", "KUBECONFIG=/etc/rancher/k3s/k3s.yaml",
		"-e", "KUBECTL_PATH=/bin/kubectl",
		// Backend images, read from the files upstream uses for the same purpose.
		"-e", "HTTPBUN_IMAGE=" + s.readImage("test/e2e/HTTPBUN_IMAGE"),
		"-e", "NGINX_BASE_IMAGE=" + s.readImage("NGINX_BASE"),
		s.k3sContainer.GetContainerID(),
		"/e2e/ginkgo",
		"--procs=" + nodes,
		"--timeout=2h",
		"--poll-progress-after=5m",
		"-v",
	}
	if focus := os.Getenv("FOCUS"); focus != "" {
		args = append(args, "--focus="+focus)
	}
	if len(skips) > 0 {
		args = append(args, "--skip="+strings.Join(skips, "|"))
	}
	args = append(args, "/e2e/e2e.test")

	// Docker is used directly, rather than the container Exec, to stream the output of the long-running suite.
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	require.NoError(s.T(), cmd.Run())
}

// buildLinux builds a binary of the ingress-nginx e2e module for the k3s node.
func (s *IngressNGINXE2ESuite) buildLinux(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = ingressNGINXE2EDir
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH, "CGO_ENABLED=0")

	out, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, string(out))
}

func (s *IngressNGINXE2ESuite) readImage(path string) string {
	image, err := os.ReadFile(filepath.Join(ingressNGINXE2EDir, path))
	require.NoError(s.T(), err)

	return strings.TrimSpace(string(image))
}

// knownFailures returns the skip regexes listed in known-failures.txt.
func (s *IngressNGINXE2ESuite) knownFailures() []string {
	file, err := os.Open(filepath.Join(ingressNGINXE2EDir, "known-failures.txt"))
	require.NoError(s.T(), err)
	defer file.Close()

	var skips []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		skips = append(skips, line)
	}
	require.NoError(s.T(), scanner.Err())

	return skips
}
