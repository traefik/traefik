package file

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/safe"
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

func TestTLSCertificateContentMissingFile(t *testing.T) {
	tempDir := t.TempDir()

	missingCert := filepath.Join(tempDir, "not-yet-issued-cert.pem")
	missingKey := filepath.Join(tempDir, "not-yet-issued-key.pem")

	fileConfig, err := os.CreateTemp(tempDir, "temp*.toml")
	require.NoError(t, err)

	content := `
[[tls.certificates]]
  certFile = "` + missingCert + `"
  keyFile = "` + missingKey + `"
`

	_, err = fileConfig.WriteString(content)
	require.NoError(t, err)

	provider := &Provider{}
	_, refFiles, err := provider.loadFileConfig(t.Context(), fileConfig.Name())
	require.NoError(t, err)

	require.ElementsMatch(t, []string{missingCert, missingKey}, refFiles)
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
	configuration, refFiles, err := provider.loadFileConfig(t.Context(), fileConfig.Name())
	require.NoError(t, err)

	// Every certificate/key/CA path referenced from the config should be reported,
	// so the caller can keep watching them for external changes.
	require.ElementsMatch(t, []string{
		fileTLS.Name(), fileTLSKey.Name(), // tls.certificates
		fileTLS.Name(),                    // tls.options.default.clientAuth.caFiles
		fileTLS.Name(), fileTLSKey.Name(), // tls.stores.default.defaultCertificate
		fileTLS.Name(),                    // http.serversTransports.default.rootCAs
		fileTLS.Name(), fileTLSKey.Name(), // http.serversTransports.default.certificates
		fileTLS.Name(),                    // tcp.serversTransports.default.tls.rootCAs
		fileTLS.Name(), fileTLSKey.Name(), // tcp.serversTransports.default.tls.certificates
	}, refFiles)

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

func TestProvideWithWatch_RenewedCertificateInPlaceReloads(t *testing.T) {
	dynamicDir := t.TempDir()
	certsDir := t.TempDir()

	firstExpiry := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	writeCertPair(t, filepath.Join(certsDir, "cert.pem"), filepath.Join(certsDir, "key.pem"), firstExpiry)
	newDynamicCertConfig(t, dynamicDir, certsDir)

	provider := &Provider{Directory: dynamicDir, Watch: true}
	configChan := make(chan dynamic.Message)
	go func() {
		assert.NoError(t, provider.Provide(configChan, safe.NewPool(t.Context())))
	}()

	waitForCertNotAfter(t, configChan, firstExpiry, 2*time.Second)

	secondExpiry := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	writeCertPair(t, filepath.Join(certsDir, "cert.pem"), filepath.Join(certsDir, "key.pem"), secondExpiry)

	waitForCertNotAfter(t, configChan, secondExpiry, 2*time.Second)
}

func TestProvideWithWatch_RenamedCertificateReloadsRepeatedly(t *testing.T) {
	dynamicDir := t.TempDir()
	certsDir := t.TempDir()

	firstExpiry := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	writeCertPair(t, filepath.Join(certsDir, "cert.pem"), filepath.Join(certsDir, "key.pem"), firstExpiry)
	newDynamicCertConfig(t, dynamicDir, certsDir)

	provider := &Provider{Directory: dynamicDir, Watch: true}
	configChan := make(chan dynamic.Message)
	go func() {
		assert.NoError(t, provider.Provide(configChan, safe.NewPool(t.Context())))
	}()

	waitForCertNotAfter(t, configChan, firstExpiry, 2*time.Second)

	secondExpiry := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	renameCertPair(t, filepath.Join(certsDir, "cert.pem"), filepath.Join(certsDir, "key.pem"), secondExpiry)
	waitForCertNotAfter(t, configChan, secondExpiry, 2*time.Second)

	thirdExpiry := time.Now().Add(72 * time.Hour).Truncate(time.Second)
	renameCertPair(t, filepath.Join(certsDir, "cert.pem"), filepath.Join(certsDir, "key.pem"), thirdExpiry)
	waitForCertNotAfter(t, configChan, thirdExpiry, 2*time.Second)
}

func TestProvideWithWatch_DeletedAndRecreatedCertificateReloads(t *testing.T) {
	dynamicDir := t.TempDir()
	certsDir := t.TempDir()

	firstExpiry := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	certPath := filepath.Join(certsDir, "cert.pem")
	keyPath := filepath.Join(certsDir, "key.pem")
	writeCertPair(t, certPath, keyPath, firstExpiry)
	newDynamicCertConfig(t, dynamicDir, certsDir)

	provider := &Provider{Directory: dynamicDir, Watch: true}
	configChan := make(chan dynamic.Message)
	go func() {
		assert.NoError(t, provider.Provide(configChan, safe.NewPool(t.Context())))
	}()

	waitForCertNotAfter(t, configChan, firstExpiry, 2*time.Second)

	require.NoError(t, os.Remove(certPath))
	require.NoError(t, os.Remove(keyPath))

	secondExpiry := time.Now().Add(96 * time.Hour).Truncate(time.Second)
	writeCertPair(t, certPath, keyPath, secondExpiry)

	waitForCertNotAfter(t, configChan, secondExpiry, 2*time.Second)
}

func TestProvideWithWatch_CertificateMissingAtStartupIsPickedUpOnceCreated(t *testing.T) {
	dynamicDir := t.TempDir()
	certsDir := t.TempDir()
	certPath := filepath.Join(certsDir, "cert.pem")
	keyPath := filepath.Join(certsDir, "key.pem")
	newDynamicCertConfig(t, dynamicDir, certsDir)

	provider := &Provider{Directory: dynamicDir, Watch: true}
	configChan := make(chan dynamic.Message)
	go func() {
		assert.NoError(t, provider.Provide(configChan, safe.NewPool(t.Context())))
	}()

	time.Sleep(200 * time.Millisecond)

	expiry := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	writeCertPair(t, certPath, keyPath, expiry)

	waitForCertNotAfter(t, configChan, expiry, 2*time.Second)
}

func TestIsBaseDir(t *testing.T) {
	tests := []struct {
		desc      string
		directory string
		filename  string
		dir       string
		want      bool
	}{
		{
			desc:      "directory mode, exact match",
			directory: "/etc/traefik/dynamic",
			dir:       "/etc/traefik/dynamic",
			want:      true,
		},
		{
			desc:      "directory mode, unrelated dir",
			directory: "/etc/traefik/dynamic",
			dir:       "/etc/traefik/other",
			want:      false,
		},
		{
			desc:     "filename mode, matches parent dir",
			filename: "/etc/traefik/dynamic/certificates.toml",
			dir:      "/etc/traefik/dynamic",
			want:     true,
		},
		{
			desc:     "filename mode, unrelated dir",
			filename: "/etc/traefik/dynamic/certificates.toml",
			dir:      "/etc/other",
			want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			p := &Provider{Directory: test.directory, Filename: test.filename}
			assert.Equal(t, test.want, p.isBaseDir(test.dir))
		})
	}
}

