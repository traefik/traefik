package file

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

type ProvideTestCase struct {
	desc                  string
	directoryPaths        []string
	filePath              string
	expectedNumRouter     int
	expectedNumService    int
	expectedNumTLSConf    int
	expectedNumTLSOptions int
}

func TestTLSCertificateContent(t *testing.T) {
	tempDir := t.TempDir()

	fileTLS, err := createTempFile("./fixtures/toml/tls_file.cert", tempDir)
	require.NoError(t, err)

	fileTLSKey, err := createTempFile("./fixtures/toml/tls_file_key.cert", tempDir)
	require.NoError(t, err)

	fileConfig, err := os.CreateTemp(tempDir, "temp*.toml")
	require.NoError(t, err)

	content := `
[[tls.certificates]]
  certFile = "` + fileTLS.Name() + `"
  keyFile = "` + fileTLSKey.Name() + `"

[tls.options.default.clientAuth]
  caFiles = ["` + fileTLS.Name() + `"]

[tls.stores.default.defaultCertificate]
  certFile = "` + fileTLS.Name() + `"
  keyFile = "` + fileTLSKey.Name() + `"

[http.serversTransports.default]
  rootCAs = ["` + fileTLS.Name() + `"]
  [[http.serversTransports.default.certificates]]
    certFile = "` + fileTLS.Name() + `"
    keyFile = "` + fileTLSKey.Name() + `"

[tcp.serversTransports.default]
  [tcp.serversTransports.default.tls]
    rootCAs = ["` + fileTLS.Name() + `"]
  	[[tcp.serversTransports.default.tls.certificates]]
      certFile = "` + fileTLS.Name() + `"
      keyFile = "` + fileTLSKey.Name() + `"
`

	_, err = fileConfig.WriteString(content)
	require.NoError(t, err)

	provider := &Provider{}
	configuration, err := provider.loadFileConfig(t.Context(), fileConfig.Name(), true)
	require.NoError(t, err)

	require.Equal(t, "CONTENT", configuration.TLS.Certificates[0].Certificate.CertFile.String())
	require.Equal(t, "CONTENTKEY", configuration.TLS.Certificates[0].Certificate.KeyFile.String())

	require.Equal(t, "CONTENT", configuration.TLS.Options["default"].ClientAuth.CAFiles[0].String())

	require.Equal(t, "CONTENT", configuration.TLS.Stores["default"].DefaultCertificate.CertFile.String())
	require.Equal(t, "CONTENTKEY", configuration.TLS.Stores["default"].DefaultCertificate.KeyFile.String())

	require.Equal(t, "CONTENT", configuration.HTTP.ServersTransports["default"].Certificates[0].CertFile.String())
	require.Equal(t, "CONTENTKEY", configuration.HTTP.ServersTransports["default"].Certificates[0].KeyFile.String())
	require.Equal(t, "CONTENT", configuration.HTTP.ServersTransports["default"].RootCAs[0].String())

	require.Equal(t, "CONTENT", configuration.TCP.ServersTransports["default"].TLS.Certificates[0].CertFile.String())
	require.Equal(t, "CONTENTKEY", configuration.TCP.ServersTransports["default"].TLS.Certificates[0].KeyFile.String())
	require.Equal(t, "CONTENT", configuration.TCP.ServersTransports["default"].TLS.RootCAs[0].String())
}

