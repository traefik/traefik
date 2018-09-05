package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func buildIngress(opts ...func(*extensionsv1beta1.Ingress)) *extensionsv1beta1.Ingress {
	i := &extensionsv1beta1.Ingress{}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

func iNamespace(value string) func(*extensionsv1beta1.Ingress) {
	return func(i *extensionsv1beta1.Ingress) {
		i.Namespace = value
	}
}

func iAnnotation(name string, value string) func(*extensionsv1beta1.Ingress) {
	return func(i *extensionsv1beta1.Ingress) {
		if i.Annotations == nil {
			i.Annotations = make(map[string]string)
		}
		i.Annotations[name] = value
	}
}

func iRules(opts ...func(*extensionsv1beta1.IngressSpec)) func(*extensionsv1beta1.Ingress) {
	return func(i *extensionsv1beta1.Ingress) {
		s := &extensionsv1beta1.IngressSpec{}
		for _, opt := range opts {
			opt(s)
		}
		i.Spec = *s
	}
}

func iSpecBackends(opts ...func(*extensionsv1beta1.IngressSpec)) func(*extensionsv1beta1.Ingress) {
	return func(i *extensionsv1beta1.Ingress) {
		s := &extensionsv1beta1.IngressSpec{}
		for _, opt := range opts {
			opt(s)
		}
		i.Spec = *s
	}
}

func iSpecBackend(opts ...func(*extensionsv1beta1.IngressBackend)) func(*extensionsv1beta1.IngressSpec) {
	return func(s *extensionsv1beta1.IngressSpec) {
		p := &extensionsv1beta1.IngressBackend{}
		for _, opt := range opts {
			opt(p)
		}
		s.Backend = p
	}
}

func iIngressBackend(name string, port intstr.IntOrString) func(*extensionsv1beta1.IngressBackend) {
	return func(p *extensionsv1beta1.IngressBackend) {
		p.ServiceName = name
		p.ServicePort = port
	}
}

func iRule(opts ...func(*extensionsv1beta1.IngressRule)) func(*extensionsv1beta1.IngressSpec) {
	return func(spec *extensionsv1beta1.IngressSpec) {
		r := &extensionsv1beta1.IngressRule{}
		for _, opt := range opts {
			opt(r)
		}
		spec.Rules = append(spec.Rules, *r)
	}
}

func iHost(name string) func(*extensionsv1beta1.IngressRule) {
	return func(rule *extensionsv1beta1.IngressRule) {
		rule.Host = name
	}
}

func iPaths(opts ...func(*extensionsv1beta1.HTTPIngressRuleValue)) func(*extensionsv1beta1.IngressRule) {
	return func(rule *extensionsv1beta1.IngressRule) {
		rule.HTTP = &extensionsv1beta1.HTTPIngressRuleValue{}
		for _, opt := range opts {
			opt(rule.HTTP)
		}
	}
}

func onePath(opts ...func(*extensionsv1beta1.HTTPIngressPath)) func(*extensionsv1beta1.HTTPIngressRuleValue) {
	return func(irv *extensionsv1beta1.HTTPIngressRuleValue) {
		p := &extensionsv1beta1.HTTPIngressPath{}
		for _, opt := range opts {
			opt(p)
		}
		irv.Paths = append(irv.Paths, *p)
	}
}

func iPath(name string) func(*extensionsv1beta1.HTTPIngressPath) {
	return func(p *extensionsv1beta1.HTTPIngressPath) {
		p.Path = name
	}
}

func iBackend(name string, port intstr.IntOrString) func(*extensionsv1beta1.HTTPIngressPath) {
	return func(p *extensionsv1beta1.HTTPIngressPath) {
		p.Backend = extensionsv1beta1.IngressBackend{
			ServiceName: name,
			ServicePort: port,
		}
	}
}

func iTLSes(opts ...func(*extensionsv1beta1.IngressTLS)) func(*extensionsv1beta1.Ingress) {
	return func(i *extensionsv1beta1.Ingress) {
		for _, opt := range opts {
			iTLS := extensionsv1beta1.IngressTLS{}
			opt(&iTLS)
			i.Spec.TLS = append(i.Spec.TLS, iTLS)
		}
	}
}

func iTLS(secret string, hosts ...string) func(*extensionsv1beta1.IngressTLS) {
	return func(i *extensionsv1beta1.IngressTLS) {
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

func sampleIngress() *extensionsv1beta1.Ingress {
	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testing",
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					Host: "foo",
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/bar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "service1",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: "/namedthing",
									Backend: extensionsv1beta1.IngressBackend{
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
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "service3",
										ServicePort: intstr.FromString("https"),
									},
								},
								{
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "service2",
										ServicePort: intstr.FromInt(802),
									},
								},
							},
						},
					},
				},
			},
			TLS: []extensionsv1beta1.IngressTLS{
				{
					Hosts:      []string{"foo"},
					SecretName: "tls-secret",
				},
			},
		},
	}
}
