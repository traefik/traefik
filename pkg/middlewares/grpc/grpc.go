package grpc

import (
	"context"
	"fmt"
	proto "github.com/containous/traefik/v2/api/middleware/grpc"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/containous/traefik/v2/pkg/types"
	"github.com/hashicorp/go-uuid"
	"github.com/opentracing/opentracing-go/ext"
	grpcLib "google.golang.org/grpc"
	"net/http"
	"time"
)

const (
	typeName       = "GRPC"
	defaultTimeout = types.Duration(time.Millisecond * 100)
)

var (
	defaultGrpcDialOptions = []grpcLib.DialOption{
		grpcLib.WithInsecure(),
	}

	defaultGrpcCallOptions = []grpcLib.CallOption{
		grpcLib.WaitForReady(false),
	}
)

type grpc struct {
	next http.Handler
	name string

	timeout time.Duration

	serverID     string
	passToClient struct {
		Method     bool
		RequestURI bool
		RemoteAddr bool
		Headers    []string
	}

	grpcDialOptions []grpcLib.DialOption
	grpcCallOptions []grpcLib.CallOption

	conn       *grpcLib.ClientConn
	grpcClient proto.MiddlewareClient
}

// New creates a new handler.
func New(ctx context.Context, next http.Handler, config dynamic.GRPC, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")
	var result *grpc

	if config.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	result = &grpc{
		next:     next,
		name:     name,
		timeout:  time.Duration(config.Timeout),
		serverID: config.ServerID,
		passToClient: struct {
			Method     bool
			RequestURI bool
			RemoteAddr bool
			Headers    []string
		}{
			Method:     config.PassToClient.Method,
			RequestURI: config.PassToClient.RequestURI,
			RemoteAddr: config.PassToClient.RemoteAddr,
			Headers:    config.PassToClient.Headers,
		},
		// todo: change DialOptions with config
		grpcDialOptions: defaultGrpcDialOptions,
		// todo: change CallOptions with config
		grpcCallOptions: defaultGrpcCallOptions,
	}

	if result.serverID == "" {
		var err error
		result.serverID, err = uuid.GenerateUUID()
		if err != nil {
			return nil, err
		}
	}

	var err error

	result.conn, err = grpcLib.Dial(config.Address, result.grpcDialOptions...)
	if err != nil {
		return nil, err
	}

	result.grpcClient = proto.NewMiddlewareClient(result.conn)

	return result, nil
}

func (g *grpc) GetTracingInformation() (string, ext.SpanKindEnum) {
	return g.name, tracing.SpanKindNoneEnum
}
