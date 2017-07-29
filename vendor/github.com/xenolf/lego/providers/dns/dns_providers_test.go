package dns

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xenolf/lego/providers/dns/exoscale"
)

var (
	apiKey    string
	apiSecret string
)

func init() {
	apiSecret = os.Getenv("EXOSCALE_API_SECRET")
	apiKey = os.Getenv("EXOSCALE_API_KEY")
}

func restoreExoscaleEnv() {
	os.Setenv("EXOSCALE_API_KEY", apiKey)
	os.Setenv("EXOSCALE_API_SECRET", apiSecret)
}

func TestKnownDNSProviderSuccess(t *testing.T) {
	os.Setenv("EXOSCALE_API_KEY", "abc")
	os.Setenv("EXOSCALE_API_SECRET", "123")
	provider, err := NewDNSChallengeProviderByName("exoscale")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	if reflect.TypeOf(provider) != reflect.TypeOf(&exoscale.DNSProvider{}) {
		t.Errorf("Not loaded correct DNS proviver: %v is not *exoscale.DNSProvider", reflect.TypeOf(provider))
	}
	restoreExoscaleEnv()
}

func TestKnownDNSProviderError(t *testing.T) {
	os.Setenv("EXOSCALE_API_KEY", "")
	os.Setenv("EXOSCALE_API_SECRET", "")
	_, err := NewDNSChallengeProviderByName("exoscale")
	assert.Error(t, err)
	restoreExoscaleEnv()
}

func TestUnknownDNSProvider(t *testing.T) {
	_, err := NewDNSChallengeProviderByName("foobar")
	assert.Error(t, err)
}
