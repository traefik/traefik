package server

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
	mauth "github.com/containous/traefik/middlewares/auth"
	"github.com/containous/traefik/middlewares/errorpages"
	"github.com/containous/traefik/middlewares/redirect"
	"github.com/containous/traefik/types"
	thoas_stats "github.com/thoas/stats"
	"github.com/unrolled/secure"
	"github.com/urfave/negroni"
)

type handlerPostConfig func(backendsHandlers map[string]http.Handler) error

type modifyResponse func(*http.Response) error

func (s *Server) buildMiddlewares(frontendName string, frontend *types.Frontend,
	backends map[string]*types.Backend,
	entryPointName string, entryPoint *configuration.EntryPoint,
	providerName string) ([]negroni.Handler, modifyResponse, handlerPostConfig, error) {

	var middle []negroni.Handler
	var postConfig handlerPostConfig

	// Error pages
	if len(frontend.Errors) > 0 {
		handlers, err := buildErrorPagesMiddleware(frontendName, frontend, backends, entryPointName, providerName)
		if err != nil {
			return nil, nil, nil, err
		}

		postConfig = errorPagesPostConfig(handlers)

		for _, handler := range handlers {
			middle = append(middle, handler)
		}
	}

	// Metrics
	if s.metricsRegistry.IsEnabled() {
		handler := middlewares.NewBackendMetricsMiddleware(s.metricsRegistry, frontend.Backend)
		middle = append(middle, handler)
	}

	// Whitelist
	ipWhitelistMiddleware, err := buildIPWhiteLister(frontend.WhiteList, frontend.WhitelistSourceRange)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating IP Whitelister: %s", err)
	}
	if ipWhitelistMiddleware != nil {
		if frontend.WhiteList != nil {
			log.Debugf("Configured IP Whitelists: %v", frontend.WhiteList.SourceRange)
		} else {
			log.Debugf("Configured IP Whitelists: %v", frontend.WhitelistSourceRange)
		}

		handler := s.tracingMiddleware.NewNegroniHandlerWrapper(
			"IP whitelist",
			s.wrapNegroniHandlerWithAccessLog(ipWhitelistMiddleware, fmt.Sprintf("ipwhitelister for %s", frontendName)),
			false)
		middle = append(middle, handler)
	}

	// Redirect
	if frontend.Redirect != nil && entryPointName != frontend.Redirect.EntryPoint {
		rewrite, err := s.buildRedirectHandler(entryPointName, frontend.Redirect)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error creating Frontend Redirect: %v", err)
		}

		handler := s.wrapNegroniHandlerWithAccessLog(rewrite, fmt.Sprintf("frontend redirect for %s", frontendName))
		middle = append(middle, handler)

		log.Debugf("Frontend %s redirect created", frontendName)
	}

	// Header
	headerMiddleware := middlewares.NewHeaderFromStruct(frontend.Headers)
	if headerMiddleware != nil {
		log.Debugf("Adding header middleware for frontend %s", frontendName)

		handler := s.tracingMiddleware.NewNegroniHandlerWrapper("Header", headerMiddleware, false)
		middle = append(middle, handler)
	}

	// Secure
	secureMiddleware := middlewares.NewSecure(frontend.Headers)
	if secureMiddleware != nil {
		log.Debugf("Adding secure middleware for frontend %s", frontendName)

		handler := negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNextForRequestOnly)
		middle = append(middle, handler)
	}

	// Basic auth
	if len(frontend.BasicAuth) > 0 {
		log.Debugf("Adding basic authentication for frontend %s", frontendName)

		authMiddleware, err := s.buildBasicAuthMiddleware(frontend.BasicAuth)
		if err != nil {
			return nil, nil, nil, err
		}

		handler := s.wrapNegroniHandlerWithAccessLog(authMiddleware, fmt.Sprintf("Basic Auth for %s", frontendName))
		middle = append(middle, handler)
	}

	// TLSClientHeaders
	tlsClientHeadersMiddleware := middlewares.NewTLSClientHeaders(frontend)
	if tlsClientHeadersMiddleware != nil {
		log.Debugf("Adding TLSClientHeaders middleware for frontend %s", frontendName)

		handler := s.tracingMiddleware.NewNegroniHandlerWrapper("TLSClientHeaders", tlsClientHeadersMiddleware, false)
		middle = append(middle, handler)
	}

	// Authentication
	if frontend.Auth != nil {
		authMiddleware, err := mauth.NewAuthenticator(frontend.Auth, s.tracingMiddleware)
		if err != nil {
			return nil, nil, nil, err
		}

		handler := s.wrapNegroniHandlerWithAccessLog(authMiddleware, fmt.Sprintf("Auth for %s", frontendName))
		middle = append(middle, handler)
	}

	return middle, buildModifyResponse(secureMiddleware, headerMiddleware), postConfig, nil
}

