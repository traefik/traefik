package server

import (
	"context"
	stdtls "crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
	ingressnginx "github.com/traefik/traefik/v3/pkg/provider/kubernetes/ingress-nginx"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	"github.com/traefik/traefik/v3/pkg/server/service"
	"github.com/traefik/traefik/v3/pkg/tcp"
	th "github.com/traefik/traefik/v3/pkg/testhelpers"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/tls/generate"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestIngressNGINXCoalescedHTTPS(t *testing.T) {
	cert, key, err := generate.KeyPair("*.example.com", time.Time{})
	require.NoError(t, err)

	objects := k8s.MustParseYaml([]byte(`
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: nginx
spec:
  controller: k8s.io/ingress-nginx
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: default
spec:
  type: ExternalName
  externalName: backend.example.com
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: shared
  namespace: default
spec:
  ingressClassName: nginx
  tls:
    - hosts: [api.example.com, cdn.example.com]
      secretName: shared
  rules:
    - host: api.example.com
      http: {paths: [{path: /, pathType: Prefix, backend: {service: {name: backend, port: {number: 80}}}}]}
    - host: cdn.example.com
      http: {paths: [{path: /, pathType: Prefix, backend: {service: {name: backend, port: {number: 80}}}}]}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: client-auth
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/auth-tls-secret: default/shared
spec:
  ingressClassName: nginx
  rules:
    - host: private.example.com
      http: {paths: [{path: /, pathType: Prefix, backend: {service: {name: backend, port: {number: 80}}}}]}
`))
	objects = append(objects, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "shared"},
		Data:       map[string][]byte{corev1.TLSCertKey: cert, corev1.TLSPrivateKeyKey: key, corev1.ServiceAccountRootCAKey: cert},
	})
	kubeClient := ingressnginx.NewClient(kubefake.NewClientset(objects...))
	p := &ingressnginx.Provider{HTTPSEntryPoint: "websecure"}
	p.SetDefaults()
	require.NoError(t, p.InitWithClient(kubeClient))
	_, err = kubeClient.WatchAll(t.Context(), "", "")
	require.NoError(t, err)

	conf := Aggregate(dynamic.Configurations{ingressnginx.ProviderName: p.BuildConfiguration(t.Context())}, []string{"websecure"}, false)

	epConfig := &static.EntryPoint{Address: "127.0.0.1:0"}
	epConfig.SetDefaults()
	staticConfig := static.Configuration{EntryPoints: static.EntryPoints{"websecure": epConfig}}
	tlsManager := tls.NewManager(nil)
	tlsManager.UpdateConfigs(t.Context(), conf.TLS.Stores, conf.TLS.Options, conf.TLS.Certificates)
	transportManager := service.NewTransportManager(nil)
	transportManager.Update(map[string]*dynamic.ServersTransport{"default@internal": {}})
	managerFactory := service.NewManagerFactory(staticConfig, nil, nil, transportManager, proxyBuilderMock{}, nil, tlsManager)
	factory, err := NewRouterFactory(staticConfig, managerFactory, tlsManager, nil, nil, tcp.NewDialerManager(nil))
	require.NoError(t, err)
	routers, _ := factory.CreateRouters(runtime.NewConfig(conf))
	t.Cleanup(factory.cancelPrevState)

	entryPoint, err := NewTCPEntryPoint(t.Context(), "websecure", epConfig, nil, nil)
	require.NoError(t, err)
	t.Cleanup(func() { entryPoint.Shutdown(context.Background()) })
	entryPoint.SwitchRouter(routers["websecure"])
	go entryPoint.Start(t.Context())

	roots := x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM(cert))
	transport := &http.Transport{
		TLSClientConfig:   &stdtls.Config{ServerName: "api.example.com", RootCAs: roots},
		ForceAttemptHTTP2: true,
	}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	for _, test := range []struct {
		host   string
		status int
		reused bool
	}{
		{host: "api.example.com", status: http.StatusOK},
		{host: "cdn.example.com", status: http.StatusOK, reused: true},
		{host: "private.example.com", status: http.StatusMisdirectedRequest, reused: true},
	} {
		t.Run(test.host, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodHead, "https://"+entryPoint.listener.Addr().String()+"/", nil)
			require.NoError(t, err)
			req.Host = test.host
			var conn httptrace.GotConnInfo
			req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
				GotConn: func(info httptrace.GotConnInfo) { conn = info },
			}))

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, 2, resp.ProtoMajor)
			assert.Equal(t, test.reused, conn.Reused)
			assert.Equal(t, test.status, resp.StatusCode)
		})
	}
}

