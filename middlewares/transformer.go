package middlewares

import (
	"net/http"
)

type Transformer struct {
	next    http.Handler
	metrics *MetricsWrapper
}

func NewTransformer(next http.Handler, metrics *MetricsWrapper) *Transformer {
	return &Transformer{next, metrics}
}

func (sb *Transformer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	sb.metrics.ServeHTTP(rw, r, sb.next.ServeHTTP)
}
