package kubernetes

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/containous/traefik/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

// MustEncodeYaml Encode object to YAML.
//
// ex:
// 	MustEncodeYaml(ingresses[0], "extensions/v1beta1", "ingress.yml")
// 	MustEncodeYaml(services[0], "v1", "service.yml")
// 	MustEncodeYaml(endpoints[0], "v1", "endpoint.yml")
func MustEncodeYaml(object runtime.Object, groupName string, w io.Writer) {
	info, ok := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		panic("oops")
	}

	gv, err := schema.ParseGroupVersion(groupName)
	if err != nil {
		panic(err)
	}

	err = scheme.Codecs.EncoderForVersion(info.Serializer, gv).Encode(object, w)
	if err != nil {
		panic(err)
	}
}

// MustDecodeYaml Decode a YAML to objects.
func MustDecodeYaml(content []byte) []runtime.Object {
	acceptedK8sTypes := regexp.MustCompile(`(Deployment|Endpoints|Service|Ingress|Secret)`)

	files := strings.Split(string(content), "---")
	retVal := make([]runtime.Object, 0, len(files))
	for _, file := range files {
		if file == "\n" || file == "" {
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(file), nil, nil)
		if err != nil {
			panic(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
		}

		if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
			log.Debugf("The custom-roles configMap contained K8s object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind)
		} else {
			retVal = append(retVal, obj)
		}
	}
	return retVal
}

func mustCreateFile(filename string) *os.File {
	fp := filepath.Join("fixtures", filename)
	err := os.MkdirAll(filepath.Dir(fp), 0777)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(fp)
	if err != nil {
		panic(err)
	}

	return file
}
