package runtimeconfig

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuoteMySQLIdentifier(t *testing.T) {
	for _, name := range []string{"events", "Events_12", strings.Repeat("a", 64)} {
		quoted, err := QuoteMySQLIdentifier(name)
		require.NoError(t, err)
		assert.Equal(t, "`"+name+"`", quoted)
	}
	for _, name := range []string{"", "1events", "_events", "events`;DROP TABLE users;--", "db.events", "events/*comment*/", "events\n", "events\x00", "events'", "events\"", "../events", "{{ .Body.table }}", "évents", strings.Repeat("a", 65)} {
		t.Run(name, func(t *testing.T) {
			quoted, err := QuoteMySQLIdentifier(name)
			require.Error(t, err)
			assert.Empty(t, quoted)
		})
	}
}
