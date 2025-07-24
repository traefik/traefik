package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	goauth "github.com/abbot/go-http-auth"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"golang.org/x/sync/singleflight"
)

const (
	typeNameBasic = "BasicAuth"
)

type basicAuth struct {
	next         http.Handler
	auth         *goauth.BasicAuth
	users        map[string]string
	headerField  string
	removeHeader bool
	name         string

	checkSecret       func(password, secret string) bool
	singleflightGroup *singleflight.Group
}

// NewBasic creates a basicAuth middleware.
func NewBasic(ctx context.Context, next http.Handler, authConfig dynamic.BasicAuth, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeNameBasic).Debug().Msg("Creating middleware")

	users, err := getUsers(authConfig.UsersFile, authConfig.Users, basicUserParser)
	if err != nil {
		return nil, err
	}

	ba := &basicAuth{
		next:              next,
		users:             users,
		headerField:       authConfig.HeaderField,
		removeHeader:      authConfig.RemoveHeader,
		name:              name,
		checkSecret:       goauth.CheckSecret,
		singleflightGroup: new(singleflight.Group),
	}

	realm := defaultRealm
	if len(authConfig.Realm) > 0 {
		realm = authConfig.Realm
	}

	ba.auth = &goauth.BasicAuth{Realm: realm, Secrets: ba.secretBasic}

	return ba, nil
}

func (b *basicAuth) GetTracingInformation() (string, string) {
	return b.name, typeNameBasic
}

func (b *basicAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), b.name, typeNameBasic)

	user, password, ok := req.BasicAuth()
	if ok {
		ok = b.checkPassword(user, password)
	}

	logData := accesslog.GetLogData(req)
	if logData != nil {
		logData.Core[accesslog.ClientUsername] = user
	}

	if !ok {
		logger.Debug().Msg("Authentication failed")
		observability.SetStatusErrorf(req.Context(), "Authentication failed")

		b.auth.RequireAuth(rw, req)
		return
	}

	logger.Debug().Msg("Authentication succeeded")
	req.URL.User = url.User(user)

	if b.headerField != "" {
		req.Header[b.headerField] = []string{user}
	}

	if b.removeHeader {
		logger.Debug().Msg("Removing authorization header")
		req.Header.Del(authorizationHeader)
	}
	b.next.ServeHTTP(rw, req)
}

func (b *basicAuth) checkPassword(user, password string) bool {
	secret := b.auth.Secrets(user, b.auth.Realm)
	if secret == "" {
		return false
	}

	key := password + secret
	match, _, _ := b.singleflightGroup.Do(key, func() (any, error) {
		return b.checkSecret(password, secret), nil
	})

	return match.(bool)
}

func (b *basicAuth) secretBasic(user, realm string) string {
	if secret, ok := b.users[user]; ok {
		return secret
	}

	return ""
}

func basicUserParser(user string) (string, string, error) {
	split := strings.Split(user, ":")
	if len(split) != 2 {
		return "", "", fmt.Errorf("error parsing BasicUser: %v", user)
	}
	return split[0], split[1], nil
}
