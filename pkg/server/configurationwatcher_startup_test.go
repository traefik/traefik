package server

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/safe"
)

func startupSnapshot(name string) dynamic.Message {
	return dynamic.Message{ProviderName: name, Configuration: &dynamic.Configuration{HTTP: &dynamic.HTTPConfiguration{Routers: map[string]*dynamic.Router{}}}}
}

func TestInitialConfigurationRequiresSnapshotsAndCompletedListeners(t *testing.T) {
	for _, names := range [][]string{{"internal", "file", "le.acme"}, {"internal", "http", "docker", "swarm", "dns.acme"}} {
		t.Run(names[1], func(t *testing.T) {
			pool := safe.NewPool(t.Context())
			t.Cleanup(pool.Stop)
			watcher := NewConfigurationWatcher(pool, nil, nil, "internal", false)
			watcher.SetRequiredProviders(names...)
			applied := make(chan struct{}, len(names))
			release := make(chan struct{})
			var releaseOnce sync.Once
			defer releaseOnce.Do(func() { close(release) })
			count := 0
			watcher.AddListener(func(dynamic.Configuration) {
				count++
				applied <- struct{}{}
				if count == len(names) {
					<-release
				}
			})
			pool.GoCtx(watcher.receiveConfigurations)
			pool.GoCtx(watcher.applyConfigurations)
			result := make(chan error, 1)
			go func() { result <- watcher.WaitInitialConfiguration(t.Context()) }()
			for i, name := range names {
				watcher.allProvidersConfigs <- startupSnapshot(name)
				// Consume acknowledgements to force independent partial applications.
				if i < len(names)-1 {
					<-applied
				}
				select {
				case err := <-result:
					t.Fatalf("ready before complete application: %v", err)
				default:
				}
			}
			// The last callback must finish before startup is released.
			<-applied
			select {
			case err := <-result:
				t.Fatalf("ready before listener completed: %v", err)
			default:
			}
			releaseOnce.Do(func() { close(release) })
			require.NoError(t, <-result)
		})
	}
}

func TestInitialConfigurationErrors(t *testing.T) {
	for _, complete := range []bool{false, true} {
		t.Run(map[bool]string{false: "partial", true: "complete"}[complete], func(t *testing.T) {
			pool := safe.NewPool(t.Context())
			t.Cleanup(pool.Stop)
			watcher := NewConfigurationWatcher(pool, nil, nil, "internal", false)
			watcher.SetRequiredProviders("internal", "file")
			failure := errors.New("invalid TLS configuration")
			feedback := make(chan struct{}, 2)
			watcher.AddCriticalListener(func(dynamic.Configuration) error { return failure })
			watcher.AddListener(func(dynamic.Configuration) { feedback <- struct{}{} })
			pool.GoCtx(watcher.receiveConfigurations)
			pool.GoCtx(watcher.applyConfigurations)
			watcher.allProvidersConfigs <- startupSnapshot("internal")
			<-feedback
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			result := make(chan error, 1)
			go func() { result <- watcher.WaitInitialConfiguration(ctx) }()
			if complete {
				watcher.allProvidersConfigs <- startupSnapshot("file")
				<-feedback
				require.ErrorIs(t, <-result, failure)
			} else {
				cancel()
				require.ErrorIs(t, <-result, context.Canceled)
			}
		})
	}
}

func TestInitialConfigurationWaitCancellation(t *testing.T) {
	pool := safe.NewPool(t.Context())
	defer pool.Stop()
	watcher := NewConfigurationWatcher(pool, nil, nil, "internal", false)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	assert.ErrorIs(t, watcher.WaitInitialConfiguration(ctx), context.Canceled)
}

func TestInitialConfigurationEmptyCertificateSnapshot(t *testing.T) {
	pool := safe.NewPool(t.Context())
	defer pool.Stop()
	watcher := NewConfigurationWatcher(pool, nil, nil, "", false)
	watcher.SetRequiredProviders("le.acme")
	pool.GoCtx(watcher.receiveConfigurations)
	pool.GoCtx(watcher.applyConfigurations)
	watcher.allProvidersConfigs <- startupSnapshot("le.acme")
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	require.NoError(t, watcher.WaitInitialConfiguration(ctx))
}
