package http

import (
	"net/http"
	"testing"

	. "github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestShallowHeaders_emptyCase(t *testing.T) {
	hdr := make(http.Header)
	flatHdr := NewHeaders(hdr).Flatten()
	assert.Equal(t, DataMap{}, flatHdr)
}

func TestShallowHeaders_SimplifyCookies(t *testing.T) {
	hdr := make(http.Header)
	hdr.Set("cookie", "a=123; b=456")
	hdr.Add("cookie", "c=789")
	flatHdr := NewHeaders(hdr).SimplifyCookies().Flatten()
	assert.Equal(t, DataMap{
		"cookie": []string{"a=123", "b=456", "c=789"},
	}, flatHdr)
}

func TestShallowHeaders_ClientAndRequestHeaders(t *testing.T) {
	hdr := make(http.Header)
	hdr.Set("x-request-id", "12345")
	hdr.Set("forwarded-for", "abc")
	hdr.Set("proxy-xyz", "def")
	hdr.Set("akamai-abc", "ghi")
	hdr.Set("foo", "bar")
	hdr.Set("x-foo", "foobar")

	var clientHeaders, requestHeaders DataMap
	clientHeaders, requestHeaders = NewHeaders(hdr).ClientAndRequestHeaders()

	assert.Equal(t, DataMap{
		"foo": "bar",
	}, clientHeaders)

	assert.Equal(t, DataMap{
		"forwarded-for": "abc",
		"proxy-xyz":     "def",
		"akamai-abc":    "ghi",
		"x-foo":         "foobar",
	}, requestHeaders)
}
