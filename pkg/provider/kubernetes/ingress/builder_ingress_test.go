package ingress

import (
	"k8s.io/api/networking/v1beta1"
)

func buildIngress(opts ...func(*v1beta1.Ingress)) *v1beta1.Ingress {
	i := &v1beta1.Ingress{}
	i.Kind = "Ingress"
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
