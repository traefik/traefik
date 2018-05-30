package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/containous/traefik/types"
	oidc "github.com/coreos/go-oidc"
	"github.com/satori/go.uuid"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"gopkg.in/square/go-jose.v2"
)

const oidcCookieName = "traefik-oidc"

// OIDCProviderRefresher is a wrapper around oidc.Provider that only allows
// discovery information to be used for up to one hour. After one hour, it
// refreshes the discovery information.
type OIDCProviderRefresher struct {
	discoveryURL string
	lock         sync.Mutex
	provider     *oidc.Provider
	expiration   time.Time
}

func NewOIDCProviderRefresher(discoveryURL string) OIDCProviderRefresher {
	return OIDCProviderRefresher{
		discoveryURL: discoveryURL,
	}
}

func (opr *OIDCProviderRefresher) Get(ctx context.Context) (*oidc.Provider, error) {
	now := time.Now()

	opr.lock.Lock()
	defer opr.lock.Unlock()

	// Return cached value if not expired.
	if opr.provider != nil && opr.expiration.After(now) {
		return opr.provider, nil
	}

	// Reload provider configuration.
	provider, err := oidc.NewProvider(ctx, opr.discoveryURL)
	if err != nil {
		return nil, err
	}

	opr.provider = provider
	opr.expiration = now.Add(time.Hour)
	return opr.provider, nil
}

// oidcCookie holds the OAuth2/OIDC session state and is stored on the client.
type oidcCookie struct {
	Type string `json:"type,omitempty"`

	// Fields for Type == "authenticating".
	State      string `json:"state,omitempty"`
	ReturnPath string `json:"return_path,omitempty"`

	// Fields for Type == "done".
	OAuth2Token oauth2.Token `json:"oauth2_token,omitempty"`
	IDToken     string       `json:"id_token,omitempty"`
}

const (
	oidcCookieTypeAuthenticating = "authenticating"
	oidcCookieTypeDone           = "done"
)

// oidcGetCookie decodes the cookie sent by the client.
func oidcGetCookie(r *http.Request, sharedKey []byte, expectedType string) (oidcCookie, error) {
	encrypted, err := r.Cookie(oidcCookieName)
	if err != nil {
		return oidcCookie{}, err
	}
	jwe, err := jose.ParseEncrypted(encrypted.Value)
	if err != nil {
		return oidcCookie{}, err
	}
	decrypted, err := jwe.Decrypt(sharedKey)
	if err != nil {
		return oidcCookie{}, err
	}
	var cookie oidcCookie
	err = json.Unmarshal(decrypted, &cookie)
	if err != nil {
		return oidcCookie{}, err
	}
	if cookie.Type != expectedType {
		return oidcCookie{}, errors.New("Cookie type mismatch")
	}
	return cookie, nil
}

