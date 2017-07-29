package namecheap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	fakeUser     = "foo"
	fakeKey      = "bar"
	fakeClientIP = "10.0.0.1"

	tlds = map[string]string{
		"com.au": "com.au",
		"com":    "com",
		"co.uk":  "co.uk",
		"uk":     "uk",
		"edu":    "edu",
		"co.com": "co.com",
		"za.com": "za.com",
	}
)

func assertEq(t *testing.T, variable, got, want string) {
	if got != want {
		t.Errorf("Expected %s to be '%s' but got '%s'", variable, want, got)
	}
}

func assertHdr(tc *testcase, t *testing.T, values *url.Values) {
	ch, _ := newChallenge(tc.domain, "", tlds)

	assertEq(t, "ApiUser", values.Get("ApiUser"), fakeUser)
	assertEq(t, "ApiKey", values.Get("ApiKey"), fakeKey)
	assertEq(t, "UserName", values.Get("UserName"), fakeUser)
	assertEq(t, "ClientIp", values.Get("ClientIp"), fakeClientIP)
	assertEq(t, "SLD", values.Get("SLD"), ch.sld)
	assertEq(t, "TLD", values.Get("TLD"), ch.tld)
}

func mockServer(tc *testcase, t *testing.T, w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		values := r.URL.Query()
		cmd := values.Get("Command")
		switch cmd {
		case "namecheap.domains.dns.getHosts":
			assertHdr(tc, t, &values)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, tc.getHostsResponse)
		case "namecheap.domains.getTldList":
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, responseGetTlds)
		default:
			t.Errorf("Unexpected GET command: %s", cmd)
		}

	case "POST":
		r.ParseForm()
		values := r.Form
		cmd := values.Get("Command")
		switch cmd {
		case "namecheap.domains.dns.setHosts":
			assertHdr(tc, t, &values)
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tc.setHostsResponse)
		default:
			t.Errorf("Unexpected POST command: %s", cmd)
		}

	default:
		t.Errorf("Unexpected http method: %s", r.Method)

	}
}

func testGetHosts(tc *testcase, t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			mockServer(tc, t, w, r)
		}))
	defer mock.Close()

	prov := &DNSProvider{
		baseURL:  mock.URL,
		apiUser:  fakeUser,
		apiKey:   fakeKey,
		clientIP: fakeClientIP,
	}

	ch, _ := newChallenge(tc.domain, "", tlds)
	hosts, err := prov.getHosts(ch)
	if tc.errString != "" {
		if err == nil || err.Error() != tc.errString {
			t.Errorf("Namecheap getHosts case %s expected error", tc.name)
		}
	} else {
		if err != nil {
			t.Errorf("Namecheap getHosts case %s failed\n%v", tc.name, err)
		}
	}

next1:
	for _, h := range hosts {
		for _, th := range tc.hosts {
			if h == th {
				continue next1
			}
		}
		t.Errorf("getHosts case %s unexpected record [%s:%s:%s]",
			tc.name, h.Type, h.Name, h.Address)
	}

next2:
	for _, th := range tc.hosts {
		for _, h := range hosts {
			if h == th {
				continue next2
			}
		}
		t.Errorf("getHosts case %s missing record [%s:%s:%s]",
			tc.name, th.Type, th.Name, th.Address)
	}
}

func mockDNSProvider(url string) *DNSProvider {
	return &DNSProvider{
		baseURL:  url,
		apiUser:  fakeUser,
		apiKey:   fakeKey,
		clientIP: fakeClientIP,
	}
}

func testSetHosts(tc *testcase, t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			mockServer(tc, t, w, r)
		}))
	defer mock.Close()

	prov := mockDNSProvider(mock.URL)
	ch, _ := newChallenge(tc.domain, "", tlds)
	hosts, err := prov.getHosts(ch)
	if tc.errString != "" {
		if err == nil || err.Error() != tc.errString {
			t.Errorf("Namecheap getHosts case %s expected error", tc.name)
		}
	} else {
		if err != nil {
			t.Errorf("Namecheap getHosts case %s failed\n%v", tc.name, err)
		}
	}
	if err != nil {
		return
	}

	err = prov.setHosts(ch, hosts)
	if err != nil {
		t.Errorf("Namecheap setHosts case %s failed", tc.name)
	}
}