func TestSyncExternalFileWatches(t *testing.T) {
	dynamicDir := t.TempDir()
	certsDirA := t.TempDir()
	certsDirB := t.TempDir()

	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	t.Cleanup(func() { _ = watcher.Close() })

	p := &Provider{Directory: dynamicDir}
	p.watcher = watcher

	certA := filepath.Join(certsDirA, "cert.pem")
	certB := filepath.Join(certsDirB, "cert.pem")
	inBaseDir := filepath.Join(dynamicDir, "inline-cert.pem")

	p.syncExternalFileWatches([]string{certA, inBaseDir})
	assert.Equal(t, map[string]struct{}{certsDirA: {}}, p.externalDirs, "certsDirA should be tracked; the base dir must not be")

	p.syncExternalFileWatches([]string{certB, inBaseDir})
	assert.Equal(t, map[string]struct{}{certsDirB: {}}, p.externalDirs)

	p.syncExternalFileWatches(nil)
	assert.Empty(t, p.externalDirs)
}

func TestIsRelevantEvent(t *testing.T) {
	tests := []struct {
		desc      string
		directory string
		filename  string
		external  map[string]struct{}
		event     string
		want      bool
	}{
		{
			desc:      "directory mode, event under directory",
			directory: "/etc/traefik/dynamic",
			event:     "/etc/traefik/dynamic/certificates.toml",
			want:      true,
		},
		{
			desc:      "directory mode, sibling dir with overlapping name prefix is not matched",
			directory: "/etc/traefik/dynamic",
			event:     "/etc/traefik/dynamic-backup/certificates.toml",
			want:      false,
		},
		{
			desc:     "filename mode, matching basename",
			filename: "/etc/traefik/dynamic/certificates.toml",
			event:    "/etc/traefik/dynamic/certificates.toml",
			want:     true,
		},
		{
			desc:     "filename mode, unrelated file in same dir",
			filename: "/etc/traefik/dynamic/certificates.toml",
			event:    "/etc/traefik/dynamic/other.toml",
			want:     false,
		},
		{
			desc:     "filename mode, event under a tracked external dir",
			filename: "/etc/traefik/dynamic/certificates.toml",
			external: map[string]struct{}{"/etc/letsencrypt/live/example.com": {}},
			event:    "/etc/letsencrypt/live/example.com/cert.pem",
			want:     true,
		},
		{
			desc:     "filename mode, event under an untracked dir",
			filename: "/etc/traefik/dynamic/certificates.toml",
			event:    "/etc/letsencrypt/live/example.com/cert.pem",
			want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			p := &Provider{Directory: test.directory, Filename: test.filename}
			p.externalDirs = test.external
			assert.Equal(t, test.want, p.isRelevantEvent(test.event))
		})
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

func generateSelfSignedCert(t *testing.T, notAfter time.Time) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "repro.local"},
		NotBefore:    time.Now(),
		NotAfter:     notAfter,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM
}

