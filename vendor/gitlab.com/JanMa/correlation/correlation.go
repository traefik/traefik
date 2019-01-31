package correlation

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lucsky/cuid"
)

const (
	// UUID generate a random UUID as correlation id.
	UUID correlationType = iota + 1
	// CUID generate a CUID as correlation id.
	CUID
	// Random generate a pseudo random Int64 as correlation id.
	Random
	// Custom pass a custom string as correlation id.
	Custom
	// Time use the elapsed nanoseconds since the Epoch as correlation id.
	Time

	ctxCorrelationHeaderKey = correlationCtxKey("CorrelationResponseHeader")
	correlationIDHeader     = "X-Correlation-ID"
)

type correlationCtxKey string
type correlationType int

//Options is a struct for specifying configuration options.
type Options struct {
	// CorrelationHeaderName the name of the header to be used as correlation id. Defaults to `X-Correlation-ID`.
	CorrelationHeaderName string
	// CorrelationIDType the type of correlation id to generate. Defaults to `correlation.UUID`.
	CorrelationIDType correlationType
	// CorrelationCustomString the value to use when using a custom correlation id. Default is empty.
	CorrelationCustomString string
}

// Correlation is a middleware that adds correlation ids to requests. A correlation.Options struct can
// be provided t override the default configuration values.
type Correlation struct {
	opt  Options
	rand rand.Source
}

//New returns a new correlation struct with the provides options.
func New(opt Options) *Correlation {
	return &Correlation{
		opt:  opt,
		rand: rand.NewSource(time.Now().UnixNano()),
	}
}

//Handler for integration with net/http.
func (c *Correlation) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := c.Process(w, r)

		if err != nil {
			return
		}

		h.ServeHTTP(w, r)
	})
}

// HandlerForRequestOnly for integration with net/http.
func (c *Correlation) HandlerForRequestOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, err := c.processRequest(w, r)
		if err != nil {
			return
		}

		ctx := context.WithValue(r.Context(), ctxCorrelationHeaderKey, response)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandlerFuncWithNext for integration with github.com/urfave/negroni.
func (c *Correlation) HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	err := c.Process(w, r)

	if err == nil && next != nil {
		next(w, r)
	}
}

// HandlerFuncWithNextForRequestOnly for integration with github.com/urfave/negroni.
func (c *Correlation) HandlerFuncWithNextForRequestOnly(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	response, err := c.processRequest(w, r)

	if err == nil && next != nil {
		ctx := context.WithValue(r.Context(), ctxCorrelationHeaderKey, response)

		next(w, r.WithContext(ctx))
	}
}

//Process processes the incoming request.
func (c *Correlation) Process(w http.ResponseWriter, r *http.Request) error {
	response, err := c.processRequest(w, r)
	if response != nil {
		for key, values := range response {
			for _, value := range values {
				w.Header().Set(key, value)
			}
		}
	}

	return err
}

func (c *Correlation) processRequest(w http.ResponseWriter, r *http.Request) (http.Header, error) {
	response := make(http.Header)

	value := func(c *Correlation) string {
		switch c.opt.CorrelationIDType {
		case UUID:
			return uuid.New().String()
		case CUID:
			return cuid.New()
		case Random:
			return strconv.FormatInt(c.rand.Int63(), 10)
		case Custom:
			return c.opt.CorrelationCustomString
		case Time:
			return strconv.FormatInt(time.Now().UnixNano(), 10)
		}
		return uuid.New().String()
	}
	correlationHeader := func(c *Correlation) string {
		if len(c.opt.CorrelationHeaderName) > 0 {
			return c.opt.CorrelationHeaderName
		}
		return correlationIDHeader
	}

	if len(r.Header.Get(correlationHeader(c))) == 0 {
		r.Header.Set(correlationHeader(c), value(c))
	}

	response.Set(correlationHeader(c), r.Header.Get(correlationHeader(c)))

	return response, nil
}

// ModifyResponseHeaders modifies the response for integration with net/http/httputil ReverseProxy.
func (c *Correlation) ModifyResponseHeaders(res *http.Response) error {
	if res != nil && res.Request != nil {
		response := res.Request.Context().Value(ctxCorrelationHeaderKey)
		if response != nil {
			for h, v := range response.(http.Header) {
				if len(v) > 0 {
					res.Header.Set(h, strings.Join(v, ","))
				}
			}
		}
	}
	return nil
}