func testPresent(tc *testcase, t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			mockServer(tc, t, w, r)
		}))
	defer mock.Close()

	prov := mockDNSProvider(mock.URL)
	err := prov.Present(tc.domain, "", "dummyKey")
	if tc.errString != "" {
		if err == nil || err.Error() != tc.errString {
			t.Errorf("Namecheap Present case %s expected error", tc.name)
		}
	} else {
		if err != nil {
			t.Errorf("Namecheap Present case %s failed\n%v", tc.name, err)
		}
	}
}

func testCleanUp(tc *testcase, t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			mockServer(tc, t, w, r)
		}))
	defer mock.Close()

	prov := mockDNSProvider(mock.URL)
	err := prov.CleanUp(tc.domain, "", "dummyKey")
	if tc.errString != "" {
		if err == nil || err.Error() != tc.errString {
			t.Errorf("Namecheap CleanUp case %s expected error", tc.name)
		}
	} else {
		if err != nil {
			t.Errorf("Namecheap CleanUp case %s failed\n%v", tc.name, err)
		}
	}
}

func TestNamecheap(t *testing.T) {
	for _, tc := range testcases {
		testGetHosts(&tc, t)
		testSetHosts(&tc, t)
		testPresent(&tc, t)
		testCleanUp(&tc, t)
	}
}

func TestNamecheapDomainSplit(t *testing.T) {
	tests := []struct {
		domain string
		valid  bool
		tld    string
		sld    string
		host   string
	}{
		{"a.b.c.test.co.uk", true, "co.uk", "test", "a.b.c"},
		{"test.co.uk", true, "co.uk", "test", ""},
		{"test.com", true, "com", "test", ""},
		{"test.co.com", true, "co.com", "test", ""},
		{"www.test.com.au", true, "com.au", "test", "www"},
		{"www.za.com", true, "za.com", "www", ""},
		{"", false, "", "", ""},
		{"a", false, "", "", ""},
		{"com", false, "", "", ""},
		{"co.com", false, "", "", ""},
		{"co.uk", false, "", "", ""},
		{"test.au", false, "", "", ""},
		{"za.com", false, "", "", ""},
		{"www.za", false, "", "", ""},
		{"www.test.au", false, "", "", ""},
		{"www.test.unk", false, "", "", ""},
	}

	for _, test := range tests {
		valid := true
		ch, err := newChallenge(test.domain, "", tlds)
		if err != nil {
			valid = false
		}

		if test.valid && !valid {
			t.Errorf("Expected '%s' to split", test.domain)
		} else if !test.valid && valid {
			t.Errorf("Expected '%s' to produce error", test.domain)
		}

		if test.valid && valid {
			assertEq(t, "domain", ch.domain, test.domain)
			assertEq(t, "tld", ch.tld, test.tld)
			assertEq(t, "sld", ch.sld, test.sld)
			assertEq(t, "host", ch.host, test.host)
		}
	}
}

type testcase struct {
	name             string
	domain           string
	hosts            []host
	errString        string
	getHostsResponse string
	setHostsResponse string
}

var testcases = []testcase{
	{
		"Test:Success:1",
		"test.example.com",
		[]host{
			{"A", "home", "10.0.0.1", "10", "1799"},
			{"A", "www", "10.0.0.2", "10", "1200"},
			{"AAAA", "a", "::0", "10", "1799"},
			{"CNAME", "*", "example.com.", "10", "1799"},
			{"MXE", "example.com", "10.0.0.5", "10", "1800"},
			{"URL", "xyz", "https://google.com", "10", "1799"},
		},
		"",
		responseGetHostsSuccess1,
		responseSetHostsSuccess1,
	},
	{
		"Test:Success:2",
		"example.com",
		[]host{
			{"A", "@", "10.0.0.2", "10", "1200"},
			{"A", "www", "10.0.0.3", "10", "60"},
		},
		"",
		responseGetHostsSuccess2,
		responseSetHostsSuccess2,
	},
	{
		"Test:Error:BadApiKey:1",
		"test.example.com",
		nil,
		"Namecheap error: API Key is invalid or API access has not been enabled [1011102]",
		responseGetHostsErrorBadAPIKey1,
		"",
	},
}

