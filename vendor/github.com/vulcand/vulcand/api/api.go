package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/scroll"
	"github.com/vulcand/vulcand/anomaly"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/plugin"
	"github.com/vulcand/vulcand/router"
)

type ProxyController struct {
	ng    engine.Engine
	stats engine.StatsProvider
	app   *scroll.App
}

func InitProxyController(ng engine.Engine, stats engine.StatsProvider, app *scroll.App) {
	c := &ProxyController{ng: ng, stats: stats, app: app}

	app.SetNotFoundHandler(c.handleError)

	app.AddHandler(scroll.Spec{Paths: []string{"/v1/status"}, Methods: []string{"GET"}, HandlerWithBody: c.getStatus})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/status"}, Methods: []string{"GET"}, HandlerWithBody: c.getStatus})

	app.AddHandler(scroll.Spec{Paths: []string{"/v2/log/severity"}, Methods: []string{"GET"}, Handler: c.getLogSeverity})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/log/severity"}, Methods: []string{"PUT"}, Handler: c.updateLogSeverity})

	// Hosts
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/hosts"}, Methods: []string{"POST"}, HandlerWithBody: c.upsertHost})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/hosts"}, Methods: []string{"GET"}, HandlerWithBody: c.getHosts})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/hosts/{hostname}"}, Methods: []string{"GET"}, Handler: c.getHost})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/hosts/{hostname}"}, Methods: []string{"DELETE"}, Handler: c.deleteHost})

	// Listeners
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/listeners"}, Methods: []string{"GET"}, Handler: c.getListeners})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/listeners"}, Methods: []string{"POST"}, HandlerWithBody: c.upsertListener})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/listeners/{id}"}, Methods: []string{"GET"}, Handler: c.getListener})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/listeners/{id}"}, Methods: []string{"DELETE"}, Handler: c.deleteListener})

	// Top provides top-style realtime statistics about frontends and servers
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/top/frontends"}, Methods: []string{"GET"}, Handler: c.getTopFrontends})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/top/servers"}, Methods: []string{"GET"}, Handler: c.getTopServers})

	// Frontends
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/frontends"}, Methods: []string{"POST"}, HandlerWithBody: c.upsertFrontend})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/frontends/{id}"}, Methods: []string{"GET"}, Handler: c.getFrontend})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/frontends"}, Methods: []string{"GET"}, Handler: c.getFrontends})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/frontends/{id}"}, Methods: []string{"DELETE"}, Handler: c.deleteFrontend})

	// Backends
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/backends"}, Methods: []string{"POST"}, HandlerWithBody: c.upsertBackend})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/backends"}, Methods: []string{"GET"}, Handler: c.getBackends})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/backends/{id}"}, Methods: []string{"DELETE"}, Handler: c.deleteBackend})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/backends/{id}"}, Methods: []string{"GET"}, Handler: c.getBackend})

	// Servers
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/backends/{backendId}/servers"}, Methods: []string{"GET"}, Handler: c.getServers})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/backends/{backendId}/servers"}, Methods: []string{"POST"}, HandlerWithBody: c.upsertServer})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/backends/{backendId}/servers/{id}"}, Methods: []string{"GET"}, Handler: c.getServer})
	app.AddHandler(scroll.Spec{Paths: []string{"/v2/backends/{backendId}/servers/{id}"}, Methods: []string{"DELETE"}, Handler: c.deleteServer})

	// Middlewares
	c.app.AddHandler(
		scroll.Spec{
			Paths:           []string{fmt.Sprintf("/v2/frontends/{frontend}/middlewares")},
			Methods:         []string{"POST"},
			HandlerWithBody: c.upsertMiddleware,
		})

	c.app.AddHandler(
		scroll.Spec{
			Paths:   []string{fmt.Sprintf("/v2/frontends/{frontend}/middlewares/{id}")},
			Methods: []string{"GET"},
			Handler: c.getMiddleware,
		})

	c.app.AddHandler(
		scroll.Spec{
			Paths:   []string{fmt.Sprintf("/v2/frontends/{frontend}/middlewares")},
			Methods: []string{"GET"},
			Handler: c.getMiddlewares,
		})

	c.app.AddHandler(
		scroll.Spec{
			Paths:   []string{fmt.Sprintf("/v2/frontends/{frontend}/middlewares/{id}")},
			Methods: []string{"DELETE"},
			Handler: c.deleteMiddleware,
		})
}

func (c *ProxyController) handleError(w http.ResponseWriter, r *http.Request) {
	scroll.ReplyError(w, scroll.NotFoundError{Description: "Object not found"})
}

