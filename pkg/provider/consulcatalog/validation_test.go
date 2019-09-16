package consulcatalog

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConsulCatalog_ValidateConfig_Entrypoints(t *testing.T) {
	var err error
	p := &Provider{
		Protocol: "http",
	}

	err = p.validateConfig()
	require.Error(t, err)
	assert.Equal(t, "default entrypoints must be specified", err.Error())

	p.Entrypoints = []string{"foo"}
	err = p.validateConfig()
	require.NoError(t, err)
}

func TestConsulCatalog_ValidateConfig_Protocol(t *testing.T) {
	var err error
	p := &Provider{
		Entrypoints: []string{"web"},
	}

	err = p.validateConfig()
	require.Error(t, err)
	assert.Equal(t, "wrong protocol specified, allowed values are http and tcp", err.Error())

	p.Protocol = "foo"
	require.Error(t, err)
	assert.Equal(t, "wrong protocol specified, allowed values are http and tcp", err.Error())

	p.Protocol = "http"
	err = p.validateConfig()
	require.NoError(t, err)

	p.Protocol = "tcp"
	err = p.validateConfig()
	require.NoError(t, err)
}
