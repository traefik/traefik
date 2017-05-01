package middlewares

import (
	"net/http"
)

//Transformer is a hacky way to get metrics middleware with frontend
type Transformer struct {
	next    http.Handler
	metrics *MetricsWrapper
}

//NewTransformer creates new proxy for given Metrics and Handler
func NewTransformer(next http.Handler, metrics *MetricsWrapper) *Transformer {
	return &Transformer{next, metrics}
}

func (sb *Transformer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	sb.metrics.ServeHTTP(rw, r, sb.next.ServeHTTP)
}
