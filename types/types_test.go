package types

import (
	"strings"
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

func TestHeaderRedactions_OneItemShouldBeOneItem(t *testing.T) {
	redactions := HeaderRedactions{}
	redactions.Set("Authorization")
	assert.True(t, len(redactions) == 1)
}

func TestHeaderRedactions_TwoItemsShouldBeTwoItems(t *testing.T) {
	redactions := HeaderRedactions{}
	redactions.Set("Authorization,User-Agent")
	assert.True(t, len(redactions) == 2)
}

func TestHeaderRedactions_EmptyItemsShouldBeSkipped(t *testing.T) {
	redactions := HeaderRedactions{}
	redactions.Set(",Authorization,")
	assert.True(t, len(redactions) == 1)
}

func TestHeaderRedactions_WhitespaceShouldBeStripped(t *testing.T) {
	redactions := HeaderRedactions{}
	redactions.Set(" Authorization  ")
	assert.False(t, strings.Contains(redactions[0], " "))
}
