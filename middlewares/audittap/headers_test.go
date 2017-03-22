package audittap

import (
	. "github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestShallowHeaders_emptyCase(t *testing.T) {
	hdr := make(http.Header)
	flat := NewHeaders(hdr).Flatten("")
	assert.Equal(t, DataMap{}, flat)
}

func TestShallowHeaders_SimplifyCookies(t *testing.T) {
	hdr := make(http.Header)
	hdr.Set("cookie", "a=123; b=456")
	hdr.Add("cookie", "c=789")
	flat := NewHeaders(hdr).SimplifyCookies().Flatten("")
	assert.Equal(t, DataMap{
		"cookie": []string{"a=123", "b=456", "c=789"},
	}, flat)
}

func TestShallowHeaders_DropHopByHopHeaders(t *testing.T) {
	hdr := make(http.Header)
	hdr.Set("connection", "a")
	hdr.Set("keep-alive", "b")
	hdr.Set("proxy-authenticate", "c")
	hdr.Set("proxy-authorization", "d")
	hdr.Set("te", "e")
	hdr.Set("trailers", "f")
	hdr.Set("transfer-encoding", "g")
	hdr.Set("upgrade", "h")
	flat := NewHeaders(hdr).DropHopByHopHeaders().Flatten("")
	assert.Equal(t, DataMap{}, flat)
}

func TestShallowHeaders_CamelCaseKeys(t *testing.T) {
	hdr := make(http.Header)
	hdr.Set("Host", "a")
	hdr.Set("Content-Type", "b")
	hdr.Set("X-Request-ID", "c")
	flat := NewHeaders(hdr).CamelCaseKeys().Flatten("")
	assert.Equal(t, DataMap{
		"host":        "a",
		"contentType": "b",
		"xRequestId":  "c",
	}, flat)
}
