package k8s

import (
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// ResourceEventHandler handles Add, Update or Delete Events for resources.
type ResourceEventHandler struct {
	Ev chan<- any
}

// OnAdd is called on Add Events.
func (reh *ResourceEventHandler) OnAdd(obj any, _ bool) {
	eventHandlerFunc(reh.Ev, obj)
}

// OnUpdate is called on Update Events.
// Ignores useless changes.
func (reh *ResourceEventHandler) OnUpdate(oldObj, newObj any) {
	if objChanged(oldObj, newObj) {
		eventHandlerFunc(reh.Ev, newObj)
	}
}

// OnDelete is called on Delete Events.
func (reh *ResourceEventHandler) OnDelete(obj any) {
	eventHandlerFunc(reh.Ev, obj)
}

// eventHandlerFunc will pass the obj on to the events channel or drop it.
// This is so passing the events along won't block in the case of high volume.
// The events are only used for signaling anyway so dropping a few is ok.
func eventHandlerFunc(events chan<- any, obj any) {
	select {
	case events <- obj:
	default:
	}
}

func objChanged(oldObj, newObj any) bool {
	if oldObj == nil || newObj == nil {
		return true
	}

	if oldObj.(metav1.Object).GetResourceVersion() == newObj.(metav1.Object).GetResourceVersion() {
		return false
	}

	switch old := oldObj.(type) {
	case *discoveryv1.EndpointSlice:
		return endpointSliceChanged(old, newObj.(*discoveryv1.EndpointSlice))
	case *corev1.Node:
		return nodeChanged(old, newObj.(*corev1.Node))
	}

	return true
}

// endpointSliceChanged reports whether two EndpointSlices differ in fields that
// Traefik consumes: ports, endpoint addresses, and per-endpoint readiness
// conditions. Metadata-only updates (ResourceVersion bumps, label/annotation
// churn, managedFields edits) are intentionally ignored to avoid spurious
// configuration rebuilds.
//
// Conditions.Ready is consumed by the TCP/UDP/Gateway providers, while
// Conditions.Serving and Conditions.Terminating are consumed by the HTTP
// provider (server filtering and the Fenced behaviour); all three are compared
// with nil / true / false treated as distinct values, since a nil value
// defaults to "true" per the Kubernetes API spec but a downgrade to false is
// load-bearing.
func endpointSliceChanged(a, b *discoveryv1.EndpointSlice) bool {
	if len(a.Ports) != len(b.Ports) {
		return true
	}

	for i, aport := range a.Ports {
		bport := b.Ports[i]
		if !ptr.Equal(aport.Name, bport.Name) {
			return true
		}
		if !ptr.Equal(aport.Port, bport.Port) {
			return true
		}
	}

	if len(a.Endpoints) != len(b.Endpoints) {
		return true
	}

	for i, ea := range a.Endpoints {
		eb := b.Endpoints[i]
		if endpointChanged(ea, eb) {
			return true
		}
	}

	return false
}

func endpointChanged(a, b discoveryv1.Endpoint) bool {
	if len(a.Addresses) != len(b.Addresses) {
		return true
	}

	for i, aaddr := range a.Addresses {
		baddr := b.Addresses[i]
		if aaddr != baddr {
			return true
		}
	}

	if !ptr.Equal(a.Conditions.Ready, b.Conditions.Ready) {
		return true
	}
	if !ptr.Equal(a.Conditions.Serving, b.Conditions.Serving) {
		return true
	}
	if !ptr.Equal(a.Conditions.Terminating, b.Conditions.Terminating) {
		return true
	}

	return false
}

// nodeChanged reports whether two Node objects differ in any field Traefik
// reads. Traefik only consumes Status.Addresses entries of type InternalIP and
// ExternalIP (for NodePort load-balancer servers and IngressLoadBalancer
// status), so kubelet status heartbeats (~10s) that rewrite conditions,
// capacity, images, allocatable etc. without touching those addresses must not
// drive a full configuration rebuild.
//
// Comparison is allocation-free: Nodes typically report 2-3 relevant
// addresses, so an O(N*M) pairwise walk beats building two maps per call and
// avoids the per-heartbeat GC churn that the map approach incurred under
// kubelet's ~10s status cadence.
func nodeChanged(a, b *corev1.Node) bool {
	if relevantNodeAddressCount(a) != relevantNodeAddressCount(b) {
		return true
	}
	for _, aa := range a.Status.Addresses {
		if aa.Type != corev1.NodeInternalIP && aa.Type != corev1.NodeExternalIP {
			continue
		}
		found := false
		for _, ba := range b.Status.Addresses {
			if ba.Type == aa.Type && ba.Address == aa.Address {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	return false
}

func relevantNodeAddressCount(n *corev1.Node) int {
	c := 0
	for _, addr := range n.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP || addr.Type == corev1.NodeExternalIP {
			c++
		}
	}
	return c
}
