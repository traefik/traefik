package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	keyBits := 32 // small value keeps test fast
	keyType := RSA2048
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		t.Fatal("Could not generate test key:", err)
	}
	user := mockUser{
		email:      "test@test.com",
		regres:     new(RegistrationResource),
		privatekey: key,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(directory{NewAuthzURL: "http://test", NewCertURL: "http://test", NewRegURL: "http://test", RevokeCertURL: "http://test"})
		w.Write(data)
	}))

	client, err := NewClient(ts.URL, user, keyType)
	if err != nil {
		t.Fatalf("Could not create client: %v", err)
	}

	if client.jws == nil {
		t.Fatalf("Expected client.jws to not be nil")
	}
	if expected, actual := key, client.jws.privKey; actual != expected {
		t.Errorf("Expected jws.privKey to be %p but was %p", expected, actual)
	}

	if client.keyType != keyType {
		t.Errorf("Expected keyType to be %s but was %s", keyType, client.keyType)
	}

	if expected, actual := 2, len(client.solvers); actual != expected {
		t.Fatalf("Expected %d solver(s), got %d", expected, actual)
	}
}

func TestClientOptPort(t *testing.T) {
	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		t.Fatal("Could not generate test key:", err)
	}
	user := mockUser{
		email:      "test@test.com",
		regres:     new(RegistrationResource),
		privatekey: key,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(directory{NewAuthzURL: "http://test", NewCertURL: "http://test", NewRegURL: "http://test", RevokeCertURL: "http://test"})
		w.Write(data)
	}))

	optPort := "1234"
	optHost := ""
	client, err := NewClient(ts.URL, user, RSA2048)
	if err != nil {
		t.Fatalf("Could not create client: %v", err)
	}
	client.SetHTTPAddress(net.JoinHostPort(optHost, optPort))
	client.SetTLSAddress(net.JoinHostPort(optHost, optPort))

	httpSolver, ok := client.solvers[HTTP01].(*httpChallenge)
	if !ok {
		t.Fatal("Expected http-01 solver to be httpChallenge type")
	}
	if httpSolver.jws != client.jws {
		t.Error("Expected http-01 to have same jws as client")
	}
	if got := httpSolver.provider.(*HTTPProviderServer).port; got != optPort {
		t.Errorf("Expected http-01 to have port %s but was %s", optPort, got)
	}
	if got := httpSolver.provider.(*HTTPProviderServer).iface; got != optHost {
		t.Errorf("Expected http-01 to have iface %s but was %s", optHost, got)
	}

	httpsSolver, ok := client.solvers[TLSSNI01].(*tlsSNIChallenge)
	if !ok {
		t.Fatal("Expected tls-sni-01 solver to be httpChallenge type")
	}
	if httpsSolver.jws != client.jws {
		t.Error("Expected tls-sni-01 to have same jws as client")
	}
	if got := httpsSolver.provider.(*TLSProviderServer).port; got != optPort {
		t.Errorf("Expected tls-sni-01 to have port %s but was %s", optPort, got)
	}
	if got := httpsSolver.provider.(*TLSProviderServer).iface; got != optHost {
		t.Errorf("Expected tls-sni-01 to have port %s but was %s", optHost, got)
	}

	// test setting different host
	optHost = "127.0.0.1"
	client.SetHTTPAddress(net.JoinHostPort(optHost, optPort))
	client.SetTLSAddress(net.JoinHostPort(optHost, optPort))

	if got := httpSolver.provider.(*HTTPProviderServer).iface; got != optHost {
		t.Errorf("Expected http-01 to have iface %s but was %s", optHost, got)
	}
	if got := httpsSolver.provider.(*TLSProviderServer).port; got != optPort {
		t.Errorf("Expected tls-sni-01 to have port %s but was %s", optPort, got)
	}
}

