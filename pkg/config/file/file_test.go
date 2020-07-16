package file

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode_TOML(t *testing.T) {
	f, err := ioutil.TempFile("", "traefik-config-*.toml")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(f.Name())
	}()

	_, err = f.Write([]byte(`
foo = "bar"
fii = "bir"
[yi]
`))
	require.NoError(t, err)

	element := &Yo{
		Fuu: "test",
	}

	err = Decode(f.Name(), element)
	require.NoError(t, err)

	expected := &Yo{
		Foo: "bar",
		Fii: "bir",
		Fuu: "test",
		Yi: &Yi{
			Foo: "foo",
			Fii: "fii",
		},
	}
	assert.Equal(t, expected, element)
}

func TestDecodeContent_TOML(t *testing.T) {
	content := `
foo = "bar"
fii = "bir"
[yi]
`

	element := &Yo{
		Fuu: "test",
	}

	err := DecodeContent(content, ".toml", element)
	require.NoError(t, err)

	expected := &Yo{
		Foo: "bar",
		Fii: "bir",
		Fuu: "test",
		Yi: &Yi{
			Foo: "foo",
			Fii: "fii",
		},
	}
	assert.Equal(t, expected, element)
}

func TestDecodeContent_TOML_rawValue(t *testing.T) {
	content := `
name = "test"
[[meta.aaa]]
	bbb = 1
`

	type Foo struct {
		Name string
		Meta map[string]interface{}
	}

	element := &Foo{}

	err := DecodeContent(content, ".toml", element)
	require.NoError(t, err)

	expected := &Foo{
		Name: "test",
		Meta: map[string]interface{}{"aaa": []interface{}{map[string]interface{}{"bbb": "1"}}},
	}
	assert.Equal(t, expected, element)
}

func TestDecode_YAML(t *testing.T) {
	f, err := ioutil.TempFile("", "traefik-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(f.Name())
	}()

	_, err = f.Write([]byte(`
foo: bar
fii: bir
yi: {}
`))
	require.NoError(t, err)

	element := &Yo{
		Fuu: "test",
	}

	err = Decode(f.Name(), element)
	require.NoError(t, err)

	expected := &Yo{
		Foo: "bar",
		Fii: "bir",
		Fuu: "test",
		Yi: &Yi{
			Foo: "foo",
			Fii: "fii",
		},
	}
	assert.Equal(t, expected, element)
}

func TestDecodeContent_YAML(t *testing.T) {
	content := `
foo: bar
fii: bir
yi: {}
`

	element := &Yo{
		Fuu: "test",
	}

	err := DecodeContent(content, ".yaml", element)
	require.NoError(t, err)

	expected := &Yo{
		Foo: "bar",
		Fii: "bir",
		Fuu: "test",
		Yi: &Yi{
			Foo: "foo",
			Fii: "fii",
		},
	}
	assert.Equal(t, expected, element)
}

func TestDecodeContent_YAML_rawValue(t *testing.T) {
	content := `
name: test
meta:
  aaa:
  - bbb: 1
`

	type Foo struct {
		Name string
		Meta map[string]interface{}
	}

	element := &Foo{}

	err := DecodeContent(content, ".yaml", element)
	require.NoError(t, err)

	expected := &Foo{
		Name: "test",
		Meta: map[string]interface{}{"aaa": []interface{}{map[string]interface{}{"bbb": "1"}}},
	}
	assert.Equal(t, expected, element)
}
