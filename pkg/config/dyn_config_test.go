package config

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepCopy(t *testing.T) {
	content, err := ioutil.ReadFile("./fixtures/sample.toml")
	require.NoError(t, err)

	cfg := &Configuration{}
	err = toml.Unmarshal(content, &cfg)
	require.NoError(t, err)

	cfgCopy := cfg
	assert.Equal(t, reflect.ValueOf(cfgCopy), reflect.ValueOf(cfg))
	assert.Equal(t, reflect.ValueOf(cfgCopy), reflect.ValueOf(cfg))
	assert.Equal(t, cfgCopy, cfg)

	cfgDeepCopy := cfg.DeepCopy()
	assert.NotEqual(t, reflect.ValueOf(cfgDeepCopy), reflect.ValueOf(cfg))
	assert.Equal(t, reflect.TypeOf(cfgDeepCopy), reflect.TypeOf(cfg))
	assert.Equal(t, cfgDeepCopy, cfg)
}
