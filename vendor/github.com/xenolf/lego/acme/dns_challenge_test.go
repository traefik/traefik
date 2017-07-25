package acme

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

var lookupNameserversTestsOK = []struct {
	fqdn string
	nss  []string
}{
	{"books.google.com.ng.",
		[]string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
	},
	{"www.google.com.",
		[]string{"ns1.google.com.", "ns2.google.com.", "ns3.google.com.", "ns4.google.com."},
	},
	{"physics.georgetown.edu.",
		[]string{"ns1.georgetown.edu.", "ns2.georgetown.edu.", "ns3.georgetown.edu."},
	},
}

var lookupNameserversTestsErr = []struct {
	fqdn  string
	error string
}{
	// invalid tld
	{"_null.n0n0.",
		"Could not determine the zone",
	},
	// invalid domain
	{"_null.com.",
		"Could not determine the zone",
	},
	// invalid domain
	{"in-valid.co.uk.",
		"Could not determine the zone",
	},
}

var findZoneByFqdnTests = []struct {
	fqdn string
	zone string
}{
	{"mail.google.com.", "google.com."}, // domain is a CNAME
	{"foo.google.com.", "google.com."},  // domain is a non-existent subdomain
}

var checkAuthoritativeNssTests = []struct {
	fqdn, value string
	ns          []string
	ok          bool
}{
	// TXT RR w/ expected value
	{"8.8.8.8.asn.routeviews.org.", "151698.8.8.024", []string{"asnums.routeviews.org."},
		true,
	},
	// No TXT RR
	{"ns1.google.com.", "", []string{"ns2.google.com."},
		false,
	},
}

var checkAuthoritativeNssTestsErr = []struct {
	fqdn, value string
	ns          []string
	error       string
}{
	// TXT RR /w unexpected value
	{"8.8.8.8.asn.routeviews.org.", "fe01=", []string{"asnums.routeviews.org."},
		"did not return the expected TXT record",
	},
	// No TXT RR
	{"ns1.google.com.", "fe01=", []string{"ns2.google.com."},
		"did not return the expected TXT record",
	},
}

var checkResolvConfServersTests = []struct {
	fixture  string
	expected []string
	defaults []string
}{
	{"testdata/resolv.conf.1", []string{"10.200.3.249:53", "10.200.3.250:5353", "[2001:4860:4860::8844]:53", "[10.0.0.1]:5353"}, []string{"127.0.0.1:53"}},
	{"testdata/resolv.conf.nonexistant", []string{"127.0.0.1:53"}, []string{"127.0.0.1:53"}},
}

func TestDNSValidServerResponse(t *testing.T) {
	PreCheckDNS = func(fqdn, value string) (bool, error) {
		return true, nil
	}
	privKey, _ := rsa.GenerateKey(rand.Reader, 512)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"dns01\",\"status\":\"valid\",\"uri\":\"http://some.url\",\"token\":\"http8\"}"))
	}))

	manualProvider, _ := NewDNSProviderManual()
	jws := &jws{privKey: privKey, directoryURL: ts.URL}
	solver := &dnsChallenge{jws: jws, validate: validate, provider: manualProvider}
	clientChallenge := challenge{Type: "dns01", Status: "pending", URI: ts.URL, Token: "http8"}

	go func() {
		time.Sleep(time.Second * 2)
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		f.WriteString("\n")
	}()

	if err := solver.Solve(clientChallenge, "example.com"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}
}

func TestPreCheckDNS(t *testing.T) {
	ok, err := PreCheckDNS("acme-staging.api.letsencrypt.org", "fe01=")
	if err != nil || !ok {
		t.Errorf("preCheckDNS failed for acme-staging.api.letsencrypt.org")
	}
}

func TestLookupNameserversOK(t *testing.T) {
	for _, tt := range lookupNameserversTestsOK {
		nss, err := lookupNameservers(tt.fqdn)
		if err != nil {
			t.Fatalf("#%s: got %q; want nil", tt.fqdn, err)
		}

		sort.Strings(nss)
		sort.Strings(tt.nss)

		if !reflect.DeepEqual(nss, tt.nss) {
			t.Errorf("#%s: got %v; want %v", tt.fqdn, nss, tt.nss)
		}
	}
}

func TestLookupNameserversErr(t *testing.T) {
	for _, tt := range lookupNameserversTestsErr {
		_, err := lookupNameservers(tt.fqdn)
		if err == nil {
			t.Fatalf("#%s: expected %q (error); got <nil>", tt.fqdn, tt.error)
		}

		if !strings.Contains(err.Error(), tt.error) {
			t.Errorf("#%s: expected %q (error); got %q", tt.fqdn, tt.error, err)
			continue
		}
	}
}

func TestFindZoneByFqdn(t *testing.T) {
	for _, tt := range findZoneByFqdnTests {
		res, err := FindZoneByFqdn(tt.fqdn, RecursiveNameservers)
		if err != nil {
			t.Errorf("FindZoneByFqdn failed for %s: %v", tt.fqdn, err)
		}
		if res != tt.zone {
			t.Errorf("%s: got %s; want %s", tt.fqdn, res, tt.zone)
		}
	}
}

func TestCheckAuthoritativeNss(t *testing.T) {
	for _, tt := range checkAuthoritativeNssTests {
		ok, _ := checkAuthoritativeNss(tt.fqdn, tt.value, tt.ns)
		if ok != tt.ok {
			t.Errorf("%s: got %t; want %t", tt.fqdn, ok, tt.ok)
		}
	}
}

func TestCheckAuthoritativeNssErr(t *testing.T) {
	for _, tt := range checkAuthoritativeNssTestsErr {
		_, err := checkAuthoritativeNss(tt.fqdn, tt.value, tt.ns)
		if err == nil {
			t.Fatalf("#%s: expected %q (error); got <nil>", tt.fqdn, tt.error)
		}
		if !strings.Contains(err.Error(), tt.error) {
			t.Errorf("#%s: expected %q (error); got %q", tt.fqdn, tt.error, err)
			continue
		}
	}
}

func TestResolveConfServers(t *testing.T) {
	for _, tt := range checkResolvConfServersTests {
		result := getNameservers(tt.fixture, tt.defaults)

		sort.Strings(result)
		sort.Strings(tt.expected)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("#%s: expected %q; got %q", tt.fixture, tt.expected, result)
		}
	}
}
