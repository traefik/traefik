package consulcatalog

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHelpers_InArray(t *testing.T) {
	assert.False(t, inArray("foo", []string{"bar", "baz"}))
	assert.False(t, inArray("foo", []string{}))
	assert.True(t, inArray("foo", []string{"foo"}))
	assert.True(t, inArray("foo", []string{"bar", "baz", "foo", "boo"}))
}

func TestHelpers_InArrayPrefix(t *testing.T) {
	var value string
	var ok bool

	value, ok = inArrayPrefix("foo=", []string{""})
	assert.False(t, ok)

	value, ok = inArrayPrefix("foo=", []string{"foo"})
	assert.False(t, ok)

	value, ok = inArrayPrefix("foo=", []string{"foo="})
	assert.True(t, ok)
	assert.Equal(t, "", value)

	value, ok = inArrayPrefix("foo=", []string{"foo=bar"})
	assert.True(t, ok)
	assert.Equal(t, "bar", value)
}

func TestConvertLabels(t *testing.T) {
	result, err := convertLabels([]string{})
	require.NoError(t, err)
	assert.Len(t, result, 0)

	result, err = convertLabels([]string{"label1", "foo=bar", "bar=baz=baz2"})
	require.NoError(t, err)
	require.Len(t, result, 3)

	e, ok := result["label1"]
	assert.True(t, ok)
	assert.Equal(t, "", e)

	e, ok = result["foo"]
	assert.True(t, ok)
	assert.Equal(t, "bar", e)

	e, ok = result["bar"]
	assert.True(t, ok)
	assert.Equal(t, "baz=baz2", e)
}
