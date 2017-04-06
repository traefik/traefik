package oidc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	jose "gopkg.in/square/go-jose.v2"
)

// IDTokenVerifier provides verification for ID Tokens.
type IDTokenVerifier struct {
	keySet *remoteKeySet
	config *verificationConfig
}

// verificationConfig is the unexported configuration for an IDTokenVerifier.
//
// Users interact with this struct using a VerificationOption.
type verificationConfig struct {
	issuer string
	// If provided, this value must be in the ID Token audiences.
	audience string
	// If not nil, check the expiry of the id token.
	checkExpiry func() time.Time
	// If specified, only these sets of algorithms may be used to sign the JWT.
	requiredAlgs []string
	// If not nil, don't verify nonce.
	nonceSource NonceSource
}

// VerificationOption provides additional checks on ID Tokens.
type VerificationOption interface {
	// Unexport this method so other packages can't implement this interface.
	updateConfig(c *verificationConfig)
}

// Verifier returns an IDTokenVerifier that uses the provider's key set to verify JWTs.
//
// The returned IDTokenVerifier is tied to the Provider's context and its behavior is
// undefined once the Provider's context is canceled.
func (p *Provider) Verifier(options ...VerificationOption) *IDTokenVerifier {
	config := &verificationConfig{issuer: p.issuer}
	for _, option := range options {
		option.updateConfig(config)
	}

	return newVerifier(p.remoteKeySet, config)
}

func newVerifier(keySet *remoteKeySet, config *verificationConfig) *IDTokenVerifier {
	// As discussed in the godocs for VerifrySigningAlg, because almost all providers
	// only support RS256, default to only allowing it.
	if len(config.requiredAlgs) == 0 {
		config.requiredAlgs = []string{RS256}
	}

	return &IDTokenVerifier{
		keySet: keySet,
		config: config,
	}
}

func parseJWT(p string) ([]byte, error) {
	parts := strings.Split(p, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("oidc: malformed jwt, expected 3 parts got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("oidc: malformed jwt payload: %v", err)
	}
	return payload, nil
}

func contains(sli []string, ele string) bool {
	for _, s := range sli {
		if s == ele {
			return true
		}
	}
	return false
}

// Verify parses a raw ID Token, verifies it's been signed by the provider, preforms
// any additional checks passed as VerifictionOptions, and returns the payload.
//
// See: https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation
//
//    oauth2Token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
//    if err != nil {
//        // handle error
//    }
//
//    // Extract the ID Token from oauth2 token.
//    rawIDToken, ok := oauth2Token.Extra("id_token").(string)
//    if !ok {
//        // handle error
//    }
//
//    token, err := verifier.Verify(ctx, rawIDToken)
//
func (v *IDTokenVerifier) Verify(ctx context.Context, rawIDToken string) (*IDToken, error) {
	jws, err := jose.ParseSigned(rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc: mallformed jwt: %v", err)
	}

	// Throw out tokens with invalid claims before trying to verify the token. This lets
	// us do cheap checks before possibly re-syncing keys.
	payload, err := parseJWT(rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc: malformed jwt: %v", err)
	}
	var token idToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return nil, fmt.Errorf("oidc: failed to unmarshal claims: %v", err)
	}

	t := &IDToken{
		Issuer:   token.Issuer,
		Subject:  token.Subject,
		Audience: []string(token.Audience),
		Expiry:   time.Time(token.Expiry),
		IssuedAt: time.Time(token.IssuedAt),
		Nonce:    token.Nonce,
		claims:   payload,
	}

	// Check issuer.
	if t.Issuer != v.config.issuer {
		return nil, fmt.Errorf("oidc: id token issued by a different provider, expected %q got %q", v.config.issuer, t.Issuer)
	}

	// If a client ID has been provided, make sure it's part of the audience.
	if v.config.audience != "" {
		if !contains(t.Audience, v.config.audience) {
			return nil, fmt.Errorf("oidc: expected audience %q got %q", v.config.audience, t.Audience)
		}
	}

	// If a set of required algorithms has been provided, ensure that the signatures use those.
	var keyIDs, gotAlgs []string
	for _, sig := range jws.Signatures {
		if len(v.config.requiredAlgs) == 0 || contains(v.config.requiredAlgs, sig.Header.Algorithm) {
			keyIDs = append(keyIDs, sig.Header.KeyID)
		} else {
			gotAlgs = append(gotAlgs, sig.Header.Algorithm)
		}
	}
	if len(keyIDs) == 0 {
		return nil, fmt.Errorf("oidc: no signatures use a require algorithm, expected %q got %q", v.config.requiredAlgs, gotAlgs)
	}

	// Get keys from the remote key set. This may trigger a re-sync.
	keys, err := v.keySet.keysWithID(ctx, keyIDs)
	if err != nil {
		return nil, fmt.Errorf("oidc: get keys for id token: %v", err)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("oidc: no keys match signature ID(s) %q", keyIDs)
	}

	// Try to use a key to validate the signature.
	var gotPayload []byte
	for _, key := range keys {
		if p, err := jws.Verify(&key); err == nil {
			gotPayload = p
		}
	}
	if len(gotPayload) == 0 {
		return nil, fmt.Errorf("oidc: failed to verify id token")
	}

	// Ensure that the payload returned by the square actually matches the payload parsed earlier.
	if !bytes.Equal(gotPayload, payload) {
		return nil, errors.New("oidc: internal error, payload parsed did not match previous payload")
	}

	// Check the nonce after we've verified the token. We don't want to allow unverified
	// payloads to trigger a nonce lookup.
	if v.config.nonceSource != nil {
		if err := v.config.nonceSource.ClaimNonce(t.Nonce); err != nil {
			return nil, err
		}
	}

	return t, nil
}

