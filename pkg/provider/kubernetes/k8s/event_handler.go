package k8s

import (
	"reflect"

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

func endpointSliceChanged(a, b *discoveryv1.EndpointSlice) bool {
	if len(a.Endpoints) != len(b.Endpoints) || len(a.Ports) != len(b.Ports) {
		return true
	}

	return !(reflect.DeepEqual(a.Endpoints, b.Endpoints) && reflect.DeepEqual(a.Ports, b.Ports))
}
