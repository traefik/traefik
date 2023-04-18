package dynamic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeConfig struct {
	Name string `json:"name"`
}

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FakeConfig) DeepCopyInto(out *FakeConfig) {
	*out = *in
}

// DeepCopy is a deepcopy function, copying the receiver, creating a new AddPrefix.
func (in *FakeConfig) DeepCopy() *FakeConfig {
	if in == nil {
		return nil
	}
	out := new(FakeConfig)
	in.DeepCopyInto(out)
	return out
}

type Foo struct {
	Name string
}

func TestPluginConf_DeepCopy_mapOfStruct(t *testing.T) {
	f := &FakeConfig{Name: "bir"}
	p := PluginConf{
		"fii": f,
	}

	clone := p.DeepCopy()
	assert.Equal(t, &p, clone)

	f.Name = "bur"

	assert.NotEqual(t, &p, clone)
}

func TestPluginConf_DeepCopy_map(t *testing.T) {
	m := map[string]interface{}{
		"name": "bar",
	}
	p := PluginConf{
		"config": map[string]interface{}{
			"foo": m,
		},
	}

	clone := p.DeepCopy()
	assert.Equal(t, &p, clone)

	p["one"] = "a"
	m["two"] = "b"

	assert.NotEqual(t, &p, clone)
}

func TestPluginConf_DeepCopy_panic(t *testing.T) {
	p := &PluginConf{
		"config": map[string]interface{}{
			"foo": &Foo{Name: "gigi"},
		},
	}

	assert.Panics(t, func() {
		p.DeepCopy()
	})
}
