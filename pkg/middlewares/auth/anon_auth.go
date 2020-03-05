package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	goauth "github.com/abbot/go-http-auth"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/middlewares/accesslog"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	anonTypeName = "AnonAuth"
)

type anonAuth struct {
	next         http.Handler
	auth         *goauth.BasicAuth
	users        []*regexp.Regexp
	headerField  string
	removeHeader bool
	name         string
}

// NewAnon creates a anonAuth middleware.
func NewAnon(ctx context.Context, next http.Handler, authConfig dynamic.AnonAuth, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, anonTypeName)).Debug("Creating middleware")
	users, err := loadUsers(authConfig.UsersFile, authConfig.Users)

	var usersRegexp []*regexp.Regexp

	usersRegexp, err = anonUsersParser(users)

	if err != nil {
		return nil, err
	}

	aa := &anonAuth{
		next:         next,
		users:        usersRegexp,
		headerField:  authConfig.HeaderField,
		removeHeader: authConfig.RemoveHeader,
		name:         name,
	}

	realm := defaultRealm
	if len(authConfig.Realm) > 0 {
		realm = authConfig.Realm
	}
	aa.auth = goauth.NewBasicAuthenticator(realm, func(user, realm string) string { return "" })

	return aa, nil
}

func (a *anonAuth) GetTracingInformation() (string, ext.SpanKindEnum) {
	return a.name, tracing.SpanKindNoneEnum
}

func (a *anonAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := log.FromContext(middlewares.GetLoggerCtx(req.Context(), a.name, anonTypeName))

	username, _, ok := req.BasicAuth()

	if ok && a.anonUserCheck(username) {
		logger.Debug("Authentication succeeded")
		req.URL.User = url.User(username)

		logData := accesslog.GetLogData(req)
		if logData != nil {
			logData.Core[accesslog.ClientUsername] = username
		}

		if a.headerField != "" {
			req.Header[a.headerField] = []string{username}
		}

		if a.removeHeader {
			logger.Debug("Removing authorization header")
			req.Header.Del(authorizationHeader)
		}
		a.next.ServeHTTP(rw, req)

	} else {
		logger.Debug("Authentication required")
		tracing.SetErrorWithEvent(req, "Authentication required")
		a.auth.RequireAuth(rw, req)

	}
}

func anonUsersParser(users []string) ([]*regexp.Regexp, error) {

	var usersRegexp []*regexp.Regexp

	for _, user := range users {

		userRegexp, err := regexp.Compile(fmt.Sprintf(`^%v$`, user))

		if err != nil {
			return nil, err
		}

		usersRegexp = append(usersRegexp, userRegexp)
	}

	return usersRegexp, nil
}

func (a *anonAuth) anonUserCheck(user string) bool {

	if len(a.users) > 0 {
		for _, userRegexp := range a.users {

			if userRegexp.MatchString(user) {
				return true
			}
		}
		return false
	} else {
		return true
	}
}
