package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestMustParseYaml_MultiDocumentCRLF(t *testing.T) {
	// A CRLF working tree (core.autocrlf=true) turns the document separators
	// of multi-document fixtures into "---\r\n"; both documents must decode.
	content := []byte("---\r\n" +
		"apiVersion: v1\r\n" +
		"kind: Service\r\n" +
		"metadata:\r\n" +
		"  name: svc-crlf\r\n" +
		"spec:\r\n" +
		"  ports:\r\n" +
		"    - name: web\r\n" +
		"      port: 80\r\n" +
		"---\r\n" +
		"apiVersion: v1\r\n" +
		"kind: Namespace\r\n" +
		"metadata:\r\n" +
		"  name: ns-crlf\r\n" +
		"---\r\n")

	objs := MustParseYaml(content)
	require.Len(t, objs, 2)

	service, ok := objs[0].(*corev1.Service)
	require.True(t, ok, "first object should be a Service, got %T", objs[0])
	assert.Equal(t, "svc-crlf", service.Name)

	namespace, ok := objs[1].(*corev1.Namespace)
	require.True(t, ok, "second object should be a Namespace, got %T", objs[1])
	assert.Equal(t, "ns-crlf", namespace.Name)
}

func TestMustParseYaml_MultiDocumentLF(t *testing.T) {
	content := []byte("---\n" +
		"apiVersion: v1\n" +
		"kind: Service\n" +
		"metadata:\n" +
		"  name: svc-lf\n" +
		"---\n" +
		"apiVersion: v1\n" +
		"kind: Namespace\n" +
		"metadata:\n" +
		"  name: ns-lf\n")

	objs := MustParseYaml(content)
	require.Len(t, objs, 2)

	service, ok := objs[0].(*corev1.Service)
	require.True(t, ok, "first object should be a Service, got %T", objs[0])
	assert.Equal(t, "svc-lf", service.Name)

	namespace, ok := objs[1].(*corev1.Namespace)
	require.True(t, ok, "second object should be a Namespace, got %T", objs[1])
	assert.Equal(t, "ns-lf", namespace.Name)
}
