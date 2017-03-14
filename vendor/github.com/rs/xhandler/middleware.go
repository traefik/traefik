package xhandler

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// CloseHandler returns a Handler cancelling the context when the client
// connection close unexpectedly.
func CloseHandler(next HandlerC) HandlerC {
	return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Cancel the context if the client closes the connection
		if wcn, ok := w.(http.CloseNotifier); ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			defer cancel()

			notify := wcn.CloseNotify()
			go func() {
				select {
				case <-notify:
					cancel()
				case <-ctx.Done():
				}
			}()
		}

		next.ServeHTTPC(ctx, w, r)
	})
}

// TimeoutHandler returns a Handler which adds a timeout to the context.
//
// Child handlers have the responsability to obey the context deadline and to return
// an appropriate error (or not) response in case of timeout.
func TimeoutHandler(timeout time.Duration) func(next HandlerC) HandlerC {
	return func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx, _ = context.WithTimeout(ctx, timeout)
			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// If is a special handler that will skip insert the condNext handler only if a condition
// applies at runtime.
func If(cond func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, condNext func(next HandlerC) HandlerC) func(next HandlerC) HandlerC {
	return func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if cond(ctx, w, r) {
				condNext(next).ServeHTTPC(ctx, w, r)
			} else {
				next.ServeHTTPC(ctx, w, r)
			}
		})
	}
}
