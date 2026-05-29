package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_detectChanges(t *testing.T) {
	portA := int32(80)
	portB := int32(8080)
	tests := []struct {
		name   string
		oldObj any
		newObj any
		want   bool
	}{
		{
			name: "With nil values",
			want: true,
		},
		{
			name:   "With empty endpointslice",
			oldObj: &discoveryv1.EndpointSlice{},
			newObj: &discoveryv1.EndpointSlice{},
		},
		{
			name:   "With old nil",
			newObj: &discoveryv1.EndpointSlice{},
			want:   true,
		},
		{
			name:   "With new nil",
			oldObj: &discoveryv1.EndpointSlice{},
			want:   true,
		},
		{
			name: "With same version",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
		},
		{
			name: "With different version",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
			},
		},
		{
			name: "Ingress With same version",
			oldObj: &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
			newObj: &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
		},
		{
			name: "Ingress With different version",
			oldObj: &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
			newObj: &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
			},
			want: true,
		},
		{
			name: "With same annotations",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
					Annotations: map[string]string{
						"test-annotation": "_",
					},
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Annotations: map[string]string{
						"test-annotation": "_",
					},
				},
			},
		},
		{
			name: "With different annotations",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
					Annotations: map[string]string{
						"test-annotation": "V",
					},
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Annotations: map[string]string{
						"test-annotation": "X",
					},
				},
			},
		},
		{
			name: "With same endpoints and ports",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{},
				Ports:     []discoveryv1.EndpointPort{},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{},
				Ports:     []discoveryv1.EndpointPort{},
			},
		},
		{
			name: "With different len of endpoints",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{}},
			},
			want: true,
		},
		{
			name: "With different endpoints",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.11"},
				}},
			},
			want: true,
		},
		{
			name: "With different len of ports",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{}},
			},
			want: true,
		},
		{
			name: "With different ports",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{{
					Port: &portA,
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{
					Port: &portB,
				}},
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, objChanged(test.oldObj, test.newObj))
		})
	}
}

func Test_endpointSliceChanged_conditions(t *testing.T) {
	bTrue := true
	bFalse := false

	makeSlice := func(ready, serving, terminating *bool) *discoveryv1.EndpointSlice {
		return &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"},
			Endpoints: []discoveryv1.Endpoint{{
				Addresses: []string{"10.0.0.1"},
				Conditions: discoveryv1.EndpointConditions{
					Ready:       ready,
					Serving:     serving,
					Terminating: terminating,
				},
			}},
		}
	}

	tests := []struct {
		name string
		a    *discoveryv1.EndpointSlice
		b    *discoveryv1.EndpointSlice
		want bool
	}{
		{
			name: "Identical conditions (all nil)",
			a:    makeSlice(nil, nil, nil),
			b:    makeSlice(nil, nil, nil),
			want: false,
		},
		{
			name: "Identical conditions (all true)",
			a:    makeSlice(&bTrue, &bTrue, &bFalse),
			b:    makeSlice(&bTrue, &bTrue, &bFalse),
			want: false,
		},
		{
			name: "Ready transitions nil -> true",
			a:    makeSlice(nil, &bTrue, &bFalse),
			b:    makeSlice(&bTrue, &bTrue, &bFalse),
			want: true,
		},
		{
			name: "Ready transitions true -> false",
			a:    makeSlice(&bTrue, &bTrue, &bFalse),
			b:    makeSlice(&bFalse, &bTrue, &bFalse),
			want: true,
		},
		{
			name: "Serving transitions true -> false",
			a:    makeSlice(&bTrue, &bTrue, &bFalse),
			b:    makeSlice(&bTrue, &bFalse, &bFalse),
			want: true,
		},
		{
			name: "Terminating transitions false -> true",
			a:    makeSlice(&bTrue, &bTrue, &bFalse),
			b:    makeSlice(&bTrue, &bTrue, &bTrue),
			want: true,
		},
		{
			name: "Terminating transitions nil -> false (distinct)",
			a:    makeSlice(&bTrue, &bTrue, nil),
			b:    makeSlice(&bTrue, &bTrue, &bFalse),
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, endpointSliceChanged(test.a, test.b))
		})
	}
}

