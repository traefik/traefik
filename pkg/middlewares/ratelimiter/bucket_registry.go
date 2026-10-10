package ratelimiter

import "sync"

// BucketRegistry caches the token buckets that are shared by every router referencing the
// same rate limit middleware.
//
// It is owned by the middleware builder rather than kept in a package-level variable,
// because a builder is created for each configuration generation. A cached limiter
// therefore cannot outlive the configuration that built it: an edit produces a new builder
// and new limiters, and the previous ones are released with the rest of the router tree.
//
// A package-level registry would break both of those properties. It would hand back a
// limiter carrying the rate captured when it was constructed, silently serving the previous
// limit after a configuration edit, and it would accumulate entries for middlewares that no
// longer exist.
type BucketRegistry struct {
	mu       sync.Mutex
	limiters map[string]limiter
}

// NewBucketRegistry creates a new BucketRegistry.
func NewBucketRegistry() *BucketRegistry {
	return &BucketRegistry{limiters: make(map[string]limiter)}
}

// getOrCreate returns the limiter stored for the given middleware name, calling create to
// build it on first use.
func (b *BucketRegistry) getOrCreate(name string, create func() (limiter, error)) (limiter, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if l, ok := b.limiters[name]; ok {
		return l, nil
	}

	l, err := create()
	if err != nil {
		return nil, err
	}

	b.limiters[name] = l

	return l, nil
}

// sharedLimiter returns the limiter shared for the given middleware name, or a new
// unshared one when no registry is available.
func sharedLimiter(registry *BucketRegistry, name string, create func() (limiter, error)) (limiter, error) {
	if registry == nil {
		return create()
	}

	return registry.getOrCreate(name, create)
}
