//go:build gatewayAPIConformanceMerged

package integration

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/gateway"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance"
	v1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	ksuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/yaml"
)

// GatewayAPIConformanceMergedSuite runs the Gateway API conformance suite
// against a single, statically deployed Traefik instance serving every Gateway.
//
// It is the counterpart to the operator-provisioned per-Gateway data planes the
// GatewayAPIConformanceSuite exercises: here one Traefik merges all Gateways
// behind a single address, so the features a single instance cannot satisfy
// (per-Gateway addresses, GatewayStaticAddresses, GatewayInfrastructurePropagation,
// and the multiple-Gateways test) are left out.
type GatewayAPIConformanceMergedSuite struct {
	BaseSuite

	k3sContainer *k3s.K3sContainer
	kubeClient   client.Client
	restConfig   *rest.Config
	clientSet    *kclientset.Clientset
}

func TestGatewayAPIConformanceMergedSuite(t *testing.T) {
	suite.Run(t, new(GatewayAPIConformanceMergedSuite))
}

func (s *GatewayAPIConformanceMergedSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	// Avoid panic.
	klog.SetLogger(zap.New())

	ctx := s.T().Context()

	provider, err := testcontainers.ProviderDocker.GetProvider()
	require.NoError(s.T(), err)

	// Ensure image is available locally.
	images, err := provider.ListImages(ctx)
	require.NoError(s.T(), err)

	if !slices.ContainsFunc(images, func(img testcontainers.ImageInfo) bool {
		return img.Name == traefikImage
	}) {
		s.T().Fatal("Traefik image is not present")
	}

	s.k3sContainer, err = k3s.Run(
		ctx,
		k3sImage,
		k3s.WithManifest("./fixtures/gateway-api-conformance/00-experimental-v1.6.1.yml"),
		k3s.WithManifest("./fixtures/gateway-api-conformance/merged/01-rbac.yml"),
		k3s.WithManifest("./fixtures/gateway-api-conformance/merged/02-traefik.yml"),
		network.WithNetwork(nil, s.network),
	)
	require.NoError(s.T(), err)

	require.NoError(s.T(), s.k3sContainer.LoadImages(ctx, traefikImage))

	exitCode, _, err := s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", traefikNamespace, traefikDeployment, "--for=condition=Available", "--timeout=30s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Traefik pod is not ready: %v", err)
	}

	kubeConfigYaml, err := s.k3sContainer.GetKubeConfig(ctx)
	require.NoError(s.T(), err)

	s.restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	require.NoError(s.T(), err)

	s.kubeClient, err = client.New(s.restConfig, client.Options{})
	require.NoError(s.T(), err)

	s.clientSet, err = kclientset.NewForConfig(s.restConfig)
	require.NoError(s.T(), err)

	require.NoError(s.T(), gatev1.Install(s.kubeClient.Scheme()))
	require.NoError(s.T(), apiextensionsv1.AddToScheme(s.kubeClient.Scheme()))
}

func (s *GatewayAPIConformanceMergedSuite) TearDownSuite() {
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

	require.NoError(s.T(), s.k3sContainer.Terminate(ctx))

	s.BaseSuite.TearDownSuite()
}

func (s *GatewayAPIConformanceMergedSuite) TestK8sGatewayAPIConformanceMerged() {
	// Wait for traefik to start
	k3sContainerIP, err := s.k3sContainer.ContainerIP(s.T().Context())
	require.NoError(s.T(), err)

	err = try.GetRequest("http://"+k3sContainerIP+":9000/api/entrypoints", 10*time.Second, try.BodyContains(`"name":"web"`))
	require.NoError(s.T(), err)

	// Traefik reconciles a resource in a couple of seconds or less.
	// They are shortened for a status Traefik will never report to fail before
	// the test binary timeout, which would discard the whole run and its report.
	timeoutConfig := config.DefaultTimeoutConfig()
	timeoutConfig.GatewayMustHaveAddress = 60 * time.Second
	timeoutConfig.GatewayMustHaveCondition = 60 * time.Second
	timeoutConfig.GWCMustBeAccepted = 60 * time.Second
	timeoutConfig.ListenerSetMustHaveCondition = 60 * time.Second
	timeoutConfig.NamespacesMustBeReady = 60 * time.Second

	cSuite, err := ksuite.NewConformanceTestSuite(ksuite.ConformanceOptions{
		Client:     s.kubeClient,
		Clientset:  s.clientSet,
		RestConfig: s.restConfig,
		ManifestFS: []fs.FS{&conformance.Manifests},
		ConfigurableOptions: ksuite.ConfigurableOptions{
			// A single instance serves every Gateway, so the report is written
			// under a dedicated mode to avoid clobbering the per-Gateway
			// operator report.
			Mode:                       "merged",
			GatewayClassName:           "traefik",
			Debug:                      true,
			CleanupBaseResources:       true,
			CleanupTestResources:       true,
			TimeoutConfig:              timeoutConfig,
			EnableAllSupportedFeatures: false,
			RunTest:                    *gatewayAPIConformanceRunTest,
			Implementation: v1.Implementation{
				Organization: "traefik",
				Project:      "traefik",
				URL:          "https://traefik.io/",
				Version:      *traefikVersion,
				Contact:      []string{"@traefik/maintainers"},
			},
			ConformanceProfiles: []ksuite.ConformanceProfileName{
				ksuite.GatewayHTTPConformanceProfileName,
				ksuite.GatewayGRPCConformanceProfileName,
				ksuite.GatewayTLSConformanceProfileName,
			},
			SupportedFeatures: gateway.SupportedFeatures(),
			// The following tests are skipped because they require features that a single Traefik instance
			// cannot satisfy in merged mode cause they enforce having an operator.
			SkipTests: []string{
				tests.HTTPRouteMultipleGateways.ShortName,
				tests.TLSRouteHostnameIntersection.ShortName,
			},
		},
	})
	require.NoError(s.T(), err)

	cSuite.Setup(s.T(), tests.ConformanceTests)

	err = cSuite.Run(s.T(), tests.ConformanceTests)
	require.NoError(s.T(), err)

	report, err := cSuite.Report()
	require.NoError(s.T(), err, "failed generating conformance report")

	// Ordering profile reports for the serialized report to be comparable.
	slices.SortFunc(report.ProfileReports, func(a, b v1.ProfileReport) int {
		return strings.Compare(a.Name, b.Name)
	})

	rawReport, err := yaml.Marshal(report)
	require.NoError(s.T(), err)
	s.T().Logf("Conformance report:\n%s", string(rawReport))

	require.NoError(s.T(), os.MkdirAll("./gateway-api-conformance-reports/"+report.GatewayAPIVersion, 0o755))
	outFile := filepath.Join("gateway-api-conformance-reports/"+report.GatewayAPIVersion, fmt.Sprintf("%s-%s-%s-report.yaml", report.GatewayAPIChannel, report.Version, report.Mode))
	require.NoError(s.T(), os.WriteFile(outFile, rawReport, 0o600))
	s.T().Logf("Report written to: %s", outFile)
}