func Test_endpointSliceChanged_ports(t *testing.T) {
	port80 := int32(80)
	port80bis := int32(80)
	port8080 := int32(8080)
	nameA := "http"
	nameAbis := "http"
	nameB := "metrics"

	makeSlice := func(ports []discoveryv1.EndpointPort) *discoveryv1.EndpointSlice {
		return &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"},
			Ports:      ports,
		}
	}

	tests := []struct {
		name string
		a    *discoveryv1.EndpointSlice
		b    *discoveryv1.EndpointSlice
		want bool
	}{
		{
			name: "Identical port values but distinct pointers",
			a:    makeSlice([]discoveryv1.EndpointPort{{Name: &nameA, Port: &port80}}),
			b:    makeSlice([]discoveryv1.EndpointPort{{Name: &nameAbis, Port: &port80bis}}),
			want: false,
		},
		{
			name: "Different port number",
			a:    makeSlice([]discoveryv1.EndpointPort{{Name: &nameA, Port: &port80}}),
			b:    makeSlice([]discoveryv1.EndpointPort{{Name: &nameA, Port: &port8080}}),
			want: true,
		},
		{
			name: "Different port name",
			a:    makeSlice([]discoveryv1.EndpointPort{{Name: &nameA, Port: &port80}}),
			b:    makeSlice([]discoveryv1.EndpointPort{{Name: &nameB, Port: &port80}}),
			want: true,
		},
		{
			name: "Port name nil vs set",
			a:    makeSlice([]discoveryv1.EndpointPort{{Port: &port80}}),
			b:    makeSlice([]discoveryv1.EndpointPort{{Name: &nameA, Port: &port80}}),
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, endpointSliceChanged(test.a, test.b))
		})
	}
}

func Test_nodeChanged(t *testing.T) {
	mkNode := func(addrs ...corev1.NodeAddress) *corev1.Node {
		return &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1", Name: "n"},
			Status: corev1.NodeStatus{
				Addresses: addrs,
			},
		}
	}

	tests := []struct {
		name string
		a    *corev1.Node
		b    *corev1.Node
		want bool
	}{
		{
			name: "Identical InternalIP",
			a:    mkNode(corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}),
			b:    mkNode(corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}),
			want: false,
		},
		{
			name: "Heartbeat-only change (hostname order, capacity, etc.)",
			a: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
						{Type: corev1.NodeHostName, Address: "old-host"},
					},
					Capacity: corev1.ResourceList{},
				},
			},
			b: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
						{Type: corev1.NodeHostName, Address: "new-host"},
					},
					Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
				},
			},
			want: false,
		},
		{
			name: "InternalIP changed",
			a:    mkNode(corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}),
			b:    mkNode(corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.2"}),
			want: true,
		},
		{
			name: "InternalIP added",
			a:    mkNode(corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}),
			b: mkNode(
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.2"},
			),
			want: true,
		},
		{
			name: "ExternalIP added",
			a:    mkNode(corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}),
			b: mkNode(
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				corev1.NodeAddress{Type: corev1.NodeExternalIP, Address: "203.0.113.1"},
			),
			want: true,
		},
		{
			name: "ExternalIP value changed",
			a: mkNode(
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				corev1.NodeAddress{Type: corev1.NodeExternalIP, Address: "203.0.113.1"},
			),
			b: mkNode(
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				corev1.NodeAddress{Type: corev1.NodeExternalIP, Address: "203.0.113.2"},
			),
			want: true,
		},
		{
			name: "Same addresses but reordered",
			a: mkNode(
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				corev1.NodeAddress{Type: corev1.NodeExternalIP, Address: "203.0.113.1"},
			),
			b: mkNode(
				corev1.NodeAddress{Type: corev1.NodeExternalIP, Address: "203.0.113.1"},
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
			),
			want: false,
		},
		{
			name: "Hostname-only change (irrelevant address type)",
			a: mkNode(
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				corev1.NodeAddress{Type: corev1.NodeHostName, Address: "old-host"},
			),
			b: mkNode(
				corev1.NodeAddress{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				corev1.NodeAddress{Type: corev1.NodeHostName, Address: "new-host"},
			),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, nodeChanged(test.a, test.b))
		})
	}
}

func Test_objChanged_dispatchesNode(t *testing.T) {
	a := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}},
		},
	}
	b := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"}, // heartbeat bumps RV
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}},
		},
	}
	// Same Internal/ExternalIPs even though RV changed - must not fire.
	assert.False(t, objChanged(a, b))
}

func Test_endpointSliceServiceNameIndexFunc(t *testing.T) {
	tests := []struct {
		name string
		obj  any
		want []string
	}{
		{
			name: "EndpointSlice with service-name label",
			obj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "es-1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "svc",
					},
				},
			},
			want: []string{"ns/svc"},
		},
		{
			name: "EndpointSlice missing service-name label",
			obj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "es-1"},
			},
		},
		{
			name: "EndpointSlice with empty service-name label",
			obj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "es-1",
					Labels:    map[string]string{discoveryv1.LabelServiceName: ""},
				},
			},
		},
		{
			name: "Non-EndpointSlice object",
			obj:  &corev1.Node{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := endpointSliceServiceNameIndexFunc(test.obj)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