var responseGetHostsSuccess1 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.dns.getHosts</RequestedCommand>
  <CommandResponse Type="namecheap.domains.dns.getHosts">
    <DomainDNSGetHostsResult Domain="example.com" EmailType="MXE" IsUsingOurDNS="true">
      <host HostId="217076" Name="www" Type="A" Address="10.0.0.2" MXPref="10" TTL="1200" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217069" Name="home" Type="A" Address="10.0.0.1" MXPref="10" TTL="1799" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217071" Name="a" Type="AAAA" Address="::0" MXPref="10" TTL="1799" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217075" Name="*" Type="CNAME" Address="example.com." MXPref="10" TTL="1799" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217073" Name="example.com" Type="MXE" Address="10.0.0.5" MXPref="10" TTL="1800" AssociatedAppTitle="MXE" FriendlyName="MXE1" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217077" Name="xyz" Type="URL" Address="https://google.com" MXPref="10" TTL="1799" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
    </DomainDNSGetHostsResult>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>3.338</ExecutionTime>
</ApiResponse>`

var responseSetHostsSuccess1 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.dns.setHosts</RequestedCommand>
  <CommandResponse Type="namecheap.domains.dns.setHosts">
    <DomainDNSSetHostsResult Domain="example.com" IsSuccess="true">
      <Warnings />
    </DomainDNSSetHostsResult>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>2.347</ExecutionTime>
</ApiResponse>`

var responseGetHostsSuccess2 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.dns.getHosts</RequestedCommand>
  <CommandResponse Type="namecheap.domains.dns.getHosts">
    <DomainDNSGetHostsResult Domain="example.com" EmailType="MXE" IsUsingOurDNS="true">
      <host HostId="217076" Name="@" Type="A" Address="10.0.0.2" MXPref="10" TTL="1200" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217069" Name="www" Type="A" Address="10.0.0.3" MXPref="10" TTL="60" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
    </DomainDNSGetHostsResult>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>3.338</ExecutionTime>
</ApiResponse>`

var responseSetHostsSuccess2 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.dns.setHosts</RequestedCommand>
  <CommandResponse Type="namecheap.domains.dns.setHosts">
    <DomainDNSSetHostsResult Domain="example.com" IsSuccess="true">
      <Warnings />
    </DomainDNSSetHostsResult>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>2.347</ExecutionTime>
</ApiResponse>`

var responseGetHostsErrorBadAPIKey1 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="ERROR" xmlns="http://api.namecheap.com/xml.response">
  <Errors>
    <Error Number="1011102">API Key is invalid or API access has not been enabled</Error>
  </Errors>
  <Warnings />
  <RequestedCommand />
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>0</ExecutionTime>
</ApiResponse>`

var responseGetTlds = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.getTldList</RequestedCommand>
  <CommandResponse Type="namecheap.domains.getTldList">
    <Tlds>
      <Tld Name="com" NonRealTime="false" MinRegisterYears="1" MaxRegisterYears="10" MinRenewYears="1" MaxRenewYears="10" RenewalMinDays="0" RenewalMaxDays="4000" ReactivateMaxDays="27" MinTransferYears="1" MaxTransferYears="1" IsApiRegisterable="true" IsApiRenewable="true" IsApiTransferable="true" IsEppRequired="true" IsDisableModContact="false" IsDisableWGAllot="false" IsIncludeInExtendedSearchOnly="false" SequenceNumber="10" Type="GTLD" SubType="" IsSupportsIDN="true" Category="A" SupportsRegistrarLock="true" AddGracePeriodDays="5" WhoisVerification="false" ProviderApiDelete="true" TldState="" SearchGroup="" Registry="">Most recognized top level domain<Categories><TldCategory Name="popular" SequenceNumber="10" /></Categories></Tld>
    </Tlds>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>0.004</ExecutionTime>
</ApiResponse>`
