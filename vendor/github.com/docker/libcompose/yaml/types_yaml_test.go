package yaml

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

type StructStringorInt struct {
	Foo StringorInt
}

func TestStringorIntYaml(t *testing.T) {
	for _, str := range []string{`{foo: 10}`, `{foo: "10"}`} {
		s := StructStringorInt{}
		yaml.Unmarshal([]byte(str), &s)

		assert.Equal(t, StringorInt(10), s.Foo)

		d, err := yaml.Marshal(&s)
		assert.Nil(t, err)

		s2 := StructStringorInt{}
		yaml.Unmarshal(d, &s2)

		assert.Equal(t, StringorInt(10), s2.Foo)
	}
}

type StructStringorslice struct {
	Foo Stringorslice
}

func TestStringorsliceYaml(t *testing.T) {
	str := `{foo: [bar, baz]}`

	s := StructStringorslice{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, Stringorslice{"bar", "baz"}, s.Foo)

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructStringorslice{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, Stringorslice{"bar", "baz"}, s2.Foo)
}

type StructSliceorMap struct {
	Foos SliceorMap `yaml:"foos,omitempty"`
	Bars []string   `yaml:"bars"`
}

func TestSliceOrMapYaml(t *testing.T) {
	str := `{foos: [bar=baz, far=faz]}`

	s := StructSliceorMap{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, SliceorMap{"bar": "baz", "far": "faz"}, s.Foos)

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructSliceorMap{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, SliceorMap{"bar": "baz", "far": "faz"}, s2.Foos)
}

var sampleStructSliceorMap = `
foos:
  io.rancher.os.bar: baz
  io.rancher.os.far: true
bars: []
`

func TestUnmarshalSliceOrMap(t *testing.T) {
	s := StructSliceorMap{}
	err := yaml.Unmarshal([]byte(sampleStructSliceorMap), &s)
	assert.Equal(t, fmt.Errorf("Cannot unmarshal 'true' of type bool into a string value"), err)
}

func TestStr2SliceOrMapPtrMap(t *testing.T) {
	s := map[string]*StructSliceorMap{"udav": {
		Foos: SliceorMap{"io.rancher.os.bar": "baz", "io.rancher.os.far": "true"},
		Bars: []string{},
	}}
	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := map[string]*StructSliceorMap{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, s, s2)
}

type StructMaporslice struct {
	Foo MaporEqualSlice
}

func contains(list []string, item string) bool {
	for _, test := range list {
		if test == item {
			return true
		}
	}
	return false
}

func TestMaporsliceYaml(t *testing.T) {
	str := `{foo: {bar: baz, far: 1, qux: null}}`

	s := StructMaporslice{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, 3, len(s.Foo))
	assert.True(t, contains(s.Foo, "bar=baz"))
	assert.True(t, contains(s.Foo, "far=1"))
	assert.True(t, contains(s.Foo, "qux"))

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructMaporslice{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, 3, len(s2.Foo))
	assert.True(t, contains(s2.Foo, "bar=baz"))
	assert.True(t, contains(s2.Foo, "far=1"))
	assert.True(t, contains(s2.Foo, "qux"))
}

func TestMapWithEmptyValue(t *testing.T) {
	str := `foo:
  bar: baz
  far: 1
  qux: null
  empty: ""
  lookup:`

	s := StructMaporslice{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, 5, len(s.Foo))
	assert.True(t, contains(s.Foo, "bar=baz"))
	assert.True(t, contains(s.Foo, "far=1"))
	assert.True(t, contains(s.Foo, "qux"))
	assert.True(t, contains(s.Foo, "empty="))
	assert.True(t, contains(s.Foo, "lookup"))

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructMaporslice{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, 5, len(s2.Foo))
	assert.True(t, contains(s2.Foo, "bar=baz"))
	assert.True(t, contains(s2.Foo, "far=1"))
	assert.True(t, contains(s2.Foo, "qux"))
	assert.True(t, contains(s2.Foo, "empty="))
	assert.True(t, contains(s2.Foo, "lookup"))
}

func TestSliceWithEmptyValue(t *testing.T) {
	str := `foo:
  - bar=baz
  - far=1
  - qux=null
  - quotes=""
  - empty=
  - lookup`

	s := StructMaporslice{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, 6, len(s.Foo))
	assert.True(t, contains(s.Foo, "bar=baz"))
	assert.True(t, contains(s.Foo, "far=1"))
	assert.True(t, contains(s.Foo, "qux=null"))
	assert.True(t, contains(s.Foo, "quotes=\"\""))
	assert.True(t, contains(s.Foo, "empty="))
	assert.True(t, contains(s.Foo, "lookup"))

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructMaporslice{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, 6, len(s2.Foo))
	assert.True(t, contains(s2.Foo, "bar=baz"))
	assert.True(t, contains(s2.Foo, "far=1"))
	assert.True(t, contains(s2.Foo, "qux=null"))
	assert.True(t, contains(s2.Foo, "quotes=\"\""))
	assert.True(t, contains(s2.Foo, "empty="))
	assert.True(t, contains(s2.Foo, "lookup"))
}

func TestEqualSliceToMapEqualSign(t *testing.T) {
	slice := MaporEqualSlice{"foo=bar=baz"}
	result := slice.ToMap()
	assert.Equal(t, "bar=baz", result["foo"])

	slice = MaporEqualSlice{"foo=bar=baz=buz"}
	result = slice.ToMap()
	assert.Equal(t, "bar=baz=buz", result["foo"])
}
