package kubernetesfields

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
)

const (
	typeName = "KubernetesFields"
)

type kubernetesFields struct {
	next           http.Handler
	middlewareName string
	namespace      string
	kind           string
	name           string
}

// New creates a Kubernetes fields middleware.
func New(ctx context.Context, next http.Handler, config dynamic.KubernetesFields, middlewareName string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, middlewareName, typeName)
	logger.Debug().Msg("Creating middleware")

	return &kubernetesFields{
		next:           next,
		middlewareName: middlewareName,
		namespace:      config.Namespace,
		kind:           config.Kind,
		name:           config.Name,
	}, nil
}

func (k *kubernetesFields) GetTracingInformation() (string, string) {
	return k.middlewareName, typeName
}

func (k *kubernetesFields) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if logDataTable, ok := req.Context().Value(accesslog.DataTableKey).(*accesslog.LogData); ok {
		if logDataTable.Core == nil {
			logDataTable.Core = make(accesslog.CoreLogData)
		}
		logDataTable.Core[accesslog.KubernetesNamespace] = k.namespace
		logDataTable.Core[accesslog.KubernetesKind] = k.kind
		logDataTable.Core[accesslog.KubernetesName] = k.name
	}

	// If there is a next, call it.
	if k.next != nil {
		k.next.ServeHTTP(rw, req)
	}
}
