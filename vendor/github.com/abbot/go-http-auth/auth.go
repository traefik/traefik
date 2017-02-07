// Package auth is an implementation of HTTP Basic and HTTP Digest authentication.
package auth

import (
	"net/http"

	"golang.org/x/net/context"
)

/*
 Request handlers must take AuthenticatedRequest instead of http.Request
*/
type AuthenticatedRequest struct {
	http.Request
	/*
	 Authenticated user name. Current API implies that Username is
	 never empty, which means that authentication is always done
	 before calling the request handler.
	*/
	Username string
}

/*
 AuthenticatedHandlerFunc is like http.HandlerFunc, but takes
 AuthenticatedRequest instead of http.Request
*/
type AuthenticatedHandlerFunc func(http.ResponseWriter, *AuthenticatedRequest)

/*
 Authenticator wraps an AuthenticatedHandlerFunc with
 authentication-checking code.

 Typical Authenticator usage is something like:

   authenticator := SomeAuthenticator(...)
   http.HandleFunc("/", authenticator(my_handler))

 Authenticator wrapper checks the user authentication and calls the
 wrapped function only after authentication has succeeded. Otherwise,
 it returns a handler which initiates the authentication procedure.
*/
type Authenticator func(AuthenticatedHandlerFunc) http.HandlerFunc

// Info contains authentication information for the request.
type Info struct {
	// Authenticated is set to true when request was authenticated
	// successfully, i.e. username and password passed in request did
	// pass the check.
	Authenticated bool

	// Username contains a user name passed in the request when
	// Authenticated is true. It's value is undefined if Authenticated
	// is false.
	Username string

	// ResponseHeaders contains extra headers that must be set by server
	// when sending back HTTP response.
	ResponseHeaders http.Header
}

// UpdateHeaders updates headers with this Info's ResponseHeaders. It is
// safe to call this function on nil Info.
func (i *Info) UpdateHeaders(headers http.Header) {
	if i == nil {
		return
	}
	for k, values := range i.ResponseHeaders {
		for _, v := range values {
			headers.Add(k, v)
		}
	}
}

type key int // used for context keys

var infoKey key = 0

type AuthenticatorInterface interface {
	// NewContext returns a new context carrying authentication
	// information extracted from the request.
	NewContext(ctx context.Context, r *http.Request) context.Context

	// Wrap returns an http.HandlerFunc which wraps
	// AuthenticatedHandlerFunc with this authenticator's
	// authentication checks.
	Wrap(AuthenticatedHandlerFunc) http.HandlerFunc
}

// FromContext returns authentication information from the context or
// nil if no such information present.
func FromContext(ctx context.Context) *Info {
	info, ok := ctx.Value(infoKey).(*Info)
	if !ok {
		return nil
	}
	return info
}

// AuthUsernameHeader is the header set by JustCheck functions. It
// contains an authenticated username (if authentication was
// successful).
const AuthUsernameHeader = "X-Authenticated-Username"

func JustCheck(auth AuthenticatorInterface, wrapped http.HandlerFunc) http.HandlerFunc {
	return auth.Wrap(func(w http.ResponseWriter, ar *AuthenticatedRequest) {
		ar.Header.Set(AuthUsernameHeader, ar.Username)
		wrapped(w, &ar.Request)
	})
}
