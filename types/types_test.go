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