// oidcSetCookie encrypts a cookie and attaches it to the HTTP response.
func oidcSetCookie(w http.ResponseWriter, sharedKey []byte, cookie *oidcCookie) {
	marshalled, err := json.Marshal(cookie)
	if err != nil {
		log.Fatal(err)
	}
	encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{
		Algorithm: jose.DIRECT,
		Key:       sharedKey,
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	encrypted, err := encrypter.Encrypt(marshalled)
	if err != nil {
		log.Fatal(err)
	}
	serialized, err := encrypted.CompactSerialize()
	if err != nil {
		log.Fatal(err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oidcCookieName,
		Value:    serialized,
		Path:     "/",
		HttpOnly: true,
	})
}

// oidcSetDownstreamHeaders verifies the OIDC ID token and sets HTTP headers
// containing some of its properties accordingly. Headers are set using a naming
// scheme similar to keycloak-proxy.
func oidcSetDownstreamHeaders(config *types.OIDC, provider *oidc.Provider, ctx context.Context, r *http.Request, idTokenStr string) error {
	// Validate the ID token.
	oidcConfig := &oidc.Config{
		ClientID: config.ClientID,
	}
	verifier := provider.Verifier(oidcConfig)
	idToken, err := verifier.Verify(ctx, idTokenStr)
	if err != nil {
		return err
	}

	// Set headers based on required ID token fields and custom claims.
	r.Header.Set("X-Auth-Subject", idToken.Subject)
	var claims struct {
		Name          string `json:"name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err == nil {
		if claims.Name != "" {
			r.Header.Set("X-Auth-Name", claims.Name)
		}
		if claims.Email != "" && claims.EmailVerified {
			r.Header.Set("X-Auth-Email", claims.Email)
		}
	}
	return nil
}

// oidcExtractIDToken extracts an OIDC ID token from an OAuth2 token. If no ID
// token is present, a previously generated ID token (fallback) may be returned.
func oidcExtractIDToken(token *oauth2.Token, fallback string) string {
	idTokenField := token.Extra("id_token")
	if idTokenField == nil {
		log.Print("OAuth2 token does not contain an ID token")
		return fallback
	}
	idToken, ok := idTokenField.(string)
	if !ok {
		log.Print("ID token field is not a string")
		return fallback
	}
	return idToken
}

func OIDC(oidcProviderRefresher *OIDCProviderRefresher, sharedKey []byte, config *types.OIDC, w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// TODO: Using r.Context() here makes the code below fail.
	ctx := context.TODO()

	// Obtain public keys and endpoint URLs of the identity provider.
	provider, err := oidcProviderRefresher.Get(ctx)
	if err != nil {
		log.Print(err)
		http.Error(w, "Failed to contact identity provider", http.StatusInternalServerError)
		return
	}

	// OAuth2 configuration.
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	callbackPath := "/.traefik-oidc-callback"
	callbackURL := url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   callbackPath,
	}
	oauth2Config := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  callbackURL.String(),
		Scopes:       config.Scopes,
	}

	if r.URL.Path == callbackPath {
		if r.Method != http.MethodGet {
			http.Error(w, "OIDC callback only processes GET requests", http.StatusMethodNotAllowed)
			return
		}

		cookie, err := oidcGetCookie(r, sharedKey, oidcCookieTypeAuthenticating)
		if err != nil {
			log.Print(err)
			http.Error(w, "CSRF validation failed", http.StatusBadRequest)
			return
		}

		// Prevent CSRF by checking that the request includes a state
		// parameter that matches the value that we set earlier.
		query := r.URL.Query()
		if cookie.State != query.Get("state") {
			log.Print("State argument does not match state cookie")
			http.Error(w, "CSRF validation failed", http.StatusBadRequest)
			return
		}

		// Obtain a new access token based on the authorization code.
		token, err := oauth2Config.Exchange(ctx, query.Get("code"))
		if err != nil {
			log.Print(err)
			http.Error(w, "Failed to obtain access token", http.StatusInternalServerError)
			return
		}

		// Redirect to the originating page.
		oidcSetCookie(w, sharedKey, &oidcCookie{
			Type:        oidcCookieTypeDone,
			OAuth2Token: *token,
			IDToken:     oidcExtractIDToken(token, ""),
		})
		originatingURL := url.URL{
			Scheme: scheme,
			Host:   r.Host,
			Path:   cookie.ReturnPath,
		}
		http.Redirect(w, r, originatingURL.String(), http.StatusSeeOther)
	} else if r.URL.Path == "/.traefik-oidc-logout" {
		// OpenID Connect Front-Channel Logout.
		http.SetCookie(w, &http.Cookie{
			Name:     oidcCookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
		})
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("You have been logged out.\n"))
	} else {
		cookie, err := oidcGetCookie(r, sharedKey, oidcCookieTypeDone)
		if err == nil {
			tokenSource := oauth2Config.TokenSource(ctx, &cookie.OAuth2Token)
			if token, err := tokenSource.Token(); err == nil {
				cookie.IDToken = oidcExtractIDToken(token, cookie.IDToken)
				if err := oidcSetDownstreamHeaders(config, provider, ctx, r, cookie.IDToken); err == nil {
					// Valid OAuth2 token and OIDC ID token found. Forward the request.
					// TODO(edsch): Only generate cookie if it has actually changed?
					oidcSetCookie(w, sharedKey, &cookie)
					next(w, r)
					return
				} else {
					log.Print(err)
				}
			}
		} else {
			log.Print(err)
		}

		if r.Method == http.MethodGet {
			// Redirect GET requests to the login page, so that they
			// can be retried after logging in.
			log.Print("No valid cookie found. Redirecting to login page.")
			state := uuid.NewV4().String()
			oidcSetCookie(w, sharedKey, &oidcCookie{
				Type:       oidcCookieTypeAuthenticating,
				State:      state,
				ReturnPath: r.URL.RequestURI(),
			})
			http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusSeeOther)
		} else {
			// Other requests cannot be retried after when
			// performing redirects, so make them fail explictily.
			http.Error(w, "Unauthorized or session expired", http.StatusUnauthorized)
		}
	}
}
