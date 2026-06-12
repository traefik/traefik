package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func Test_detectChanges(t *testing.T) {
	portA := int32(80)
	portACopy := int32(80)
	portB := int32(8080)
	portName := "http"
	portNameCopy := "http"
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
			name: "With different service name label",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "foo",
					},
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "bar",
					},
				},
			},
			want: true,
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
			name: "With same ports from different pointers",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{{
					Name: &portName,
					Port: &portA,
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{
					Name: &portNameCopy,
					Port: &portACopy,
				}},
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
		{
			name: "With same endpoint conditions",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Ready:       ptr.To(true),
						Serving:     ptr.To(true),
						Terminating: ptr.To(false),
					},
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Ready:       ptr.To(true),
						Serving:     ptr.To(true),
						Terminating: ptr.To(false),
					},
				}},
			},
		},
		{
			name: "With different endpoint conditions",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Ready: ptr.To(true),
					},
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Ready: ptr.To(false),
					},
				}},
			},
			want: true,
		},
		{
			name: "With nil and false endpoint conditions",
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
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Ready: ptr.To(false),
					},
				}},
			},
			want: true,
		},
		{
			name: "Node with same internal and external addresses",
			oldObj: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
						{Type: corev1.NodeExternalIP, Address: "192.0.2.1"},
						{Type: corev1.NodeHostName, Address: "node-a"},
					},
				},
			},
			newObj: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeHostName, Address: "node-b"},
						{Type: corev1.NodeExternalIP, Address: "192.0.2.1"},
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
					},
					Capacity: corev1.ResourceList{},
				},
			},
		},
		{
			name: "Node with different internal address",
			oldObj: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
					},
				},
			},
			newObj: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.2"},
					},
				},
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
