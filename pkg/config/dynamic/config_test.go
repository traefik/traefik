package dynamic

import (
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepCopy(t *testing.T) {
	cfg := &Configuration{}
	_, err := toml.DecodeFile("./fixtures/sample.toml", &cfg)
	require.NoError(t, err)

	cfgCopy := cfg
	assert.Equal(t, reflect.ValueOf(cfgCopy), reflect.ValueOf(cfg))
	assert.Equal(t, reflect.ValueOf(cfgCopy), reflect.ValueOf(cfg))
	assert.Equal(t, cfgCopy, cfg)

	cfgDeepCopy := cfg.DeepCopy()
	assert.NotEqual(t, reflect.ValueOf(cfgDeepCopy), reflect.ValueOf(cfg))
	assert.Equal(t, reflect.TypeOf(cfgDeepCopy), reflect.TypeOf(cfg))
	assert.Equal(t, cfgDeepCopy, cfg)

	// Update cfg
	cfg.HTTP.Routers["powpow"] = &Router{}

	assert.Equal(t, reflect.ValueOf(cfgCopy), reflect.ValueOf(cfg))
	assert.Equal(t, reflect.ValueOf(cfgCopy), reflect.ValueOf(cfg))
	assert.Equal(t, cfgCopy, cfg)

	assert.NotEqual(t, reflect.ValueOf(cfgDeepCopy), reflect.ValueOf(cfg))
	assert.Equal(t, reflect.TypeOf(cfgDeepCopy), reflect.TypeOf(cfg))
	assert.NotEqual(t, cfgDeepCopy, cfg)
}
