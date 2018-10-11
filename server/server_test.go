package server

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/safe"
	th "github.com/containous/traefik/testhelpers"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unrolled/secure"
)

// The below certs and keys are PEM-encoded TLS certs with SAN IPs
// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
// generated from src/crypto/tls:
// go run /usr/share/go-1.10/src/crypto/tls/generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var (
	customCert1 = []byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQVAnjKvZDP3mAxWyoeTYgtTANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDrCy4w0TcQWJoUIutMFZfeGbe96iNEF8W2q2DRRnWDnLx/aiokfEodjPIK
aQPK1zAEjvq2Ok3E3oDHcOnbCKmQOvz6f8n/irnw8hUjRe8YZxD/5eJqNW4v0Ymo
31NluFzaOY0F3jULHbhRog3l97XVEc/FS5Vi7RSVeZzDtKWlAQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQATAfWcAwivjRu+2BL+znyuJsy6SFaYv+hq
09hmw7or9nGWGG9KyondB2hbekIhJB0Q1q54ea6gTioyQKXrvIiWzMuKARryvaKX
U3rS0Val2xWw+u0Pusg6repvwx596sg6YFonqjDmCx33pIBQVs/IXiTmRiCh/IjI
WVqV1GyORA==
-----END CERTIFICATE-----`)

	customKey1 = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDrCy4w0TcQWJoUIutMFZfeGbe96iNEF8W2q2DRRnWDnLx/aiok
fEodjPIKaQPK1zAEjvq2Ok3E3oDHcOnbCKmQOvz6f8n/irnw8hUjRe8YZxD/5eJq
NW4v0Ymo31NluFzaOY0F3jULHbhRog3l97XVEc/FS5Vi7RSVeZzDtKWlAQIDAQAB
AoGAAi8jab639UXtgJxmdVmKBL1WcMRZOYvDAZSMHMW719JACisRYy9ofOfPY/tf
1qWzQ4eUmtbl3Bt5NOE+uxNUiAcETU1S9Qa4YdovKejfoRGZ0IfPEEAKtxSUtpjB
QHilJhiq86sKQU9AlJkLWelw7I776ZhEr3otSbs8EAKPDtUCQQDrxeJ8DvI+X7fs
OstTp8U68FYv6OfdtZyWVZDoaV5RfYXrvJk7cvx2VZ8T5T2h9wRk9BNZYEnbuWFm
JEgYz+YrAkEA/zVHQJL4anJc75EpRUicN/JeNEDyHU4sHkkJebXPHeOkP+wG6Wk2
y5U0Ly24ZV4h1/UYBJXjUcUsSjRIED8XgwJBANWAIjGZDz/wSYq/SvP8DpvqmwFT
dPPNy3hPD6OGFwTQF/96j3/IBlnZ+u13PzJ1jyMj6omaqgcwfcSSwj7FtHUCQQC+
0a1nAP0xSjVnAxjirvnvcw8w7uaZNtwSAPZOxLwKUy16hhZc68iGzBbqt7rKQGn5
uU6uDwybFVyaVyES1LnVAkBuOzGVxAbqvp4LbWUYKju6J9u3BDSzDF1vS85fU1ym
xGa/980Gj1ps+4/FkHGe6pXHHs902Xpxfpyv9hKFqNHx
-----END RSA PRIVATE KEY-----`)

	customCert2 = []byte(`-----BEGIN CERTIFICATE-----
MIICFDCCAX2gAwIBAgIRAI1CA9k6ki63x5MkOdTnPbkwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAw
gYkCgYEAw5OhEEQwOp5vYU+JgcaxPPkbzKwVakczQOFlcrPDeZ3/EUU8H8rn78Th
mR5q0C2uYThf7gPi3cK+b4rdLemJTXZhKkfFyeuNAmsSFJ4/+gHA+YRGgSUFg58V
C1hVsH2qnmeOE+33z0Tri1iCcE6XtO1ciadfnyqVui98UXFNspUCAwEAAaNoMGYw
DgYDVR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQF
MAMBAf8wLgYDVR0RBCcwJYILZXhhbXBsZS5jb22HBH8AAAGHEAAAAAAAAAAAAAAA
AAAAAAEwDQYJKoZIhvcNAQELBQADgYEAF2b1CT8rVFqthWEk5pqJc7w59FNTT9gU
d3bHy8ec7TWSoyElD2ijdnc6G2J3TE52Ls7iMNxRu0YhEDgwQNH0VbIRluUFmEWz
eEfTrsjUysVgjLQwOjSeyDCJVOe6tAnuJgZ55hyGyzsCTWXjnxXG2fUwRKeq78AN
qLOHyejIgWY=
-----END CERTIFICATE-----`)

	customKey2 = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDDk6EQRDA6nm9hT4mBxrE8+RvMrBVqRzNA4WVys8N5nf8RRTwf
