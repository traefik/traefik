package tls

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
)

func generateTestCert(t *testing.T, domain string) (types.FileOrContent, types.FileOrContent, []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: domain},
		DNSNames:     []string{domain},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return types.FileOrContent(certPEM), types.FileOrContent(keyPEM), der
}

// A handshake holds the store pointer captured by Manager.Get, and resolves
// certificates without the manager lock while UpdateConfigs updates that same
// store in place. No cache entry may point to a certificate from the replaced
// configuration once UpdateConfigs has returned.
func TestManager_UpdateConfigs_ReusedStoreNoStaleCache(t *testing.T) {
	type gen struct {
		configs []*CertAndStores
		der     []byte
	}

	var padding []*CertAndStores
	for i := range 20 {
		cert, key, _ := generateTestCert(t, fmt.Sprintf("pad%d.other.test", i))
		padding = append(padding, &CertAndStores{Certificate: Certificate{CertFile: cert, KeyFile: key}})
	}

	var gens [2]gen
	for i := range gens {
		cert, key, der := generateTestCert(t, "*.example.com")
		configs := append([]*CertAndStores{{Certificate: Certificate{CertFile: cert, KeyFile: key}}}, padding...)
		gens[i] = gen{configs: configs, der: der}
	}

	tlsManager := NewManager(nil)
	tlsManager.UpdateConfigs(t.Context(), nil, nil, gens[0].configs)

	store := tlsManager.GetStore(DefaultTLSStoreName)
	require.NotNil(t, store)

	var sni atomic.Uint64
	staleUpdates := 0
	const updates = 50

	for i := 1; i <= updates; i++ {
		want := gens[i%2]

		stop := make(chan struct{})
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				for {
					select {
					case <-stop:
						return
					default:
					}
					store.GetBestCertificate(&tls.ClientHelloInfo{ServerName: fmt.Sprintf("h%d.example.com", sni.Add(1))})
					store.GetDefaultCertificate()
				}
			})
		}

		tlsManager.UpdateConfigs(t.Context(), nil, nil, want.configs)
		close(stop)
		wg.Wait()

		require.Same(t, store, tlsManager.GetStore(DefaultTLSStoreName))

		for _, item := range store.CertCache.Items() {
			if !bytes.Equal(item.Object.(*CertificateData).Certificate.Certificate[0], want.der) {
				staleUpdates++
				break
			}
		}
	}

	t.Logf("updates leaving a stale cache entry: %d/%d", staleUpdates, updates)
	require.Zero(t, staleUpdates)
}
