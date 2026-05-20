package tcp

import (
	"context"
	"crypto/tls"
)

// TLSHandler handles TLS connections.
type TLSHandler struct {
	Next       Handler
	Config     *tls.Config
	ConfigName string
}

// ServeTCP terminates the TLS connection.
func (t *TLSHandler) ServeTCP(conn WriteCloser) {
	t.Next.ServeTCP(tls.Server(TLSConnWithName{WriteCloser: conn, ConfigName: t.ConfigName}, t.Config))
}

type TLSConnWithName struct {
	WriteCloser

	ConfigName string
}

func (t *TLSConnWithName) GetConfigName() string {
	return t.ConfigName
}

type optionKey struct{}

func AddTLSOptionsNameInContext(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, optionKey{}, name)
}

func GetTLSOptionsNameInContext(ctx context.Context) string {
	if name, ok := ctx.Value(optionKey{}).(string); ok {
		return name
	}

	return ""
}