func TestNotHoldingLockWhileMakingHTTPRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(250 * time.Millisecond)
		w.Header().Add("Replay-Nonce", "12345")
		w.Header().Add("Retry-After", "0")
		writeJSONResponse(w, &challenge{Type: "http-01", Status: "Valid", URI: "http://example.com/", Token: "token"})
	}))
	defer ts.Close()

	privKey, _ := rsa.GenerateKey(rand.Reader, 512)
	j := &jws{privKey: privKey, directoryURL: ts.URL}
	ch := make(chan bool)
	resultCh := make(chan bool)
	go func() {
		j.Nonce()
		ch <- true
	}()
	go func() {
		j.Nonce()
		ch <- true
	}()
	go func() {
		<-ch
		<-ch
		resultCh <- true
	}()
	select {
	case <-resultCh:
	case <-time.After(400 * time.Millisecond):
		t.Fatal("JWS is probably holding a lock while making HTTP request")
	}
}

func TestValidate(t *testing.T) {
	var statuses []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Minimal stub ACME server for validation.
		w.Header().Add("Replay-Nonce", "12345")
		w.Header().Add("Retry-After", "0")
		switch r.Method {
		case "HEAD":
		case "POST":
			st := statuses[0]
			statuses = statuses[1:]
			writeJSONResponse(w, &challenge{Type: "http-01", Status: st, URI: "http://example.com/", Token: "token"})

		case "GET":
			st := statuses[0]
			statuses = statuses[1:]
			writeJSONResponse(w, &challenge{Type: "http-01", Status: st, URI: "http://example.com/", Token: "token"})

		default:
			http.Error(w, r.Method, http.StatusMethodNotAllowed)
		}
	}))
	defer ts.Close()

	privKey, _ := rsa.GenerateKey(rand.Reader, 512)
	j := &jws{privKey: privKey, directoryURL: ts.URL}

	tsts := []struct {
		name     string
		statuses []string
		want     string
	}{
		{"POST-unexpected", []string{"weird"}, "unexpected"},
		{"POST-valid", []string{"valid"}, ""},
		{"POST-invalid", []string{"invalid"}, "Error Detail"},
		{"GET-unexpected", []string{"pending", "weird"}, "unexpected"},
		{"GET-valid", []string{"pending", "valid"}, ""},
		{"GET-invalid", []string{"pending", "invalid"}, "Error Detail"},
	}

	for _, tst := range tsts {
		statuses = tst.statuses
		if err := validate(j, "example.com", ts.URL, challenge{Type: "http-01", Token: "token"}); err == nil && tst.want != "" {
			t.Errorf("[%s] validate: got error %v, want something with %q", tst.name, err, tst.want)
		} else if err != nil && !strings.Contains(err.Error(), tst.want) {
			t.Errorf("[%s] validate: got error %v, want something with %q", tst.name, err, tst.want)
		}
	}
}

func TestGetChallenges(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET", "HEAD":
			w.Header().Add("Replay-Nonce", "12345")
			w.Header().Add("Retry-After", "0")
			writeJSONResponse(w, directory{NewAuthzURL: ts.URL, NewCertURL: ts.URL, NewRegURL: ts.URL, RevokeCertURL: ts.URL})
		case "POST":
			writeJSONResponse(w, authorization{})
		}
	}))
	defer ts.Close()

	keyBits := 512 // small value keeps test fast
	keyType := RSA2048
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		t.Fatal("Could not generate test key:", err)
	}
	user := mockUser{
		email:      "test@test.com",
		regres:     &RegistrationResource{NewAuthzURL: ts.URL},
		privatekey: key,
	}

	client, err := NewClient(ts.URL, user, keyType)
	if err != nil {
		t.Fatalf("Could not create client: %v", err)
	}

	_, failures := client.getChallenges([]string{"example.com"})
	if failures["example.com"] == nil {
		t.Fatal("Expecting \"Server did not provide next link to proceed\" error, got nil")
	}
}

// writeJSONResponse marshals the body as JSON and writes it to the response.
func writeJSONResponse(w http.ResponseWriter, body interface{}) {
	bs, err := json.Marshal(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(bs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// stubValidate is like validate, except it does nothing.
func stubValidate(j *jws, domain, uri string, chlng challenge) error {
	return nil
}

type mockUser struct {
	email      string
	regres     *RegistrationResource
	privatekey *rsa.PrivateKey
}

func (u mockUser) GetEmail() string                       { return u.email }
func (u mockUser) GetRegistration() *RegistrationResource { return u.regres }
func (u mockUser) GetPrivateKey() crypto.PrivateKey       { return u.privatekey }
