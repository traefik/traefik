// +build ignore

/*
 Example application using NewContext/FromContext

 Build with:

 go build context.go
*/

package main

import (
	"fmt"
	"net/http"

	auth ".."
	"golang.org/x/net/context"
)

func Secret(user, realm string) string {
	if user == "john" {
		// password is "hello"
		return "b98e16cbc3d01734b264adba7baa3bf9"
	}
	return ""
}

type ContextHandler interface {
	ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

type ContextHandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)

func (f ContextHandlerFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	authInfo := auth.FromContext(ctx)
	authInfo.UpdateHeaders(w.Header())
	if authInfo == nil || !authInfo.Authenticated {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "<html><body><h1>Hello, %s!</h1></body></html>", authInfo.Username)
}

func authenticatedHandler(a auth.AuthenticatorInterface, h ContextHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := a.NewContext(context.Background(), r)
		h.ServeHTTP(ctx, w, r)
	})
}

func main() {
	authenticator := auth.NewDigestAuthenticator("example.com", Secret)
	http.Handle("/", authenticatedHandler(authenticator, ContextHandlerFunc(handle)))
	http.ListenAndServe(":8080", nil)
}
