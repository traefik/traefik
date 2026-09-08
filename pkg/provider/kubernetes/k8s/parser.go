package k8s

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
)

// documentSeparator matches the line separating two YAML documents. The line
// ending depends on the platform the repository is checked out on, and a
// separator missed here silently drops every document after it.
var documentSeparator = regexp.MustCompile(`(?m)^---[ \t]*\r?\n`)

// MustParseYaml parses a YAML to objects.
func MustParseYaml(content []byte) []runtime.Object {
	acceptedK8sTypes := regexp.MustCompile(`^(Namespace|Deployment|EndpointSlice|Node|Service|ConfigMap|Ingress|IngressRoute|IngressRouteTCP|IngressRouteUDP|Middleware|MiddlewareTCP|Secret|TLSOption|TLSStore|TraefikService|IngressClass|ServersTransport|ServersTransportTCP|GatewayClass|Gateway|GRPCRoute|HTTPRoute|TCPRoute|TLSRoute|ReferenceGrant|BackendTLSPolicy)$`)

	files := documentSeparator.Split(string(content), -1)
	retVal := make([]runtime.Object, 0, len(files))
	for _, file := range files {
		if strings.TrimSpace(file) == "" {
			continue
		}

		decode := kscheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(file), nil, nil)
		if err != nil {
			panic(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
		}

		if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
			log.Debug().Msgf("The custom-roles configMap contained K8s object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind)
		} else {
			retVal = append(retVal, obj)
		}
	}
	return retVal
}
