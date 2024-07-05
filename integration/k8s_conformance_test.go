package integration

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/version"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatev1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance"
	v1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	ksuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
	"sigs.k8s.io/yaml"
)

const (
	k3sImage          = "docker.io/rancher/k3s:v1.29.3-k3s1"
	traefikImage      = "traefik/traefik:latest"
	traefikDeployment = "deployments/traefik"
	traefikNamespace  = "traefik"
)

// K8sConformanceSuite tests suite.
type K8sConformanceSuite struct {
	BaseSuite

	k3sContainer *k3s.K3sContainer
	kubeClient   client.Client
	clientSet    *kclientset.Clientset
}

func TestK8sConformanceSuite(t *testing.T) {
	suite.Run(t, new(K8sConformanceSuite))
}

func (s *K8sConformanceSuite) SetupSuite() {
	if !*k8sConformance {
		s.T().Skip("Skip because it can take a long time to execute. To enable pass the `k8sConformance` flag.")
	}

	s.BaseSuite.SetupSuite()

	// Avoid panic.
	klog.SetLogger(zap.New())

	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		s.T().Fatal(err)
	}

	ctx := context.Background()

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

	s.k3sContainer, err = k3s.RunContainer(ctx,
		testcontainers.WithImage(k3sImage),
		k3s.WithManifest("./fixtures/k8s-conformance/00-experimental-v1.1.0.yml"),
		k3s.WithManifest("./fixtures/k8s-conformance/01-rbac.yml"),
		k3s.WithManifest("./fixtures/k8s-conformance/02-traefik.yml"),
		network.WithNetwork(nil, s.network),
	)
	if err != nil {
		s.T().Fatal(err)
	}

	if err = s.k3sContainer.LoadImages(ctx, traefikImage); err != nil {
		s.T().Fatal(err)
	}

	exitCode, _, err := s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", traefikNamespace, traefikDeployment, "--for=condition=Available", "--timeout=30s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Traefik pod is not ready: %v", err)
	}

	kubeConfigYaml, err := s.k3sContainer.GetKubeConfig(ctx)
	if err != nil {
		s.T().Fatal(err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		s.T().Fatalf("Error loading Kubernetes config: %v", err)
	}

	s.kubeClient, err = client.New(restConfig, client.Options{})
	if err != nil {
		s.T().Fatalf("Error initializing Kubernetes client: %v", err)
	}

	s.clientSet, err = kclientset.NewForConfig(restConfig)
	if err != nil {
		s.T().Fatalf("Error initializing Kubernetes REST client: %v", err)
	}

	if err = gatev1alpha2.Install(s.kubeClient.Scheme()); err != nil {
		s.T().Fatal(err)
	}

	if err = gatev1beta1.Install(s.kubeClient.Scheme()); err != nil {
		s.T().Fatal(err)
	}

	if err = gatev1.Install(s.kubeClient.Scheme()); err != nil {
		s.T().Fatal(err)
	}

	if err = apiextensionsv1.AddToScheme(s.kubeClient.Scheme()); err != nil {
		s.T().Fatal(err)
	}
}

func (s *K8sConformanceSuite) TearDownSuite() {
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

func (s *K8sConformanceSuite) TestK8sGatewayAPIConformance() {
	// Wait for traefik to start
	k3sContainerIP, err := s.k3sContainer.ContainerIP(context.Background())
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+k3sContainerIP+":9000/api/entrypoints", 10*time.Second, try.BodyContains(`"name":"web"`))
	require.NoError(s.T(), err)

	cSuite, err := ksuite.NewConformanceTestSuite(ksuite.ConformanceOptions{
		Client:                     s.kubeClient,
		Clientset:                  s.clientSet,
		GatewayClassName:           "traefik",
		Debug:                      true,
		CleanupBaseResources:       true,
		TimeoutConfig:              config.DefaultTimeoutConfig(),
		ManifestFS:                 []fs.FS{&conformance.Manifests},
		EnableAllSupportedFeatures: false,
		RunTest:                    *k8sConformanceRunTest,
		Implementation: v1.Implementation{
			Organization: "traefik",
			Project:      "traefik",
			URL:          "https://traefik.io/",
			Version:      version.Version,
			Contact:      []string{"@traefik/maintainers"},
		},
		ConformanceProfiles: sets.New(ksuite.GatewayHTTPConformanceProfileName),
		SupportedFeatures: sets.New(
			features.SupportGateway,
			features.SupportGatewayPort8080,
			features.SupportHTTPRoute,
			features.SupportHTTPRouteQueryParamMatching,
			features.SupportHTTPRouteMethodMatching,
			features.SupportHTTPRoutePortRedirect,
			features.SupportHTTPRouteSchemeRedirect,
			features.SupportHTTPRouteHostRewrite,
			features.SupportHTTPRoutePathRewrite,
			features.SupportHTTPRoutePathRedirect,
		),
		ExemptFeatures: sets.New(
			features.SupportHTTPRouteRequestTimeout,
			features.SupportHTTPRouteBackendTimeout,
			features.SupportHTTPRouteResponseHeaderModification,
			features.SupportHTTPRouteRequestMirror,
			features.SupportHTTPRouteRequestMultipleMirrors,
		),
	})
	require.NoError(s.T(), err)

	cSuite.Setup(s.T(), tests.ConformanceTests)

	err = cSuite.Run(s.T(), tests.ConformanceTests)
	require.NoError(s.T(), err)

	report, err := cSuite.Report()
	require.NoError(s.T(), err, "failed generating conformance report")

	rawReport, err := yaml.Marshal(report)
	require.NoError(s.T(), err)
	s.T().Logf("Conformance report:\n%s", string(rawReport))

	require.NoError(s.T(), os.MkdirAll("./conformance-reports", 0o755))
	outFile := filepath.Join("conformance-reports", fmt.Sprintf("%s-%s-%s-report.yaml", report.GatewayAPIChannel, report.Version, report.Mode))
	require.NoError(s.T(), os.WriteFile(outFile, rawReport, 0o600))
	s.T().Logf("Report written to: %s", outFile)
}
