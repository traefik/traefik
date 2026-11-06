package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/safe"
)

func TestInitialStoredSnapshotPrecedesRenewal(t *testing.T) {
	pool := safe.NewPool(t.Context())
	release := make(chan struct{})
	contacted := make(chan struct{}, 1)
	ca := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case contacted <- struct{}{}:
		default:
		}
		<-release
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ca.Close()
	defer pool.Stop()
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	config := &Configuration{}
	config.SetDefaults()
	config.CAServer = ca.URL
	caPath := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Certificate().Raw}), 0o600))
	config.CACertificates = []string{caPath}
	config.KeyType = "EC256"
	cached := &CertAndStore{Certificate: Certificate{Certificate: []byte("invalid cached certificate"), Key: []byte("cached key")}, Store: "default"}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	account := &Account{Email: "startup@example.invalid", PrivateKey: x509.MarshalPKCS1PrivateKey(key)}
	provider := &Provider{Configuration: config, ResolverName: "stored", Store: &startupStore{account: account, certs: []*CertAndStore{cached}}}
	require.NoError(t, provider.Init())
	messages := make(chan dynamic.Message, 1)
	finished := make(chan error, 1)
	go func() { finished <- provider.Provide(messages, pool) }()
	// An invalid cached certificate forces the subsequent renewal path. The local
	// CA blocks it so moving renewal before publication makes this test fail.
	select {
	case msg := <-messages:
		require.Equal(t, "stored.acme", msg.ProviderName)
		require.Len(t, msg.Configuration.TLS.Certificates, 1)
		require.Equal(t, string(cached.Certificate.Certificate), string(msg.Configuration.TLS.Certificates[0].CertFile))
	case <-time.After(time.Second):
		t.Error("initial stored certificate snapshot waited for renewal")
	}
	select {
	case <-contacted:
	case <-time.After(time.Second):
		t.Error("renewal did not reach the blocking CA positive control")
	}
	// Release and collect Provide before closing the pool or its HTTP fixture.
	releaseOnce.Do(func() { close(release) })
	require.NoError(t, <-finished)
}

type startupStore struct {
	account *Account
	certs   []*CertAndStore
}

func (s *startupStore) GetAccount(string) (*Account, error)             { return s.account, nil }
func (s *startupStore) SaveAccount(string, *Account) error              { return nil }
func (s *startupStore) GetCertificates(string) ([]*CertAndStore, error) { return s.certs, nil }
func (s *startupStore) SaveCertificates(string, []*CertAndStore) error  { return nil }
