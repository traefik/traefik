package ingress

import (
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

func buildIngress(opts ...func(*extensionsv1beta1.Ingress)) *extensionsv1beta1.Ingress {
	i := &extensionsv1beta1.Ingress{}
	i.Kind = "Ingress"
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

func iRules(opts ...func(*extensionsv1beta1.IngressSpec)) func(*extensionsv1beta1.Ingress) {
	return func(i *extensionsv1beta1.Ingress) {
		s := &extensionsv1beta1.IngressSpec{}
		for _, opt := range opts {
			opt(s)
		}
		i.Spec = *s
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
