package webroot

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestHTTPProvider(t *testing.T) {
	webroot := "webroot"
	domain := "domain"
	token := "token"
	keyAuth := "keyAuth"
	challengeFilePath := webroot + "/.well-known/acme-challenge/" + token

	os.MkdirAll(webroot+"/.well-known/acme-challenge", 0777)
	defer os.RemoveAll(webroot)

	provider, err := NewHTTPProvider(webroot)
	if err != nil {
		t.Errorf("Webroot provider error: got %v, want nil", err)
	}

	err = provider.Present(domain, token, keyAuth)
	if err != nil {
		t.Errorf("Webroot provider present() error: got %v, want nil", err)
	}

	if _, err := os.Stat(challengeFilePath); os.IsNotExist(err) {
		t.Error("Challenge file was not created in webroot")
	}

	data, err := ioutil.ReadFile(challengeFilePath)
	if err != nil {
		t.Errorf("Webroot provider ReadFile() error: got %v, want nil", err)
	}
	dataStr := string(data)
	if dataStr != keyAuth {
		t.Errorf("Challenge file content: got %q, want %q", dataStr, keyAuth)
	}

	err = provider.CleanUp(domain, token, keyAuth)
	if err != nil {
		t.Errorf("Webroot provider CleanUp() error: got %v, want nil", err)
	}
}
