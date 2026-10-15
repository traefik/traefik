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

func FuzzQuoteMySQLIdentifier(f *testing.F) {
	for _, value := range []string{
		"events", "Events_12", "", "_events", "9events", "db.events",
		"events`;DROP TABLE users;--", "events/*comment*/", "events\n",
		"events\x00", "évents", strings.Repeat("a", 64), strings.Repeat("a", 65),
	} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		valid := len(value) > 0 && len(value) <= 64
		for index := range len(value) {
			char := value[index]
			letter := char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
			if index == 0 {
				valid = valid && letter
			} else {
				valid = valid && (letter || char >= '0' && char <= '9' || char == '_')
			}
		}
		quoted, err := QuoteMySQLIdentifier(value)
		if !valid {
			require.Error(t, err)
			require.Empty(t, quoted)
			return
		}
		require.NoError(t, err)
		require.Len(t, quoted, len(value)+2)
		require.Equal(t, value, quoted[1:len(quoted)-1])
		require.Equal(t, 2, strings.Count(quoted, "`"))
		require.Equal(t, byte('`'), quoted[0])
		require.Equal(t, byte('`'), quoted[len(quoted)-1])
	})
}