func (s *Server) buildServerEntryPointMiddlewares(serverEntryPointName string, serverEntryPoint *serverEntryPoint) ([]negroni.Handler, error) {
	serverMiddlewares := []negroni.Handler{middlewares.NegroniRecoverHandler()}

	if s.tracingMiddleware.IsEnabled() {
		serverMiddlewares = append(serverMiddlewares, s.tracingMiddleware.NewEntryPoint(serverEntryPointName))
	}

	if s.accessLoggerMiddleware != nil {
		serverMiddlewares = append(serverMiddlewares, s.accessLoggerMiddleware)
	}

	if s.metricsRegistry.IsEnabled() {
		serverMiddlewares = append(serverMiddlewares, middlewares.NewEntryPointMetricsMiddleware(s.metricsRegistry, serverEntryPointName))
	}

	if s.globalConfiguration.API != nil {
		if s.globalConfiguration.API.Stats == nil {
			s.globalConfiguration.API.Stats = thoas_stats.New()
		}
		serverMiddlewares = append(serverMiddlewares, s.globalConfiguration.API.Stats)
		if s.globalConfiguration.API.Statistics != nil {
			if s.globalConfiguration.API.StatsRecorder == nil {
				s.globalConfiguration.API.StatsRecorder = middlewares.NewStatsRecorder(s.globalConfiguration.API.Statistics.RecentErrors)
			}
			serverMiddlewares = append(serverMiddlewares, s.globalConfiguration.API.StatsRecorder)
		}
	}

	if s.entryPoints[serverEntryPointName].Configuration.Redirect != nil {
		redirectHandlers, err := s.buildEntryPointRedirect()
		if err != nil {
			return nil, fmt.Errorf("failed to create redirect middleware: %v", err)
		}
		serverMiddlewares = append(serverMiddlewares, redirectHandlers[serverEntryPointName])
	}

	if s.entryPoints[serverEntryPointName].Configuration.Auth != nil {
		authMiddleware, err := mauth.NewAuthenticator(s.entryPoints[serverEntryPointName].Configuration.Auth, s.tracingMiddleware)
		if err != nil {
			return nil, fmt.Errorf("failed to create authentication middleware: %v", err)
		}
		serverMiddlewares = append(serverMiddlewares, s.wrapNegroniHandlerWithAccessLog(authMiddleware, fmt.Sprintf("Auth for entrypoint %s", serverEntryPointName)))
	}

	if s.entryPoints[serverEntryPointName].Configuration.Compress {
		serverMiddlewares = append(serverMiddlewares, &middlewares.Compress{})
	}

	ipWhitelistMiddleware, err := buildIPWhiteLister(
		s.entryPoints[serverEntryPointName].Configuration.WhiteList,
		s.entryPoints[serverEntryPointName].Configuration.WhitelistSourceRange)
	if err != nil {
		return nil, fmt.Errorf("failed to create ip whitelist middleware: %v", err)
	}
	if ipWhitelistMiddleware != nil {
		serverMiddlewares = append(serverMiddlewares, s.wrapNegroniHandlerWithAccessLog(ipWhitelistMiddleware, fmt.Sprintf("ipwhitelister for entrypoint %s", serverEntryPointName)))
	}

	// RequestHost Cannonizer
	serverMiddlewares = append(serverMiddlewares, &middlewares.RequestHost{})

	return serverMiddlewares, nil
}

func errorPagesPostConfig(epHandlers []*errorpages.Handler) handlerPostConfig {
	return func(backendsHandlers map[string]http.Handler) error {
		for _, errorPageHandler := range epHandlers {
			if handler, ok := backendsHandlers[errorPageHandler.BackendName]; ok {
				err := errorPageHandler.PostLoad(handler)
				if err != nil {
					return fmt.Errorf("failed to configure error pages for backend %s: %v", errorPageHandler.BackendName, err)
				}
			} else {
				err := errorPageHandler.PostLoad(nil)
				if err != nil {
					return fmt.Errorf("failed to configure error pages for %s: %v", errorPageHandler.FallbackURL, err)
				}
			}
		}
		return nil
	}
}

