package dnsmadeeasy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testLive      bool
	testAPIKey    string
	testAPISecret string
	testDomain    string
)

func init() {
	testAPIKey = os.Getenv("DNSMADEEASY_API_KEY")
	testAPISecret = os.Getenv("DNSMADEEASY_API_SECRET")
	testDomain = os.Getenv("DNSMADEEASY_DOMAIN")
	os.Setenv("DNSMADEEASY_SANDBOX", "true")
	testLive = len(testAPIKey) > 0 && len(testAPISecret) > 0
}

func TestPresentAndCleanup(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()

	err = provider.Present(testDomain, "", "123d==")
	assert.NoError(t, err)

	err = provider.CleanUp(testDomain, "", "123d==")
	assert.NoError(t, err)
}