func TestReuseService(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	staticConfig := static.Configuration{
		EntryPoints: map[string]*static.EntryPoint{
			"web": {},
		},
	}

	dynamicConfigs := th.BuildConfiguration(
		th.WithRouters(
			th.WithRouter("foo",
				th.WithEntryPoints("web"),
				th.WithServiceName("bar"),
				th.WithRule("Path(`/ok`)")),
			th.WithRouter("foo2",
				th.WithEntryPoints("web"),
				th.WithRule("Path(`/unauthorized`)"),
				th.WithServiceName("bar"),
				th.WithRouterMiddlewares("basicauth")),
		),
		th.WithMiddlewares(th.WithMiddleware("basicauth",
			th.WithBasicAuth(&dynamic.BasicAuth{Users: []string{"foo:bar"}}),
		)),
		th.WithServices(
			th.WithService("bar", th.WithServiceServersLoadBalancer(th.WithServers(th.WithServer(testServer.URL)))),
		),
	)

	transportManager := service.NewTransportManager(nil)
	transportManager.Update(map[string]*dynamic.ServersTransport{"default@internal": {}})

	managerFactory := service.NewManagerFactory(staticConfig, nil, nil, transportManager, proxyBuilderMock{}, nil, nil)
	tlsManager := tls.NewManager(nil)

	dialerManager := tcp.NewDialerManager(nil)
	dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
	factory, err := NewRouterFactory(staticConfig, managerFactory, tlsManager, nil, nil, dialerManager)
	require.NoError(t, err)

	entryPointsHandlers, _ := factory.CreateRouters(runtime.NewConfig(dynamic.Configuration{HTTP: dynamicConfigs}))

	// Test that the /ok path returns a status 200.
	responseRecorderOk := &httptest.ResponseRecorder{}
	requestOk := httptest.NewRequest(http.MethodGet, testServer.URL+"/ok", nil)
	entryPointsHandlers["web"].GetHTTPHandler().ServeHTTP(responseRecorderOk, requestOk)

	assert.Equal(t, http.StatusOK, responseRecorderOk.Result().StatusCode, "status code")

	// Test that the /unauthorized path returns a 401 because of
	// the basic authentication defined on the frontend.
	responseRecorderUnauthorized := &httptest.ResponseRecorder{}
	requestUnauthorized := httptest.NewRequest(http.MethodGet, testServer.URL+"/unauthorized", nil)
	entryPointsHandlers["web"].GetHTTPHandler().ServeHTTP(responseRecorderUnauthorized, requestUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, responseRecorderUnauthorized.Result().StatusCode, "status code")
}

