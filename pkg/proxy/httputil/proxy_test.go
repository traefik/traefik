package httputil

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectorBuilder(t *testing.T) {
	tests := []struct {
		name            string
		target          *url.URL
		passHostHeader  bool
		preservePath    bool
		incomingURL     string
		expectedScheme  string
		expectedHost    string
		expectedPath    string
		expectedRawPath string
		expectedQuery   string
	}{
		{
			name:           "Basic proxy",
			target:         mustParseURL("http://example.com"),
			passHostHeader: false,
			preservePath:   false,
			incomingURL:    "http://localhost/test?param=value",
			expectedScheme: "http",
			expectedHost:   "example.com",
			expectedPath:   "/test",
			expectedQuery:  "param=value",
		},
		{
			name:           "HTTPS target",
			target:         mustParseURL("https://secure.example.com"),
			passHostHeader: false,
			preservePath:   false,
			incomingURL:    "http://localhost/secure",
			expectedScheme: "https",
			expectedHost:   "secure.example.com",
			expectedPath:   "/secure",
		},
		{
			name:           "Pass host header",
			target:         mustParseURL("http://example.com"),
			passHostHeader: true,
			preservePath:   false,
			incomingURL:    "http://original.host/test",
			expectedScheme: "http",
			expectedHost:   "example.com",
			expectedPath:   "/test",
		},
		{
			name:           "Keep path",
			target:         mustParseURL("http://example.com/base"),
			passHostHeader: false,
			preservePath:   true,
			incomingURL:    "http://localhost/test",
			expectedScheme: "http",
			expectedHost:   "example.com",
			expectedPath:   "/base/test",
		},
		{
			name:           "Handle semicolons in query",
			target:         mustParseURL("http://example.com"),
			passHostHeader: false,
			preservePath:   false,
			incomingURL:    "http://localhost/test?param1=value1;param2=value2",
			expectedScheme: "http",
			expectedHost:   "example.com",
			expectedPath:   "/test",
			expectedQuery:  "param1=value1&param2=value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			director := directorBuilder(tt.target, tt.passHostHeader, tt.preservePath)

			incomingURL := mustParseURL(tt.incomingURL)
			req := &http.Request{
				URL:        incomingURL,
				Host:       incomingURL.Host,
				RequestURI: tt.incomingURL,
			}

			director(req)

			assert.Equal(t, tt.expectedScheme, req.URL.Scheme)
			assert.Equal(t, tt.expectedHost, req.URL.Host)
			assert.Equal(t, tt.expectedPath, req.URL.Path)
			assert.Equal(t, tt.expectedRawPath, req.URL.RawPath)
			assert.Equal(t, tt.expectedQuery, req.URL.RawQuery)
			assert.Empty(t, req.RequestURI)
			assert.Equal(t, "HTTP/1.1", req.Proto)
			assert.Equal(t, 1, req.ProtoMajor)
			assert.Equal(t, 1, req.ProtoMinor)
			assert.False(t, !tt.passHostHeader && req.Host != req.URL.Host)
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
