//go:build gatewayAPIConformance

package integration

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"maps"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclientset "k8s.io/client-go/kubernetes"
)

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
