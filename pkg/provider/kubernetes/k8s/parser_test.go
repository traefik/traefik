package k8s

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestMustParseYaml(t *testing.T) {
	content := `apiVersion: v1
kind: Service
metadata:
  name: first
  namespace: default

---
apiVersion: v1
kind: Service
metadata:
  name: second
  namespace: default
`

	testCases := []struct {
		desc    string
		content string
	}{
		{
			desc:    "LF line endings",
			content: content,
		},
		{
			desc:    "CRLF line endings",
			content: strings.ReplaceAll(content, "\n", "\r\n"),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			objects := MustParseYaml([]byte(test.content))
			require.Len(t, objects, 2)

			var names []string
			for _, object := range objects {
				service, ok := object.(*corev1.Service)
				require.True(t, ok)

				names = append(names, service.Name)
			}

			assert.Equal(t, []string{"first", "second"}, names)
		})
	}
}
