package k8s

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
)

// MustParseYaml parses a YAML to objects.
func MustParseYaml(content []byte) []runtime.Object {
	acceptedK8sTypes := regexp.MustCompile(`^(Namespace|Deployment|EndpointSlice|Node|Service|ConfigMap|Ingress|IngressRoute|IngressRouteTCP|IngressRouteUDP|Middleware|MiddlewareTCP|Secret|TLSOption|TLSStore|TraefikService|IngressClass|ServersTransport|ServersTransportTCP|GatewayClass|Gateway|GRPCRoute|HTTPRoute|TCPRoute|TLSRoute|ReferenceGrant|BackendTLSPolicy)$`)

	// Normalize line endings so document separators match even when the
	// working tree uses CRLF (e.g. a checkout made with core.autocrlf=true);
	// otherwise multi-document fixtures silently decode as one document.
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	files := strings.Split(string(content), "---\n")
	retVal := make([]runtime.Object, 0, len(files))
	for _, file := range files {
		if file == "\n" || file == "" {
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