func writeCertPair(t *testing.T, certPath, keyPath string, notAfter time.Time) {
	t.Helper()

	certPEM, keyPEM := generateSelfSignedCert(t, notAfter)
	require.NoError(t, os.WriteFile(certPath, certPEM, 0o600))
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0o600))
}

func renameCertPair(t *testing.T, certPath, keyPath string, notAfter time.Time) {
	t.Helper()

	dir := t.TempDir()
	tmpCert := filepath.Join(dir, "cert.pem")
	tmpKey := filepath.Join(dir, "key.pem")
	writeCertPair(t, tmpCert, tmpKey, notAfter)

	require.NoError(t, os.Rename(tmpCert, certPath))
	require.NoError(t, os.Rename(tmpKey, keyPath))
}

func waitForCertNotAfter(t *testing.T, configChan chan dynamic.Message, want time.Time, timeout time.Duration) {
	t.Helper()

	deadline := time.After(timeout)
	for {
		select {
		case conf := <-configChan:
			if conf.Configuration.TLS == nil || len(conf.Configuration.TLS.Certificates) == 0 {
				continue
			}

			certPEM := conf.Configuration.TLS.Certificates[0].Certificate.CertFile.String()
			block, _ := pem.Decode([]byte(certPEM))
			if block == nil {
				continue
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				continue
			}

			if cert.NotAfter.Truncate(time.Second).Equal(want.Truncate(time.Second)) {
				return
			}
		case <-deadline:
			t.Fatalf("timeout waiting for certificate with NotAfter=%s", want)
		}
	}
}

func newDynamicCertConfig(t *testing.T, dir, certsDir string) {
	t.Helper()

	confPath := filepath.Join(dir, "certificates.toml")
	content := "[[tls.certificates]]\n" +
		"  certFile = \"" + filepath.Join(certsDir, "cert.pem") + "\"\n" +
		"  keyFile = \"" + filepath.Join(certsDir, "key.pem") + "\"\n"
	require.NoError(t, os.WriteFile(confPath, []byte(content), 0o600))
}
