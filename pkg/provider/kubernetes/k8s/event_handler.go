package k8s

import (
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceEventHandler handles Add, Update or Delete Events for resources.
type ResourceEventHandler struct {
	Ev chan<- interface{}
}

// OnAdd is called on Add Events.
func (reh *ResourceEventHandler) OnAdd(obj interface{}, isInInitialList bool) {
	eventHandlerFunc(reh.Ev, obj)
}

// OnUpdate is called on Update Events.
// Ignores useless changes.
func (reh *ResourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	if objChanged(oldObj, newObj) {
		eventHandlerFunc(reh.Ev, newObj)
	}
}

// OnDelete is called on Delete Events.
func (reh *ResourceEventHandler) OnDelete(obj interface{}) {
	eventHandlerFunc(reh.Ev, obj)
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

func objChanged(oldObj, newObj interface{}) bool {
	if oldObj == nil || newObj == nil {
		return true
	}

	if oldObj.(metav1.Object).GetResourceVersion() == newObj.(metav1.Object).GetResourceVersion() {
		return false
	}

	if _, ok := oldObj.(*discoveryv1.EndpointSlice); ok {
		return endpointSliceChanged(oldObj.(*discoveryv1.EndpointSlice), newObj.(*discoveryv1.EndpointSlice))
	}

	return true
}

// In some Kubernetes versions leader election is done by updating an endpoint annotation every second,
// if there are no changes to the endpoints addresses, ports, and there are no addresses defined for an endpoint
// the event can safely be ignored and won't cause unnecessary config reloads.
// TODO: check if Kubernetes is still using EndpointSlice for leader election, which seems to not be the case anymore.
func endpointSliceChanged(a, b *discoveryv1.EndpointSlice) bool {
	if len(a.Ports) != len(b.Ports) {
		return true
	}

	for i, aport := range a.Ports {
		bport := b.Ports[i]
		if aport.Name != bport.Name {
			return true
		}
		if aport.Port != bport.Port {
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

	return false
}
