package plugins

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/http-wasm/http-wasm-host-go/handler"
	wasm "github.com/http-wasm/http-wasm-host-go/handler/nethttp"
	"github.com/rs/zerolog"
	"github.com/tetratelabs/wazero"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
)

// cachedWasmInstance holds a built http-wasm middleware (with its wazero
// runtime, compiled module and guest instance pool), the host-context applier
// and the hash of the guest config the instance was built from.
//
// Lifetime is reference-counted: each handler built from the instance via
// wrapWithRelease holds a ref; the instance is closed when (a) it has been
// evicted from the cache (its guest config changed, or the middleware was
// replaced by a new entry) and (b) the last handler has been released
// (i.e., garbage-collected). This keeps the wazero runtime alive while any
// in-flight request may still be served by an outgoing handler, while still
// reclaiming runtimes deterministically once all references are gone.
type cachedWasmInstance struct {
	confHash [sha256.Size]byte
	mw       wasm.Middleware
	applyCtx func(context.Context) context.Context
	logger   *zerolog.Logger

	mu       sync.Mutex
	refcount int
	evicted  bool
	closed   bool
}

func (ci *cachedWasmInstance) acquire() {
	ci.mu.Lock()
	ci.refcount++
	ci.mu.Unlock()
}

func (ci *cachedWasmInstance) release(ctx context.Context) {
	ci.mu.Lock()
	ci.refcount--
	shouldClose := ci.refcount == 0 && ci.evicted && !ci.closed
	if shouldClose {
		ci.closed = true
	}
	ci.mu.Unlock()
	if shouldClose {
		ci.doClose(ctx)
	}
}

// markEvicted marks the instance as no longer cached. If no handlers reference
// it any longer, the underlying wazero runtime is closed immediately;
// otherwise close is deferred until the last handler is released (GC'd).
func (ci *cachedWasmInstance) markEvicted(ctx context.Context) {
	ci.mu.Lock()
	if ci.evicted {
		ci.mu.Unlock()
		return
	}
	ci.evicted = true
	shouldClose := ci.refcount == 0 && !ci.closed
	if shouldClose {
		ci.closed = true
	}
	ci.mu.Unlock()
	if shouldClose {
		ci.doClose(ctx)
	}
}

func (ci *cachedWasmInstance) doClose(ctx context.Context) {
	if err := ci.mw.Close(ctx); err != nil {
		ci.logger.Err(err).Msg("[wasm] middleware Close failed")
	} else {
		ci.logger.Debug().Msg("[wasm] middleware Close ok")
	}
}

// wasmInstanceCache memoizes built http-wasm middlewares per middleware name.
//
// A dynamic-configuration reload re-publishes every middleware and rebuilds the
// whole handler chain. Without memoization each reload allocates a brand-new
// wazero runtime (compiling the guest module and creating a new linear-memory
// instance) per middleware, and the previous runtimes are only reclaimed via a
// best-effort finalizer that cannot keep up under frequent reloads — causing an
// unbounded memory leak (see https://github.com/traefik/traefik/issues/11119
// and https://github.com/traefik/traefik/issues/13235).
//
// Entries are keyed by middleware name. When the guest config for a name is
// unchanged across reloads the existing instance is reused and only the
// downstream handler is rebound, so no new runtime is created. When the guest
// config does change, the previous instance is evicted and closed once its
// last outgoing handler has drained (see cachedWasmInstance).
//
// The cache is shared by pointer so it survives the value-copy of
// wasmMiddlewareBuilder that happens when a WasmMiddleware is created.
type wasmInstanceCache struct {
	mu sync.Mutex
	m  map[string]*cachedWasmInstance
}

type wasmMiddlewareBuilder struct {
	path      string
	cache     wazero.CompilationCache
	settings  Settings
	instances *wasmInstanceCache
}

func newWasmMiddlewareBuilder(goPath, moduleName, wasmPath string, settings Settings) (*wasmMiddlewareBuilder, error) {
	ctx := context.Background()
	path := filepath.Join(goPath, "src", moduleName, wasmPath)
	cache := wazero.NewCompilationCache()

	code, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading Wasm binary: %w", err)
	}

	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCompilationCache(cache))
	if _, err = rt.CompileModule(ctx, code); err != nil {
		return nil, fmt.Errorf("compiling guest module: %w", err)
	}
	// This runtime is only used to validate that the guest compiles; it is not
	// reused for serving, so release it instead of leaking it.
	_ = rt.Close(ctx)

	return &wasmMiddlewareBuilder{
		path:      path,
		cache:     cache,
		settings:  settings,
		instances: &wasmInstanceCache{m: make(map[string]*cachedWasmInstance)},
	}, nil
}

func (b wasmMiddlewareBuilder) newMiddleware(config map[string]any, middlewareName string) (pluginMiddleware, error) {
	return &WasmMiddleware{
		middlewareName: middlewareName,
		config:         reflect.ValueOf(config),
		builder:        b,
	}, nil
}

func (b wasmMiddlewareBuilder) newHandler(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, error) {
	h, applyCtx, err := b.buildMiddleware(ctx, next, cfg, middlewareName)
	if err != nil {
		return nil, fmt.Errorf("building Wasm middleware: %w", err)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(rw, req.WithContext(applyCtx(req.Context())))
	}), nil
}

