package ingress

import netv1 "k8s.io/api/networking/v1"

func buildIngress(opts ...func(*netv1.Ingress)) *netv1.Ingress {
	i := &netv1.Ingress{}
	i.Kind = "Ingress"
	for _, opt := range opts {
		opt(i)
	}
	return i
}

func iNamespace(value string) func(*netv1.Ingress) {
	return func(i *netv1.Ingress) {
		i.Namespace = value
	}
}

func iRules(opts ...func(*netv1.IngressSpec)) func(*netv1.Ingress) {
	return func(i *netv1.Ingress) {
		s := &netv1.IngressSpec{}
		for _, opt := range opts {
			opt(s)
		}
		i.Spec = *s
	}
}

func iRule(opts ...func(*netv1.IngressRule)) func(*netv1.IngressSpec) {
	return func(spec *netv1.IngressSpec) {
		r := &netv1.IngressRule{}
		for _, opt := range opts {
			opt(r)
		}
		spec.Rules = append(spec.Rules, *r)
	}
}

func iHost(name string) func(*netv1.IngressRule) {
	return func(rule *netv1.IngressRule) {
		rule.Host = name
	}
}

func iTLSes(opts ...func(*netv1.IngressTLS)) func(*netv1.Ingress) {
	return func(i *netv1.Ingress) {
		for _, opt := range opts {
			iTLS := netv1.IngressTLS{}
			opt(&iTLS)
			i.Spec.TLS = append(i.Spec.TLS, iTLS)
		}
	}
}

func iTLS(secret string, hosts ...string) func(*netv1.IngressTLS) {
	return func(i *netv1.IngressTLS) {
		i.SecretName = secret
		i.Hosts = hosts
	}
}
