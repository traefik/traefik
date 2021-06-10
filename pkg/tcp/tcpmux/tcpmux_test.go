package tcpmux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTCPMux(t *testing.T) {
	router, err := NewTCPMux()
	require.NoError(t, err)
	assert.NotNil(t, router)
}

func TestCreateRoute(t *testing.T) {
	route := NewRoute()
	assert.NotNil(t, route)
}

func TestCreateMatchers(t *testing.T) {
	c := NewClientIP("10.1")
	assert.NotNil(t, c)
	s := NewSNIHost("foo")
	assert.NotNil(t, s)
}

func TestAddMatchers(t *testing.T) {
	route := NewRoute()
	require.NotNil(t, route)

	route.AddMatcher(NewClientIP("10.1"))
}
