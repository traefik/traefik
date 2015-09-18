package kingpin

import (
	"io/ioutil"
	"net"
	"net/url"
	"os"

	"github.com/stretchr/testify/assert"

	"testing"
)

func TestParseStrings(t *testing.T) {
	p := parserMixin{}
	v := p.Strings()
	p.value.Set("a")
	p.value.Set("b")
	assert.Equal(t, []string{"a", "b"}, *v)
}

func TestStringsStringer(t *testing.T) {
	target := []string{}
	v := newAccumulator(&target, func(v interface{}) Value { return newStringValue(v.(*string)) })
	v.Set("hello")
	v.Set("world")
	assert.Equal(t, "hello,world", v.String())
}

func TestParseStringMap(t *testing.T) {
	p := parserMixin{}
	v := p.StringMap()
	p.value.Set("a:b")
	p.value.Set("b:c")
	assert.Equal(t, map[string]string{"a": "b", "b": "c"}, *v)
}

func TestParseIP(t *testing.T) {
	p := parserMixin{}
	v := p.IP()
	p.value.Set("10.1.1.2")
	ip := net.ParseIP("10.1.1.2")
	assert.Equal(t, ip, *v)
}

func TestParseURL(t *testing.T) {
	p := parserMixin{}
	v := p.URL()
	p.value.Set("http://w3.org")
	u, err := url.Parse("http://w3.org")
	assert.NoError(t, err)
	assert.Equal(t, *u, **v)
}

func TestParseExistingFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	p := parserMixin{}
	v := p.ExistingFile()
	err = p.value.Set(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, f.Name(), *v)
	err = p.value.Set("/etc/hostsDEFINITELYMISSING")
	assert.Error(t, err)
}

func TestParseTCPAddr(t *testing.T) {
	p := parserMixin{}
	v := p.TCP()
	err := p.value.Set("127.0.0.1:1234")
	assert.NoError(t, err)
	expected, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1234")
	assert.NoError(t, err)
	assert.Equal(t, *expected, **v)
}

func TestParseTCPAddrList(t *testing.T) {
	p := parserMixin{}
	_ = p.TCPList()
	err := p.value.Set("127.0.0.1:1234")
	assert.NoError(t, err)
	err = p.value.Set("127.0.0.1:1235")
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:1234,127.0.0.1:1235", p.value.String())
}