func TestTLSCertificateMissingFile(t *testing.T) {
	t.Parallel()

	const (
		pemCert = "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"
		pemKey  = "-----BEGIN PRIVATE KEY-----\nMIIE\n-----END PRIVATE KEY-----\n"
	)

	testCases := []struct {
		desc          string
		certFile      string
		keyFile       string
		expected      *tls.Certificate
		expectedStore *tls.Certificate
		expectedError string
	}{
		{
			desc:     "existing non-PEM files",
			certFile: "fixtures/toml/tls_file.cert",
			keyFile:  "fixtures/toml/tls_file_key.cert",
			expected: &tls.Certificate{CertFile: "CONTENT", KeyFile: "CONTENTKEY"},
		},
		{
			desc:     "inline PEM content after a preamble",
			certFile: "Bag Attributes\n" + pemCert,
			keyFile:  pemKey,
			expected: &tls.Certificate{CertFile: "Bag Attributes\n" + pemCert, KeyFile: pemKey},
		},
		{
			desc:          "missing certFile",
			certFile:      "fixtures/missing.crt",
			keyFile:       "fixtures/toml/tls_file_key.cert",
			expectedStore: &tls.Certificate{CertFile: "fixtures/missing.crt", KeyFile: "CONTENTKEY"},
			expectedError: "reading certFile fixtu[...]g.crt: file does not exist and value is not PEM data",
		},
		{
			desc:          "missing keyFile",
			certFile:      "fixtures/toml/tls_file.cert",
			keyFile:       "fixtures/missing.key",
			expectedStore: &tls.Certificate{CertFile: "CONTENT", KeyFile: "fixtures/missing.key"},
			expectedError: "reading keyFile of certificate fixtures/toml/tls_file.cert: file does not exist and value is not PEM data",
		},
		{
			desc:     "empty keyFile",
			certFile: "fixtures/toml/tls_file.cert",
			expected: &tls.Certificate{CertFile: "CONTENT"},
		},
		{
			desc:     "path through a regular file",
			certFile: "fixtures/toml/tls_file.cert/missing.crt",
			keyFile:  pemKey,
			expected: &tls.Certificate{CertFile: "fixtures/toml/tls_file.cert/missing.crt", KeyFile: pemKey},
		},
		{
			desc:     "name too long",
			certFile: strings.Repeat("a", 300),
			keyFile:  pemKey,
			expected: &tls.Certificate{CertFile: types.FileOrContent(strings.Repeat("a", 300)), KeyFile: pemKey},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			content := fmt.Sprintf(`
tls:
  certificates:
    - certFile: %[1]s
      keyFile: %[2]s
  stores:
    default:
      defaultCertificate:
        certFile: %[1]s
        keyFile: %[2]s
http:
  serversTransports:
    default:
      certificates:
        - certFile: %[1]s
          keyFile: %[2]s
tcp:
  serversTransports:
    default:
      tls:
        certificates:
          - certFile: %[1]s
            keyFile: %[2]s
`, strconv.Quote(test.certFile), strconv.Quote(test.keyFile))

			configFile := filepath.Join(t.TempDir(), "config.yml")
			require.NoError(t, os.WriteFile(configFile, []byte(content), 0o600))

			var logs bytes.Buffer
			ctx := zerolog.New(&logs).WithContext(t.Context())

			configuration, err := (&Provider{}).loadFileConfig(ctx, configFile, false)
			require.NoError(t, err)

			var expectedErrors []string
			var expectedCertAndStores []*tls.CertAndStores
			var expectedCertificates tls.Certificates
			expectedStore := test.expected
			if test.expectedError != "" {
				expectedErrors = slices.Repeat([]string{test.expectedError}, 4)
				expectedStore = test.expectedStore
			} else {
				expectedCertAndStores = []*tls.CertAndStores{{Certificate: *test.expected}}
				expectedCertificates = tls.Certificates{*test.expected}
			}

			var loggedErrors []string
			for line := range strings.Lines(logs.String()) {
				var entry struct {
					Level string `json:"level"`
					Error string `json:"error"`
				}
				require.NoError(t, json.Unmarshal([]byte(line), &entry))
				assert.Equal(t, "error", entry.Level)
				loggedErrors = append(loggedErrors, entry.Error)
			}

			assert.Equal(t, expectedErrors, loggedErrors)
			assert.Equal(t, expectedCertAndStores, configuration.TLS.Certificates)
			assert.Equal(t, expectedStore, configuration.TLS.Stores["default"].DefaultCertificate)
			assert.Equal(t, expectedCertificates, configuration.HTTP.ServersTransports["default"].Certificates)
			assert.Equal(t, expectedCertificates, configuration.TCP.ServersTransports["default"].TLS.Certificates)
		})
	}
}

