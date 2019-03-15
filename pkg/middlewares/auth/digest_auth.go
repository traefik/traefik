package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	goauth "github.com/abbot/go-http-auth"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/middlewares/accesslog"
	"github.com/containous/traefik/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	digestTypeName = "digestAuth"
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
func NewDigest(ctx context.Context, next http.Handler, authConfig config.DigestAuth, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, digestTypeName).Debug("Creating middleware")
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

func (d *digestAuth) GetTracingInformation() (string, ext.SpanKindEnum) {
	return d.name, tracing.SpanKindNoneEnum
}

func (d *digestAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), d.name, digestTypeName)

	if username, _ := d.auth.CheckAuth(req); username == "" {
		logger.Debug("Digest authentication failed")
		tracing.SetErrorWithEvent(req, "Digest authentication failed")
		d.auth.RequireAuth(rw, req)
	} else {
		logger.Debug("Digest authentication succeeded")
		req.URL.User = url.User(username)

		logData := accesslog.GetLogData(req)
		if logData != nil {
			logData.Core[accesslog.ClientUsername] = username
		}

		if d.headerField != "" {
			req.Header[d.headerField] = []string{username}
		}

		if d.removeHeader {
			logger.Debug("Removing the Authorization header")
			req.Header.Del(authorizationHeader)
		}
		d.next.ServeHTTP(rw, req)
	}
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
