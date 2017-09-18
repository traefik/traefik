package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	goauth "github.com/abbot/go-http-auth"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/urfave/negroni"
)

// Authenticator is a middleware that provides HTTP basic and digest authentication
type Authenticator struct {
	handler negroni.Handler
	users   map[string]string
}

// NewAuthenticator builds a new Autenticator given a config
func NewAuthenticator(authConfig *types.Auth) (*Authenticator, error) {
	if authConfig == nil {
		return nil, fmt.Errorf("Error creating Authenticator: auth is nil")
	}
	var err error
	authenticator := Authenticator{}
	if authConfig.Basic != nil {
		authenticator.users, err = parserBasicUsers(authConfig.Basic)
		if err != nil {
			return nil, err
		}
		basicAuth := goauth.NewBasicAuthenticator("traefik", authenticator.secretBasic)
		authenticator.handler = negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			if username := basicAuth.CheckAuth(r); username == "" {
				log.Debug("Basic auth failed...")
				basicAuth.RequireAuth(w, r)
			} else {
				log.Debug("Basic auth success...")
				if authConfig.HeaderField != "" {
					r.Header[authConfig.HeaderField] = []string{username}
				}
				next.ServeHTTP(w, r)
			}
		})
	} else if authConfig.Digest != nil {
		authenticator.users, err = parserDigestUsers(authConfig.Digest)
		if err != nil {
			return nil, err
		}
		digestAuth := goauth.NewDigestAuthenticator("traefik", authenticator.secretDigest)
		authenticator.handler = negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			if username, _ := digestAuth.CheckAuth(r); username == "" {
				log.Debug("Digest auth failed...")
				digestAuth.RequireAuth(w, r)
			} else {
				log.Debug("Digest auth success...")
				if authConfig.HeaderField != "" {
					r.Header[authConfig.HeaderField] = []string{username}
				}
				next.ServeHTTP(w, r)
			}
		})
	} else if authConfig.Forward != nil {
		authenticator.handler = negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			Forward(authConfig.Forward, w, r, next)
		})
	}
	return &authenticator, nil
}

func getLinesFromFile(filename string) ([]string, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// Trim lines and filter out blanks
	rawLines := strings.Split(string(dat), "\n")
	var filteredLines []string
	for _, rawLine := range rawLines {
		line := strings.TrimSpace(rawLine)
		if line != "" {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines, nil
}

func (a *Authenticator) secretBasic(user, realm string) string {
	if secret, ok := a.users[user]; ok {
		return secret
	}
	log.Debugf("User not found: %s", user)
	return ""
}

func (a *Authenticator) secretDigest(user, realm string) string {
	if secret, ok := a.users[user+":"+realm]; ok {
		return secret
	}
	log.Debugf("User not found: %s:%s", user, realm)
	return ""
}

func (a *Authenticator) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	a.handler.ServeHTTP(rw, r, next)
}
