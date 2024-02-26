package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/version"
	"gopkg.in/yaml.v3"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatev1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	conformanceV1alpha1 "sigs.k8s.io/gateway-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	ksuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
)

// K8sConformanceSuite tests suite.
type K8sConformanceSuite struct{ BaseSuite }

func TestK8sConformanceSuite(t *testing.T) {
	suite.Run(t, new(K8sConformanceSuite))
}

func (s *K8sConformanceSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("k8s")
	s.composeUp()

	abs, err := filepath.Abs("./fixtures/k8s/config.skip/kubeconfig.yaml")
	require.NoError(s.T(), err)

	err = try.Do(60*time.Second, func() error {
		_, err := os.Stat(abs)
		return err
	})
	require.NoError(s.T(), err)

	data, err := os.ReadFile(abs)
	require.NoError(s.T(), err)

	content := strings.ReplaceAll(string(data), "https://server:6443", fmt.Sprintf("https://%s", net.JoinHostPort(s.getComposeServiceIP("server"), "6443")))

	err = os.WriteFile(abs, []byte(content), 0o644)
	require.NoError(s.T(), err)

	err = os.Setenv("KUBECONFIG", abs)
	require.NoError(s.T(), err)
}

func (s *K8sConformanceSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()

	generatedFiles := []string{
		"./fixtures/k8s/config.skip/kubeconfig.yaml",
		"./fixtures/k8s/config.skip/k3s.log",
		"./fixtures/k8s/rolebindings.yaml",
		"./fixtures/k8s/ccm.yaml",
	}

	for _, filename := range generatedFiles {
		if err := os.Remove(filename); err != nil {
			log.Warn().Err(err).Send()
		}
	}
}

func (s *K8sConformanceSuite) TestK8sGatewayAPIConformance() {
	if !*k8sConformance {
		s.T().Skip("Skip because it can take a long time to execute. To enable pass the `k8sConformance` flag.")
	}

	configFromFlags, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		s.T().Fatal(err)
	}

	kClient, err := client.New(configFromFlags, client.Options{})
	if err != nil {
		s.T().Fatalf("Error initializing Kubernetes client: %v", err)
	}

	kClientSet, err := kclientset.NewForConfig(configFromFlags)
	if err != nil {
		s.T().Fatal(err)
	}

	err = gatev1alpha2.AddToScheme(kClient.Scheme())
	require.NoError(s.T(), err)
	err = gatev1beta1.AddToScheme(kClient.Scheme())
	require.NoError(s.T(), err)
	err = gatev1.AddToScheme(kClient.Scheme())
	require.NoError(s.T(), err)

	s.traefikCmd(withConfigFile("fixtures/k8s_gateway_conformance.toml"))

	// Wait for traefik to start
	err = try.GetRequest("http://127.0.0.1:8080/api/entrypoints", 10*time.Second, try.BodyContains(`"name":"web"`))
	require.NoError(s.T(), err)

	err = try.Do(10*time.Second, func() error {
		gwc := &gatev1.GatewayClass{}
		err := kClient.Get(context.Background(), ktypes.NamespacedName{Name: "my-gateway-class"}, gwc)
		if err != nil {
			return fmt.Errorf("error fetching GatewayClass: %w", err)
		}

		return nil
	})
	require.NoError(s.T(), err)

	opts := ksuite.Options{
		Client:               kClient,
		RestConfig:           configFromFlags,
		Clientset:            kClientSet,
		GatewayClassName:     "my-gateway-class",
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
