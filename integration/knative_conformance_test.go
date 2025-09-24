// Use a build tag to include and run Knative conformance tests.
// The Knative conformance toolkit redefines the skip-tests flag,
// which conflicts with the testing library and causes a panic.
//go:build knativeConformance

package integration

import (
	"context"
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
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/networking/test/conformance/ingress"
	"sigs.k8s.io/controller-runtime/pkg/client"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	knativeNamespace        = "knative-serving"
	knativeActivator        = "deployment/activator"
	knativeAutoscaler       = "deployment/autoscaler"
	knativeController       = "deployment/controller"
	knativeWebhook          = "deployment/webhook"
	knativeNetworkConfigMap = "configmap/config-network"
	knativeDomainConfigMap  = "configmap/config-domain"
	knativeSkipTests        = "visibility/split,visibility/path,visibility,update,headers/probe,hosts/multiple,tls"
)

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
	kubeClient   client.Client
	restConfig   *rest.Config
	clientSet    *kclientset.Clientset
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
		k3s.WithManifest("./fixtures/knative/00-knative-crd-v1.17.0.yml"),
		k3s.WithManifest("./fixtures/knative/01-rbac.yml"),
		k3s.WithManifest("./fixtures/knative/02-traefik.yml"),
		k3s.WithManifest("./fixtures/knative/03-knative-serving-v1.17.0.yaml"),
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

	exitCode, _, err := s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", traefikNamespace, traefikDeployment, "--for=condition=Available", "--timeout=6000s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Traefik pod is not ready: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "patch", "-n", knativeNamespace, knativeNetworkConfigMap, "--type", "merge", "-p", `{"data":{"ingress.class":"traefik.ingress.networking.knative.dev"}}`})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Failed to update network config map: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "patch", "-n", knativeNamespace, knativeDomainConfigMap, "--type", "merge", "-p", `{"data":{"example.com":""}}`})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Failed to update network config map: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", knativeNamespace, knativeActivator, "--for=condition=Available", "--timeout=30s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Activator pod is not ready: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", knativeNamespace, knativeController, "--for=condition=Available", "--timeout=30s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Controller pod is not ready: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", knativeNamespace, knativeAutoscaler, "--for=condition=Available", "--timeout=30s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Autoscaler pod is not ready: %v", err)
	}

	exitCode, _, err = s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", knativeNamespace, knativeWebhook, "--for=condition=Available", "--timeout=30s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Webhook pod is not ready: %v", err)
	}
	kubeConfigYaml, err := s.k3sContainer.GetKubeConfig(ctx)
	if err != nil {
		s.T().Fatal(err)
	}

	s.restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		s.T().Fatalf("Error loading Kubernetes config: %v", err)
	}

	s.kubeClient, err = client.New(s.restConfig, client.Options{})
	if err != nil {
		s.T().Fatalf("Error initializing Kubernetes client: %v", err)
	}

	s.clientSet, err = kclientset.NewForConfig(s.restConfig)
	if err != nil {
		s.T().Fatalf("Error initializing Kubernetes REST client: %v", err)
	}
}

func (s *KnativeConformanceSuite) TearDownSuite() {
	ctx := context.Background()

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
	k3sContainerIP, err := s.k3sContainer.ContainerIP(context.Background())
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+k3sContainerIP+":9000/api/entrypoints", 10*time.Second, try.BodyContains(`"name":"web"`))
	require.NoError(s.T(), err)

	config, err := s.k3sContainer.GetKubeConfig(context.Background())
	if err != nil {
		s.T().Fatal(err)
	}

	// Ensure the directory exists

	homeDir, err := os.UserHomeDir()
	if err != nil {
		s.T().Fatal(err)
	}

	err = os.MkdirAll(homeDir+"/work/traefik/k3s/", 0o755)
	if err != nil {
		s.T().Fatal(err)
	}

	// Write the byte array to the file
	err = os.WriteFile(homeDir+"/work/traefik/k3s/k3s.yaml", config, 0o644)
	if err != nil {
		s.T().Fatal(err)
	}

	err = flag.CommandLine.Set("kubeconfig", homeDir+"/work/traefik/k3s/k3s.yaml")
	if err != nil {
		s.T().Fatal(err)
	}

	err = flag.CommandLine.Set("ingressendpoint", k3sContainerIP)
	if err != nil {
		s.T().Fatal(err)
	}

	err = flag.CommandLine.Set("skip-tests", knativeSkipTests)
	if err != nil {
		s.T().Fatal(err)
	}

	ingress.RunConformance(s.T())
}
