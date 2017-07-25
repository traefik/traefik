package http

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Server wraps an endpoint and implements http.Handler.
type Server struct {
	ctx          context.Context
	e            endpoint.Endpoint
	dec          DecodeRequestFunc
	enc          EncodeResponseFunc
	before       []RequestFunc
	after        []ServerResponseFunc
	errorEncoder ErrorEncoder
	logger       log.Logger
}

// NewServer constructs a new server, which implements http.Server and wraps
// the provided endpoint.
func NewServer(
	ctx context.Context,
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...ServerOption,
) *Server {
	s := &Server{
		ctx:          ctx,
		e:            e,
		dec:          dec,
		enc:          enc,
		errorEncoder: defaultErrorEncoder,
		logger:       log.NewNopLogger(),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// ServerOption sets an optional parameter for servers.
type ServerOption func(*Server)

// ServerBefore functions are executed on the HTTP request object before the
// request is decoded.
func ServerBefore(before ...RequestFunc) ServerOption {
	return func(s *Server) { s.before = before }
}

// ServerAfter functions are executed on the HTTP response writer after the
// endpoint is invoked, but before anything is written to the client.
func ServerAfter(after ...ServerResponseFunc) ServerOption {
	return func(s *Server) { s.after = after }
}

// ServerErrorEncoder is used to encode errors to the http.ResponseWriter
// whenever they're encountered in the processing of a request. Clients can
// use this to provide custom error formatting and response codes. By default,
// errors will be written as plain text with an appropriate, if generic,
// status code.
func ServerErrorEncoder(ee ErrorEncoder) ServerOption {
	return func(s *Server) { s.errorEncoder = ee }
}

// ServerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged.
func ServerErrorLogger(logger log.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

// ServeHTTP implements http.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := s.ctx

	for _, f := range s.before {
		ctx = f(ctx, r)
	}

	request, err := s.dec(ctx, r)
	if err != nil {
		s.logger.Log("err", err)
		s.errorEncoder(ctx, Error{Domain: DomainDecode, Err: err}, w)
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		s.errorEncoder(ctx, Error{Domain: DomainDo, Err: err}, w)
		return
	}

	for _, f := range s.after {
		ctx = f(ctx, w)
	}

	if err := s.enc(ctx, w, response); err != nil {
		s.logger.Log("err", err)
		s.errorEncoder(ctx, Error{Domain: DomainEncode, Err: err}, w)
		return
	}
}

// ErrorEncoder is responsible for encoding an error to the ResponseWriter.
//
// In the server implementation, only kit/transport/http.Error values are ever
// passed to this function, so you might be tempted to have this function take
// one of those directly. But, users are encouraged to use custom ErrorEncoders
// to encode all HTTP errors to their clients, and so may want to pass and check
// for their own error types. See the example shipping/handling service.
type ErrorEncoder func(ctx context.Context, err error, w http.ResponseWriter)

func defaultErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	switch e := err.(type) {
	case Error:
		switch e.Domain {
		case DomainDecode:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case DomainDo:
			http.Error(w, err.Error(), http.StatusServiceUnavailable) // too aggressive?
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
