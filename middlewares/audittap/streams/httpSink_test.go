package streams

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func TestHttpSink(t *testing.T) {
	var got string

	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		got = string(body)
	}))
	defer stub.Close()

	w1, err := NewHTTPSink("PUT", stub.URL)
	assert.NoError(t, err)

	err = w1.Audit(encodedJSONSample)
	assert.NoError(t, err)

	assert.Equal(t, string(encodedJSONSample.Bytes), got)
}
