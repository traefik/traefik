package server

import (
	"encoding/json"
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/codegangsta/negroni"
	"github.com/containous/mux"
	"github.com/containous/traefik/autogen"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	thoas_stats "github.com/thoas/stats"
	"github.com/unrolled/render"
)

var (
	metrics       = thoas_stats.New()
	statsRecorder *middlewares.StatsRecorder
)

// WebProvider is a provider.Provider implementation that provides the UI.
// FIXME to be handled another way.
type WebProvider struct {
	Address    string            `description:"Web administration port"`
	CertFile   string            `description:"SSL certificate"`
	KeyFile    string            `description:"SSL certificate"`
	ReadOnly   bool              `description:"Enable read only API"`
	Statistics *types.Statistics `description:"Enable more detailed statistics"`
	Metrics    *types.Metrics    `description:"Enable a metrics exporter"`
	Path       string            `description:"Root path for dashboard and API"`
	server     *Server
	Auth       *types.Auth
}

var (
	templatesRenderer = render.New(render.Options{
		Directory: "nowhere",
	})
)

func init() {
	expvar.Publish("Goroutines", expvar.Func(goroutines))
}

func goroutines() interface{} {
	return runtime.NumGoroutine()
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *WebProvider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, _ types.Constraints) error {

	systemRouter := mux.NewRouter()

	if provider.Path == "" {
		provider.Path = "/"
	}

	if provider.Path != "/" {
		if provider.Path[len(provider.Path)-1:] != "/" {
			provider.Path += "/"
		}
		systemRouter.Methods("GET").Path("/").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			http.Redirect(response, request, provider.Path, 302)
		})
	}

	// Prometheus route
	if provider.Metrics != nil && provider.Metrics.Prometheus != nil {
		systemRouter.Methods("GET").Path(provider.Path + "metrics").Handler(promhttp.Handler())
	}

	// health route
	systemRouter.Methods("GET").Path(provider.Path + "health").HandlerFunc(provider.getHealthHandler)

	// ping route
	systemRouter.Methods("GET").Path(provider.Path + "ping").HandlerFunc(provider.getPingHandler)
	// API routes
	systemRouter.Methods("GET").Path(provider.Path + "api").HandlerFunc(provider.getConfigHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/version").HandlerFunc(provider.getVersionHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers").HandlerFunc(provider.getConfigHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}").HandlerFunc(provider.getProviderHandler)
	systemRouter.Methods("PUT").Path(provider.Path + "api/providers/{provider}").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if provider.ReadOnly {
			response.WriteHeader(http.StatusForbidden)
			fmt.Fprint(response, "REST API is in read-only mode")
			return
		}
		vars := mux.Vars(request)
		if vars["provider"] != "web" {
			response.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(response, "Only 'web' provider can be updated through the REST API")
			return
		}

		configuration := new(types.Configuration)
		body, _ := ioutil.ReadAll(request.Body)
		err := json.Unmarshal(body, configuration)
		if err == nil {
			configurationChan <- types.ConfigMessage{ProviderName: "web", Configuration: configuration}
			provider.getConfigHandler(response, request)
		} else {
			log.Errorf("Error parsing configuration %+v", err)
			http.Error(response, fmt.Sprintf("%+v", err), http.StatusBadRequest)
		}
	})
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}/backends").HandlerFunc(provider.getBackendsHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}/backends/{backend}").HandlerFunc(provider.getBackendHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}/backends/{backend}/servers").HandlerFunc(provider.getServersHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}/backends/{backend}/servers/{server}").HandlerFunc(provider.getServerHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}/frontends").HandlerFunc(provider.getFrontendsHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}/frontends/{frontend}").HandlerFunc(provider.getFrontendHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}/frontends/{frontend}/routes").HandlerFunc(provider.getRoutesHandler)
	systemRouter.Methods("GET").Path(provider.Path + "api/providers/{provider}/frontends/{frontend}/routes/{route}").HandlerFunc(provider.getRouteHandler)

	// Expose dashboard
	systemRouter.Methods("GET").Path(provider.Path).HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, provider.Path+"dashboard/", 302)
	})
	systemRouter.Methods("GET").PathPrefix(provider.Path + "dashboard/").Handler(http.StripPrefix(provider.Path+"dashboard/", http.FileServer(&assetfs.AssetFS{Asset: autogen.Asset, AssetInfo: autogen.AssetInfo, AssetDir: autogen.AssetDir, Prefix: "static"})))

	// expvars
	if provider.server.globalConfiguration.Debug {
		systemRouter.Methods("GET").Path(provider.Path + "debug/vars").HandlerFunc(expvarHandler)
	}

	go func() {
		var err error
		var negroni = negroni.New()
		if provider.Auth != nil {
			authMiddleware, err := middlewares.NewAuthenticator(provider.Auth)
			if err != nil {
				log.Fatal("Error creating Auth: ", err)
			}
			negroni.Use(authMiddleware)
		}
		negroni.UseHandler(systemRouter)

		if len(provider.CertFile) > 0 && len(provider.KeyFile) > 0 {
			err = http.ListenAndServeTLS(provider.Address, provider.CertFile, provider.KeyFile, negroni)
		} else {
			err = http.ListenAndServe(provider.Address, negroni)
		}

		if err != nil {
			log.Fatal("Error creating server: ", err)
		}
	}()
	return nil
}

// healthResponse combines data returned by thoas/stats with statistics (if
// they are enabled).
type healthResponse struct {
	*thoas_stats.Data
	*middlewares.Stats
}

func (provider *WebProvider) getHealthHandler(response http.ResponseWriter, request *http.Request) {
	health := &healthResponse{Data: metrics.Data()}
	if statsRecorder != nil {
		health.Stats = statsRecorder.Data()
	}
	templatesRenderer.JSON(response, http.StatusOK, health)
}

func (provider *WebProvider) getPingHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprint(response, "OK")
}

func (provider *WebProvider) getConfigHandler(response http.ResponseWriter, request *http.Request) {
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	templatesRenderer.JSON(response, http.StatusOK, currentConfigurations)
}

func (provider *WebProvider) getVersionHandler(response http.ResponseWriter, request *http.Request) {
	v := struct {
		Version  string
		Codename string
	}{
		Version:  version.Version,
		Codename: version.Codename,
	}
	templatesRenderer.JSON(response, http.StatusOK, v)
}

func (provider *WebProvider) getProviderHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *WebProvider) getBackendsHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider.Backends)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *WebProvider) getBackendHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, backend)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getServersHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, backend.Servers)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getServerHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	serverID := vars["server"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			if server, ok := backend.Servers[serverID]; ok {
				templatesRenderer.JSON(response, http.StatusOK, server)
				return
			}
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getFrontendsHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider.Frontends)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *WebProvider) getFrontendHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, frontend)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getRoutesHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, frontend.Routes)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getRouteHandler(response http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	routeID := vars["route"]
	currentConfigurations := provider.server.currentConfigurations.Get().(configs)
	if provider, ok := currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			if route, ok := frontend.Routes[routeID]; ok {
				templatesRenderer.JSON(response, http.StatusOK, route)
				return
			}
		}
	}
	http.NotFound(response, request)
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprint(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprint(w, "\n}\n")
}