func TestErrorWhenEmptyConfig(t *testing.T) {
	provider := &Provider{}
	configChan := make(chan dynamic.Message)
	errorChan := make(chan struct{})
	go func() {
		err := provider.Provide(configChan, safe.NewPool(t.Context()))
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

func TestBuildConfigurationErrorIncludesFilename(t *testing.T) {
	tempDir := t.TempDir()

	file, err := os.CreateTemp(tempDir, "temp*.yml")
	require.NoError(t, err)

	// passHostHeader is intentionally misindented to produce a YAML syntax error.
	content := `
http:
  services:
    Service01:
      loadBalancer:
        servers:
          - url: "http://localhost:8000"
       passHostHeader: true
`
	_, err = file.WriteString(content)
	require.NoError(t, err)

	provider := &Provider{Filename: file.Name()}

	_, err = provider.buildConfiguration()
	require.Error(t, err)
	assert.ErrorContains(t, err, file.Name())
}

func TestProvideWithoutWatch(t *testing.T) {
	for _, test := range getTestCases() {
		t.Run(test.desc+" without watch", func(t *testing.T) {
			provider := createProvider(t, test, false)
			configChan := make(chan dynamic.Message)

			provider.DebugLogGeneratedTemplate = true

			go func() {
				err := provider.Provide(configChan, safe.NewPool(t.Context()))
				assert.NoError(t, err)
			}()

			timeout := time.After(time.Second)
			select {
			case conf := <-configChan:

				require.NotNil(t, conf.Configuration.HTTP)
				numServices := len(conf.Configuration.HTTP.Services) + len(conf.Configuration.TCP.Services) + len(conf.Configuration.UDP.Services)
				numRouters := len(conf.Configuration.HTTP.Routers) + len(conf.Configuration.TCP.Routers) + len(conf.Configuration.UDP.Routers)
				assert.Equal(t, test.expectedNumService, numServices)
				assert.Equal(t, test.expectedNumRouter, numRouters)
				require.NotNil(t, conf.Configuration.TLS)
				assert.Len(t, conf.Configuration.TLS.Certificates, test.expectedNumTLSConf)
				assert.Len(t, conf.Configuration.TLS.Options, test.expectedNumTLSOptions)
			case <-timeout:
				t.Errorf("timeout while waiting for config")
			}
		})
	}
}

func TestProvideWithWatch(t *testing.T) {
	for _, test := range getTestCases() {
		t.Run(test.desc+" with watch", func(t *testing.T) {
			provider := createProvider(t, test, true)
			configChan := make(chan dynamic.Message)

			go func() {
				err := provider.Provide(configChan, safe.NewPool(t.Context()))
				assert.NoError(t, err)
			}()

			timeout := time.After(time.Second)
			select {
			case conf := <-configChan:
				require.NotNil(t, conf.Configuration.HTTP)
				numServices := len(conf.Configuration.HTTP.Services) + len(conf.Configuration.TCP.Services) + len(conf.Configuration.UDP.Services)
				numRouters := len(conf.Configuration.HTTP.Routers) + len(conf.Configuration.TCP.Routers) + len(conf.Configuration.UDP.Routers)
				assert.Equal(t, 0, numServices)
				assert.Equal(t, 0, numRouters)
				require.NotNil(t, conf.Configuration.TLS)
				assert.Empty(t, conf.Configuration.TLS.Certificates)
			case <-timeout:
				t.Errorf("timeout while waiting for config")
			}

			if len(test.filePath) > 0 {
				err := copyFile(test.filePath, provider.Filename)
				require.NoError(t, err)
			}

			if len(test.directoryPaths) > 0 {
				for i, filePath := range test.directoryPaths {
					err := copyFile(filePath, filepath.Join(provider.Directory, strconv.Itoa(i)+filepath.Ext(filePath)))
					require.NoError(t, err)
				}
			}

			timeout = time.After(1 * time.Second)
			var numUpdates, numServices, numRouters, numTLSConfs int
			for {
				select {
				case conf := <-configChan:
					numUpdates++
					numServices = len(conf.Configuration.HTTP.Services) + len(conf.Configuration.TCP.Services) + len(conf.Configuration.UDP.Services)
					numRouters = len(conf.Configuration.HTTP.Routers) + len(conf.Configuration.TCP.Routers) + len(conf.Configuration.UDP.Routers)
					numTLSConfs = len(conf.Configuration.TLS.Certificates)
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

func TestProvideWatchWithNonConfigDanglingSymlink(t *testing.T) {
	tempDir := t.TempDir()

	err := copyFile("./fixtures/yaml/simple_file_01.yml", filepath.Join(tempDir, "simple_file_01.yml"))
	require.NoError(t, err)

	err = os.Symlink(filepath.Join(tempDir, "non_existent_file.txt"), filepath.Join(tempDir, "dangling_symlink.txt"))
	require.NoError(t, err)

	provider := &Provider{
		Directory: tempDir,
		Watch:     true,
	}
	configChan := make(chan dynamic.Message)
	go func() {
		err := provider.Provide(configChan, safe.NewPool(t.Context()))
		assert.NoError(t, err)
	}()

	timeout := time.After(time.Second)
	select {
	case conf := <-configChan:
		require.NotNil(t, conf.Configuration.HTTP)
		numServices := len(conf.Configuration.HTTP.Services) + len(conf.Configuration.TCP.Services) + len(conf.Configuration.UDP.Services)
		numRouters := len(conf.Configuration.HTTP.Routers) + len(conf.Configuration.TCP.Routers) + len(conf.Configuration.UDP.Routers)
		assert.Equal(t, 6, numServices)
		assert.Equal(t, 3, numRouters)
	case <-timeout:
		t.Errorf("timeout while waiting for config")
	}
}

func getTestCases() []ProvideTestCase {
	return []ProvideTestCase{
		{
			desc:               "simple file",
			filePath:           "./fixtures/toml/simple_file_01.toml",
			expectedNumRouter:  3,
			expectedNumService: 6,
			expectedNumTLSConf: 5,
		},
		{
			desc:               "simple file with tcp and udp",
			filePath:           "./fixtures/toml/simple_file_02.toml",
			expectedNumRouter:  5,
			expectedNumService: 8,
			expectedNumTLSConf: 5,
		},
		{
			desc:               "simple file yaml",
			filePath:           "./fixtures/yaml/simple_file_01.yml",
			expectedNumRouter:  3,
			expectedNumService: 6,
			expectedNumTLSConf: 5,
		},
		{
			desc:              "template file",
			filePath:          "./fixtures/toml/template_file.toml",
			expectedNumRouter: 20,
		},
		{
			desc:              "template file yaml",
			filePath:          "./fixtures/yaml/template_file.yml",
			expectedNumRouter: 20,
		},
		{
			desc: "simple directory",
			directoryPaths: []string{
				"./fixtures/toml/dir01_file01.toml",
				"./fixtures/toml/dir01_file02.toml",
				"./fixtures/toml/dir01_file03.toml",
			},
			expectedNumRouter:     2,
			expectedNumService:    3,
			expectedNumTLSConf:    4,
			expectedNumTLSOptions: 1,
		},
		{
			desc: "simple directory yaml",
			directoryPaths: []string{
				"./fixtures/yaml/dir01_file01.yml",
				"./fixtures/yaml/dir01_file02.yml",
				"./fixtures/yaml/dir01_file03.yml",
			},
			expectedNumRouter:     2,
			expectedNumService:    3,
			expectedNumTLSConf:    4,
			expectedNumTLSOptions: 1,
		},
		{
			desc: "template in directory",
			directoryPaths: []string{
				"./fixtures/toml/template_in_directory_file01.toml",
				"./fixtures/toml/template_in_directory_file02.toml",
			},
			expectedNumRouter:  20,
			expectedNumService: 20,
		},
		{
			desc: "template in directory yaml",
			directoryPaths: []string{
				"./fixtures/yaml/template_in_directory_file01.yml",
				"./fixtures/yaml/template_in_directory_file02.yml",
			},
			expectedNumRouter:  20,
			expectedNumService: 20,
		},
		{
			desc:               "simple file with empty store yaml",
			filePath:           "./fixtures/yaml/simple_empty_store.yml",
			expectedNumRouter:  0,
			expectedNumService: 0,
			expectedNumTLSConf: 0,
		},
	}
}

func createProvider(t *testing.T, test ProvideTestCase, watch bool) *Provider {
	t.Helper()

	tempDir := t.TempDir()

	provider := &Provider{}
	provider.Watch = true

	if len(test.directoryPaths) > 0 {
		if !watch {
			for _, filePath := range test.directoryPaths {
				var err error
				_, err = createTempFile(filePath, tempDir)
				require.NoError(t, err)
			}
		}
		provider.Directory = tempDir
	}

	if len(test.filePath) > 0 {
		var file *os.File
		if watch {
			var err error
			file, err = os.CreateTemp(tempDir, "temp*"+filepath.Ext(test.filePath))
			require.NoError(t, err)
		} else {
			var err error
			file, err = createTempFile(test.filePath, tempDir)
			require.NoError(t, err)
		}

		provider.Filename = file.Name()
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return provider
}

func copyFile(srcPath, dstPath string) error {
	dst, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}
	defer dst.Close()

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = io.Copy(dst, src)
	return err
}

func createTempFile(srcPath, tempDir string) (*os.File, error) {
	file, err := os.CreateTemp(tempDir, "temp*"+filepath.Ext(srcPath))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	src, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	_, err = io.Copy(file, src)
	return file, err
}
