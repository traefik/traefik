package ns1

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	liveTest bool
	apiKey   string
	domain   string
)

func init() {
	apiKey = os.Getenv("NS1_API_KEY")
	domain = os.Getenv("NS1_DOMAIN")
	if len(apiKey) > 0 && len(domain) > 0 {
		liveTest = true
	}
}

func restoreNS1Env() {
	os.Setenv("NS1_API_KEY", apiKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	os.Setenv("NS1_API_KEY", "")
	_, err := NewDNSProviderCredentials("123")
	assert.NoError(t, err)
	restoreNS1Env()
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	os.Setenv("NS1_API_KEY", "")
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "NS1 credentials missing")
	restoreNS1Env()
}

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(apiKey)
	assert.NoError(t, err)

	err = provider.Present(domain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProviderCredentials(apiKey)
	assert.NoError(t, err)

	err = provider.CleanUp(domain, "", "123d==")
	assert.NoError(t, err)
}
