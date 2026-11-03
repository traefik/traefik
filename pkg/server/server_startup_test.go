package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/safe"
	tcprouter "github.com/traefik/traefik/v3/pkg/server/router/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestServerStartupWaitsForInitialConfiguration(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	pool := safe.NewPool(ctx)
	watcher := NewConfigurationWatcher(pool, &mockProvider{messages: []dynamic.Message{{
		ProviderName:  "internal",
		Configuration: &dynamic.Configuration{HTTP: &dynamic.HTTPConfiguration{Routers: map[string]*dynamic.Router{}}},
	}}}, nil, "internal", false)
	entered, release := make(chan struct{}), make(chan struct{})
	watcher.AddListener(func(dynamic.Configuration) {
		close(entered)
		<-release
	})
	srv := NewServer(pool, nil, nil, watcher, nil)
	started := make(chan struct{})
	var startErr error
	go func() {
		startErr = srv.Start(ctx)
		close(started)
	}()
	<-entered
	select {
	case <-started:
		t.Error("server started before the initial configuration was applied")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Error("server did not start after configuration was applied")
	}
	require.NoError(t, startErr)
	cancel()
	srv.Wait()
	srv.Close()
}

// startupProvider lets tests pause between independent provider snapshots.
type startupProvider struct{ messages chan dynamic.Message }

func (p *startupProvider) Init() error { return nil }
func (p *startupProvider) Provide(out chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-p.messages:
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	})
	return nil
}

type startupListener struct {
	net.Listener

	accepted chan struct{}
}

func (l *startupListener) Accept() (net.Conn, error) {
	select {
	case l.accepted <- struct{}{}:
	default:
	}
	return l.Listener.Accept()
}

func TestServerFirstHandshakeUsesConfiguredCertificate(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	pool := safe.NewPool(ctx)
	config := &static.EntryPoint{Address: "127.0.0.1:0"}
	config.SetDefaults()
	ep, err := NewTCPEntryPoint(ctx, "https", config, nil, nil)
	require.NoError(t, err)
	listener := &startupListener{Listener: ep.listener, accepted: make(chan struct{}, 1)}
	ep.listener = listener
	p := &startupProvider{messages: make(chan dynamic.Message)}
	watcher := NewConfigurationWatcher(pool, p, nil, "internal", false)
	watcher.SetRequiredProviders("internal", "file")
	manager := traefiktls.NewManager(nil)
	watcher.AddCriticalListener(func(conf dynamic.Configuration) error {
		return manager.UpdateConfigsWithError(ctx, conf.TLS.Stores, conf.TLS.Options, conf.TLS.Certificates)
	})
	watcher.AddListener(func(dynamic.Configuration) {
		cfg, err := manager.Get("default", "default")
		if err != nil {
			return
		}
		router, err := tcprouter.NewRouter(nil)
		if err != nil {
			return
		}
		router.SetHTTPSHandler(http.NotFoundHandler(), cfg)
		ep.SwitchRouter(router)
	})
	srv := NewServer(pool, TCPEntryPoints{"https": ep}, nil, watcher, nil)
	done := make(chan struct{})
	var startErr error
	go func() { startErr = srv.Start(ctx); close(done) }()
	t.Cleanup(func() {
		cancel()
		<-done
		if startErr == nil {
			srv.Wait()
		}
		srv.Close()
	})
	p.messages <- startupSnapshot("internal")
	select {
	case <-listener.accepted:
		t.Fatal("Accept called before file snapshot")
	case <-time.After(20 * time.Millisecond):
	}
	cert, key := startupCertificate(t)
	snapshot := startupSnapshot("file")
	snapshot.Configuration.TLS = &dynamic.TLSConfiguration{Certificates: []*traefiktls.CertAndStores{{CertFile: types.FileOrContent(cert), KeyFile: types.FileOrContent(key)}}}
	p.messages <- snapshot
	<-done
	require.NoError(t, startErr)
	roots := x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM(cert))
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Second}, "tcp", ep.listener.Addr().String(), &tls.Config{RootCAs: roots, ServerName: "startup.example"})
	require.NoError(t, err)
	defer conn.Close()
	require.Equal(t, "startup.example", conn.ConnectionState().PeerCertificates[0].DNSNames[0])
}

func startupCertificate(t *testing.T) ([]byte, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{SerialNumber: big.NewInt(1), DNSNames: []string{"startup.example"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, BasicConstraintsValid: true, IsCA: true}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	encoded, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encoded})
}

func TestServerStartupCancellationClosesBoundListener(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	pool := safe.NewPool(ctx)
	config := &static.EntryPoint{Address: "127.0.0.1:0"}
	config.SetDefaults()
	ep, err := NewTCPEntryPoint(ctx, "https", config, nil, nil)
	require.NoError(t, err)
	address := ep.listener.Addr().String()
	watcher := NewConfigurationWatcher(pool, &startupProvider{messages: make(chan dynamic.Message)}, nil, "internal", false)
	srv := NewServer(pool, TCPEntryPoints{"https": ep}, nil, watcher, nil)
	done := make(chan error, 1)
	go func() { done <- srv.Start(ctx) }()
	cancel()
	require.ErrorIs(t, <-done, context.Canceled)
	srv.Close()
	listener, err := net.Listen("tcp", address)
	require.NoError(t, err)
	require.NoError(t, listener.Close())
}
