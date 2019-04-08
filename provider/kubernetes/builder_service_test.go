package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func buildService(opts ...func(*corev1.Service)) *corev1.Service {
	s := &corev1.Service{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func sNamespace(value string) func(*corev1.Service) {
	return func(i *corev1.Service) {
		i.Namespace = value
	}
}

func sName(value string) func(*corev1.Service) {
	return func(i *corev1.Service) {
		i.Name = value
	}
}

func sUID(value types.UID) func(*corev1.Service) {
	return func(i *corev1.Service) {
		i.UID = value
	}
}

func sSpec(opts ...func(*corev1.ServiceSpec)) func(*corev1.Service) {
	return func(s *corev1.Service) {
		spec := &corev1.ServiceSpec{}
		for _, opt := range opts {
			opt(spec)
		}
		s.Spec = *spec
	}
}

func clusterIP(ip string) func(*corev1.ServiceSpec) {
	return func(spec *corev1.ServiceSpec) {
		spec.ClusterIP = ip
	}
}

func sType(value corev1.ServiceType) func(*corev1.ServiceSpec) {
	return func(spec *corev1.ServiceSpec) {
		spec.Type = value
	}
}

func sExternalName(name string) func(*corev1.ServiceSpec) {
	return func(spec *corev1.ServiceSpec) {
		spec.ExternalName = name
	}
}

func sPorts(opts ...func(*corev1.ServicePort)) func(*corev1.ServiceSpec) {
	return func(spec *corev1.ServiceSpec) {
		for _, opt := range opts {
			p := &corev1.ServicePort{}
			opt(p)
			spec.Ports = append(spec.Ports, *p)
		}
	}
}

func sPort(port int32, name string) func(*corev1.ServicePort) {
	return func(sp *corev1.ServicePort) {
		sp.Port = port
		sp.Name = name
	}
}

// Test

func TestBuildService(t *testing.T) {
	actual1 := buildService(
		sName("service1"),
		sNamespace("testing"),
		sUID("1"),
		sSpec(
			clusterIP("10.0.0.1"),
			sPorts(sPort(80, "")),
		),
	)

	assert.EqualValues(t, sampleService1(), actual1)

	actual2 := buildService(
		sName("service2"),
		sNamespace("testing"),
		sUID("2"),
		sSpec(
			clusterIP("10.0.0.2"),
			sType("ExternalName"),
			sExternalName("example.com"),
			sPorts(
				sPort(80, "http"),
				sPort(443, "https"),
			),
		),
	)

	assert.EqualValues(t, sampleService2(), actual2)

	actual3 := buildService(
		sName("service3"),
		sNamespace("testing"),
		sUID("3"),
		sSpec(
			clusterIP("10.0.0.3"),
			sType("ExternalName"),
			sExternalName("example.com"),
			sPorts(
				sPort(8080, "http"),
				sPort(8443, "https"),
			),
		),
	)

	assert.EqualValues(t, sampleService3(), actual3)
}

func sampleService1() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service1",
			UID:       "1",
			Namespace: "testing",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.0.0.1",
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}
}

func sampleService2() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service2",
			UID:       "2",
			Namespace: "testing",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:    "10.0.0.2",
			Type:         "ExternalName",
			ExternalName: "example.com",
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
				{
					Name: "https",
					Port: 443,
				},
			},
		},
	}
}

func sampleService3() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service3",
			UID:       "3",
			Namespace: "testing",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:    "10.0.0.3",
			Type:         "ExternalName",
			ExternalName: "example.com",
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 8080,
				},
				{
					Name: "https",
					Port: 8443,
				},
			},
		},
	}
}
