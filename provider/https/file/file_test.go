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
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestProvideSingleFileAndWatch(t *testing.T) {
	tempDir := createTempDir(t, "testhttpsfile")
	defer os.RemoveAll(tempDir)

	expectedTLSConfiguration := 2

	tempFile := createFile(t,
		tempDir, "simpleHttps.toml",
		createTLSConfiguration(expectedTLSConfiguration))
	configurationChan, signal := createConfigurationRoutine(t, &expectedTLSConfiguration)

	provide(configurationChan, watch, withFile(tempFile))

	// Wait for initial message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now test again with single frontend and backend
	expectedTLSConfiguration = 1

	createFile(t,
		tempDir, "simpleHttps.toml",
		createTLSConfiguration(expectedTLSConfiguration))
	err = waitForSignal(signal, 2*time.Second, "single TLS configuration")
	assert.NoError(t, err)
}

func TestProvideSingleFileAndNotWatch(t *testing.T) {
	tempDir := createTempDir(t, "testhttpsfile")
	defer os.RemoveAll(tempDir)

	expectedNumConfigurations := 2

	tempFile := createFile(t,
		tempDir, "simpleHttps.toml",
		createTLSConfiguration(expectedNumConfigurations))
	configurationChan, signal := createConfigurationRoutine(t, &expectedNumConfigurations)

	provide(configurationChan, withFile(tempFile))

	// Wait for initial message to be tested
	err := waitForSignal(signal, 2*time.Second, "initial config")
	assert.NoError(t, err)

	// Now test again with single frontend and backend
	expectedNumConfigurations = 1

	createFile(t,
		tempDir, "simpleHttps.toml",
		createTLSConfiguration(expectedNumConfigurations))
	// Must fail because we don't watch the changes
	err = waitForSignal(signal, 2*time.Second, "single TLS configuration")
	assert.Error(t, err)
}

func createConfigurationRoutine(t *testing.T, expectedNumCertificates *int) (chan types.ConfigMessage, chan interface{}) {
	configurationChan := make(chan types.ConfigMessage)
	signal := make(chan interface{})

	safe.Go(func() {
		for {
			data := <-configurationChan
			assert.Equal(t, ProviderHTTPSFile, data.ProviderName)
			tlsConfigMap := make(map[string][]*tls.Certificate)
			for ep, conf := range *data.TLSConfiguration {
				tlsConfigMap[ep] = conf
			}
			assert.Len(t, tlsConfigMap["https"], *expectedNumCertificates)
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
		return fmt.Errorf("timed out waiting for assertions to be tested: %s", caseName)
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

func withFile(tempFile *os.File) func(*Provider) {
	return func(p *Provider) {
		p.ConfigurationFile = tempFile.Name()
	}
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

// createTLSConfiguration Helper
func createTLSConfiguration(n int) string {
	var conf string
	for i := 1; i <= n; i++ {
		conf += fmt.Sprintf(`
		[[Tls]]
		EntryPoints = ["https"]
			[Tls.Certificate]
			CertFile = "integration/fixtures/https/snitest%[1]d.com.cert"
			KeyFile = "integration/fixtures/https/snitest%[1]d.com.key"`, i)
	}
	return conf
}