func TestServerResponseEmptyBackend(t *testing.T) {
	const requestPath = "/path"
	const routeRule = "Path(`" + requestPath + "`)"

	testCases := []struct {
		desc               string
		config             func(testServerURL string) *dynamic.HTTPConfiguration
		expectedStatusCode int
	}{
		{
			desc: "Ok",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("web"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithServices(
						th.WithService("bar", th.WithServiceServersLoadBalancer(th.WithServers(th.WithServer(testServerURL)))),
					),
				)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "No Frontend",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			desc: "Empty Backend LB",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("web"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithServices(
						th.WithService("bar", th.WithServiceServersLoadBalancer()),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB Sticky",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("web"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithServices(
						th.WithService("bar", th.WithServiceServersLoadBalancer(th.WithSticky("test"))),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB Sticky",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("web"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithServices(
						th.WithService("bar", th.WithServiceServersLoadBalancer(th.WithSticky("test"))),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			defer testServer.Close()

			staticConfig := static.Configuration{
				EntryPoints: map[string]*static.EntryPoint{
					"web": {},
				},
			}

			transportManager := service.NewTransportManager(nil)
			transportManager.Update(map[string]*dynamic.ServersTransport{"default@internal": {}})

			managerFactory := service.NewManagerFactory(staticConfig, nil, nil, transportManager, proxyBuilderMock{}, nil, nil)
			tlsManager := tls.NewManager(nil)

			dialerManager := tcp.NewDialerManager(nil)
			dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
			observabiltyMgr := middleware.NewObservabilityMgr(staticConfig, nil, nil, nil, nil, nil)
			factory, err := NewRouterFactory(staticConfig, managerFactory, tlsManager, observabiltyMgr, nil, dialerManager)
			require.NoError(t, err)

			entryPointsHandlers, _ := factory.CreateRouters(runtime.NewConfig(dynamic.Configuration{HTTP: test.config(testServer.URL)}))

			responseRecorder := &httptest.ResponseRecorder{}
			request := httptest.NewRequest(http.MethodGet, testServer.URL+requestPath, nil)

			entryPointsHandlers["web"].GetHTTPHandler().ServeHTTP(responseRecorder, request)

			assert.Equal(t, test.expectedStatusCode, responseRecorder.Result().StatusCode, "status code")
		})
	}
}

func TestInternalServices(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	staticConfig := static.Configuration{
		API: &static.API{},
		EntryPoints: map[string]*static.EntryPoint{
			"web": {},
		},
	}

	dynamicConfigs := th.BuildConfiguration(
		th.WithRouters(
			th.WithRouter("foo",
				th.WithEntryPoints("web"),
				th.WithServiceName("api@internal"),
				th.WithRule("PathPrefix(`/api`)")),
		),
	)

	transportManager := service.NewTransportManager(nil)
	transportManager.Update(map[string]*dynamic.ServersTransport{"default@internal": {}})

	managerFactory := service.NewManagerFactory(staticConfig, nil, nil, transportManager, nil, nil, nil)
	tlsManager := tls.NewManager(nil)

	dialerManager := tcp.NewDialerManager(nil)
	dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
	factory, err := NewRouterFactory(staticConfig, managerFactory, tlsManager, nil, nil, dialerManager)
	require.NoError(t, err)

	entryPointsHandlers, _ := factory.CreateRouters(runtime.NewConfig(dynamic.Configuration{HTTP: dynamicConfigs}))

	// Test that the /ok path returns a status 200.
	responseRecorderOk := &httptest.ResponseRecorder{}
	requestOk := httptest.NewRequest(http.MethodGet, testServer.URL+"/api/rawdata", nil)
	entryPointsHandlers["web"].GetHTTPHandler().ServeHTTP(responseRecorderOk, requestOk)

	assert.Equal(t, http.StatusOK, responseRecorderOk.Result().StatusCode, "status code")
}

func TestRecursionService(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	staticConfig := static.Configuration{
		EntryPoints: map[string]*static.EntryPoint{
			"web": {},
		},
	}

	dynamicConfigs := th.BuildConfiguration(
		th.WithRouters(
			th.WithRouter("foo@provider1",
				th.WithEntryPoints("web"),
				th.WithServiceName("bar"),
				th.WithRule("Path(`/ok`)")),
		),
		th.WithMiddlewares(th.WithMiddleware("customerror",
			th.WithErrorPage(&dynamic.ErrorPage{Service: "bar"}),
		)),
		th.WithServices(
			th.WithService("bar@provider1", th.WithServiceWRR(th.WithWRRServices(th.WithWRRService("foo")))),
			th.WithService("foo@provider1", th.WithServiceWRR(th.WithWRRServices(th.WithWRRService("bar")))),
		),
	)

	transportManager := service.NewTransportManager(nil)
	transportManager.Update(map[string]*dynamic.ServersTransport{"default@internal": {}})

	tlsManager := tls.NewManager(nil)
	managerFactory := service.NewManagerFactory(staticConfig, nil, nil, transportManager, nil, nil, tlsManager)

	dialerManager := tcp.NewDialerManager(nil)
	dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
	factory, err := NewRouterFactory(staticConfig, managerFactory, tlsManager, nil, nil, dialerManager)
	require.NoError(t, err)

	rtConf := runtime.NewConfig(dynamic.Configuration{HTTP: dynamicConfigs})
	entryPointsHandlers, _ := factory.CreateRouters(rtConf)

	// Test that the /ok path returns a status 404.
	responseRecorderOk := &httptest.ResponseRecorder{}
	requestOk := httptest.NewRequest(http.MethodGet, testServer.URL+"/ok", nil)
	entryPointsHandlers["web"].GetHTTPHandler().ServeHTTP(responseRecorderOk, requestOk)

	assert.Equal(t, http.StatusNotFound, responseRecorderOk.Result().StatusCode, "status code")

	require.NotNil(t, rtConf.Routers["foo@provider1"])
	assert.Contains(t, rtConf.Routers["foo@provider1"].Err, "building HTTP service \"foo\": building HTTP service \"bar\": could not instantiate service bar@provider1: recursion detected in service:bar@provider1->service:foo@provider1->service:bar@provider1")
	require.NotNil(t, rtConf.Services["bar@provider1"])
	assert.Contains(t, rtConf.Services["bar@provider1"].Err, "could not instantiate service bar@provider1: recursion detected in service:bar@provider1->service:foo@provider1->service:bar@provider1")
}

type proxyBuilderMock struct{}

func (p proxyBuilderMock) Build(_ string, _ *url.URL, _, _ bool, _ time.Duration) (http.Handler, error) {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {}), nil
}

func (p proxyBuilderMock) Update(_ map[string]*dynamic.ServersTransport) {
	panic("implement me")
}
