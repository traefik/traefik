package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_detectChanges(t *testing.T) {
	portA := int32(80)
	portB := int32(8080)
	tests := []struct {
		name   string
		oldObj interface{}
		newObj interface{}
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