func (c *ProxyController) getStatus(w http.ResponseWriter, r *http.Request, params map[string]string, body []byte) (interface{}, error) {
	return scroll.Response{
		"Status": "ok",
	}, nil
}

func (c *ProxyController) getLogSeverity(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	return scroll.Response{
		"severity": c.ng.GetLogSeverity().String(),
	}, nil
}

func (c *ProxyController) updateLogSeverity(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	sev, err := log.ParseLevel(strings.ToLower(r.Form.Get("severity")))
	if err != nil {
		return nil, formatError(err)
	}
	c.ng.SetLogSeverity(sev)
	return scroll.Response{"message": fmt.Sprintf("Severity has been updated to %v", sev.String())}, nil
}

func (c *ProxyController) getHosts(w http.ResponseWriter, r *http.Request, params map[string]string, body []byte) (interface{}, error) {
	hosts, err := c.ng.GetHosts()
	return scroll.Response{
		"Hosts": hosts,
	}, err
}

func (c *ProxyController) getHost(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	h, err := c.ng.GetHost(engine.HostKey{Name: params["hostname"]})
	if err != nil {
		return nil, formatError(err)
	}
	return formatResult(h, err)
}

func (c *ProxyController) getFrontends(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	fs, err := c.ng.GetFrontends()
	if err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{
		"Frontends": fs,
	}, nil
}

func (c *ProxyController) getTopFrontends(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	limit, err := strconv.Atoi(r.Form.Get("limit"))
	if err != nil {
		return nil, formatError(err)
	}
	var bk *engine.BackendKey
	if key := r.Form.Get("backendId"); key != "" {
		bk = &engine.BackendKey{Id: key}
	}
	frontends, err := c.stats.TopFrontends(bk)
	if err != nil {
		return nil, formatError(err)
	}
	if limit > 0 && limit < len(frontends) {
		frontends = frontends[:limit]
	}
	return scroll.Response{
		"Frontends": frontends,
	}, nil
}

func (c *ProxyController) getFrontend(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	return formatResult(c.ng.GetFrontend(engine.FrontendKey{Id: params["id"]}))
}

func (c *ProxyController) upsertHost(w http.ResponseWriter, r *http.Request, params map[string]string, body []byte) (interface{}, error) {
	host, err := parseHostPack(body)
	if err != nil {
		return nil, formatError(err)
	}
	log.Infof("Upsert %s", host)
	return formatResult(host, c.ng.UpsertHost(*host))
}

func (c *ProxyController) getListeners(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	ls, err := c.ng.GetListeners()
	return scroll.Response{
		"Listeners": ls,
	}, err
}

func (c *ProxyController) upsertListener(w http.ResponseWriter, r *http.Request, params map[string]string, body []byte) (interface{}, error) {
	listener, err := parseListenerPack(body)
	if err != nil {
		return nil, formatError(err)
	}
	log.Infof("Upsert %s", listener)
	return formatResult(listener, c.ng.UpsertListener(*listener))
}

func (c *ProxyController) getListener(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	log.Infof("Get Listener(id=%s)", params["id"])
	return formatResult(c.ng.GetListener(engine.ListenerKey{Id: params["id"]}))
}

func (c *ProxyController) deleteListener(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	log.Infof("Delete Listener(id=%s)", params["id"])
	if err := c.ng.DeleteListener(engine.ListenerKey{Id: params["id"]}); err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{"message": "Listener deleted"}, nil
}

func (c *ProxyController) deleteHost(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	hostname := params["hostname"]
	log.Infof("Delete host: %s", hostname)
	if err := c.ng.DeleteHost(engine.HostKey{Name: hostname}); err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{"message": fmt.Sprintf("Host '%s' deleted", hostname)}, nil
}

func (c *ProxyController) upsertBackend(w http.ResponseWriter, r *http.Request, params map[string]string, body []byte) (interface{}, error) {
	b, err := parseBackendPack(body)
	if err != nil {
		return nil, formatError(err)
	}
	log.Infof("Upsert Backend: %s", b)
	return formatResult(b, c.ng.UpsertBackend(*b))
}

func (c *ProxyController) deleteBackend(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	backendId := params["id"]
	log.Infof("Delete Backend(id=%s)", backendId)
	if err := c.ng.DeleteBackend(engine.BackendKey{Id: backendId}); err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{"message": "Backend deleted"}, nil
}

func (c *ProxyController) getBackends(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	backends, err := c.ng.GetBackends()
	return scroll.Response{
		"Backends": backends,
	}, err
}

