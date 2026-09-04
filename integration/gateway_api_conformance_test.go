//go:build gatewayAPIConformance

package integration

import (
	"context"
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
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/gateway"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	// traefikConformanceImage tags the image built from the working tree for
	// the data plane pods. The tag is fixed, and not latest, so that the pods
	// keep the image side loaded into the node: a latest tag defaults
	// imagePullPolicy to Always, which pulls the released Traefik instead.
	traefikConformanceImage = "traefik/traefik:conformance"

	operatorImage      = "traefik/gateway-operator:latest"
	operatorNamespace  = "traefik-gateway-operator-system"
	operatorDeployment = "deployments/traefik-gateway-operator"
)

// GatewayAPIConformanceSuite runs the Gateway API conformance suite against the
// Traefik data planes the Traefik Gateway API operator provisions per Gateway.
//
// The operator owns the infrastructure part of the specification, which a
// single statically deployed instance cannot satisfy: a per-Gateway
// status.addresses, GatewayStaticAddresses, GatewayInfrastructurePropagation,
// and the tests needing two Gateways reachable at two distinct addresses.
type GatewayAPIConformanceSuite struct {
	BaseSuite

	k3sContainer *k3s.K3sContainer
	kubeClient   client.Client
	restConfig   *rest.Config
	clientSet    *kclientset.Clientset
	loadBalancer *nodeLoadBalancer

	cancelLoadBalancer context.CancelFunc
}

func TestGatewayAPIConformanceSuite(t *testing.T) {
	suite.Run(t, new(GatewayAPIConformanceSuite))
}

func (s *GatewayAPIConformanceSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	// Avoid panic.
	klog.SetLogger(zap.New())

	ctx := s.T().Context()

	provider, err := testcontainers.ProviderDocker.GetProvider()
	require.NoError(s.T(), err)

	// Ensure images are available locally: make test-gateway-api-conformance
	// builds the Traefik one and pulls the operator one.
	images, err := provider.ListImages(ctx)
	require.NoError(s.T(), err)

	for _, image := range []string{traefikConformanceImage, operatorImage} {
		if !slices.ContainsFunc(images, func(img testcontainers.ImageInfo) bool {
			return img.Name == image
		}) {
			s.T().Fatalf("Image %s is not present", image)
		}
	}

	s.k3sContainer, err = k3s.Run(
		ctx,
		k3sImage,
		// The k3s service load balancer exposes Services through host ports,
		// which a single node cannot do for the several port 80 Services the
		// operator provisions. nodeLoadBalancer assigns their addresses.
		testcontainers.WithCmdArgs("--disable=servicelb"),
		k3s.WithManifest("./fixtures/gateway-api-conformance/00-experimental-v1.6.1.yml"),
		k3s.WithManifest("./fixtures/gateway-api-conformance/01-operator.yml"),
		k3s.WithManifest("./fixtures/gateway-api-conformance/02-gatewayclass.yml"),
		k3s.WithManifest("./fixtures/gateway-api-conformance/03-bootstrap-gateway.yml"),
		network.WithNetwork(nil, s.network),
	)
	require.NoError(s.T(), err)

	require.NoError(s.T(), s.k3sContainer.LoadImages(ctx, traefikConformanceImage, operatorImage))

	exitCode, _, err := s.k3sContainer.Exec(ctx, []string{"kubectl", "wait", "-n", operatorNamespace, operatorDeployment, "--for=condition=Available", "--timeout=120s"})
	if err != nil || exitCode > 0 {
		s.T().Fatalf("Operator pod is not ready: %v", err)
	}

	kubeConfigYaml, err := s.k3sContainer.GetKubeConfig(ctx)
	require.NoError(s.T(), err)

	s.restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	require.NoError(s.T(), err)

	s.kubeClient, err = client.New(s.restConfig, client.Options{})
	require.NoError(s.T(), err)

	s.clientSet, err = kclientset.NewForConfig(s.restConfig)
	require.NoError(s.T(), err)

	require.NoError(s.T(), gatev1alpha2.Install(s.kubeClient.Scheme()))
	require.NoError(s.T(), gatev1beta1.Install(s.kubeClient.Scheme()))
	require.NoError(s.T(), gatev1.Install(s.kubeClient.Scheme()))
	require.NoError(s.T(), apiextensionsv1.AddToScheme(s.kubeClient.Scheme()))

	s.loadBalancer, err = newNodeLoadBalancer(ctx, s.k3sContainer, s.clientSet)
	require.NoError(s.T(), err)

	// The load balancer must keep reconciling across the whole suite, not just
	// for the lifetime of whichever test happens to be running when a Service
	// needs an address, so it gets a context of its own.
	lbCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	s.cancelLoadBalancer = cancel
	s.loadBalancer.Start(lbCtx)
}

