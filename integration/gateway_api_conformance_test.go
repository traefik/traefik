//go:build gatewayAPIConformance

package integration

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/gateway"
	corev1 "k8s.io/api/core/v1"
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
			// Here we are concatenating the features supported by the Traefik Gateway API implementation with the
			// features supported by the Traefik Gateway API operator.
			// TODO: support static addresses feature.
			SupportedFeatures: slices.Concat(gateway.SupportedFeatures(), []features.FeatureName{
				features.GatewayEmptyAddressFeature.Name,
				features.GatewayInfrastructurePropagationFeature.Name,
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

const (
	// nodeLBPoolSize is the number of addresses reserved at the top of the node
	// subnet. The Docker IPAM allocates container addresses from the bottom of
	// the subnet, so the top stays free.
	nodeLBPoolSize = 48

	// nodeLBUnusableAddress is a TEST-NET-1 (RFC 5737) address: routable
	// nowhere, and never part of the pool. The GatewayStaticAddresses
	// conformance test needs an address the infrastructure cannot assign.
	nodeLBUnusableAddress = "192.0.2.1"

	nodeLBReconcileInterval = 500 * time.Millisecond
)

// nodeLoadBalancer assigns addresses to LoadBalancer Services on the single-node
// k3s cluster the operator conformance suite runs on.
//
// The k3s built-in ServiceLB is disabled for that suite: it exposes Services
// through host ports, and the operator provisions one Service per Gateway, most
// of them on port 80, which a single node cannot satisfy. Addresses are taken
// from the unallocated end of the node subnet, added to the node interface, and
// published in the Service status, which is all kube-proxy needs to route them.
// This is the role MetalLB plays in the reference conformance environment.
type nodeLoadBalancer struct {
	container *k3s.K3sContainer
	client    kclientset.Interface

	iface  string
	prefix netip.Prefix

	// staticAddress is held out of the automatic rotation, so that it stays
	// available to a Gateway requesting it through spec.addresses.
	staticAddress string
	pool          []string

	mu       sync.Mutex
	assigned map[string]string
	bound    map[string]struct{}
}

// newNodeLoadBalancer discovers the node interface and carves an address pool
// out of its subnet.
func newNodeLoadBalancer(ctx context.Context, container *k3s.K3sContainer, client kclientset.Interface) (*nodeLoadBalancer, error) {
	nodeIP, err := container.ContainerIP(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting node address: %w", err)
	}

	iface, prefix, err := nodeInterface(ctx, container, nodeIP)
	if err != nil {
		return nil, err
	}

	pool, err := addressPool(prefix, nodeLBPoolSize)
	if err != nil {
		return nil, err
	}

	return &nodeLoadBalancer{
		container:     container,
		client:        client,
		iface:         iface,
		prefix:        prefix,
		staticAddress: pool[0],
		pool:          pool[1:],
		assigned:      map[string]string{},
		bound:         map[string]struct{}{},
	}, nil
}

// Start reconciles the LoadBalancer Services until ctx is done.
func (lb *nodeLoadBalancer) Start(ctx context.Context) {
	go func() {
		tick := time.Tick(nodeLBReconcileInterval)

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick:
				// Errors are transient (a Service deleted mid-reconciliation,
				// a conflicting status update) and resolve on the next tick.
				_ = lb.reconcile(ctx)
			}
		}
	}()
}

// StaticAddress returns the address reserved for the Gateways requesting one
// through spec.addresses.
func (lb *nodeLoadBalancer) StaticAddress() string {
	return lb.staticAddress
}

func (lb *nodeLoadBalancer) reconcile(ctx context.Context) error {
	services, err := lb.client.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	live := map[string]struct{}{}
	for _, service := range services.Items {
		if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
			continue
		}

		live[service.Namespace+"/"+service.Name] = struct{}{}

		if err := lb.reconcileService(ctx, &service); err != nil {
			return err
		}
	}

	// The conformance suite creates and deletes Gateways throughout the run,
	// and each of them holds an address while it exists. Reclaiming the
	// addresses of the deleted ones keeps the pool from running out.
	lb.mu.Lock()
	maps.DeleteFunc(lb.assigned, func(key, _ string) bool {
		_, ok := live[key]
		return !ok
	})
	lb.mu.Unlock()

	return nil
}

