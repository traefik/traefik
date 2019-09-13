package consulcatalog

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConsulCatalog_ValidateConfig_Protocol(t *testing.T) {
	var err error
	p := &Provider{}

	p.Protocol = ""
	err = p.validateConfig()
	require.Error(t, err)
	assert.Equal(t, "wrong protocol '', allowed 'http' or 'tcp'", err.Error())

	p.Protocol = "wrong value"
	err = p.validateConfig()
	require.Error(t, err)
	assert.Equal(t, "wrong protocol 'wrong value', allowed 'http' or 'tcp'", err.Error())

	p.Protocol = "http"
	err = p.validateConfig()
	require.NoError(t, err)

	p.Protocol = "tcp"
	err = p.validateConfig()
	require.NoError(t, err)
}
