package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalActionDocumentedFileConfiguration(t *testing.T) {
	content, err := os.ReadFile("../../../docs/content/reference/routing-configuration/http/middlewares/externalaction.md")
	require.NoError(t, err)
	_, example, found := strings.Cut(string(content), "```yaml tab=\"Structured (YAML)\"\n")
	require.True(t, found)
	example, _, found = strings.Cut(example, "\n```")
	require.True(t, found)
	filename := filepath.Join(t.TempDir(), "routing.yml")
	require.NoError(t, os.WriteFile(filename, []byte(example), 0o600))
	p := &Provider{Filename: filename}
	configuration, err := p.buildConfiguration()
	require.NoError(t, err)
	require.NotNil(t, configuration.HTTP.Middlewares["action"])
	action := configuration.HTTP.Middlewares["action"].ExternalAction
	require.NotNil(t, action)
	require.Len(t, action.Rules, 1)
	assert.Contains(t, action.Rules[0].Body.Template, `{{ jsonPath .Body "$.deployment.id" | toJson }}`)
	assert.Equal(t, `deployment:{{ jsonPath .Response.Body "$.data.id" }}`, action.Rules[0].Deduplication.KeyTemplate)
	assert.Equal(t, `{{ .Request.Host }}`, action.Rules[0].HTTP.Headers["X-Source-Host"])
}