func (c *ProxyController) getTopServers(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	limit, err := strconv.Atoi(r.Form.Get("limit"))
	if err != nil {
		return nil, formatError(err)
	}
	var bk *engine.BackendKey
	if key := r.Form.Get("backendId"); key != "" {
		bk = &engine.BackendKey{Id: key}
	}
	servers, err := c.stats.TopServers(bk)
	if err != nil {
		return nil, formatError(err)
	}
	if bk != nil {
		anomaly.MarkServerAnomalies(servers)
	}
	if limit > 0 && limit < len(servers) {
		servers = servers[:limit]
	}
	return scroll.Response{
		"Servers": servers,
	}, nil
}

func (c *ProxyController) getBackend(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	return formatResult(c.ng.GetBackend(engine.BackendKey{Id: params["id"]}))
}

func (c *ProxyController) upsertFrontend(w http.ResponseWriter, r *http.Request, params map[string]string, body []byte) (interface{}, error) {
	frontend, ttl, err := parseFrontendPack(c.ng.GetRegistry().GetRouter(), body)
	if err != nil {
		return nil, formatError(err)
	}
	log.Infof("Upsert %s", frontend)
	return formatResult(frontend, c.ng.UpsertFrontend(*frontend, ttl))
}

func (c *ProxyController) deleteFrontend(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	log.Infof("Delete Frontend(id=%s)", params["id"])
	if err := c.ng.DeleteFrontend(engine.FrontendKey{Id: params["id"]}); err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{"message": "Frontend deleted"}, nil
}

func (c *ProxyController) upsertServer(w http.ResponseWriter, r *http.Request, params map[string]string, body []byte) (interface{}, error) {
	backendId := params["backendId"]
	srv, ttl, err := parseServerPack(body)
	if err != nil {
		return nil, formatError(err)
	}
	bk := engine.BackendKey{Id: backendId}
	log.Infof("Upsert %v %v", bk, srv)
	return formatResult(srv, c.ng.UpsertServer(bk, *srv, ttl))
}

func (c *ProxyController) getServer(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	sk := engine.ServerKey{BackendKey: engine.BackendKey{Id: params["backendId"]}, Id: params["id"]}
	log.Infof("getServer %v", sk)
	srv, err := c.ng.GetServer(sk)
	if err != nil {
		return nil, formatError(err)
	}
	return formatResult(srv, err)
}

func (c *ProxyController) getServers(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	srvs, err := c.ng.GetServers(engine.BackendKey{Id: params["backendId"]})
	if err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{
		"Servers": srvs,
	}, nil
}

func (c *ProxyController) deleteServer(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	sk := engine.ServerKey{BackendKey: engine.BackendKey{Id: params["backendId"]}, Id: params["id"]}
	log.Infof("Delete %v", sk)
	if err := c.ng.DeleteServer(sk); err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{"message": "Server deleted"}, nil
}

func (c *ProxyController) upsertMiddleware(w http.ResponseWriter, r *http.Request, params map[string]string, body []byte) (interface{}, error) {
	frontend := params["frontend"]
	m, ttl, err := parseMiddlewarePack(body, c.ng.GetRegistry())
	if err != nil {
		return nil, formatError(err)
	}
	return formatResult(m, c.ng.UpsertMiddleware(engine.FrontendKey{Id: frontend}, *m, ttl))
}

func (c *ProxyController) getMiddleware(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	fk := engine.MiddlewareKey{Id: params["id"], FrontendKey: engine.FrontendKey{Id: params["frontend"]}}
	return formatResult(c.ng.GetMiddleware(fk))
}

func (c *ProxyController) getMiddlewares(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	fk := engine.FrontendKey{Id: params["frontend"]}
	out, err := c.ng.GetMiddlewares(fk)
	if err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{
		"Middlewares": out,
	}, nil
}

func (c *ProxyController) deleteMiddleware(w http.ResponseWriter, r *http.Request, params map[string]string) (interface{}, error) {
	fk := engine.MiddlewareKey{Id: params["id"], FrontendKey: engine.FrontendKey{Id: params["frontend"]}}
	if err := c.ng.DeleteMiddleware(fk); err != nil {
		return nil, formatError(err)
	}
	return scroll.Response{"message": "Middleware deleted"}, nil
}

