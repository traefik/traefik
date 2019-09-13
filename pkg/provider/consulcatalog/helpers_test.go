package consulcatalog

import (
	"github.com/stretchr/testify/assert"
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