func (lb *nodeLoadBalancer) reconcileService(ctx context.Context, service *corev1.Service) error {
	address, ok := lb.address(service)
	if !ok {
		// A requested address outside the pool cannot be assigned: the Service
		// is left without an ingress address, as a real load balancer would.
		if len(service.Status.LoadBalancer.Ingress) == 0 {
			return nil
		}

		service.Status.LoadBalancer.Ingress = nil

		_, err := lb.client.CoreV1().Services(service.Namespace).UpdateStatus(ctx, service, metav1.UpdateOptions{})
		return err
	}

	if len(service.Status.LoadBalancer.Ingress) == 1 && service.Status.LoadBalancer.Ingress[0].IP == address {
		return nil
	}

	if err := lb.bind(ctx, address); err != nil {
		return err
	}

	service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: address}}

	_, err := lb.client.CoreV1().Services(service.Namespace).UpdateStatus(ctx, service, metav1.UpdateOptions{})
	return err
}

// address returns the address to publish for a Service. A Service requesting a
// specific address only gets it when it belongs to the pool.
func (lb *nodeLoadBalancer) address(service *corev1.Service) (string, bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	key := service.Namespace + "/" + service.Name

	if requested := service.Spec.LoadBalancerIP; requested != "" {
		if requested != lb.staticAddress {
			return "", false
		}

		lb.assigned[key] = requested
		return requested, true
	}

	if address, ok := lb.assigned[key]; ok {
		return address, true
	}

	used := slices.Collect(maps.Values(lb.assigned))

	for _, address := range lb.pool {
		if slices.Contains(used, address) {
			continue
		}

		lb.assigned[key] = address
		return address, true
	}

	return "", false
}

// bind adds the address to the node interface, so that the node answers for it
// and kube-proxy sees the traffic.
func (lb *nodeLoadBalancer) bind(ctx context.Context, address string) error {
	lb.mu.Lock()
	if _, ok := lb.bound[address]; ok {
		lb.mu.Unlock()
		return nil
	}
	lb.mu.Unlock()

	cidr := fmt.Sprintf("%s/%d", address, lb.prefix.Bits())

	exitCode, reader, err := lb.container.Exec(ctx, []string{"ip", "addr", "add", cidr, "dev", lb.iface}, exec.Multiplexed())
	if err != nil {
		return fmt.Errorf("binding %s: %w", cidr, err)
	}

	if exitCode != 0 {
		output, _ := io.ReadAll(reader)
		return fmt.Errorf("binding %s: exit code %d: %s", cidr, exitCode, output)
	}

	lb.mu.Lock()
	lb.bound[address] = struct{}{}
	lb.mu.Unlock()

	return nil
}

// nodeInterface returns the name and subnet of the interface holding nodeIP.
func nodeInterface(ctx context.Context, container *k3s.K3sContainer, nodeIP string) (string, netip.Prefix, error) {
	exitCode, reader, err := container.Exec(ctx, []string{"ip", "-o", "-4", "addr", "show"}, exec.Multiplexed())
	if err != nil {
		return "", netip.Prefix{}, fmt.Errorf("listing node addresses: %w", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		return "", netip.Prefix{}, fmt.Errorf("reading node addresses: %w", err)
	}

	if exitCode != 0 {
		return "", netip.Prefix{}, fmt.Errorf("listing node addresses: exit code %d: %s", exitCode, output)
	}

	// Lines look like: 2: eth0    inet 172.28.0.2/16 brd 172.28.255.255 scope global eth0
	for line := range strings.Lines(string(output)) {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[2] != "inet" {
			continue
		}

		prefix, err := netip.ParsePrefix(fields[3])
		if err != nil || prefix.Addr().String() != nodeIP {
			continue
		}

		return strings.TrimSuffix(fields[1], ":"), prefix, nil
	}

	return "", netip.Prefix{}, fmt.Errorf("no interface holding node address %s in:\n%s", nodeIP, output)
}

// addressPool returns the size highest usable addresses of an IPv4 prefix,
// lowest first.
func addressPool(prefix netip.Prefix, size int) ([]string, error) {
	if !prefix.Addr().Is4() {
		return nil, fmt.Errorf("not an IPv4 prefix: %s", prefix)
	}

	hostCount := uint64(1) << (32 - prefix.Bits())
	if hostCount < uint64(size)+2 {
		return nil, fmt.Errorf("prefix %s is too small for a pool of %d addresses", prefix, size)
	}

	network := prefix.Masked().Addr().As4()
	// The last address of the subnet is the broadcast address.
	last := binary.BigEndian.Uint32(network[:]) + uint32(hostCount) - 2

	first := last - uint32(size) + 1

	pool := make([]string, 0, size)
	for i := range size {
		var address [4]byte
		binary.BigEndian.PutUint32(address[:], first+uint32(i))
		pool = append(pool, netip.AddrFrom4(address).String())
	}

	return pool, nil
}