func formatError(e error) error {
	switch err := e.(type) {
	case *engine.AlreadyExistsError:
		return scroll.ConflictError{Description: err.Error()}
	case *engine.NotFoundError:
		return scroll.NotFoundError{Description: err.Error()}
	case *engine.InvalidFormatError:
		return scroll.InvalidParameterError{Value: err.Error()}
	case scroll.GenericAPIError, scroll.MissingFieldError,
		scroll.InvalidFormatError, scroll.InvalidParameterError,
		scroll.NotFoundError, scroll.ConflictError:
		return e
	}
	return scroll.GenericAPIError{Reason: e.Error()}
}

func formatResult(in interface{}, err error) (interface{}, error) {
	if err != nil {
		return nil, formatError(err)
	}
	return in, nil
}

type backendPack struct {
	Backend engine.Backend
}

type backendReadPack struct {
	Backend json.RawMessage
}

type hostPack struct {
	Host engine.Host
}

type hostReadPack struct {
	Host json.RawMessage
}

type listenerPack struct {
	Listener engine.Listener
	TTL      string
}

type listenerReadPack struct {
	Listener json.RawMessage
}

type frontendReadPack struct {
	Frontend json.RawMessage
	TTL      string
}

type frontendPack struct {
	Frontend engine.Frontend
	TTL      string
}

type middlewareReadPack struct {
	Middleware json.RawMessage
	TTL        string
}

type middlewarePack struct {
	Middleware engine.Middleware
	TTL        string
}

type serverReadPack struct {
	Server json.RawMessage
	TTL    string
}

type serverPack struct {
	Server engine.Server
	TTL    string
}

func parseListenerPack(v []byte) (*engine.Listener, error) {
	var lp listenerReadPack
	if err := json.Unmarshal(v, &lp); err != nil {
		return nil, err
	}
	if len(lp.Listener) == 0 {
		return nil, &scroll.MissingFieldError{Field: "Listener"}
	}
	return engine.ListenerFromJSON(lp.Listener)
}

func parseHostPack(v []byte) (*engine.Host, error) {
	var hp hostReadPack
	if err := json.Unmarshal(v, &hp); err != nil {
		return nil, err
	}
	if len(hp.Host) == 0 {
		return nil, &scroll.MissingFieldError{Field: "Host"}
	}
	return engine.HostFromJSON(hp.Host)
}

func parseBackendPack(v []byte) (*engine.Backend, error) {
	var bp *backendReadPack
	if err := json.Unmarshal(v, &bp); err != nil {
		return nil, err
	}
	if bp == nil || len(bp.Backend) == 0 {
		return nil, &scroll.MissingFieldError{Field: "Backend"}
	}
	return engine.BackendFromJSON(bp.Backend)
}

func parseFrontendPack(router router.Router, v []byte) (*engine.Frontend, time.Duration, error) {
	var fp frontendReadPack
	if err := json.Unmarshal(v, &fp); err != nil {
		return nil, 0, err
	}
	if len(fp.Frontend) == 0 {
		return nil, 0, &scroll.MissingFieldError{Field: "Frontend"}
	}
	f, err := engine.FrontendFromJSON(router, fp.Frontend)
	if err != nil {
		return nil, 0, err
	}

	var ttl time.Duration
	if fp.TTL != "" {
		ttl, err = time.ParseDuration(fp.TTL)
		if err != nil {
			return nil, 0, err
		}
	}
	return f, ttl, nil
}

func parseMiddlewarePack(v []byte, r *plugin.Registry) (*engine.Middleware, time.Duration, error) {
	var mp middlewareReadPack
	if err := json.Unmarshal(v, &mp); err != nil {
		return nil, 0, err
	}
	if len(mp.Middleware) == 0 {
		return nil, 0, &scroll.MissingFieldError{Field: "Middleware"}
	}
	f, err := engine.MiddlewareFromJSON(mp.Middleware, r.GetSpec)
	if err != nil {
		return nil, 0, err
	}
	var ttl time.Duration
	if mp.TTL != "" {
		ttl, err = time.ParseDuration(mp.TTL)
		if err != nil {
			return nil, 0, err
		}
	}
	return f, ttl, nil
}

func parseServerPack(v []byte) (*engine.Server, time.Duration, error) {
	var sp serverReadPack
	if err := json.Unmarshal(v, &sp); err != nil {
		return nil, 0, err
	}
	if len(sp.Server) == 0 {
		return nil, 0, &scroll.MissingFieldError{Field: "Server"}
	}
	s, err := engine.ServerFromJSON(sp.Server)
	if err != nil {
		return nil, 0, err
	}
	var ttl time.Duration
	if sp.TTL != "" {
		ttl, err = time.ParseDuration(sp.TTL)
		if err != nil {
			return nil, 0, err
		}
	}
	return s, ttl, nil
}
