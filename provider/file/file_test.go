package file

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestProvideSingleFileAndWatch(t *testing.T) {
	tempDir := createTempDir(t, "testfile")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2
	expectedNumTLSConf := 2

	tempFile := createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLS(expectedNumTLSConf))

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends, &expectedNumTLSConf)

	provide(configurationChan, watch, withFile(tempFile))

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
		createTLS(expectedNumTLSConf))

	err = waitForSignal(signal, 2*time.Second, "single frontend, backend, TLS configuration")
	assert.NoError(t, err)
}

func TestProvideSingleFileAndNotWatch(t *testing.T) {
	tempDir := createTempDir(t, "testfile")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2
	expectedNumTLSConf := 2

	tempFile := createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends),
		createTLS(expectedNumTLSConf))

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends, &expectedNumTLSConf)

	provide(configurationChan, withFile(tempFile))

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
		createTLS(expectedNumTLSConf))

	// Must fail because we don't watch the changes
	err = waitForSignal(signal, 2*time.Second, "single frontend, backend and TLS configuration")
	assert.Error(t, err)
}

func TestProvideDirectoryAndWatch(t *testing.T) {
	tempDir := createTempDir(t, "testdir")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2
	expectedNumTLSConf := 2

	tempFile1 := createRandomFile(t, tempDir, createFrontendConfiguration(expectedNumFrontends))
	tempFile2 := createRandomFile(t, tempDir, createBackendConfiguration(expectedNumBackends))
	tempFile3 := createRandomFile(t, tempDir, createTLS(expectedNumTLSConf))

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends, &expectedNumTLSConf)

	provide(configurationChan, watch, withDirectory(tempDir))

	// Wait for initial config message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now remove the backends file
	expectedNumFrontends = 2
	expectedNumBackends = 0
	expectedNumTLSConf = 2
	os.Remove(tempFile2.Name())
	err = waitForSignal(signal, 2*time.Second, "remove the backends file")
	assert.NoError(t, err)

	// Now remove the frontends file
	expectedNumFrontends = 0
	expectedNumBackends = 0
	expectedNumTLSConf = 2
	os.Remove(tempFile1.Name())
	err = waitForSignal(signal, 2*time.Second, "remove the frontends file")
	assert.NoError(t, err)

	// Now remove the TLS configuration file
	expectedNumFrontends = 0
	expectedNumBackends = 0
	expectedNumTLSConf = 0
	os.Remove(tempFile3.Name())
	err = waitForSignal(signal, 2*time.Second, "remove the TLS configuration file")
	assert.NoError(t, err)
}

func TestProvideDirectoryAndNotWatch(t *testing.T) {
	tempDir := createTempDir(t, "testdir")
	tempTLSDir := createSubDir(t, tempDir, "tls")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2
	expectedNumTLSConf := 2

	createRandomFile(t, tempDir, createFrontendConfiguration(expectedNumFrontends))
	tempFile2 := createRandomFile(t, tempDir, createBackendConfiguration(expectedNumBackends))
	createRandomFile(t, tempTLSDir, createTLS(expectedNumTLSConf))

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends, &expectedNumTLSConf)

	provide(configurationChan, withDirectory(tempDir))

	// Wait for initial config message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now remove the backends file
	expectedNumFrontends = 2
	expectedNumBackends = 0
	expectedNumTLSConf = 2
	os.Remove(tempFile2.Name())

	// Must fail because we don't watch the changes
	err = waitForSignal(signal, 2*time.Second, "remove the backends file")
	assert.Error(t, err)

}

func createConfigurationRoutine(t *testing.T, expectedNumFrontends *int, expectedNumBackends *int, expectedNumTLSes *int) (chan types.ConfigMessage, chan interface{}) {
	configurationChan := make(chan types.ConfigMessage)
	signal := make(chan interface{})

	safe.Go(func() {
		for {
			data := <-configurationChan
			assert.Equal(t, "file", data.ProviderName)
			assert.Len(t, data.Configuration.Frontends, *expectedNumFrontends)
			assert.Len(t, data.Configuration.Backends, *expectedNumBackends)
			assert.Len(t, data.Configuration.TLS, *expectedNumTLSes)
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

func provide(configurationChan chan types.ConfigMessage, builders ...func(p *Provider)) {
	pvd := &Provider{}

	for _, builder := range builders {
		builder(pvd)
	}

	pvd.Provide(configurationChan, safe.NewPool(context.Background()), nil)
}

func watch(pvd *Provider) {
	pvd.Watch = true
}

func withDirectory(name string) func(*Provider) {
	return func(pvd *Provider) {
		pvd.Directory = name
	}
}

func withFile(tempFile *os.File) func(*Provider) {
	return func(p *Provider) {
		p.Filename = tempFile.Name()
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

// createDir Helper
func createSubDir(t *testing.T, rootDir, dir string) string {
	t.Helper()
	err := os.Mkdir(rootDir+"/"+dir, 0775)
	if err != nil {
		t.Fatal(err)
	}
	return rootDir + "/" + dir
}

// createFrontendConfiguration Helper
func createFrontendConfiguration(n int) string {
	conf := "{{$home := env \"HOME\"}}\n[frontends]\n"
	for i := 1; i <= n; i++ {
		conf += fmt.Sprintf(`  [frontends."frontend%[1]d"]
  backend = "backend%[1]d"
`, i)
		conf += fmt.Sprintf(`    [frontends."frontend%[1]d".headers]
    "PublicKey" = "{{$home}}/pub.key"
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

// createTLS Helper
func createTLS(n int) string {
	var conf string
	for i := 1; i <= n; i++ {
		conf += fmt.Sprintf(`[[TLS]]
	EntryPoints = ["https"]
	[TLS.Certificate]
	CertFile = "integration/fixtures/https/snitest%[1]d.com.cert"
	KeyFile = "integration/fixtures/https/snitest%[1]d.com.key"
`, i)
	}
	return conf
}
