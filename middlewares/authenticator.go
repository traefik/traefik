package middlewares

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/abbot/go-http-auth"
	"github.com/codegangsta/negroni"
	"github.com/containous/traefik/types"
	"net/http"
	"strings"
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
		authenticator.users, err = parserBasicUsers(authConfig.Basic.Users)
		if err != nil {
			return nil, err
		}
		basicAuth := auth.NewBasicAuthenticator("traefik", authenticator.secretBasic)
		authenticator.handler = negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			if username := basicAuth.CheckAuth(r); username == "" {
				log.Debugf("Auth failed...")
				basicAuth.RequireAuth(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	} else if authConfig.Digest != nil {
		authenticator.users, err = parserDigestUsers(authConfig.Digest.Users)
		if err != nil {
			return nil, err
		}
		digestAuth := auth.NewDigestAuthenticator("traefik", authenticator.secretDigest)
		authenticator.handler = negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			if username, _ := digestAuth.CheckAuth(r); username == "" {
				digestAuth.RequireAuth(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
	return &authenticator, nil
}

func parserBasicUsers(users types.Users) (map[string]string, error) {
	userMap := make(map[string]string)
	for _, user := range users {
		split := strings.Split(user, ":")
		if len(split) != 2 {
			return nil, fmt.Errorf("Error parsing Authenticator user: %v", user)
		}
		userMap[split[0]] = split[1]
	}
	return userMap, nil
}

func parserDigestUsers(users types.Users) (map[string]string, error) {
	userMap := make(map[string]string)
	for _, user := range users {
		split := strings.Split(user, ":")
		if len(split) != 3 {
			return nil, fmt.Errorf("Error parsing Authenticator user: %v", user)
		}
		userMap[split[0]+":"+split[1]] = split[2]
	}
	return userMap, nil
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