// VerifyAudience ensures that an ID Token was issued for the specific client.
//
// Note that a verified token may be valid for other clients, as OpenID Connect allows a token to have
// multiple audiences.
func VerifyAudience(clientID string) VerificationOption {
	return clientVerifier{clientID}
}

type clientVerifier struct {
	clientID string
}

func (v clientVerifier) updateConfig(c *verificationConfig) {
	c.audience = v.clientID
}

// VerifyExpiry ensures that an ID Token has not expired.
func VerifyExpiry() VerificationOption {
	return expiryVerifier{}
}

type expiryVerifier struct{}

func (v expiryVerifier) updateConfig(c *verificationConfig) {
	c.checkExpiry = time.Now
}

// VerifySigningAlg enforces that an ID Token is signed by a specific signing algorithm.
//
// Because so many providers only support RS256, if this verifiction option isn't used,
// the IDTokenVerifier defaults to only allowing RS256.
func VerifySigningAlg(allowedAlgs ...string) VerificationOption {
	return algVerifier{allowedAlgs}
}

type algVerifier struct {
	algs []string
}

func (v algVerifier) updateConfig(c *verificationConfig) {
	c.requiredAlgs = v.algs
}

// Nonce returns an auth code option which requires the ID Token created by the
// OpenID Connect provider to contain the specified nonce.
func Nonce(nonce string) oauth2.AuthCodeOption {
	return oauth2.SetAuthURLParam("nonce", nonce)
}

// NonceSource represents a source which can verify a nonce is valid and has not
// been claimed before.
type NonceSource interface {
	ClaimNonce(nonce string) error
}

// VerifyNonce ensures that the ID Token contains a nonce which can be claimed by the nonce source.
func VerifyNonce(source NonceSource) VerificationOption {
	return nonceVerifier{source}
}

type nonceVerifier struct {
	nonceSource NonceSource
}

func (n nonceVerifier) updateConfig(c *verificationConfig) {
	c.nonceSource = n.nonceSource
}