func (s *GatewayAPIConformanceSuite) TearDownSuite() {
	ctx := s.T().Context()

	if s.cancelLoadBalancer != nil {
		s.cancelLoadBalancer()
	}

	if s.T().Failed() || *showLog {
		k3sLogs, err := s.k3sContainer.Logs(ctx)
		if err == nil {
			if res, err := io.ReadAll(k3sLogs); err == nil {
				s.T().Log(string(res))
			}
		}

		s.logCommand(ctx, "kubectl", "logs", "-n", operatorNamespace, operatorDeployment)

		// The data planes are spread over the conformance namespaces, and are
		// where a routing failure shows up.
		s.logCommand(ctx, "kubectl", "get", "gateways,pods,services", "--all-namespaces")
		s.logDataPlanes(ctx)
	}

	require.NoError(s.T(), s.k3sContainer.Terminate(ctx))

	s.BaseSuite.TearDownSuite()
}

func (s *GatewayAPIConformanceSuite) TestK8sGatewayAPIConformance() {
	// Provisioning a data plane adds a Deployment rollout to every Gateway
	// reconciliation, and the GatewayClass is only accepted once the first data
	// plane runs, so the timeouts are longer than for a statically deployed
	// instance. They stay shortened for a status Traefik will never report, to
	// fail before the test binary timeout, which would discard the whole run
	// and its report.
	timeoutConfig := config.DefaultTimeoutConfig()
	timeoutConfig.GatewayMustHaveAddress = 120 * time.Second
	timeoutConfig.GatewayMustHaveCondition = 120 * time.Second
	timeoutConfig.GWCMustBeAccepted = 120 * time.Second
	timeoutConfig.ListenerSetMustHaveCondition = 120 * time.Second
	timeoutConfig.NamespacesMustBeReady = 300 * time.Second

	cSuite, err := ksuite.NewConformanceTestSuite(ksuite.ConformanceOptions{
		Client:     s.kubeClient,
		Clientset:  s.clientSet,
		RestConfig: s.restConfig,
		ManifestFS: []fs.FS{&conformance.Manifests},
		ConfigurableOptions: ksuite.ConfigurableOptions{
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
			// The data plane is the single writer of the GatewayClass status, so
			// it only publishes its set once a Gateway, and the data plane
			// serving it, exist. Reading it back would come too early. The
			// operator adds GatewayInfrastructurePropagation, the one Gateway
			// feature it implements itself.
			//
			// GatewayStaticAddresses is deliberately left out: neither Traefik
			// nor the operator reject an unsupported spec.addresses entry or
			// report Programmed=False/AddressNotUsable for one that can't be
			// realized, so the test can never pass, only fail or hang.
			SupportedFeatures: slices.Concat(gateway.SupportedFeatures(), []features.FeatureName{
				features.SupportGatewayInfrastructurePropagation,
			}),
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

// logDataPlanes logs the data plane pods one by one: kubectl logs cannot select
// pods across namespaces.
func (s *GatewayAPIConformanceSuite) logDataPlanes(ctx context.Context) {
	if s.clientSet == nil {
		return
	}

	pods, err := s.clientSet.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/managed-by=traefik-gateway-operator",
	})
	if err != nil {
		s.T().Logf("listing data plane pods: %v", err)
		return
	}

	for _, pod := range pods.Items {
		s.logCommand(ctx, "kubectl", "logs", "-n", pod.Namespace, pod.Name, "--tail=200")
	}
}

// logCommand runs a command in the k3s container and logs its output.
func (s *GatewayAPIConformanceSuite) logCommand(ctx context.Context, command ...string) {
	exitCode, result, err := s.k3sContainer.Exec(ctx, command, exec.Multiplexed())
	if err != nil {
		s.T().Logf("%v: %v", command, err)
		return
	}

	output, err := io.ReadAll(result)
	if err != nil {
		s.T().Logf("%v: %v", command, err)
		return
	}

	s.T().Logf("%v (exit code %d):\n%s", command, exitCode, output)
}
