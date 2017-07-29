package azure

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	azureLiveTest       bool
	azureClientID       string
	azureClientSecret   string
	azureSubscriptionID string
	azureTenantID       string
	azureResourceGroup  string
	azureDomain         string
)

func init() {
	azureClientID = os.Getenv("AZURE_CLIENT_ID")
	azureClientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	azureSubscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	azureTenantID = os.Getenv("AZURE_TENANT_ID")
	azureResourceGroup = os.Getenv("AZURE_RESOURCE_GROUP")
	azureDomain = os.Getenv("AZURE_DOMAIN")
	if len(azureClientID) > 0 && len(azureClientSecret) > 0 {
		azureLiveTest = true
	}
}

func restoreAzureEnv() {
	os.Setenv("AZURE_CLIENT_ID", azureClientID)
	os.Setenv("AZURE_SUBSCRIPTION_ID", azureSubscriptionID)
}

func TestNewDNSProviderValid(t *testing.T) {
	if !azureLiveTest {
		t.Skip("skipping live test (requires credentials)")
	}
	os.Setenv("AZURE_CLIENT_ID", "")
	_, err := NewDNSProviderCredentials(azureClientID, azureClientSecret, azureSubscriptionID, azureTenantID, azureResourceGroup)
	assert.NoError(t, err)
	restoreAzureEnv()
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	if !azureLiveTest {
		t.Skip("skipping live test (requires credentials)")
	}
	os.Setenv("AZURE_CLIENT_ID", "other")
	_, err := NewDNSProvider()
	assert.NoError(t, err)
	restoreAzureEnv()
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	os.Setenv("AZURE_SUBSCRIPTION_ID", "")
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "Azure configuration missing")
	restoreAzureEnv()
}

func TestLiveAzurePresent(t *testing.T) {
	if !azureLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(azureClientID, azureClientSecret, azureSubscriptionID, azureTenantID, azureResourceGroup)
	assert.NoError(t, err)

	err = provider.Present(azureDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveAzureCleanUp(t *testing.T) {
	if !azureLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(azureClientID, azureClientSecret, azureSubscriptionID, azureTenantID, azureResourceGroup)
	time.Sleep(time.Second * 1)

	assert.NoError(t, err)

	err = provider.CleanUp(azureDomain, "", "123d==")
	assert.NoError(t, err)
}
