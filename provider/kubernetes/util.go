package kubernetes

import (
	"strings"

	"github.com/containous/traefik/types"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

func splitKeyValue(str string) (key, val string) {
	parts := strings.SplitN(str, "=", 2)
	key = parts[0]
	if len(parts) == 2 {
		val = parts[1]
	}
	return
}

func getRuleForPath(pa v1beta1.HTTPIngressPath, i *v1beta1.Ingress) string {
	if len(pa.Path) == 0 {
		return ""
	}

	ruleType := i.Annotations[types.LabelFrontendRuleType]
	if ruleType == "" {
		ruleType = ruleTypePathPrefix
	}

	rule := ruleType + ":" + pa.Path

	if rewriteTarget := i.Annotations[annotationKubernetesRewriteTarget]; rewriteTarget != "" {
		rule = ruleTypeReplacePath + ":" + rewriteTarget
	}

	return rule
}

func endpointPortNumber(servicePort v1.ServicePort, endpointPorts []v1.EndpointPort) int {
	if len(endpointPorts) > 0 {
		//name is optional if there is only one port
		port := endpointPorts[0]
		for _, endpointPort := range endpointPorts {
			if servicePort.Name == endpointPort.Name {
				port = endpointPort
			}
		}
		return int(port.Port)
	}
	return int(servicePort.Port)
}

func equalPorts(servicePort v1.ServicePort, ingressPort intstr.IntOrString) bool {
	if int(servicePort.Port) == ingressPort.IntValue() {
		return true
	}
	if servicePort.Name != "" && servicePort.Name == ingressPort.String() {
		return true
	}
	return false
}

func shouldProcessIngress(ingressClass string) bool {
	switch ingressClass {
	case "", "traefik":
		return true
	default:
		return false
	}
}
