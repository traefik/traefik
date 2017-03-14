package xhandler

import (
	"net/http"

	"golang.org/x/net/context"
)

// Chain is an helper to chain middleware handlers together for an easier
// management.
type Chain []func(next HandlerC) HandlerC

// UseC appends a context-aware handler to the middleware chain.
func (c *Chain) UseC(f func(next HandlerC) HandlerC) {
	*c = append(*c, f)
}

// Use appends a standard http.Handler to the middleware chain without
// lossing track of the context when inserted between two context aware handlers.
//
// Caveat: the f function will be called on each request so you are better to put
// any initialization sequence outside of this function.
func (c *Chain) Use(f func(next http.Handler) http.Handler) {
	xf := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			n := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTPC(ctx, w, r)
			})
			f(n).ServeHTTP(w, r)
		})
	}
	*c = append(*c, xf)
}

// Handler wraps the provided final handler with all the middleware appended to
// the chain and return a new standard http.Handler instance.
// The context.Background() context is injected automatically.
func (c Chain) Handler(xh HandlerC) http.Handler {
	ctx := context.Background()
	return c.HandlerCtx(ctx, xh)
}

// HandlerFC is an helper to provide a function (HandlerFuncC) to Handler().
//
// HandlerFC is equivalent to:
//  c.Handler(xhandler.HandlerFuncC(xhc))
func (c Chain) HandlerFC(xhf HandlerFuncC) http.Handler {
	ctx := context.Background()
	return c.HandlerCtx(ctx, HandlerFuncC(xhf))
}

// HandlerH is an helper to provide a standard http handler (http.HandlerFunc)
// to Handler(). Your final handler won't have access the context though.
func (c Chain) HandlerH(h http.Handler) http.Handler {
	ctx := context.Background()
	return c.HandlerCtx(ctx, HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}))
}

// HandlerF is an helper to provide a standard http handler function
// (http.HandlerFunc) to Handler(). Your final handler won't have access
// the context though.
func (c Chain) HandlerF(hf http.HandlerFunc) http.Handler {
	ctx := context.Background()
	return c.HandlerCtx(ctx, HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		hf(w, r)
	}))
}

// HandlerCtx wraps the provided final handler with all the middleware appended to
// the chain and return a new standard http.Handler instance.
func (c Chain) HandlerCtx(ctx context.Context, xh HandlerC) http.Handler {
	return New(ctx, c.HandlerC(xh))
}

// HandlerC wraps the provided final handler with all the middleware appended to
// the chain and returns a HandlerC instance.
func (c Chain) HandlerC(xh HandlerC) HandlerC {
	for i := len(c) - 1; i >= 0; i-- {
		xh = c[i](xh)
	}
	return xh
}

// HandlerCF wraps the provided final handler func with all the middleware appended to
// the chain and returns a HandlerC instance.
//
// HandlerCF is equivalent to:
//  c.HandlerC(xhandler.HandlerFuncC(xhc))
func (c Chain) HandlerCF(xhc HandlerFuncC) HandlerC {
	return c.HandlerC(HandlerFuncC(xhc))
}
