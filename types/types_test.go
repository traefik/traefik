package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaders_ShouldReturnFalseWhenNotHasCustomHeadersDefined(t *testing.T) {
	headers := Headers{}

	assert.False(t, headers.HasCustomHeadersDefined())
}

func TestHeaders_ShouldReturnTrueWhenHasCustomHeadersDefined(t *testing.T) {
	headers := Headers{}

	headers.CustomRequestHeaders = map[string]string{
		"foo": "bar",
	}

	assert.True(t, headers.HasCustomHeadersDefined())
}

func TestHeaders_ShouldReturnFalseWhenNotHasSecureHeadersDefined(t *testing.T) {
	headers := Headers{}

	assert.False(t, headers.HasSecureHeadersDefined())
}

func TestHeaders_ShouldReturnTrueWhenHasSecureHeadersDefined(t *testing.T) {
	headers := Headers{}

	headers.SSLRedirect = true

	assert.True(t, headers.HasSecureHeadersDefined())
}

func TestServerUnmarshalTOML(t *testing.T) {
	var tests = []struct {
		err string
		in  interface{}
		out Server
	}{
		{
			err: "",
			in:  map[string]interface{}{},
			out: Server{
				URL:    "",
				Weight: 1,
			},
		},
		{
			err: "",
			in: map[string]interface{}{
				"url": "http://example.com",
			},
			out: Server{
				URL:    "http://example.com",
				Weight: 1,
			},
		},
		{
			err: "",
			in: map[string]interface{}{
				"weight": int64(0),
			},
			out: Server{
				URL:    "",
				Weight: 0,
			},
		},
		{
			err: "",
			in: map[string]interface{}{
				"url":    "http://example.com",
				"weight": int64(0),
			},
			out: Server{
				URL:    "http://example.com",
				Weight: 0,
			},
		},
		{
			err: "Server UnmarshalTOML want (map[string]interface{})",
			in:  "invalid type",
			out: Server{},
		},
		{
			err: "toml: server url must be string",
			in: map[string]interface{}{
				"url": []string{"invalid type"},
			},
			out: Server{},
		},
		{
			err: "toml: server weight must be int64",
			in: map[string]interface{}{
				"weight": "invalid type",
			},
			out: Server{},
		},
	}

	for _, tt := range tests {
		var s Server

		if err := s.UnmarshalTOML(tt.in); err != nil && err.Error() != tt.err {
			t.Fatalf("UnmarshalTOML(%#v) got err %q, want err %q", tt.in, tt.err, err)
		}

		if s != tt.out {
			t.Fatalf("UnmarshalTOML(%#v) got %+v, want %+v", tt.in, s, tt.out)
		}
	}
}