func (b *wasmMiddlewareBuilder) buildMiddleware(ctx context.Context, next http.Handler, cfg reflect.Value, middlewareName string) (http.Handler, func(ctx context.Context) context.Context, error) {
	guestConfig, err := marshalGuestConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	confHash := sha256.Sum256(guestConfig)

	// Fast path: reuse the cached instance when the guest config is unchanged,
	// rebinding only the (cheap) downstream handler. No new runtime is created.
	b.instances.mu.Lock()
	if ci, ok := b.instances.m[middlewareName]; ok && ci.confHash == confHash {
		ci.acquire()
		b.instances.mu.Unlock()
		return wrapWithRelease(ci, ci.mw.NewHandler(ctx, next)), ci.applyCtx, nil
	}
	b.instances.mu.Unlock()

	// Slow path: build a fresh instance outside the lock (compilation and host
	// instantiation are expensive and must not serialize unrelated middlewares).
	newCI, err := b.newInstance(ctx, guestConfig, confHash, middlewareName)
	if err != nil {
		return nil, nil, err
	}

	b.instances.mu.Lock()
	// Another goroutine may have built the same instance while we were
	// building; prefer the already-cached one and discard the redundant build.
	if ci, ok := b.instances.m[middlewareName]; ok && ci.confHash == confHash {
		ci.acquire()
		b.instances.mu.Unlock()
		_ = newCI.mw.Close(ctx)
		return wrapWithRelease(ci, ci.mw.NewHandler(ctx, next)), ci.applyCtx, nil
	}
	// Replace any existing entry for this name and arrange for the previous
	// instance's wazero runtime to be closed once all of its remaining handlers
	// have been released (see cachedWasmInstance.markEvicted).
	old := b.instances.m[middlewareName]
	b.instances.m[middlewareName] = newCI
	newCI.acquire()
	b.instances.mu.Unlock()

	if old != nil {
		old.markEvicted(ctx)
	}

	return wrapWithRelease(newCI, newCI.mw.NewHandler(ctx, next)), newCI.applyCtx, nil
}

// instanceHandler is the per-Build handler returned to traefik's middleware
// chain. It exists as a concrete struct (rather than an http.HandlerFunc) so
// that a finalizer can be attached: when this handler is no longer referenced
// anywhere in the chain (e.g., a previous dynamic configuration's chain has
// been swapped out), the finalizer releases its hold on the cached wasm
// instance, allowing an evicted instance to be closed once its last handler
// drains.
type instanceHandler struct {
	h        http.Handler
	instance *cachedWasmInstance
}

func (ih *instanceHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ih.h.ServeHTTP(rw, req)
}

func wrapWithRelease(ci *cachedWasmInstance, h http.Handler) http.Handler {
	ih := &instanceHandler{h: h, instance: ci}
	runtime.SetFinalizer(ih, func(ih *instanceHandler) {
		ih.instance.release(context.Background())
	})
	return ih
}

// marshalGuestConfig serializes the plugin guest config to the JSON form passed
// to the wasm guest, or returns nil when no config is provided.
func marshalGuestConfig(cfg reflect.Value) ([]byte, error) {
	i := cfg.Interface()
	if i == nil {
		return nil, nil
	}

	config, ok := i.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("could not type assert config: %T", i)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}

	return data, nil
}

// newInstance builds a fresh http-wasm middleware backed by its own wazero
// runtime. It is only called on a cache miss (new middleware) or when a guest
// config changes.
func (b *wasmMiddlewareBuilder) newInstance(ctx context.Context, guestConfig []byte, confHash [sha256.Size]byte, middlewareName string) (*cachedWasmInstance, error) {
	code, err := os.ReadFile(b.path)
	if err != nil {
		return nil, fmt.Errorf("loading binary: %w", err)
	}

	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCompilationCache(b.cache))

	guestModule, err := rt.CompileModule(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("compiling guest module: %w", err)
	}

	applyCtx, err := InstantiateHost(ctx, rt, guestModule, b.settings)
	if err != nil {
		return nil, fmt.Errorf("instantiating host module: %w", err)
	}

	logger := middlewares.GetLogger(ctx, middlewareName, "wasm")

	config := wazero.NewModuleConfig().WithSysWalltime().WithStartFunctions("_start", "_initialize")
	for _, env := range b.settings.Envs {
		config = config.WithEnv(env, os.Getenv(env))
	}

	if len(b.settings.Mounts) > 0 {
		fsConfig := wazero.NewFSConfig()
		for _, mount := range b.settings.Mounts {
			withDir := fsConfig.WithDirMount
			prefix, readOnly := strings.CutSuffix(mount, ":ro")
			if readOnly {
				withDir = fsConfig.WithReadOnlyDirMount
			}
			parts := strings.Split(prefix, ":")
			switch {
			case len(parts) == 1:
				fsConfig = withDir(parts[0], parts[0])
			case len(parts) == 2:
				fsConfig = withDir(parts[0], parts[1])
			default:
				return nil, fmt.Errorf("invalid directory %q", mount)
			}
		}
		config = config.WithFSConfig(fsConfig)
	}

	opts := []handler.Option{
		handler.ModuleConfig(config),
		handler.Logger(logs.NewWasmLogger(logger)),
	}

	if guestConfig != nil {
		opts = append(opts, handler.GuestConfig(guestConfig))
	}

	opts = append(opts, handler.Runtime(func(ctx context.Context) (wazero.Runtime, error) {
		return rt, nil
	}))

	mw, err := wasm.NewMiddleware(applyCtx(ctx), code, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating middleware: %w", err)
	}

	return &cachedWasmInstance{
		confHash: confHash,
		mw:       mw,
		applyCtx: applyCtx,
		logger:   logger,
	}, nil
}

// WasmMiddleware is an HTTP handler plugin wrapper.
type WasmMiddleware struct {
	middlewareName string
	config         reflect.Value
	builder        wasmMiddlewareBuilder
}

// NewHandler creates a new HTTP handler.
func (m WasmMiddleware) NewHandler(ctx context.Context, next http.Handler) (http.Handler, error) {
	return m.builder.newHandler(ctx, next, m.config, m.middlewareName)
}
