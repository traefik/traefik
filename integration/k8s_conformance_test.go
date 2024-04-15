package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/version"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/sets"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatev1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	conformanceV1alpha1 "sigs.k8s.io/gateway-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	ksuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
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

	k3sContainerIP string
	kubeClient     client.Client
	clientSet      *kclientset.Clientset
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

	ctx := context.Background()

	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		s.T().Fatal(err)
	}

	// ensure image is available locally
	images, err := provider.ListImages(ctx)
	if err != nil {
		s.T().Fatal(err)
	}

	if !slices.ContainsFunc(images, func(img testcontainers.ImageInfo) bool {
		return img.Name == traefikImage
	}) {
		s.T().Fatal("Traefik image is not present")
	}

	k3sContainer, err := k3s.RunContainer(ctx,
		testcontainers.WithImage(k3sImage),
		k3s.WithManifest("./fixtures/k8s-conformance/00-experimental-v1.0.0.yml"),
		k3s.WithManifest("./fixtures/k8s-conformance/01-rbac.yml"),
		k3s.WithManifest("./fixtures/k8s-conformance/02-traefik.yml"),
	)
	if err != nil {
		s.T().Fatal(err)
	}

	s.T().Cleanup(func() {
		if s.T().Failed() || *showLog {
			k3sLogs, err := k3sContainer.Logs(context.Background())
			if err == nil {
				if res, err := io.ReadAll(k3sLogs); err == nil {
					s.T().Log(string(res))
				}
			}

			exitCode, result, err := k3sContainer.Exec(context.Background(), []string{"kubectl", "logs", "-n", traefikNamespace, traefikDeployment})
			if err == nil || exitCode == 0 {
				if res, err := io.ReadAll(result); err == nil {
					s.T().Log(string(res))
				}
			}
		}

		err = k3sContainer.Terminate(context.Background())
		if err != nil {
			s.T().Fatal(err)
		}
	})

	s.k3sContainerIP, err = k3sContainer.ContainerIP(ctx)
	if err != nil {
		s.T().Fatal(err)
	}

	err = k3sContainer.LoadImages(ctx, traefikImage)
	if err != nil {
		s.T().Fatal(err)
	}

	exitCode, _, err := k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", traefikNamespace, traefikDeployment, "--for=condition=Available", "--timeout=30s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Traefik pod is not ready: %v", err)
	}

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
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

	err = gatev1alpha2.AddToScheme(s.kubeClient.Scheme())
	if err != nil {
		s.T().Fatal(err)
	}

	err = gatev1beta1.AddToScheme(s.kubeClient.Scheme())
	if err != nil {
		s.T().Fatal(err)
	}

	err = gatev1.AddToScheme(s.kubeClient.Scheme())
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *K8sConformanceSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *K8sConformanceSuite) TestK8sGatewayAPIConformance() {
	// Wait for traefik to start
	err := try.GetRequest("http://"+s.k3sContainerIP+":8080/api/entrypoints", 10*time.Second, try.BodyContains(`"name":"web"`))
	require.NoError(s.T(), err)

	opts := ksuite.Options{
		Client:               s.kubeClient,
		Clientset:            s.clientSet,
		GatewayClassName:     "traefik",
		Debug:                true,
		CleanupBaseResources: true,
		TimeoutConfig: config.TimeoutConfig{
			CreateTimeout:                     5 * time.Second,
			DeleteTimeout:                     5 * time.Second,
			GetTimeout:                        5 * time.Second,
			GatewayMustHaveAddress:            5 * time.Second,
			GatewayMustHaveCondition:          5 * time.Second,
			GatewayStatusMustHaveListeners:    10 * time.Second,
			GatewayListenersMustHaveCondition: 5 * time.Second,
			GWCMustBeAccepted:                 60 * time.Second, // Pod creation in k3s cluster can be long.
			HTTPRouteMustNotHaveParents:       5 * time.Second,
			HTTPRouteMustHaveCondition:        5 * time.Second,
			TLSRouteMustHaveCondition:         5 * time.Second,
			RouteMustHaveParents:              5 * time.Second,
			ManifestFetchTimeout:              5 * time.Second,
			MaxTimeToConsistency:              5 * time.Second,
			NamespacesMustBeReady:             60 * time.Second, // Pod creation in k3s cluster can be long.
			RequestTimeout:                    5 * time.Second,
			LatestObservedGenerationSet:       5 * time.Second,
			RequiredConsecutiveSuccesses:      0,
		},
		SupportedFeatures: sets.New[ksuite.SupportedFeature]().
			Insert(ksuite.GatewayCoreFeatures.UnsortedList()...).
			Insert(ksuite.ReferenceGrantCoreFeatures.UnsortedList()...),
		EnableAllSupportedFeatures: false,
		RunTest:                    *k8sConformanceRunTest,
		// Until the feature are all supported, following tests are skipped.
		SkipTests: []string{
			"HTTPExactPathMatching",
			"HTTPRouteHostnameIntersection",
			"HTTPRouteListenerHostnameMatching",
			"HTTPRouteRequestHeaderModifier",
			"GatewayClassObservedGenerationBump",
			"HTTPRouteInvalidNonExistentBackendRef",
			"GatewayWithAttachedRoutes",
			"HTTPRouteCrossNamespace",
			"HTTPRouteDisallowedKind",
			"HTTPRouteInvalidReferenceGrant",
			"HTTPRouteObservedGenerationBump",
			"TLSRouteSimpleSameNamespace",
			"TLSRouteInvalidReferenceGrant",
			"HTTPRouteInvalidCrossNamespaceParentRef",
			"HTTPRouteInvalidParentRefNotMatchingSectionName",
			"GatewayModifyListeners",
			"GatewayInvalidTLSConfiguration",
			"HTTPRouteInvalidCrossNamespaceBackendRef",
			"HTTPRouteMatchingAcrossRoutes",
			"HTTPRoutePartiallyInvalidViaInvalidReferenceGrant",
			"HTTPRouteRedirectHostAndStatus",
			"HTTPRouteInvalidBackendRefUnknownKind",
			"HTTPRoutePathMatchOrder",
			"HTTPRouteSimpleSameNamespace",
			"HTTPRouteMatching",
			"HTTPRouteHeaderMatching",
			"HTTPRouteReferenceGrant",
		},
	}

	cSuite, err := ksuite.NewExperimentalConformanceTestSuite(ksuite.ExperimentalConformanceOptions{
		Options: opts,
		Implementation: conformanceV1alpha1.Implementation{
			Organization: "traefik",
			Project:      "traefik",
			URL:          "https://traefik.io/",
			Version:      version.Version,
			Contact:      []string{"@traefik/maintainers"},
		},
		ConformanceProfiles: sets.New[ksuite.ConformanceProfileName](
			ksuite.HTTPConformanceProfileName,
			ksuite.TLSConformanceProfileName,
		),
	})
	require.NoError(s.T(), err)

	cSuite.Setup(s.T())
	err = cSuite.Run(s.T(), tests.ConformanceTests)
	require.NoError(s.T(), err)

	report, err := cSuite.Report()
	require.NoError(s.T(), err, "failed generating conformance report")

	report.GatewayAPIVersion = "1.0.0"

	rawReport, err := yaml.Marshal(report)
	require.NoError(s.T(), err)
	s.T().Logf("Conformance report:\n%s", string(rawReport))

	require.NoError(s.T(), os.MkdirAll("./conformance-reports", 0o755))
	outFile := filepath.Join("conformance-reports", fmt.Sprintf("traefik-traefik-%d.yaml", time.Now().UnixNano()))
	require.NoError(s.T(), os.WriteFile(outFile, rawReport, 0o600))
	s.T().Logf("Report written to: %s", outFile)
}
