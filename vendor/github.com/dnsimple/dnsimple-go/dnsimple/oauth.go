package dnsimple

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GrantType is a string that identifies a particular grant type in the exchange request.
type GrantType string

const (
	// AuthorizationCodeGrant is the type of access token request
	// for an Authorization Code Grant flow.
	// https://tools.ietf.org/html/rfc6749#section-4.1
	AuthorizationCodeGrant = GrantType("authorization_code")
)

// OauthService handles communication with the authorization related
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/oauth/
type OauthService struct {
	client *Client
}

// AccessToken represents a DNSimple Oauth access token.
type AccessToken struct {
	Token     string `json:"access_token"`
	Type      string `json:"token_type"`
	AccountID int    `json:"account_id"`
}

// ExchangeAuthorizationRequest represents a request to exchange
// an authorization code for an access token.
// RedirectURI is optional, all the other fields are mandatory.
type ExchangeAuthorizationRequest struct {
	Code         string    `json:"code"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	RedirectURI  string    `json:"redirect_uri,omitempty"`
	State        string    `json:"state,omitempty"`
	GrantType    GrantType `json:"grant_type,omitempty"`
}

// ExchangeAuthorizationError represents a failed request to exchange
// an authorization code for an access token.
type ExchangeAuthorizationError struct {
	// HTTP response
	HttpResponse *http.Response

	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Error implements the error interface.
func (r *ExchangeAuthorizationError) Error() string {
	return fmt.Sprintf("%v %v: %v %v",
		r.HttpResponse.Request.Method, r.HttpResponse.Request.URL,
		r.ErrorCode, r.ErrorDescription)
}

// ExchangeAuthorizationForToken exchanges the short-lived authorization code for an access token
// you can use to authenticate your API calls.
func (s *OauthService) ExchangeAuthorizationForToken(authorization *ExchangeAuthorizationRequest) (*AccessToken, error) {
	path := versioned("/oauth/access_token")

	req, err := s.client.NewRequest("POST", path, authorization)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorResponse := &ExchangeAuthorizationError{}
		errorResponse.HttpResponse = resp
		json.NewDecoder(resp.Body).Decode(errorResponse)
		return nil, errorResponse
	}

	accessToken := &AccessToken{}
	err = json.NewDecoder(resp.Body).Decode(accessToken)

	return accessToken, err
}

// AuthorizationOptions represents the option you can use to generate an authorization URL.
type AuthorizationOptions struct {
	RedirectURI string `url:"redirect_uri,omitempty"`
	// A randomly generated string to verify the validity of the request.
	// Currently "state" is required by the DNSimple OAuth implementation, so you must specify it.
	State string `url:"state,omitempty"`
}

// AuthorizeURL generates the URL to authorize an user for an application via the OAuth2 flow.
func (s *OauthService) AuthorizeURL(clientID string, options *AuthorizationOptions) string {
	uri, _ := url.Parse(strings.Replace(s.client.BaseURL, "api.", "", 1))
	uri.Path = "/oauth/authorize"
	query := uri.Query()
	query.Add("client_id", clientID)
	query.Add("response_type", "code")
	uri.RawQuery = query.Encode()

	path, _ := addURLQueryOptions(uri.String(), options)
	return path
}
