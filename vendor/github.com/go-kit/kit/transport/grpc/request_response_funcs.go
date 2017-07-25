package grpc

import (
	"encoding/base64"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

const (
	binHdrSuffix = "-bin"
)

// RequestFunc may take information from an gRPC request and put it into a
// request context. In Servers, BeforeFuncs are executed prior to invoking the
// endpoint. In Clients, BeforeFuncs are executed after creating the request
// but prior to invoking the gRPC client.
type RequestFunc func(context.Context, *metadata.MD) context.Context

// ResponseFunc may take information from a request context and use it to
// manipulate the gRPC metadata header. ResponseFuncs are only executed in
// servers, after invoking the endpoint but prior to writing a response.
type ResponseFunc func(context.Context, *metadata.MD)

// SetResponseHeader returns a ResponseFunc that sets the specified metadata
// key-value pair.
func SetResponseHeader(key, val string) ResponseFunc {
	return func(_ context.Context, md *metadata.MD) {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
	}
}

// SetRequestHeader returns a RequestFunc that sets the specified metadata
// key-value pair.
func SetRequestHeader(key, val string) RequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
		return ctx
	}
}

// EncodeKeyValue sanitizes a key-value pair for use in gRPC metadata headers.
func EncodeKeyValue(key, val string) (string, string) {
	key = strings.ToLower(key)
	if strings.HasSuffix(key, binHdrSuffix) {
		v := base64.StdEncoding.EncodeToString([]byte(val))
		val = string(v)
	}
	return key, val
}
