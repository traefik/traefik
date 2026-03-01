// Use a build tag to include and run Knative conformance tests.
// The Knative conformance toolkit redefines the skip-tests flag,
// which conflicts with the testing library and causes a panic.
//go:build knativeConformance

package integration

import (
	"flag"
	"io"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/traefik/traefik/v3/integration/try"
	"knative.dev/networking/test/conformance/ingress"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const knativeNamespace = "knative-serving"

var imageNames = []string{
	traefikImage,
	"ko.local/grpc-ping:latest",
	"ko.local/httpproxy:latest",
	"ko.local/retry:latest",
	"ko.local/runtime:latest",
	"ko.local/wsserver:latest",
	"ko.local/timeout:latest",
}

type KnativeConformanceSuite struct {
	BaseSuite

	k3sContainer *k3s.K3sContainer
}

func TestKnativeConformanceSuite(t *testing.T) {
	suite.Run(t, new(KnativeConformanceSuite))
}

func (s *KnativeConformanceSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	// Avoid panic.
	klog.SetLogger(zap.New())

	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		s.T().Fatal(err)
	}

	ctx := s.T().Context()

	// Ensure image is available locally.
	images, err := provider.ListImages(ctx)
	if err != nil {
		s.T().Fatal(err)
	}

	if !slices.ContainsFunc(images, func(img testcontainers.ImageInfo) bool {
		return img.Name == traefikImage
	}) {
		s.T().Fatal("Traefik image is not present")
	}

	s.k3sContainer, err = k3s.Run(ctx,
		k3sImage,
		k3s.WithManifest("./fixtures/knative/00-knative-crd-v1.19.0.yml"),
		k3s.WithManifest("./fixtures/knative/01-rbac.yml"),
		k3s.WithManifest("./fixtures/knative/02-traefik.yml"),
		k3s.WithManifest("./fixtures/knative/03-knative-serving-v1.19.0.yaml"),
		k3s.WithManifest("./fixtures/knative/04-serving-tests-namespace.yaml"),
		network.WithNetwork(nil, s.network),
	)
	if err != nil {
		s.T().Fatal(err)
	}

	for _, imageName := range imageNames {
		if err = s.k3sContainer.LoadImages(ctx, imageName); err != nil {
			s.T().Fatal(err)
		}
	}

	exitCode, _, err := s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", traefikNamespace, traefikDeployment, "--for=condition=Available", "--timeout=10s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Traefik pod is not ready: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", knativeNamespace, "deployment/activator", "--for=condition=Available", "--timeout=10s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Activator pod is not ready: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", knativeNamespace, "deployment/controller", "--for=condition=Available", "--timeout=10s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Controller pod is not ready: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", knativeNamespace, "deployment/autoscaler", "--for=condition=Available", "--timeout=10s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Autoscaler pod is not ready: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", knativeNamespace, "deployment/webhook", "--for=condition=Available", "--timeout=10s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Webhook pod is not ready: %v", err)
	}
}

func (s *KnativeConformanceSuite) TearDownSuite() {
	ctx := s.T().Context()

	if s.T().Failed() || *showLog {
		k3sLogs, err := s.k3sContainer.Logs(ctx)
		if err == nil {
			if res, err := io.ReadAll(k3sLogs); err == nil {
				s.T().Log(string(res))
			}
		}

		exitCode, result, err := s.k3sContainer.Exec(ctx, []string{"kubectl", "logs", "-n", traefikNamespace, traefikDeployment})
		if err == nil || exitCode == 0 {
			if res, err := io.ReadAll(result); err == nil {
				s.T().Log(string(res))
			}
		}
	}

	if err := s.k3sContainer.Terminate(ctx); err != nil {
		s.T().Fatal(err)
	}

	s.BaseSuite.TearDownSuite()
}

func (s *KnativeConformanceSuite) TestKnativeConformance() {
	// Wait for traefik to start
	k3sContainerIP, err := s.k3sContainer.ContainerIP(s.T().Context())
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+k3sContainerIP+":9000/api/entrypoints", 10*time.Second, try.BodyContains(`"name":"pweb"`))
	require.NoError(s.T(), err)

	kubeconfig, err := s.k3sContainer.GetKubeConfig(s.T().Context())
	if err != nil {
		s.T().Fatal(err)
	}

	// Write the kubeconfig.yaml in a temp file.
	kubeconfigFile := s.T().TempDir() + "/kubeconfig.yaml"

	if err = os.WriteFile(kubeconfigFile, kubeconfig, 0o644); err != nil {
		s.T().Fatal(err)
	}

	if err = flag.CommandLine.Set("kubeconfig", kubeconfigFile); err != nil {
		s.T().Fatal(err)
	}

	if err = flag.CommandLine.Set("ingressClass", "traefik.ingress.networking.knative.dev"); err != nil {
		s.T().Fatal(err)
	}

	if err = flag.CommandLine.Set("skip-tests", "headers/probe"); err != nil {
		s.T().Fatal(err)
	}

	ingress.RunConformance(s.T())
}
