package spnegoclient

import (
	"context"
	"net/http"

	"github.com/containous/alice"
        "github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"

        "gopkg.in/jcmturner/gokrb5.v7/client"
        "gopkg.in/jcmturner/gokrb5.v7/config"
        "gopkg.in/jcmturner/gokrb5.v7/credentials"
        "gopkg.in/jcmturner/gokrb5.v7/spnego"
)

const (
	typeName       = "SpnegoClient"
	nameService    = "spnegoclient-service"
)

type spnegoClientMiddleware struct {
	next    http.Handler
        config	*dynamic.SpnegoClient
        client  *client.Client
}

// NewSpnegoClientMiddleware creates a new middleware to add SPNEGO header to outgoing http requests.
func NewSpnegoClientMiddleware(ctx context.Context, next http.Handler, conf *dynamic.SpnegoClient) (http.Handler, error) {
        logger := log.FromContext(middlewares.GetLoggerCtx(ctx, nameService, typeName))
	logger.Debug("Creating middleware")

        ccache, err := credentials.LoadCCache(conf.CCachePath)
        if err != nil {
                logger.Errorf("error loading CredentialCache from file %s: %s", conf.CCachePath, err)
                return nil, err
        }
        krbconf, err := config.Load(conf.KrbConfPath)
        if err != nil {
                logger.Errorf("error loading Kerberos config from file %s: %s", conf.KrbConfPath, err)
		return nil, err
        }
        krbclient, err := client.NewClientFromCCache(ccache, krbconf)
        if err != nil {
                logger.Errorf("error creating Kerberos Client %s", err)
		return nil, err
        }
        return &spnegoClientMiddleware{
                next:           next,
                config:         conf,
                client:         krbclient,
        }, nil

}

// WrapHandler Wraps the handler to alice.Constructor.
func WrapHandler(ctx context.Context, config *dynamic.SpnegoClient) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return NewSpnegoClientMiddleware(ctx, next, config)
	}
}

func (m *spnegoClientMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
        err := spnego.SetSPNEGOHeader(m.client, req, m.config.Spn)
        if err != nil {
                log.WithoutContext().Errorf("error setting SPNEGO Header %s", err)
        }
	m.next.ServeHTTP(rw, req)
}

