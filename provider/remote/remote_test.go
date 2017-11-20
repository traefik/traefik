package remote

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
)

func TestProvideAndWatch(t *testing.T) {
	tempDir := createTempDir(t, "testfile")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2
	expectedNumTLSConf := 2

	tempFile := createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLSConfiguration(expectedNumTLSConf))

	server := createHTTPServer(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadFile(tempFile.Name())
		assert.NoError(t, err)
		w.Write(data)
	})
	defer server.Close()

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends, &expectedNumTLSConf)

	pool := provide(configurationChan, watch, withURL(server.URL))
	defer pool.Stop()

	// Wait for initial message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now test again with single frontend and backend
	expectedNumFrontends = 1
	expectedNumBackends = 1
	expectedNumTLSConf = 1

	createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLSConfiguration(expectedNumTLSConf))

	err = waitForSignal(signal, 12*time.Second, "single frontend, backend, TLS configuration")
	assert.NoError(t, err)
}

func TestProvideAndNotWatch(t *testing.T) {
	tempDir := createTempDir(t, "testfile")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2
	expectedNumTLSConf := 2

	tempFile := createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLSConfiguration(expectedNumTLSConf))

	server := createHTTPServer(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadFile(tempFile.Name())
		assert.NoError(t, err)
		w.Write(data)
	})
	defer server.Close()

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends, &expectedNumTLSConf)

	pool := provide(configurationChan, withURL(server.URL))
	defer pool.Stop()

	// Wait for initial message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now test again with single frontend and backend
	expectedNumFrontends = 1
	expectedNumBackends = 1
	expectedNumTLSConf = 1

	createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLSConfiguration(expectedNumTLSConf))

	// Must fail because we don't watch the changes
	err = waitForSignal(signal, 2*time.Second, "single frontend, backend and TLS configuration")
	assert.Error(t, err)
}

func TestProvideAndWatchWithLongPoll(t *testing.T) {
	tempDir := createTempDir(t, "testfile")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2
	expectedNumTLSConf := 2

	tempFile := createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLSConfiguration(expectedNumTLSConf))

	duration := 5 * time.Second
	server := createHTTPServer(func(w http.ResponseWriter, r *http.Request) {
		checkQuery := fmt.Sprintf("wait=%ds", int(duration.Seconds()))
		assert.True(t, strings.HasSuffix(r.RequestURI, checkQuery), "Expected query of %s, RequestURI was: %s", checkQuery, r.RequestURI)
		data, err := ioutil.ReadFile(tempFile.Name())
		assert.NoError(t, err)
		w.Write(data)
	})
	defer server.Close()

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends, &expectedNumTLSConf)

	pool := provide(configurationChan, watch, withURL(server.URL), withLongPollDuration(duration))
	defer pool.Stop()

	// Wait for initial message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now test again with single frontend and backend
	expectedNumFrontends = 1
	expectedNumBackends = 1
	expectedNumTLSConf = 1

	createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLSConfiguration(expectedNumTLSConf))

	err = waitForSignal(signal, 7*time.Second, "single frontend, backend, TLS configuration")
	assert.NoError(t, err)
}

func TestProvideAndWatchWithInterval(t *testing.T) {
	tempDir := createTempDir(t, "testfile")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2
	expectedNumTLSConf := 2

	tempFile := createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLSConfiguration(expectedNumTLSConf))

	duration := 3 * time.Second
	server := createHTTPServer(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadFile(tempFile.Name())
		assert.NoError(t, err)
		w.Write(data)
	})
	defer server.Close()

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends, &expectedNumTLSConf)

	pool := provide(configurationChan, watch, withURL(server.URL), withRepeatInterval(duration))
	defer pool.Stop()

	// Wait for initial message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now test again with single frontend and backend
	expectedNumFrontends = 1
	expectedNumBackends = 1
	expectedNumTLSConf = 1

	createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLSConfiguration(expectedNumTLSConf))

	err = waitForSignal(signal, 5*time.Second, "single frontend, backend, TLS configuration")
	assert.NoError(t, err)
}

func createHTTPServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func createConfigurationRoutine(t *testing.T, expectedNumFrontends *int, expectedNumBackends *int, expectedNumTLSConfigurations *int) (chan types.ConfigMessage, chan interface{}) {
	configurationChan := make(chan types.ConfigMessage)
	signal := make(chan interface{})

	safe.Go(func() {
		for {
			data := <-configurationChan
			assert.Equal(t, "remote", data.ProviderName)
			assert.Len(t, data.Configuration.Frontends, *expectedNumFrontends)
			assert.Len(t, data.Configuration.Backends, *expectedNumBackends)
			assert.Len(t, data.Configuration.TLSConfiguration, *expectedNumTLSConfigurations)
			signal <- nil
		}
	})

	return configurationChan, signal
}

func waitForSignal(signal chan interface{}, timeout time.Duration, caseName string) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-signal:

	case <-timer.C:
		return fmt.Errorf("Timed out waiting for assertions to be tested: %s", caseName)
	}
	return nil
}

func provide(configurationChan chan types.ConfigMessage, builders ...func(p *Provider)) *safe.Pool {
	pvd := &Provider{}

	for _, builder := range builders {
		builder(pvd)
	}

	pool := safe.NewPool(context.Background())
	pvd.Provide(configurationChan, pool, nil)
	return pool
}

func watch(pvd *Provider) {
	pvd.Watch = true
}

func withURL(url string) func(*Provider) {
	return func(pvd *Provider) {
		pvd.URL = url
	}
}

func withRepeatInterval(interval time.Duration) func(*Provider) {
	return func(pvd *Provider) {
		pvd.RepeatInterval = interval.String()
	}
}

func withLongPollDuration(duration time.Duration) func(*Provider) {
	return func(pvd *Provider) {
		pvd.LongPollDuration = duration.String()
	}
}

// createRandomFile Helper
func createRandomFile(t *testing.T, tempDir string, contents ...string) *os.File {
	return createFile(t, tempDir, fmt.Sprintf("temp%d.toml", time.Now().UnixNano()), contents...)
}

// createFile Helper
func createFile(t *testing.T, tempDir string, name string, contents ...string) *os.File {
	t.Helper()
	fileName := path.Join(tempDir, name)

	tempFile, err := os.Create(fileName)
	if err != nil {
		t.Fatal(err)
	}

	for _, content := range contents {
		_, err := tempFile.WriteString(content)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = tempFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	return tempFile
}

// createTempDir Helper
func createTempDir(t *testing.T, dir string) string {
	t.Helper()
	d, err := ioutil.TempDir("", dir)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// createFrontendConfiguration Helper
func createFrontendConfiguration(n int) string {
	conf := "[frontends]\n"
	for i := 1; i <= n; i++ {
		conf += fmt.Sprintf(`  [frontends.frontend%[1]d]
  backend = "backend%[1]d"
`, i)
	}
	return conf
}

// createBackendConfiguration Helper
func createBackendConfiguration(n int) string {
	conf := "[backends]\n"
	for i := 1; i <= n; i++ {
		conf += fmt.Sprintf(`  [backends.backend%[1]d]
    [backends.backend%[1]d.servers.server1]
    url = "http://172.17.0.%[1]d:80"
`, i)
	}
	return conf
}

// createTLSConfiguration Helper
func createTLSConfiguration(n int) string {
	var conf string
	for i := 1; i <= n; i++ {
		conf += fmt.Sprintf(`[[TLSConfiguration]]
	EntryPoints = ["https"]
	[TLSConfiguration.Certificate]
	CertFile = "integration/fixtures/https/snitest%[1]d.com.cert"
	KeyFile = "integration/fixtures/https/snitest%[1]d.com.key"
`, i)
	}
	return conf
}