yufvxOGZHmrQLa5hOF/uA+Ldwr5vit0t6YlNdmEqR8XJ640CaxIUnj/6AcD5hEaB
JQWDnxULWFWwfaqeZ44T7ffPROuLWIJwTpe07VyJp1+fKpW6L3xRcU2ylQIDAQAB
AoGAdfmexb4sTZ/21f9xliwyC/LE1zDS9joe67tLQ+a2Oq2ZCGT4QMFYKaVc5M2Z
Zxy3PQQRsfT8LANmdsiQZTqjzFuR+YSK3Kj2sNSTmQFI0XJlDYxR80SCT0Sbkn7U
YNb0EKWVP927bEPOjNSlXbiGZeCQDV0YVZh8bySU5S01XR0CQQDG59KH0T/T+aLW
bwBJx0KVpBovAdHwqnP58wzMSdfPUjYTo5pDqfa7ywDQIUPXB29IOaobm9EglMP5
yl9VBwrDAkEA+7cw0K3jqEOnTyNZ0gc1F8wVSQ79yhmsAw/8KVOXfS7lIL1qRDHl
rFRmbhMMth37ReS9hAWhn8PLrYnw58MHxwJBALd6hQPwC/bXolQ31IY6Hru2wsh1
31kngxAgGcAgpciCx4taMSUVlZopariS1ud13js7piUNmN17HURAX6wpcM0CQEJW
rU7SBUW7TsTUlD9+FsgGyTVP9iLlUSgddl+N4EblrQ1L3k3KuLUKKVSpQJhennJ1
Ll00/ruUZoF98TejdtECQEym1q162ja5lIcbjrtKBBXGXRgzXhngsp8+y3yh8vNT
dT1sfx9fkl414bNDIP6Ohz4TGRn+Xvnouk5Urj6W1nk=
-----END RSA PRIVATE KEY-----`)
)

func TestPrepareServerTimeouts(t *testing.T) {
	testCases := []struct {
		desc                 string
		globalConfig         configuration.GlobalConfiguration
		expectedIdleTimeout  time.Duration
		expectedReadTimeout  time.Duration
		expectedWriteTimeout time.Duration
	}{
		{
			desc: "full configuration",
			globalConfig: configuration.GlobalConfiguration{
				RespondingTimeouts: &configuration.RespondingTimeouts{
					IdleTimeout:  flaeg.Duration(10 * time.Second),
					ReadTimeout:  flaeg.Duration(12 * time.Second),
					WriteTimeout: flaeg.Duration(14 * time.Second),
				},
			},
			expectedIdleTimeout:  time.Duration(10 * time.Second),
			expectedReadTimeout:  time.Duration(12 * time.Second),
			expectedWriteTimeout: time.Duration(14 * time.Second),
		},
		{
			desc:                 "using defaults",
			globalConfig:         configuration.GlobalConfiguration{},
			expectedIdleTimeout:  time.Duration(180 * time.Second),
			expectedReadTimeout:  time.Duration(0 * time.Second),
			expectedWriteTimeout: time.Duration(0 * time.Second),
		},
		{
			desc: "deprecated IdleTimeout configured",
			globalConfig: configuration.GlobalConfiguration{
				IdleTimeout: flaeg.Duration(45 * time.Second),
			},
			expectedIdleTimeout:  time.Duration(45 * time.Second),
			expectedReadTimeout:  time.Duration(0 * time.Second),
			expectedWriteTimeout: time.Duration(0 * time.Second),
		},
		{
			desc: "deprecated and new IdleTimeout configured",
			globalConfig: configuration.GlobalConfiguration{
				IdleTimeout: flaeg.Duration(45 * time.Second),
				RespondingTimeouts: &configuration.RespondingTimeouts{
					IdleTimeout: flaeg.Duration(80 * time.Second),
				},
			},
			expectedIdleTimeout:  time.Duration(45 * time.Second),
			expectedReadTimeout:  time.Duration(0 * time.Second),
			expectedWriteTimeout: time.Duration(0 * time.Second),
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			entryPointName := "http"
			entryPoint := &configuration.EntryPoint{
				Address:          "localhost:0",
				ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
			}
			router := middlewares.NewHandlerSwitcher(mux.NewRouter())

			srv := NewServer(test.globalConfig, nil, nil)
			httpServer, _, err := srv.prepareServer(entryPointName, entryPoint, router, nil)
			require.NoError(t, err, "Unexpected error when preparing srv")

			assert.Equal(t, test.expectedIdleTimeout, httpServer.IdleTimeout, "IdleTimeout")
			assert.Equal(t, test.expectedReadTimeout, httpServer.ReadTimeout, "ReadTimeout")
			assert.Equal(t, test.expectedWriteTimeout, httpServer.WriteTimeout, "WriteTimeout")
		})
	}
}

func TestListenProvidersSkipsEmptyConfigs(t *testing.T) {
	server, stop, invokeStopChan := setupListenProvider(10 * time.Millisecond)
	defer invokeStopChan()

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-server.configurationValidatedChan:
				t.Error("An empty configuration was published but it should not")
			}
		}
	}()

	server.configurationChan <- types.ConfigMessage{ProviderName: "kubernetes"}

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)
}

func TestListenProvidersSkipsSameConfigurationForProvider(t *testing.T) {
	server, stop, invokeStopChan := setupListenProvider(10 * time.Millisecond)
	defer invokeStopChan()

	publishedConfigCount := 0
	go func() {
		for {
			select {
			case <-stop:
				return
			case config := <-server.configurationValidatedChan:
				// set the current configuration
				// this is usually done in the processing part of the published configuration
				// so we have to emulate the behaviour here
				currentConfigurations := server.currentConfigurations.Get().(types.Configurations)
				currentConfigurations[config.ProviderName] = config.Configuration
				server.currentConfigurations.Set(currentConfigurations)

				publishedConfigCount++
				if publishedConfigCount > 1 {
					t.Error("Same configuration should not be published multiple times")
				}
			}
		}
	}()

	config := th.BuildConfiguration(
		th.WithFrontends(th.WithFrontend("backend")),
		th.WithBackends(th.WithBackendNew("backend")),
	)

	// provide a configuration
	server.configurationChan <- types.ConfigMessage{ProviderName: "kubernetes", Configuration: config}

	// give some time so that the configuration can be processed
	time.Sleep(20 * time.Millisecond)

	// provide the same configuration a second time
	server.configurationChan <- types.ConfigMessage{ProviderName: "kubernetes", Configuration: config}

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)
}

func TestListenProvidersPublishWhenTLSCertChange(t *testing.T) {
	server, stop, invokeStopChan := setupListenProvider(10 * time.Millisecond)
	defer invokeStopChan()

	// Temp cert files
	certFile, err := ioutil.TempFile("", "testcert.crt")
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Unlink(certFile.Name())
	keyFile, err := ioutil.TempFile("", "testkey.key")
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Unlink(keyFile.Name())

	publishedProviderConfigCount := map[string]int{}
	publishedConfigCount := 0
	consumeFirstConfigsDone := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			case config := <-server.configurationValidatedChan:
				// set the current configuration
				// this is usually done in the processing part of the published configuration
				// so we have to emulate the behavior here
				currentConfigurations := server.currentConfigurations.Get().(types.Configurations)
				currentConfigurations[config.ProviderName] = config.Configuration
				server.currentConfigurations.Set(currentConfigurations)

				// Load certs into conf
				cert, _ := tls.LoadX509KeyPair(certFile.Name(), keyFile.Name())
				certMap := server.buildNameOrIPToCertificate([]tls.Certificate{cert})
				certs := &traefiktls.CertificateStore{
					DynamicCerts: &safe.Safe{},
				}
				certs.DynamicCerts.Set(certMap)
				entrypoint := &serverEntryPoint{
					certs: certs,
				}
				server.serverEntryPoints["https"] = entrypoint

				publishedProviderConfigCount[config.ProviderName]++
				publishedConfigCount++

				if publishedConfigCount == 1 {
					consumeFirstConfigsDone <- true
				}
				if publishedConfigCount > 2 {
					t.Error("Same configuration should not be published multiple times")
				}
			}
		}
	}()

	// Write self signed certs
	ioutil.WriteFile(certFile.Name(), customCert1, 0644)
	ioutil.WriteFile(keyFile.Name(), customKey1, 0644)

	msgConfig := &types.Configuration{
		Frontends: map[string]*types.Frontend{},
		Backends:  map[string]*types.Backend{},
		TLS: []*traefiktls.Configuration{
			{
				Certificate: &traefiktls.Certificate{
					CertFile: traefiktls.FileOrContent(certFile.Name()),
					KeyFile:  traefiktls.FileOrContent(keyFile.Name()),
				},
				EntryPoints: []string{"https"},
			},
		},
	}

	server.configurationChan <- types.ConfigMessage{ProviderName: "file", Configuration: msgConfig}
	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)

	select {
	case <-consumeFirstConfigsDone:
		if val := publishedProviderConfigCount["file"]; val == 1 {
			break
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Published configurations were not consumed in time")
	}

	//  provide the same configuration more times, that should not get published
	server.configurationChan <- types.ConfigMessage{ProviderName: "file", Configuration: msgConfig}
	server.configurationChan <- types.ConfigMessage{ProviderName: "file", Configuration: msgConfig}
	// give some time so that the configuration can be processed
	time.Sleep(20 * time.Millisecond)

	// Alter cert and key and  provide the same configuration a third time
	ioutil.WriteFile(certFile.Name(), customCert2, 0644)
	ioutil.WriteFile(keyFile.Name(), customKey2, 0644)
	server.configurationChan <- types.ConfigMessage{ProviderName: "file", Configuration: msgConfig}

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)

	if publishedConfigCount < 2 {
		t.Error("TLS change not published")
	}
}

func TestListenProvidersPublishesConfigForEachProvider(t *testing.T) {
	server, stop, invokeStopChan := setupListenProvider(10 * time.Millisecond)
	defer invokeStopChan()

	publishedProviderConfigCount := map[string]int{}
	publishedConfigCount := 0
	consumePublishedConfigsDone := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			case newConfig := <-server.configurationValidatedChan:
				publishedProviderConfigCount[newConfig.ProviderName]++
				publishedConfigCount++
				if publishedConfigCount == 2 {
					consumePublishedConfigsDone <- true
					return
				}
			}
		}
	}()

	config := th.BuildConfiguration(
		th.WithFrontends(th.WithFrontend("backend")),
		th.WithBackends(th.WithBackendNew("backend")),
	)
	server.configurationChan <- types.ConfigMessage{ProviderName: "kubernetes", Configuration: config}
	server.configurationChan <- types.ConfigMessage{ProviderName: "marathon", Configuration: config}

	select {
	case <-consumePublishedConfigsDone:
		if val := publishedProviderConfigCount["kubernetes"]; val != 1 {
			t.Errorf("Got %d configuration publication(s) for provider %q, want 1", val, "kubernetes")
		}
		if val := publishedProviderConfigCount["marathon"]; val != 1 {
			t.Errorf("Got %d configuration publication(s) for provider %q, want 1", val, "marathon")
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Published configurations were not consumed in time")
	}
}

// setupListenProvider configures the Server and starts listenProviders
func setupListenProvider(throttleDuration time.Duration) (server *Server, stop chan bool, invokeStopChan func()) {
	stop = make(chan bool)
	invokeStopChan = func() {
		stop <- true
	}

	globalConfig := configuration.GlobalConfiguration{
		EntryPoints: configuration.EntryPoints{
			"http": &configuration.EntryPoint{},
		},
		ProvidersThrottleDuration: flaeg.Duration(throttleDuration),
	}

	server = NewServer(globalConfig, nil, nil)
	go server.listenProviders(stop)

	return server, stop, invokeStopChan
}

func TestServerResponseEmptyBackend(t *testing.T) {
	const requestPath = "/path"
	const routeRule = "Path:" + requestPath

	testCases := []struct {
		desc               string
		config             func(testServerURL string) *types.Configuration
		expectedStatusCode int
	}{
		{
			desc: "Ok",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("wrr"),
						th.WithServersNew(th.WithServerNew(testServerURL))),
					),
				)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "No Frontend",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			desc: "Empty Backend LB-Drr",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("drr")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Drr Sticky",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("drr"), th.WithLBSticky("test")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("wrr")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr Sticky",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("wrr"), th.WithLBSticky("test")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			defer testServer.Close()

			globalConfig := configuration.GlobalConfiguration{}
			entryPointsConfig := map[string]EntryPoint{
				"http": {Configuration: &configuration.EntryPoint{ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true}}},
			}
			dynamicConfigs := types.Configurations{"config": test.config(testServer.URL)}

			srv := NewServer(globalConfig, nil, entryPointsConfig)
			entryPoints := srv.loadConfig(dynamicConfigs, globalConfig)

			responseRecorder := &httptest.ResponseRecorder{}
			request := httptest.NewRequest(http.MethodGet, testServer.URL+requestPath, nil)

			entryPoints["http"].httpRouter.ServeHTTP(responseRecorder, request)

			assert.Equal(t, test.expectedStatusCode, responseRecorder.Result().StatusCode, "status code")
		})
	}
}

type mockContext struct {
	headers http.Header
}

func (c mockContext) Deadline() (deadline time.Time, ok bool) {
	return deadline, ok
}

func (c mockContext) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (c mockContext) Err() error {
	return context.DeadlineExceeded
}

func (c mockContext) Value(key interface{}) interface{} {
	return c.headers
}

func TestNewServerWithResponseModifiers(t *testing.T) {
	testCases := []struct {
		desc             string
		headerMiddleware *middlewares.HeaderStruct
		secureMiddleware *secure.Secure
		ctx              context.Context
		expected         map[string]string
	}{
		{
			desc:             "header and secure nil",
			headerMiddleware: nil,
			secureMiddleware: nil,
			ctx:              mockContext{},
			expected: map[string]string{
				"X-Default":       "powpow",
				"Referrer-Policy": "same-origin",
			},
		},
		{
			desc: "header middleware not nil",
			headerMiddleware: middlewares.NewHeaderFromStruct(&types.Headers{
				CustomResponseHeaders: map[string]string{
					"X-Default": "powpow",
				},
			}),
			secureMiddleware: nil,
			ctx:              mockContext{},
			expected: map[string]string{
				"X-Default":       "powpow",
				"Referrer-Policy": "same-origin",
			},
		},
		{
			desc:             "secure middleware not nil",
			headerMiddleware: nil,
			secureMiddleware: middlewares.NewSecure(&types.Headers{
				ReferrerPolicy: "no-referrer",
			}),
			ctx: mockContext{
				headers: http.Header{"Referrer-Policy": []string{"no-referrer"}},
			},
			expected: map[string]string{
				"X-Default":       "powpow",
				"Referrer-Policy": "no-referrer",
			},
		},
		{
			desc: "header and secure middleware not nil",
			headerMiddleware: middlewares.NewHeaderFromStruct(&types.Headers{
				CustomResponseHeaders: map[string]string{
					"Referrer-Policy": "powpow",
				},
			}),
			secureMiddleware: middlewares.NewSecure(&types.Headers{
				ReferrerPolicy: "no-referrer",
			}),
			ctx: mockContext{
				headers: http.Header{"Referrer-Policy": []string{"no-referrer"}},
			},
			expected: map[string]string{
				"X-Default":       "powpow",
				"Referrer-Policy": "powpow",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			headers := make(http.Header)
			headers.Add("X-Default", "powpow")
			headers.Add("Referrer-Policy", "same-origin")

			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

			res := &http.Response{
				Request: req.WithContext(test.ctx),
				Header:  headers,
			}

			responseModifier := buildModifyResponse(test.secureMiddleware, test.headerMiddleware)
			err := responseModifier(res)

			assert.NoError(t, err)
			assert.Equal(t, len(test.expected), len(res.Header))

			for k, v := range test.expected {
				assert.Equal(t, v, res.Header.Get(k))
			}
		})
	}
}
