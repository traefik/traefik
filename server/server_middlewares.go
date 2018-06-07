package server

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
	mauth "github.com/containous/traefik/middlewares/auth"
	"github.com/containous/traefik/middlewares/redirect"
	"github.com/containous/traefik/types"
	thoas_stats "github.com/thoas/stats"
	"github.com/unrolled/secure"
	"github.com/urfave/negroni"
)

func (s *Server) setupServerEntryPoint(newServerEntryPointName string, newServerEntryPoint *serverEntryPoint) *serverEntryPoint {
	serverMiddlewares := []negroni.Handler{middlewares.NegroniRecoverHandler()}

	if s.tracingMiddleware.IsEnabled() {
		serverMiddlewares = append(serverMiddlewares, s.tracingMiddleware.NewEntryPoint(newServerEntryPointName))
	}

	if s.accessLoggerMiddleware != nil {
		serverMiddlewares = append(serverMiddlewares, s.accessLoggerMiddleware)
	}

	if s.metricsRegistry.IsEnabled() {
		serverMiddlewares = append(serverMiddlewares, middlewares.NewEntryPointMetricsMiddleware(s.metricsRegistry, newServerEntryPointName))
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

	if s.entryPoints[newServerEntryPointName].Configuration.Auth != nil {
		authMiddleware, err := mauth.NewAuthenticator(s.entryPoints[newServerEntryPointName].Configuration.Auth, s.tracingMiddleware)
		if err != nil {
			log.Fatal("Error starting server: ", err)
		}
		serverMiddlewares = append(serverMiddlewares, s.wrapNegroniHandlerWithAccessLog(authMiddleware, fmt.Sprintf("Auth for entrypoint %s", newServerEntryPointName)))
	}

	if s.entryPoints[newServerEntryPointName].Configuration.Compress {
		serverMiddlewares = append(serverMiddlewares, &middlewares.Compress{})
	}

	ipWhitelistMiddleware, err := buildIPWhiteLister(
		s.entryPoints[newServerEntryPointName].Configuration.WhiteList,
		s.entryPoints[newServerEntryPointName].Configuration.WhitelistSourceRange)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
	if ipWhitelistMiddleware != nil {
		serverMiddlewares = append(serverMiddlewares, s.wrapNegroniHandlerWithAccessLog(ipWhitelistMiddleware, fmt.Sprintf("ipwhitelister for entrypoint %s", newServerEntryPointName)))
	}

	newSrv, listener, err := s.prepareServer(newServerEntryPointName, s.entryPoints[newServerEntryPointName].Configuration, newServerEntryPoint.httpRouter, serverMiddlewares)
	if err != nil {
		log.Fatal("Error preparing server: ", err)
	}

	serverEntryPoint := s.serverEntryPoints[newServerEntryPointName]
	serverEntryPoint.httpServer = newSrv
	serverEntryPoint.listener = listener

	return serverEntryPoint
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

func (s *Server) wrapNegroniHandlerWithAccessLog(handler negroni.Handler, frontendName string) negroni.Handler {
	if s.accessLoggerMiddleware != nil {
		saveBackend := accesslog.NewSaveNegroniBackend(handler, "Træfik")
		saveFrontend := accesslog.NewSaveNegroniFrontend(saveBackend, frontendName)
		return saveFrontend
	}
	return handler
}

func (s *Server) wrapHTTPHandlerWithAccessLog(handler http.Handler, frontendName string) http.Handler {
	if s.accessLoggerMiddleware != nil {
		saveBackend := accesslog.NewSaveBackend(handler, "Træfik")
		saveFrontend := accesslog.NewSaveFrontend(saveBackend, frontendName)
		return saveFrontend
	}
	return handler
}

func buildModifyResponse(secure *secure.Secure, header *middlewares.HeaderStruct) func(res *http.Response) error {
	return func(res *http.Response) error {
		if secure != nil {
			err := secure.ModifyResponseHeaders(res)
			if err != nil {
				return err
			}
		}
		if header != nil {
			err := header.ModifyResponseHeaders(res)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
