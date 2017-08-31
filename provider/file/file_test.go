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

	tempFile := createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends))

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends)

	provide(configurationChan, watch, withFile(tempFile))

	// Wait for initial message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now test again with single frontend and backend
	expectedNumFrontends = 1
	expectedNumBackends = 1

	createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends))

	// Must fail because we don't watch the change
	err = waitForSignal(signal, 2*time.Second, "single frontend and backend")
	assert.NoError(t, err)
}

func TestProvideSingleFileAndNotWatch(t *testing.T) {
	tempDir := createTempDir(t, "testfile")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2

	tempFile := createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends))

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends)

	provide(configurationChan, withFile(tempFile))

	// Wait for initial message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now test again with single frontend and backend
	expectedNumFrontends = 1
	expectedNumBackends = 1

	createFile(t,
		tempDir, "simple.toml",
		createFrontendConfiguration(expectedNumFrontends),
		createBackendConfiguration(expectedNumBackends))

	// Must fail because we don't watch the changes
	err = waitForSignal(signal, 2*time.Second, "single frontend and backend")
	assert.Error(t, err)
}

func TestProvideDirectoryAndWatch(t *testing.T) {
	tempDir := createTempDir(t, "testdir")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2

	tempFile1 := createRandomFile(t, tempDir, createFrontendConfiguration(expectedNumFrontends))
	tempFile2 := createRandomFile(t, tempDir, createBackendConfiguration(expectedNumBackends))

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends)

	provide(configurationChan, watch, withDirectory(tempDir))

	// Wait for initial config message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now remove the backends file
	expectedNumFrontends = 2
	expectedNumBackends = 0
	os.Remove(tempFile2.Name())
	err = waitForSignal(signal, 2*time.Second, "remove the backends file")
	assert.NoError(t, err)

	// Now remove the frontends file
	expectedNumFrontends = 0
	expectedNumBackends = 0
	os.Remove(tempFile1.Name())
	err = waitForSignal(signal, 2*time.Second, "remove the frontends file")
	assert.NoError(t, err)
}

func TestProvideDirectoryAndNotWatch(t *testing.T) {
	tempDir := createTempDir(t, "testdir")
	defer os.RemoveAll(tempDir)

	expectedNumFrontends := 2
	expectedNumBackends := 2

	createRandomFile(t, tempDir, createFrontendConfiguration(expectedNumFrontends))
	tempFile2 := createRandomFile(t, tempDir, createBackendConfiguration(expectedNumBackends))

	configurationChan, signal := createConfigurationRoutine(t, &expectedNumFrontends, &expectedNumBackends)

	provide(configurationChan, withDirectory(tempDir))

	// Wait for initial config message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now remove the backends file
	expectedNumFrontends = 2
	expectedNumBackends = 0
	os.Remove(tempFile2.Name())

	// Must fail because we don't watch the changes
	err = waitForSignal(signal, 2*time.Second, "remove the backends file")
	assert.Error(t, err)

}

func createConfigurationRoutine(t *testing.T, expectedNumFrontends *int, expectedNumBackends *int) (chan types.ConfigMessage, chan interface{}) {
	configurationChan := make(chan types.ConfigMessage)
	signal := make(chan interface{})

	safe.Go(func() {
		for {
			data := <-configurationChan
			assert.Equal(t, "file", data.ProviderName)
			assert.Len(t, data.Configuration.Frontends, *expectedNumFrontends)
			assert.Len(t, data.Configuration.Backends, *expectedNumBackends)
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
