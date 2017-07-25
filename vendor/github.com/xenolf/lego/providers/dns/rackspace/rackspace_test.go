package rackspace

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	rackspaceLiveTest bool
	rackspaceUser     string
	rackspaceAPIKey   string
	rackspaceDomain   string
	testAPIURL        string
)

func init() {
	rackspaceUser = os.Getenv("RACKSPACE_USER")
	rackspaceAPIKey = os.Getenv("RACKSPACE_API_KEY")
	rackspaceDomain = os.Getenv("RACKSPACE_DOMAIN")
	if len(rackspaceUser) > 0 && len(rackspaceAPIKey) > 0 && len(rackspaceDomain) > 0 {
		rackspaceLiveTest = true
	}
}

func testRackspaceEnv() {
	rackspaceAPIURL = testAPIURL
	os.Setenv("RACKSPACE_USER", "testUser")
	os.Setenv("RACKSPACE_API_KEY", "testKey")
}

func liveRackspaceEnv() {
	rackspaceAPIURL = "https://identity.api.rackspacecloud.com/v2.0/tokens"
	os.Setenv("RACKSPACE_USER", rackspaceUser)
	os.Setenv("RACKSPACE_API_KEY", rackspaceAPIKey)
}

func startTestServers() (identityAPI, dnsAPI *httptest.Server) {
	dnsAPI = httptest.NewServer(dnsMux())
	dnsEndpoint := dnsAPI.URL + "/123456"

	identityAPI = httptest.NewServer(identityHandler(dnsEndpoint))
	testAPIURL = identityAPI.URL + "/"
	return
}

func closeTestServers(identityAPI, dnsAPI *httptest.Server) {
	identityAPI.Close()
	dnsAPI.Close()
}

func identityHandler(dnsEndpoint string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp, found := jsonMap[string(reqBody)]
		if !found {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp = strings.Replace(resp, "https://dns.api.rackspacecloud.com/v1.0/123456", dnsEndpoint, 1)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, resp)
	})
}

func dnsMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Used by `getHostedZoneID()` finding `zoneID` "?name=example.com"
	mux.HandleFunc("/123456/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") == "example.com" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["zoneDetails"])
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	})

	mux.HandleFunc("/123456/domains/112233/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// Used by `Present()` creating the TXT record
		case http.MethodPost:
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp, found := jsonMap[string(reqBody)]
			if !found {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(w, resp)
		// Used by `findTxtRecord()` finding `record.ID` "?type=TXT&name=_acme-challenge.example.com"
		case http.MethodGet:
			if r.URL.Query().Get("type") == "TXT" && r.URL.Query().Get("name") == "_acme-challenge.example.com" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, jsonMap["recordDetails"])
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		// Used by `CleanUp()` deleting the TXT record "?id=445566"
		case http.MethodDelete:
			if r.URL.Query().Get("id") == "TXT-654321" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, jsonMap["recordDelete"])
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Not Found for Request: (%+v)\n\n", r)
	})

	return mux
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	testRackspaceEnv()
	_, err := NewDNSProviderCredentials("", "")
	assert.EqualError(t, err, "Rackspace credentials missing")
}

func TestOfflineRackspaceValid(t *testing.T) {
	testRackspaceEnv()
	provider, err := NewDNSProviderCredentials(os.Getenv("RACKSPACE_USER"), os.Getenv("RACKSPACE_API_KEY"))

	assert.NoError(t, err)
	assert.Equal(t, provider.token, "testToken", "The token should match")
}

func TestOfflineRackspacePresent(t *testing.T) {
	testRackspaceEnv()
	provider, err := NewDNSProvider()

	if assert.NoError(t, err) {
		err = provider.Present("example.com", "token", "keyAuth")
		assert.NoError(t, err)
	}
}

func TestOfflineRackspaceCleanUp(t *testing.T) {
	testRackspaceEnv()
	provider, err := NewDNSProvider()

	if assert.NoError(t, err) {
		err = provider.CleanUp("example.com", "token", "keyAuth")
		assert.NoError(t, err)
	}
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

	liveRackspaceEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)
	assert.Contains(t, provider.cloudDNSEndpoint, "https://dns.api.rackspacecloud.com/v1.0/", "The endpoint URL should contain the base")
}

func TestRackspacePresent(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

	liveRackspaceEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(rackspaceDomain, "", "112233445566==")
	assert.NoError(t, err)
}

func TestRackspaceCleanUp(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 15)

	liveRackspaceEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(rackspaceDomain, "", "112233445566==")
	assert.NoError(t, err)
}

func TestMain(m *testing.M) {
	identityAPI, dnsAPI := startTestServers()
	defer closeTestServers(identityAPI, dnsAPI)
	os.Exit(m.Run())
}

var jsonMap = map[string]string{
	`{"auth":{"RAX-KSKEY:apiKeyCredentials":{"username":"testUser","apiKey":"testKey"}}}`: `{"access":{"token":{"id":"testToken","expires":"1970-01-01T00:00:00.000Z","tenant":{"id":"123456","name":"123456"},"RAX-AUTH:authenticatedBy":["APIKEY"]},"serviceCatalog":[{"type":"rax:dns","endpoints":[{"publicURL":"https://dns.api.rackspacecloud.com/v1.0/123456","tenantId":"123456"}],"name":"cloudDNS"}],"user":{"id":"fakeUseID","name":"testUser"}}}`,
	"zoneDetails": `{"domains":[{"name":"example.com","id":112233,"emailAddress":"hostmaster@example.com","updated":"1970-01-01T00:00:00.000+0000","created":"1970-01-01T00:00:00.000+0000"}],"totalEntries":1}`,
	`{"records":[{"name":"_acme-challenge.example.com","type":"TXT","data":"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM","ttl":300}]}`: `{"request":"{\"records\":[{\"name\":\"_acme-challenge.example.com\",\"type\":\"TXT\",\"data\":\"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM\",\"ttl\":300}]}","status":"RUNNING","verb":"POST","jobId":"00000000-0000-0000-0000-0000000000","callbackUrl":"https://dns.api.rackspacecloud.com/v1.0/123456/status/00000000-0000-0000-0000-0000000000","requestUrl":"https://dns.api.rackspacecloud.com/v1.0/123456/domains/112233/records"}`,
	"recordDetails": `{"records":[{"name":"_acme-challenge.example.com","id":"TXT-654321","type":"TXT","data":"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM","ttl":300,"updated":"1970-01-01T00:00:00.000+0000","created":"1970-01-01T00:00:00.000+0000"}]}`,
	"recordDelete":  `{"status":"RUNNING","verb":"DELETE","jobId":"00000000-0000-0000-0000-0000000000","callbackUrl":"https://dns.api.rackspacecloud.com/v1.0/123456/status/00000000-0000-0000-0000-0000000000","requestUrl":"https://dns.api.rackspacecloud.com/v1.0/123456/domains/112233/recordsid=TXT-654321"}`,
}
