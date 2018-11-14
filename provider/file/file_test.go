package file

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/safe"
	"github.com/stretchr/testify/assert"
)

type ProvideTestCase struct {
	desc               string
	directoryContent   []string
	fileContent        string
	traefikFileContent string
	expectedNumRouter  int
	expectedNumService int
	expectedNumTLSConf int
}

func TestProvideWithoutWatch(t *testing.T) {
	for _, test := range getTestCases() {
		t.Run(test.desc+" without watch", func(t *testing.T) {
			provider, clean := createProvider(t, test, false)
			defer clean()
			configChan := make(chan config.Message)

			provider.DebugLogGeneratedTemplate = true

			go func() {
				err := provider.Provide(configChan, safe.NewPool(context.Background()))
				assert.NoError(t, err)
			}()

			timeout := time.After(time.Second)
			select {
			case conf := <-configChan:
				assert.Len(t, conf.Configuration.Services, test.expectedNumService)
				assert.Len(t, conf.Configuration.Routers, test.expectedNumRouter)
				assert.Len(t, conf.Configuration.TLS, test.expectedNumTLSConf)
			case <-timeout:
				t.Errorf("timeout while waiting for config")
			}
		})
	}
}

func TestProvideWithWatch(t *testing.T) {
	for _, test := range getTestCases() {
		t.Run(test.desc+" with watch", func(t *testing.T) {
			provider, clean := createProvider(t, test, true)
			defer clean()
			configChan := make(chan config.Message)

			go func() {
				err := provider.Provide(configChan, safe.NewPool(context.Background()))
				assert.NoError(t, err)
			}()

			timeout := time.After(time.Second)
			select {
			case conf := <-configChan:
				assert.Len(t, conf.Configuration.Services, 0)
				assert.Len(t, conf.Configuration.Routers, 0)
				assert.Len(t, conf.Configuration.TLS, 0)
			case <-timeout:
				t.Errorf("timeout while waiting for config")
			}

			if len(test.fileContent) > 0 {
				if err := ioutil.WriteFile(provider.Filename, []byte(test.fileContent), 0755); err != nil {
					t.Error(err)
				}
			}

			if len(test.traefikFileContent) > 0 {
				if err := ioutil.WriteFile(provider.TraefikFile, []byte(test.traefikFileContent), 0755); err != nil {
					t.Error(err)
				}
			}

			if len(test.directoryContent) > 0 {
				for _, fileContent := range test.directoryContent {
					createRandomFile(t, provider.Directory, fileContent)
				}
			}

			timeout = time.After(time.Second * 1)
			var numUpdates, numServices, numRouters, numTLSConfs int
			for {
				select {
				case conf := <-configChan:
					numUpdates++
					numServices = len(conf.Configuration.Services)
					numRouters = len(conf.Configuration.Routers)
					numTLSConfs = len(conf.Configuration.TLS)
					t.Logf("received update #%d: services %d/%d, routers %d/%d, TLS configs %d/%d", numUpdates, numServices, test.expectedNumService, numRouters, test.expectedNumRouter, numTLSConfs, test.expectedNumTLSConf)

					if numServices == test.expectedNumService && numRouters == test.expectedNumRouter && numTLSConfs == test.expectedNumTLSConf {
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
	configChan := make(chan config.Message)
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

func getTestCases() []ProvideTestCase {
	return []ProvideTestCase{
		{
			desc:               "simple file",
			fileContent:        createRoutersConfiguration(3) + createServicesConfiguration(6) + createTLS(5),
			expectedNumRouter:  3,
			expectedNumService: 6,
			expectedNumTLSConf: 5,
		},
		{
			desc:        "simple file and a traefik file",
			fileContent: createRoutersConfiguration(4) + createServicesConfiguration(8) + createTLS(4),
			traefikFileContent: `
			debug=true
`,
			expectedNumRouter:  4,
			expectedNumService: 8,
			expectedNumTLSConf: 4,
		},
		{
			desc: "template file",
			fileContent: `
[routers]
{{ range $i, $e := until 20 }}
  [routers.router{{ $e }}]
  service = "application"  
{{ end }}
`,
			expectedNumRouter: 20,
		},
		{
			desc: "simple directory",
			directoryContent: []string{
				createRoutersConfiguration(2),
				createServicesConfiguration(3),
				createTLS(4),
			},
			expectedNumRouter:  2,
			expectedNumService: 3,
			expectedNumTLSConf: 4,
		},
		{
			desc: "template in directory",
			directoryContent: []string{
				`
[routers]
{{ range $i, $e := until 20 }}
  [routers.router{{ $e }}]
  service = "application"  
{{ end }}
`,
				`
[services]
{{ range $i, $e := until 20 }}
  [services.application-{{ $e }}]
	[[services.application-{{ $e }}.servers]]
	url="http://127.0.0.1"
	weight = 1
{{ end }}
`,
			},
			expectedNumRouter:  20,
			expectedNumService: 20,
		},
		{
			desc: "simple traefik file",
			traefikFileContent: `
				debug=true
				[file]	
				` + createRoutersConfiguration(2) + createServicesConfiguration(3) + createTLS(4),
			expectedNumRouter:  2,
			expectedNumService: 3,
			expectedNumTLSConf: 4,
		},
		{
			desc: "simple traefik file with templating",
			traefikFileContent: `
				temp="{{ getTag \"test\" }}"
				[file]	
				` + createRoutersConfiguration(2) + createServicesConfiguration(3) + createTLS(4),
			expectedNumRouter:  2,
			expectedNumService: 3,
			expectedNumTLSConf: 4,
		},
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
		os.Remove(tempDir)
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
		_, err = tempFile.WriteString(content)
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

// createRoutersConfiguration Helper
func createRoutersConfiguration(n int) string {
	conf := "[routers]\n"
	for i := 1; i <= n; i++ {
		conf += fmt.Sprintf(` 
[routers."router%[1]d"]
	service = "application-%[1]d"
`, i)
	}
	return conf
}

// createServicesConfiguration Helper
func createServicesConfiguration(n int) string {
	conf := "[services]\n"
	for i := 1; i <= n; i++ {
		conf += fmt.Sprintf(`
[services.application-%[1]d.loadbalancer]
   [[services.application-%[1]d.loadbalancer.servers]]
     url = "http://172.17.0.%[1]d:80"
     weight = 1
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
