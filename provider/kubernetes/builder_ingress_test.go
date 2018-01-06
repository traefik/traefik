package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

func buildIngress(opts ...func(*v1beta1.Ingress)) *v1beta1.Ingress {
	i := &v1beta1.Ingress{}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

func iNamespace(value string) func(*v1beta1.Ingress) {
	return func(i *v1beta1.Ingress) {
		i.Namespace = value
	}
}

func iAnnotation(name string, value string) func(*v1beta1.Ingress) {
	return func(i *v1beta1.Ingress) {
		if i.Annotations == nil {
			i.Annotations = make(map[string]string)
		}
		i.Annotations[name] = value
	}
}

func iRules(opts ...func(*v1beta1.IngressSpec)) func(*v1beta1.Ingress) {
	return func(i *v1beta1.Ingress) {
		s := &v1beta1.IngressSpec{}
		for _, opt := range opts {
			opt(s)
		}
		i.Spec = *s
	}
}

func iRule(opts ...func(*v1beta1.IngressRule)) func(*v1beta1.IngressSpec) {
	return func(spec *v1beta1.IngressSpec) {
		r := &v1beta1.IngressRule{}
		for _, opt := range opts {
			opt(r)
		}
		spec.Rules = append(spec.Rules, *r)
	}
}

func iHost(name string) func(*v1beta1.IngressRule) {
	return func(rule *v1beta1.IngressRule) {
		rule.Host = name
	}
}

func iPaths(opts ...func(*v1beta1.HTTPIngressRuleValue)) func(*v1beta1.IngressRule) {
	return func(rule *v1beta1.IngressRule) {
		rule.HTTP = &v1beta1.HTTPIngressRuleValue{}
		for _, opt := range opts {
			opt(rule.HTTP)
		}
	}
}

func onePath(opts ...func(*v1beta1.HTTPIngressPath)) func(*v1beta1.HTTPIngressRuleValue) {
	return func(irv *v1beta1.HTTPIngressRuleValue) {
		p := &v1beta1.HTTPIngressPath{}
		for _, opt := range opts {
			opt(p)
		}
		irv.Paths = append(irv.Paths, *p)
	}
}

func iPath(name string) func(*v1beta1.HTTPIngressPath) {
	return func(p *v1beta1.HTTPIngressPath) {
		p.Path = name
	}
}

func iBackend(name string, port intstr.IntOrString) func(*v1beta1.HTTPIngressPath) {
	return func(p *v1beta1.HTTPIngressPath) {
		p.Backend = v1beta1.IngressBackend{
			ServiceName: name,
			ServicePort: port,
		}
	}
}

func iTLSes(opts ...func(*v1beta1.IngressTLS)) func(*v1beta1.Ingress) {
	return func(i *v1beta1.Ingress) {
		for _, opt := range opts {
			iTLS := v1beta1.IngressTLS{}
			opt(&iTLS)
			i.Spec.TLS = append(i.Spec.TLS, iTLS)
		}
	}
}

func iTLS(secret string, hosts ...string) func(*v1beta1.IngressTLS) {
	return func(i *v1beta1.IngressTLS) {
		i.SecretName = secret
		i.Hosts = hosts
	}
}

// Test

func TestBuildIngress(t *testing.T) {
	i := buildIngress(
		iNamespace("testing"),
		iRules(
			iRule(iHost("foo"), iPaths(
				onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))),
				onePath(iPath("/namedthing"), iBackend("service4", intstr.FromString("https")))),
			),
			iRule(iHost("bar"), iPaths(
				onePath(iBackend("service3", intstr.FromString("https"))),
				onePath(iBackend("service2", intstr.FromInt(802))),
			),
			),
		),
		iTLSes(
			iTLS("tls-secret", "foo"),
		),
	)

	assert.EqualValues(t, sampleIngress(), i)
}

func sampleIngress() *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testing",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: "foo",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/bar",
									Backend: v1beta1.IngressBackend{
										ServiceName: "service1",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: "/namedthing",
									Backend: v1beta1.IngressBackend{
										ServiceName: "service4",
										ServicePort: intstr.FromString("https"),
									},
								},
							},
						},
					},
				},
				{
					Host: "bar",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "service3",
										ServicePort: intstr.FromString("https"),
									},
								},
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "service2",
										ServicePort: intstr.FromInt(802),
									},
								},
							},
						},
					},
				},
			},
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      []string{"foo"},
					SecretName: "tls-secret",
				},
			},
		},
	}
}
