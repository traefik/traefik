package dyn

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	dynLiveTest     bool
	dynCustomerName string
	dynUserName     string
	dynPassword     string
	dynDomain       string
)

func init() {
	dynCustomerName = os.Getenv("DYN_CUSTOMER_NAME")
	dynUserName = os.Getenv("DYN_USER_NAME")
	dynPassword = os.Getenv("DYN_PASSWORD")
	dynDomain = os.Getenv("DYN_DOMAIN")
	if len(dynCustomerName) > 0 && len(dynUserName) > 0 && len(dynPassword) > 0 && len(dynDomain) > 0 {
		dynLiveTest = true
	}
}

func TestLiveDynPresent(t *testing.T) {
	if !dynLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(dynDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveDynCleanUp(t *testing.T) {
	if !dynLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(dynDomain, "", "123d==")
	assert.NoError(t, err)
}
