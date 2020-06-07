// +build go1.8

package brotlihandler

import "net/http"

// Push initiates an HTTP/2 server push. This constructs a synthetic
// request using the given target and options, serializes that request
// into a PUSH_PROMISE frame, then dispatches that request using the
// server's request handler. If opts is nil, default options are used.
//
// The target must either be an absolute path (like "/path") or an absolute
// URL that contains a valid host and the same scheme as the parent request.
// If the target is a path, it will inherit the scheme and host of the
// parent request.
//
// The HTTP/2 spec disallows recursive pushes and cross-authority pushes.
// Push may or may not detect these invalid pushes; however, invalid
// pushes will be detected and canceled by conforming clients.
//
// Handlers that wish to push URL X should call Push before sending any
// data that may trigger a request for URL X. This avoids a race where the
// client issues requests for X before receiving the PUSH_PROMISE for X.
//
// Push will run in a separate goroutine making the order of arrival
// non-deterministic. Any required synchronization needs to be implemented
// by the caller.
//
// Push returns ErrNotSupported if the client has disabled push or if push
// is not supported on the underlying connection.
func (w *compressResponseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if ok && pusher != nil {
		return pusher.Push(target, setAcceptEncodingForPushOptions(opts))
	}
	return http.ErrNotSupported
}

func setAcceptEncodingForPushOptions(opts *http.PushOptions) *http.PushOptions {
	if opts == nil {
		opts = &http.PushOptions{
			Header: http.Header{
				"Accept-Encoding": []string{"br", "gzip"},
			},
		}
		return opts
	}

	if opts.Header == nil {
		opts.Header = http.Header{
			"Accept-Encoding": []string{"br", "gzip"},
		}
		return opts
	}

	if encoding := opts.Header.Get("Accept-Encoding"); encoding == "" {
		opts.Header.Add("Accept-Encoding", "br,gzip")
		return opts
	}

	return opts
}
