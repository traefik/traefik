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
)

const (
	typeNameDigest = "digestAuth"
)

type digestAuth struct {
	next         http.Handler
	auth         *goauth.DigestAuth
	users        map[string]string
	headerField  string
	removeHeader bool
	name         string
}

// NewDigest creates a digest auth middleware.
func NewDigest(ctx context.Context, next http.Handler, authConfig dynamic.DigestAuth, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeNameDigest).Debug().Msg("Creating middleware")

	users, err := getUsers(authConfig.UsersFile, authConfig.Users, digestUserParser)
	if err != nil {
		return nil, err
	}

	da := &digestAuth{
		next:         next,
		users:        users,
		headerField:  authConfig.HeaderField,
		removeHeader: authConfig.RemoveHeader,
		name:         name,
	}

	realm := defaultRealm
	if len(authConfig.Realm) > 0 {
		realm = authConfig.Realm
	}
	da.auth = goauth.NewDigestAuthenticator(realm, da.secretDigest)

	return da, nil
}

func (d *digestAuth) GetTracingInformation() (string, string) {
	return d.name, typeNameDigest
}

func (d *digestAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), d.name, typeNameDigest)

	username, authinfo := d.auth.CheckAuth(req)
	if username == "" {
		headerField := d.headerField
		if d.headerField == "" {
			headerField = "Authorization"
		}

		auth := goauth.DigestAuthParams(req.Header.Get(headerField))
		if auth["username"] != "" {
			logData := accesslog.GetLogData(req)
			if logData != nil {
				logData.Core[accesslog.ClientUsername] = auth["username"]
			}
		}

		if authinfo != nil && *authinfo == "stale" {
			logger.Debug().Msg("Digest authentication failed, possibly because out of order requests")
			observability.SetStatusErrorf(req.Context(), "Digest authentication failed, possibly because out of order requests")
			d.auth.RequireAuthStale(rw, req)
			return
		}

		logger.Debug().Msg("Digest authentication failed")
		observability.SetStatusErrorf(req.Context(), "Digest authentication failed")
		d.auth.RequireAuth(rw, req)
		return
	}

	logger.Debug().Msg("Digest authentication succeeded")
	req.URL.User = url.User(username)

	logData := accesslog.GetLogData(req)
	if logData != nil {
		logData.Core[accesslog.ClientUsername] = username
	}

	if d.headerField != "" {
		req.Header[d.headerField] = []string{username}
	}

	if d.removeHeader {
		logger.Debug().Msg("Removing the Authorization header")
		req.Header.Del(authorizationHeader)
	}
	d.next.ServeHTTP(rw, req)
}

func (d *digestAuth) secretDigest(user, realm string) string {
	if secret, ok := d.users[user+":"+realm]; ok {
		return secret
	}

	return ""
}

func digestUserParser(user string) (string, string, error) {
	split := strings.Split(user, ":")
	if len(split) != 3 {
		return "", "", fmt.Errorf("error parsing DigestUser: %v", user)
	}
	return split[0] + ":" + split[1], split[2], nil
}
