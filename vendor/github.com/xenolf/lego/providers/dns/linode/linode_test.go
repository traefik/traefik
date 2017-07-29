package linode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timewasted/linode"
	"github.com/timewasted/linode/dns"
)

type (
	LinodeResponse struct {
		Action string                 `json:"ACTION"`
		Data   interface{}            `json:"DATA"`
		Errors []linode.ResponseError `json:"ERRORARRAY"`
	}
	MockResponse struct {
		Response interface{}
		Errors   []linode.ResponseError
	}
	MockResponseMap map[string]MockResponse
)

var (
	apiKey     string
	isTestLive bool
)

func init() {
	apiKey = os.Getenv("LINODE_API_KEY")
	isTestLive = len(apiKey) != 0
}

func restoreEnv() {
	os.Setenv("LINODE_API_KEY", apiKey)
}

func newMockServer(t *testing.T, responses MockResponseMap) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure that we support the requested action.
		action := r.URL.Query().Get("api_action")
		resp, ok := responses[action]
		if !ok {
			msg := fmt.Sprintf("Unsupported mock action: %s", action)
			require.FailNow(t, msg)
		}

		// Build the response that the server will return.
		linodeResponse := LinodeResponse{
			Action: action,
			Data:   resp.Response,
			Errors: resp.Errors,
		}
		rawResponse, err := json.Marshal(linodeResponse)
		if err != nil {
			msg := fmt.Sprintf("Failed to JSON encode response: %v", err)
			require.FailNow(t, msg)
		}

		// Send the response.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(rawResponse)
	}))

	time.Sleep(100 * time.Millisecond)
	return srv
}

func TestNewDNSProviderWithEnv(t *testing.T) {
	os.Setenv("LINODE_API_KEY", "testing")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderWithoutEnv(t *testing.T) {
	os.Setenv("LINODE_API_KEY", "")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "Linode credentials missing")
}

func TestNewDNSProviderCredentialsWithKey(t *testing.T) {
	_, err := NewDNSProviderCredentials("testing")
	assert.NoError(t, err)
}

func TestNewDNSProviderCredentialsWithoutKey(t *testing.T) {
	_, err := NewDNSProviderCredentials("")
	assert.EqualError(t, err, "Linode credentials missing")
}

func TestDNSProvider_Present(t *testing.T) {
	os.Setenv("LINODE_API_KEY", "testing")
	defer restoreEnv()
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"domain.list": MockResponse{
			Response: []dns.Domain{
				dns.Domain{
					Domain:   domain,
					DomainID: 1234,
				},
			},
		},
		"domain.resource.create": MockResponse{
			Response: dns.ResourceResponse{
				ResourceID: 1234,
			},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()
	p.linode.ToLinode().SetEndpoint(mockSrv.URL)

	err = p.Present(domain, "", keyAuth)
	assert.NoError(t, err)
}

func TestDNSProvider_PresentNoDomain(t *testing.T) {
	os.Setenv("LINODE_API_KEY", "testing")
	defer restoreEnv()
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"domain.list": MockResponse{
			Response: []dns.Domain{
				dns.Domain{
					Domain:   "foobar.com",
					DomainID: 1234,
				},
			},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()
	p.linode.ToLinode().SetEndpoint(mockSrv.URL)

	err = p.Present(domain, "", keyAuth)
	assert.EqualError(t, err, "dns: requested domain not found")
}

func TestDNSProvider_PresentCreateFailed(t *testing.T) {
	os.Setenv("LINODE_API_KEY", "testing")
	defer restoreEnv()
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"domain.list": MockResponse{
			Response: []dns.Domain{
				dns.Domain{
					Domain:   domain,
					DomainID: 1234,
				},
			},
		},
		"domain.resource.create": MockResponse{
			Response: nil,
			Errors: []linode.ResponseError{
				linode.ResponseError{
					Code:    1234,
					Message: "Failed to create domain resource",
				},
			},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()
	p.linode.ToLinode().SetEndpoint(mockSrv.URL)

	err = p.Present(domain, "", keyAuth)
	assert.EqualError(t, err, "Failed to create domain resource")
}

func TestDNSProvider_PresentLive(t *testing.T) {
	if !isTestLive {
		t.Skip("Skipping live test")
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	os.Setenv("LINODE_API_KEY", "testing")
	defer restoreEnv()
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"domain.list": MockResponse{
			Response: []dns.Domain{
				dns.Domain{
					Domain:   domain,
					DomainID: 1234,
				},
			},
		},
		"domain.resource.list": MockResponse{
			Response: []dns.Resource{
				dns.Resource{
					DomainID:   1234,
					Name:       "_acme-challenge",
					ResourceID: 1234,
					Target:     "ElbOJKOkFWiZLQeoxf-wb3IpOsQCdvoM0y_wn0TEkxM",
					Type:       "TXT",
				},
			},
		},
		"domain.resource.delete": MockResponse{
			Response: dns.ResourceResponse{
				ResourceID: 1234,
			},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()
	p.linode.ToLinode().SetEndpoint(mockSrv.URL)

	err = p.CleanUp(domain, "", keyAuth)
	assert.NoError(t, err)
}

func TestDNSProvider_CleanUpNoDomain(t *testing.T) {
	os.Setenv("LINODE_API_KEY", "testing")
	defer restoreEnv()
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"domain.list": MockResponse{
			Response: []dns.Domain{
				dns.Domain{
					Domain:   "foobar.com",
					DomainID: 1234,
				},
			},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()
	p.linode.ToLinode().SetEndpoint(mockSrv.URL)

	err = p.CleanUp(domain, "", keyAuth)
	assert.EqualError(t, err, "dns: requested domain not found")
}

func TestDNSProvider_CleanUpDeleteFailed(t *testing.T) {
	os.Setenv("LINODE_API_KEY", "testing")
	defer restoreEnv()
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"domain.list": MockResponse{
			Response: []dns.Domain{
				dns.Domain{
					Domain:   domain,
					DomainID: 1234,
				},
			},
		},
		"domain.resource.list": MockResponse{
			Response: []dns.Resource{
				dns.Resource{
					DomainID:   1234,
					Name:       "_acme-challenge",
					ResourceID: 1234,
					Target:     "ElbOJKOkFWiZLQeoxf-wb3IpOsQCdvoM0y_wn0TEkxM",
					Type:       "TXT",
				},
			},
		},
		"domain.resource.delete": MockResponse{
			Response: nil,
			Errors: []linode.ResponseError{
				linode.ResponseError{
					Code:    1234,
					Message: "Failed to delete domain resource",
				},
			},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()
	p.linode.ToLinode().SetEndpoint(mockSrv.URL)

	err = p.CleanUp(domain, "", keyAuth)
	assert.EqualError(t, err, "Failed to delete domain resource")
}

func TestDNSProvider_CleanUpLive(t *testing.T) {
	if !isTestLive {
		t.Skip("Skipping live test")
	}
}
