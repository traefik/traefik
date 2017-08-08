package web

import (
	"encoding/json"
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

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
	"github.com/urfave/negroni"
)

// Provider is a provider.Provider implementation that provides the UI
type Provider struct {
	Address               string            `description:"Web administration port"`
	CertFile              string            `description:"SSL certificate"`
	KeyFile               string            `description:"SSL certificate"`
	ReadOnly              bool              `description:"Enable read only API"`
	Statistics            *types.Statistics `description:"Enable more detailed statistics"`
	Metrics               *types.Metrics    `description:"Enable a metrics exporter"`
	Path                  string            `description:"Root path for dashboard and API"`
	Auth                  *types.Auth
	Debug                 bool
	CurrentConfigurations *safe.Safe
	Stats                 *thoas_stats.Stats
	StatsRecorder         *middlewares.StatsRecorder
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
func (provider *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, _ types.Constraints) error {

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
	systemRouter.Methods("GET", "HEAD").Path(provider.Path + "ping").HandlerFunc(provider.getPingHandler)
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
	systemRouter.Methods("GET").PathPrefix(provider.Path + "dashboard/").
		Handler(http.StripPrefix(provider.Path+"dashboard/", http.FileServer(&assetfs.AssetFS{Asset: autogen.Asset, AssetInfo: autogen.AssetInfo, AssetDir: autogen.AssetDir, Prefix: "static"})))

	// expvars
	if provider.Debug {
		systemRouter.Methods("GET").Path(provider.Path + "debug/vars").HandlerFunc(expVarHandler)
	}

	safe.Go(func() {
		var err error
		var negroniInstance = negroni.New()
		if provider.Auth != nil {
			authMiddleware, err := middlewares.NewAuthenticator(provider.Auth)
			if err != nil {
				log.Fatal("Error creating Auth: ", err)
			}
			authMiddlewareWrapper := negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				if r.URL.Path == "/ping" {
					next.ServeHTTP(w, r)
				} else {
					authMiddleware.ServeHTTP(w, r, next)
				}
			})
			negroniInstance.Use(authMiddlewareWrapper)
		}
		negroniInstance.UseHandler(systemRouter)

		if len(provider.CertFile) > 0 && len(provider.KeyFile) > 0 {
			err = http.ListenAndServeTLS(provider.Address, provider.CertFile, provider.KeyFile, negroniInstance)
		} else {
			err = http.ListenAndServe(provider.Address, negroniInstance)
		}

		if err != nil {
			log.Fatal("Error creating server: ", err)
		}
	})
	return nil
}

// healthResponse combines data returned by thoas/stats with statistics (if
// they are enabled).
type healthResponse struct {
	*thoas_stats.Data
	*middlewares.Stats
}

func (provider *Provider) getHealthHandler(response http.ResponseWriter, request *http.Request) {
	health := &healthResponse{Data: provider.Stats.Data()}
	if provider.StatsRecorder != nil {
		health.Stats = provider.StatsRecorder.Data()
	}
	templatesRenderer.JSON(response, http.StatusOK, health)
}

func (provider *Provider) getPingHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprint(response, "OK")
}

func (provider *Provider) getConfigHandler(response http.ResponseWriter, request *http.Request) {
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
	templatesRenderer.JSON(response, http.StatusOK, currentConfigurations)
}

func (provider *Provider) getVersionHandler(response http.ResponseWriter, request *http.Request) {
	v := struct {
		Version  string
		Codename string
	}{
		Version:  version.Version,
		Codename: version.Codename,
	}
	templatesRenderer.JSON(response, http.StatusOK, v)
}

func (provider *Provider) getProviderHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *Provider) getBackendsHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider.Backends)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *Provider) getBackendHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, backend)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *Provider) getServersHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, backend.Servers)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *Provider) getServerHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	serverID := vars["server"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
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

func (provider *Provider) getFrontendsHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider.Frontends)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *Provider) getFrontendHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, frontend)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *Provider) getRoutesHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, frontend.Routes)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *Provider) getRouteHandler(response http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	routeID := vars["route"]
	currentConfigurations := provider.CurrentConfigurations.Get().(types.Configurations)
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

func expVarHandler(w http.ResponseWriter, _ *http.Request) {
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
