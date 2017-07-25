package dnsimple

import (
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestOauthService_ExchangeAuthorizationForToken(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	code := "1234567890"
	clientID := "a1b2c3"
	clientSecret := "thisisasecret"

	mux.HandleFunc("/v2/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/oauthAccessToken/success.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		want := map[string]interface{}{"code": code, "client_id": clientID, "client_secret": clientSecret, "grant_type": "authorization_code"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	token, err := client.Oauth.ExchangeAuthorizationForToken(&ExchangeAuthorizationRequest{Code: code, ClientID: clientID, ClientSecret: clientSecret, GrantType: AuthorizationCodeGrant})
	if err != nil {
		t.Fatalf("Oauth.ExchangeAuthorizationForToken() returned error: %v", err)
	}

	want := &AccessToken{Token: "zKQ7OLqF5N1gylcJweA9WodA000BUNJD", Type: "Bearer", AccountID: 1}
	if !reflect.DeepEqual(token, want) {
		t.Errorf("Oauth.ExchangeAuthorizationForToken() returned %+v, want %+v", token, want)
	}
}

func TestOauthService_ExchangeAuthorizationForToken_Error(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/oauthAccessToken/error-invalid-request.http")

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Oauth.ExchangeAuthorizationForToken(&ExchangeAuthorizationRequest{Code: "1234567890", ClientID: "a1b2c3", ClientSecret: "thisisasecret", GrantType: "authorization_code"})
	if err == nil {
		t.Fatalf("Oauth.ExchangeAuthorizationForToken() expected to return an error")
	}

	switch v := err.(type) {
	case *ExchangeAuthorizationError:
		if want := `Invalid "state": value doesn't match the "state" in the authorization request`; v.ErrorDescription != want {
			t.Errorf("Oauth.ExchangeAuthorizationForToken() error is %v, want %v", v, want)
		}
	default:
		t.Fatalf("Oauth.ExchangeAuthorizationForToken() error type unknown: %v", v.Error())
	}
}

func TestOauthService_AuthorizeURL(t *testing.T) {
	clientID := "a1b2c3"
	client.BaseURL = "https://api.host.test"

	if want, got := "https://host.test/oauth/authorize?client_id=a1b2c3&response_type=code", client.Oauth.AuthorizeURL(clientID, nil); want != got {
		t.Errorf("AuthorizeURL = %v, want %v", got, want)
	}

	if want, got := "https://host.test/oauth/authorize?client_id=a1b2c3&response_type=code&state=randomstate", client.Oauth.AuthorizeURL(clientID, &AuthorizationOptions{State: "randomstate"}); want != got {
		t.Errorf("AuthorizeURL = %v, want %v", got, want)
	}
}
