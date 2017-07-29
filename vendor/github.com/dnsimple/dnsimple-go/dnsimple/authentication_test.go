package dnsimple

import (
	"fmt"
	"reflect"
	"testing"
)

func testCredentials(t *testing.T, credentials Credentials, headers map[string]string) {

	if want, got := headers, credentials.Headers(); !reflect.DeepEqual(want, got) {
		t.Errorf("Header %v, want %v", got, want)
	}
}

func TestDomainTokenCredentialsHttpHeader(t *testing.T) {
	domainToken := "domain-token"
	credentials := NewDomainTokenCredentials(domainToken)
	testCredentials(t, credentials, map[string]string{httpHeaderDomainToken: domainToken})
}

func TestHttpBasicCredentialsHttpHeader(t *testing.T) {
	email, password := "email", "password"
	credentials := NewHTTPBasicCredentials(email, password)
	expectedHeaderValue := "Basic ZW1haWw6cGFzc3dvcmQ="
	testCredentials(t, credentials, map[string]string{httpHeaderAuthorization: expectedHeaderValue})
}

func TestOauthTokenCredentialsHttpHeader(t *testing.T) {
	oauthToken := "oauth-token"
	credentials := NewOauthTokenCredentials(oauthToken)
	expectedHeaderValue := fmt.Sprintf("Bearer %v", oauthToken)
	testCredentials(t, credentials, map[string]string{httpHeaderAuthorization: expectedHeaderValue})
}
