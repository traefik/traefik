package otc

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var fakeOTCUserName = "test"
var fakeOTCPassword = "test"
var fakeOTCDomainName = "test"
var fakeOTCProjectName = "test"
var fakeOTCToken = "62244bc21da68d03ebac94e6636ff01f"

type DNSMock struct {
	t      *testing.T
	Server *httptest.Server
	Mux    *http.ServeMux
}

func NewDNSMock(t *testing.T) *DNSMock {
	return &DNSMock{
		t: t,
	}
}

// Setup creates the mock server
func (m *DNSMock) Setup() {
	m.Mux = http.NewServeMux()
	m.Server = httptest.NewServer(m.Mux)
}

// ShutdownServer creates the mock server
func (m *DNSMock) ShutdownServer() {
	m.Server.Close()
}

func (m *DNSMock) HandleAuthSuccessfully() {
	m.Mux.HandleFunc("/v3/auth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Subject-Token", fakeOTCToken)

		fmt.Fprintf(w, `{
		  "token": {
		    "catalog": [
		      {
			"type": "dns",
			"id": "56cd81db1f8445d98652479afe07c5ba",
			"name": "",
			"endpoints": [
			  {
			    "url": "%s",
			    "region": "eu-de",
			    "region_id": "eu-de",
			    "interface": "public",
			    "id": "0047a06690484d86afe04877074efddf"
			  }
			]
		      }
		    ]
		  }}`, m.Server.URL)
	})
}

func (m *DNSMock) HandleListZonesSuccessfully() {
	m.Mux.HandleFunc("/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
		  "zones":[{
		    "id":"123123"
		  }]}
		`)

		assert.Equal(m.t, r.Method, "GET")
		assert.Equal(m.t, r.URL.Path, "/v2/zones")
		assert.Equal(m.t, r.URL.RawQuery, "name=example.com.")
		assert.Equal(m.t, r.Header.Get("Content-Type"), "application/json")
	})
}

func (m *DNSMock) HandleListZonesEmpty() {
	m.Mux.HandleFunc("/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
		  "zones":[
		  ]}
		`)

		assert.Equal(m.t, r.Method, "GET")
		assert.Equal(m.t, r.URL.Path, "/v2/zones")
		assert.Equal(m.t, r.URL.RawQuery, "name=example.com.")
		assert.Equal(m.t, r.Header.Get("Content-Type"), "application/json")
	})
}

func (m *DNSMock) HandleDeleteRecordsetsSuccessfully() {
	m.Mux.HandleFunc("/v2/zones/123123/recordsets/321321", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
		  "zones":[{
		    "id":"123123"
		  }]}
		`)

		assert.Equal(m.t, r.Method, "DELETE")
		assert.Equal(m.t, r.URL.Path, "/v2/zones/123123/recordsets/321321")
		assert.Equal(m.t, r.Header.Get("Content-Type"), "application/json")
	})
}

func (m *DNSMock) HandleListRecordsetsEmpty() {
	m.Mux.HandleFunc("/v2/zones/123123/recordsets", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
		  "recordsets":[
		  ]}
		`)

		assert.Equal(m.t, r.URL.Path, "/v2/zones/123123/recordsets")
		assert.Equal(m.t, r.URL.RawQuery, "type=TXT&name=_acme-challenge.example.com.")
	})
}
func (m *DNSMock) HandleListRecordsetsSuccessfully() {
	m.Mux.HandleFunc("/v2/zones/123123/recordsets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprintf(w, `{
			  "recordsets":[{
			    "id":"321321"
			  }]}
			`)

			assert.Equal(m.t, r.URL.Path, "/v2/zones/123123/recordsets")
			assert.Equal(m.t, r.URL.RawQuery, "type=TXT&name=_acme-challenge.example.com.")

		} else if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)

			assert.Nil(m.t, err)
			exceptedString := "{\"name\":\"_acme-challenge.example.com.\",\"description\":\"Added TXT record for ACME dns-01 challenge using lego client\",\"type\":\"TXT\",\"ttl\":300,\"records\":[\"\\\"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI\\\"\"]}"
			assert.Equal(m.t, string(body), exceptedString)

			fmt.Fprintf(w, `{
			  "recordsets":[{
                            "id":"321321"
			  }]}
			`)

		} else {
			m.t.Errorf("Expected method to be 'GET' or 'POST' but got '%s'", r.Method)
		}

		assert.Equal(m.t, r.Header.Get("Content-Type"), "application/json")
	})
}
