package auth

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
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

	notFoundSecret    string
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

	// To prevent timing attacks, we need to compute a hash even if the user is not found.
	// We assume it to be safe only when the users hashes are all from the same algorithm,
	// so we can pick the first one as a random hash to compute.
	notFoundSecret := users[slices.Collect(maps.Values(users))[0]]

	ba := &basicAuth{
		next:              next,
		users:             users,
		headerField:       authConfig.HeaderField,
		removeHeader:      authConfig.RemoveHeader,
		name:              name,
		notFoundSecret:    notFoundSecret,
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
	var authenticated bool
	if ok {
		authenticated = b.checkPassword(user, password)
	}

	logData := accesslog.GetLogData(req)
	if logData != nil {
		logData.Core[accesslog.ClientUsername] = user
	}

	if !authenticated {
		logger.Debug().Msg("Authentication failed")
		observability.SetStatusErrorf(req.Context(), "Authentication failed")

		b.auth.RequireAuth(rw, req)
		return
	}

	logger.Debug().Msg("Authentication succeeded")
	req.URL.User = url.User(user)

	if b.headerField != "" {
		// TODO Deprecated we should add the header with canonical key.
		req.Header.Del(b.headerField)
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

	key := password + secret
	match, _, _ := b.singleflightGroup.Do(key, func() (any, error) {
		if secret == "" {
			_ = b.checkSecret(password, b.notFoundSecret)
			return false, nil
		}

		return b.checkSecret(password, secret), nil
	})

	return match.(bool)
}

func (b *basicAuth) secretBasic(user, _ string) string {
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
