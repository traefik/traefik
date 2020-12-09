package k8s

import (
	"github.com/traefik/traefik/v2/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceEventHandler handles Add, Update or Delete Events for resources.
type ResourceEventHandler struct {
	Ev chan<- interface{}
}

// OnAdd is called on Add Events.
func (reh *ResourceEventHandler) OnAdd(obj interface{}) {
	eventHandlerFunc(reh.Ev, obj)
}

// OnUpdate is called on Update Events.
func (reh *ResourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	if !detectChanges(oldObj, newObj) {
		return
	}
	eventHandlerFunc(reh.Ev, newObj)
}

// OnDelete is called on Delete Events.
func (reh *ResourceEventHandler) OnDelete(obj interface{}) {
	eventHandlerFunc(reh.Ev, obj)
}

// code from https://github.com/coredns/coredns/blob/d902e859199e4085cd27453f30367fd1b0799bc5/plugin/kubernetes/controller.go#L421
// modified to work with traefik event handler.
// Kubernetes does leader election by updating an endpoint annotation every second,
// if there are no changes to the endpoints addresses or there are no addresses defined for an endpoint
// the event can safely be ignored and won't cause unnecessary config reloads.
func detectChanges(oldObj, newObj interface{}) bool {
	// If both objects have the same resource version, they are identical.
	if newObj != nil && oldObj != nil && (oldObj.(metav1.Object).GetResourceVersion() == newObj.(metav1.Object).GetResourceVersion()) {
		return false
	}
	obj := newObj
	if obj == nil {
		obj = oldObj
	}

	if _, ok := obj.(*corev1.Endpoints); ok {
		if endpointsEquivalent(oldObj.(*corev1.Endpoints), newObj.(*corev1.Endpoints)) {
			log.Debugf("endpoint %s has no changes, ignoring", newObj.(*corev1.Endpoints).Name)
			return false
		}
	}
	return true
}

// endpointsEquivalent checks if the update to an endpoint is something
// that matters to us or if they are effectively equivalent.
func endpointsEquivalent(a, b *corev1.Endpoints) bool {
	if a == nil || b == nil {
		return false
	}

	if len(a.Subsets) != len(b.Subsets) {
		return false
	}

	// we should be able to rely on
	// these being sorted and able to be compared
	// they are supposed to be in a canonical format
	for i, sa := range a.Subsets {
		sb := b.Subsets[i]
		if !subsetsEquivalent(sa, sb) {
			return false
		}
	}
	return true
}

// subsetsEquivalent checks if two endpoint subsets are significantly equivalent
// I.e. that they have the same ready addresses, host names, ports (including protocol
// and service names for SRV).
func subsetsEquivalent(sa, sb corev1.EndpointSubset) bool {
	if len(sa.Addresses) != len(sb.Addresses) {
		return false
	}
	if len(sa.Ports) != len(sb.Ports) {
		return false
	}

	// in Addresses and Ports, we should be able to rely on
	// these being sorted and able to be compared
	// they are supposed to be in a canonical format
	for addr, aaddr := range sa.Addresses {
		baddr := sb.Addresses[addr]
		if aaddr.IP != baddr.IP {
			return false
		}
		if aaddr.Hostname != baddr.Hostname {
			return false
		}
	}

	for port, aport := range sa.Ports {
		bport := sb.Ports[port]
		if aport.Name != bport.Name {
			return false
		}
		if aport.Port != bport.Port {
			return false
		}
		if aport.Protocol != bport.Protocol {
			return false
		}
	}
	return true
}

// eventHandlerFunc will pass the obj on to the events channel or drop it.
// This is so passing the events along won't block in the case of high volume.
// The events are only used for signaling anyway so dropping a few is ok.
func eventHandlerFunc(events chan<- interface{}, obj interface{}) {
	select {
	case events <- obj:
	default:
	}
}
