package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	casbinTypeName   = "CasbinAuth"
	CasbinAuthHeader = "X-Casbin-Authorization"
)

type casbinAuth struct {
	next     http.Handler
	enforcer *casbin.Enforcer
	name     string
}

func NewCasbin(ctx context.Context, next http.Handler, authConfig dynamic.CasbinAuth, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, casbinTypeName)).Debug("Creating middleware")

	enforcer, err := casbin.NewEnforcer(authConfig.ModelPath, authConfig.PolicyPath)
	if err != nil {
		return nil, err
	}

	ca := &casbinAuth{
		next:     next,
		enforcer: enforcer,
		name:     name,
	}
	return ca, nil
}

func (c *casbinAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := log.FromContext(middlewares.GetLoggerCtx(req.Context(), c.name, casbinTypeName))
	user, path, method := getParam(req)

	ok, err := c.enforcer.Enforce(user, path, method)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "could not operate authorization: %v", err)
		return
	}

	if !ok {
		logger.Debug("Authorization failed")
		tracing.SetErrorWithEvent(req, "Authorization failed")

		rw.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(rw, "authorization faile: %s, %s, %s", user, path, method)
		return
	}

	logger.Debug("Authorization succeeded")

	c.next.ServeHTTP(rw, req)
}

func getParam(req *http.Request) (string, string, string) {
	user := req.Header.Get(CasbinAuthHeader)
	path := req.URL.Path
	method := req.Method
	return user, path, method
}