func buildErrorPagesMiddleware(frontendName string, frontend *types.Frontend, backends map[string]*types.Backend, entryPointName string, providerName string) ([]*errorpages.Handler, error) {
	var errorPageHandlers []*errorpages.Handler

	for errorPageName, errorPage := range frontend.Errors {
		if frontend.Backend == errorPage.Backend {
			log.Errorf("Error when creating error page %q for frontend %q: error pages backend %q is the same as backend for the frontend (infinite call risk).",
				errorPageName, frontendName, errorPage.Backend)
		} else if backends[errorPage.Backend] == nil {
			log.Errorf("Error when creating error page %q for frontend %q: the backend %q doesn't exist.",
				errorPageName, frontendName, errorPage.Backend)
		} else {
			errorPagesHandler, err := errorpages.NewHandler(errorPage, entryPointName+providerName+errorPage.Backend)
			if err != nil {
				return nil, fmt.Errorf("error creating error pages: %v", err)
			}

			if errorPageServer, ok := backends[errorPage.Backend].Servers["error"]; ok {
				errorPagesHandler.FallbackURL = errorPageServer.URL
			}

			errorPageHandlers = append(errorPageHandlers, errorPagesHandler)
		}
	}

	return errorPageHandlers, nil
}

func (s *Server) buildBasicAuthMiddleware(authData []string) (*mauth.Authenticator, error) {
	users := types.Users{}
	for _, user := range authData {
		users = append(users, user)
	}

	auth := &types.Auth{}
	auth.Basic = &types.Basic{
		Users: users,
	}

	authMiddleware, err := mauth.NewAuthenticator(auth, s.tracingMiddleware)
	if err != nil {
		return nil, fmt.Errorf("error creating Basic Auth: %v", err)
	}

	return authMiddleware, nil
}

func (s *Server) buildEntryPointRedirect() (map[string]negroni.Handler, error) {
	redirectHandlers := map[string]negroni.Handler{}

	for entryPointName, ep := range s.entryPoints {
		entryPoint := ep.Configuration

		if entryPoint.Redirect != nil && entryPointName != entryPoint.Redirect.EntryPoint {
			handler, err := s.buildRedirectHandler(entryPointName, entryPoint.Redirect)
			if err != nil {
				return nil, fmt.Errorf("error loading configuration for entrypoint %s: %v", entryPointName, err)
			}

			handlerToUse := s.wrapNegroniHandlerWithAccessLog(handler, fmt.Sprintf("entrypoint redirect for %s", entryPointName))
			redirectHandlers[entryPointName] = handlerToUse
		}
	}

	return redirectHandlers, nil
}

func (s *Server) buildRedirectHandler(srcEntryPointName string, opt *types.Redirect) (negroni.Handler, error) {
	// entry point redirect
	if len(opt.EntryPoint) > 0 {
		entryPoint := s.entryPoints[opt.EntryPoint].Configuration
		if entryPoint == nil {
			return nil, fmt.Errorf("unknown target entrypoint %q", srcEntryPointName)
		}
		log.Debugf("Creating entry point redirect %s -> %s", srcEntryPointName, opt.EntryPoint)
		return redirect.NewEntryPointHandler(entryPoint, opt.Permanent)
	}

	// regex redirect
	redirection, err := redirect.NewRegexHandler(opt.Regex, opt.Replacement, opt.Permanent)
	if err != nil {
		return nil, err
	}
	log.Debugf("Creating regex redirect %s -> %s -> %s", srcEntryPointName, opt.Regex, opt.Replacement)

	return redirection, nil
}

func buildIPWhiteLister(whiteList *types.WhiteList, wlRange []string) (*middlewares.IPWhiteLister, error) {
	if whiteList != nil &&
		len(whiteList.SourceRange) > 0 {
		return middlewares.NewIPWhiteLister(whiteList.SourceRange, whiteList.UseXForwardedFor)
	} else if len(wlRange) > 0 {
		return middlewares.NewIPWhiteLister(wlRange, false)
	}
	return nil, nil
}

func (s *Server) wrapNegroniHandlerWithAccessLog(handler negroni.Handler, frontendName string) negroni.Handler {
	if s.accessLoggerMiddleware != nil {
		saveUsername := accesslog.NewSaveNegroniUsername(handler)
		saveBackend := accesslog.NewSaveNegroniBackend(saveUsername, "Traefik")
		saveFrontend := accesslog.NewSaveNegroniFrontend(saveBackend, frontendName)
		return saveFrontend
	}
	return handler
}

func (s *Server) wrapHTTPHandlerWithAccessLog(handler http.Handler, frontendName string) http.Handler {
	if s.accessLoggerMiddleware != nil {
		saveUsername := accesslog.NewSaveUsername(handler)
		saveBackend := accesslog.NewSaveBackend(saveUsername, "Traefik")
		saveFrontend := accesslog.NewSaveFrontend(saveBackend, frontendName)
		return saveFrontend
	}
	return handler
}

func buildModifyResponse(secure *secure.Secure, header *middlewares.HeaderStruct) func(res *http.Response) error {
	return func(res *http.Response) error {
		if secure != nil {
			if err := secure.ModifyResponseHeaders(res); err != nil {
				return err
			}
		}

		if header != nil {
			if err := header.ModifyResponseHeaders(res); err != nil {
				return err
			}
		}
		return nil
	}
}
