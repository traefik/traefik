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
	"github.com/stretchr/testify/require"
)

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
		conf += fmt.Sprintf(`  [frontends."frontend%[1]d"]
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

type ProvideTestCase struct {
	desc                string
	directoryContent    []string
	fileContent         string
	traefikFileContent  string
	expectedNumFrontend int
	expectedNumBackend  int
	expectedNumTLSConf  int
}

func getTestCases() []ProvideTestCase {
	return []ProvideTestCase{
		{
			desc:                "simple file",
			fileContent:         createFrontendConfiguration(2) + createBackendConfiguration(3) + createTLS(4),
			expectedNumFrontend: 2,
			expectedNumBackend:  3,
			expectedNumTLSConf:  4,
		},
		{
			desc:        "simple file and a traefik file",
			fileContent: createFrontendConfiguration(2) + createBackendConfiguration(3) + createTLS(4),
			traefikFileContent: `
			debug=true
`,
			expectedNumFrontend: 2,
			expectedNumBackend:  3,
			expectedNumTLSConf:  4,
		},
		{
			desc: "template file",
			fileContent: `
[frontends]
{{ range $i, $e := until 20 }}
  [frontends.frontend{{ $e }}]
  backend = "backend"  
{{ end }}
`,
			expectedNumFrontend: 20,
		},
		{
			desc: "simple directory",
			directoryContent: []string{
				createFrontendConfiguration(2),
				createBackendConfiguration(3),
				createTLS(4),
			},
			expectedNumFrontend: 2,
			expectedNumBackend:  3,
			expectedNumTLSConf:  4,
		},
		{
			desc: "template in directory",
			directoryContent: []string{
				`
[frontends]
{{ range $i, $e := until 20 }}
  [frontends.frontend{{ $e }}]
  backend = "backend"  
{{ end }}
`,
				`
[backends]
{{ range $i, $e := until 20 }}
  [backends.backend{{ $e }}]
 [backends.backend{{ $e }}.servers.server1]
	url="http://127.0.0.1"
{{ end }}
`,
			},
			expectedNumFrontend: 20,
			expectedNumBackend:  20,
		},
		{
			desc: "simple traefik file",
			traefikFileContent: `
				debug=true
				[file]	
				` + createFrontendConfiguration(2) + createBackendConfiguration(3) + createTLS(4),
			expectedNumFrontend: 2,
			expectedNumBackend:  3,
			expectedNumTLSConf:  4,
		},
		{
			desc: "simple traefik file with templating",
			traefikFileContent: `
				temp="{{ getTag \"test\" }}"
				[file]	
				` + createFrontendConfiguration(2) + createBackendConfiguration(3) + createTLS(4),
			expectedNumFrontend: 2,
			expectedNumBackend:  3,
			expectedNumTLSConf:  4,
		},
	}
}

func TestProvideWithoutWatch(t *testing.T) {
	for _, test := range getTestCases() {
		test := test
		t.Run(test.desc+" without watch", func(t *testing.T) {
			t.Parallel()

			provider, clean := createProvider(t, test, false)
			defer clean()
			configChan := make(chan types.ConfigMessage)

			go func() {
				err := provider.Provide(configChan, safe.NewPool(context.Background()))
				assert.NoError(t, err)
			}()

			timeout := time.After(time.Second)
			select {
			case config := <-configChan:
				assert.Len(t, config.Configuration.Backends, test.expectedNumBackend)
				assert.Len(t, config.Configuration.Frontends, test.expectedNumFrontend)
				assert.Len(t, config.Configuration.TLS, test.expectedNumTLSConf)
			case <-timeout:
				t.Errorf("timeout while waiting for config")
			}
		})
	}
}

func TestProvideWithWatch(t *testing.T) {
	for _, test := range getTestCases() {
		test := test
		t.Run(test.desc+" with watch", func(t *testing.T) {
			t.Parallel()

			provider, clean := createProvider(t, test, true)
			defer clean()
			configChan := make(chan types.ConfigMessage)

			go func() {
				err := provider.Provide(configChan, safe.NewPool(context.Background()))
				assert.NoError(t, err)
			}()

			timeout := time.After(time.Second)
			select {
			case config := <-configChan:
				assert.Len(t, config.Configuration.Backends, 0)
				assert.Len(t, config.Configuration.Frontends, 0)
				assert.Len(t, config.Configuration.TLS, 0)
			case <-timeout:
				t.Errorf("timeout while waiting for config")
			}

			if len(test.fileContent) > 0 {
				ioutil.WriteFile(provider.Filename, []byte(test.fileContent), 0755)
			}

			if len(test.traefikFileContent) > 0 {
				ioutil.WriteFile(provider.TraefikFile, []byte(test.traefikFileContent), 0755)
			}

			if len(test.directoryContent) > 0 {
				for _, fileContent := range test.directoryContent {
					createRandomFile(t, provider.Directory, fileContent)
				}
			}

			timeout = time.After(time.Second * 1)
			var numUpdates, numBackends, numFrontends, numTLSConfs int
			for {
				select {
				case config := <-configChan:
					numUpdates++
					numBackends = len(config.Configuration.Backends)
					numFrontends = len(config.Configuration.Frontends)
					numTLSConfs = len(config.Configuration.TLS)
					t.Logf("received update #%d: backends %d/%d, frontends %d/%d, TLS configs %d/%d", numUpdates, numBackends, test.expectedNumBackend, numFrontends, test.expectedNumFrontend, numTLSConfs, test.expectedNumTLSConf)

					if numBackends == test.expectedNumBackend && numFrontends == test.expectedNumFrontend && numTLSConfs == test.expectedNumTLSConf {
						return
					}
				case <-timeout:
					t.Fatal("timeout while waiting for config")
				}
			}
		})
	}
}

func TestErrorWhenEmptyConfig(t *testing.T) {
	provider := &Provider{}
	configChan := make(chan types.ConfigMessage)
	errorChan := make(chan struct{})
	go func() {
		err := provider.Provide(configChan, safe.NewPool(context.Background()))
		assert.Error(t, err)
		close(errorChan)
	}()

	timeout := time.After(time.Second)
	select {
	case <-configChan:
		t.Fatal("We should not receive config message")
	case <-timeout:
		t.Fatal("timeout while waiting for config")
	case <-errorChan:
	}
}

func createProvider(t *testing.T, test ProvideTestCase, watch bool) (*Provider, func()) {
	tempDir := createTempDir(t, "testdir")

	provider := &Provider{}
	provider.Watch = watch

	if len(test.directoryContent) > 0 {
		if !watch {
			for _, fileContent := range test.directoryContent {
				createRandomFile(t, tempDir, fileContent)
			}
		}
		provider.Directory = tempDir
	}

	if len(test.fileContent) > 0 {
		if watch {
			test.fileContent = ""
		}
		filename := createRandomFile(t, tempDir, test.fileContent)
		provider.Filename = filename.Name()

	}

	if len(test.traefikFileContent) > 0 {
		if watch {
			test.traefikFileContent = ""
		}
		filename := createRandomFile(t, tempDir, test.traefikFileContent)
		provider.TraefikFile = filename.Name()
	}

	return provider, func() {
		os.RemoveAll(tempDir)
	}
}

func TestTLSContent(t *testing.T) {
	tempDir := createTempDir(t, "testdir")
	defer os.RemoveAll(tempDir)

	fileTLS := createRandomFile(t, tempDir, "CONTENT")
	fileConfig := createRandomFile(t, tempDir, `
[[tls]]
entryPoints = ["https"]
  [tls.certificate]
    certFile = "`+fileTLS.Name()+`"
    keyFile = "`+fileTLS.Name()+`"
`)

	provider := &Provider{}
	configuration, err := provider.loadFileConfig(fileConfig.Name(), true)
	require.NoError(t, err)

	require.Equal(t, "CONTENT", configuration.TLS[0].Certificate.CertFile.String())
	require.Equal(t, "CONTENT", configuration.TLS[0].Certificate.KeyFile.String())
}
