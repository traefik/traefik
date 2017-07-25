package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	stdopentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc"
	addsvcgrpcclient "github.com/go-kit/kit/examples/addsvc/client/grpc"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc"
)

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		consulAddr   = flag.String("consul.addr", "", "Consul agent address")
		retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)
	flag.Parse()

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		if len(*consulAddr) > 0 {
			consulConfig.Address = *consulAddr
		}
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	// Transport domain.
	tracer := stdopentracing.GlobalTracer() // no-op
	ctx := context.Background()
	r := mux.NewRouter()

	// Now we begin installing the routes. Each route corresponds to a single
	// method: sum, concat, uppercase, and count.

	// addsvc routes.
	{
		// Each method gets constructed with a factory. Factories take an
		// instance string, and return a specific endpoint. In the factory we
		// dial the instance string we get from Consul, and then leverage an
		// addsvc client package to construct a complete service. We can then
		// leverage the addsvc.Make{Sum,Concat}Endpoint constructors to convert
		// the complete service to specific endpoint.

		var (
			tags        = []string{}
			passingOnly = true
			endpoints   = addsvc.Endpoints{}
		)
		{
			factory := addsvcFactory(addsvc.MakeSumEndpoint, tracer, logger)
			subscriber := consulsd.NewSubscriber(client, factory, logger, "addsvc", tags, passingOnly)
			balancer := lb.NewRoundRobin(subscriber)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			endpoints.SumEndpoint = retry
		}
		{
			factory := addsvcFactory(addsvc.MakeConcatEndpoint, tracer, logger)
			subscriber := consulsd.NewSubscriber(client, factory, logger, "addsvc", tags, passingOnly)
			balancer := lb.NewRoundRobin(subscriber)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			endpoints.ConcatEndpoint = retry
		}

		// Here we leverage the fact that addsvc comes with a constructor for an
		// HTTP handler, and just install it under a particular path prefix in
		// our router.

		r.PathPrefix("addsvc/").Handler(addsvc.MakeHTTPHandler(ctx, endpoints, tracer, logger))
	}

	// stringsvc routes.
	{
		// addsvc had lots of nice importable Go packages we could leverage.
		// With stringsvc we are not so fortunate, it just has some endpoints
		// that we assume will exist. So we have to write that logic here. This
		// is by design, so you can see two totally different methods of
		// proxying to a remote service.

		var (
			tags        = []string{}
			passingOnly = true
			uppercase   endpoint.Endpoint
			count       endpoint.Endpoint
		)
		{
			factory := stringsvcFactory(ctx, "GET", "/uppercase")
			subscriber := consulsd.NewSubscriber(client, factory, logger, "stringsvc", tags, passingOnly)
			balancer := lb.NewRoundRobin(subscriber)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			uppercase = retry
		}
		{
			factory := stringsvcFactory(ctx, "GET", "/count")
			subscriber := consulsd.NewSubscriber(client, factory, logger, "stringsvc", tags, passingOnly)
			balancer := lb.NewRoundRobin(subscriber)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			count = retry
		}

		// We can use the transport/http.Server to act as our handler, all we
		// have to do provide it with the encode and decode functions for our
		// stringsvc methods.

		r.Handle("/stringsvc/uppercase", httptransport.NewServer(ctx, uppercase, decodeUppercaseRequest, encodeJSONResponse))
		r.Handle("/stringsvc/count", httptransport.NewServer(ctx, count, decodeCountRequest, encodeJSONResponse))
	}

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// Run!
	logger.Log("exit", <-errc)
}

func addsvcFactory(makeEndpoint func(addsvc.Service) endpoint.Endpoint, tracer stdopentracing.Tracer, logger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		// We could just as easily use the HTTP or Thrift client package to make
		// the connection to addsvc. We've chosen gRPC arbitrarily. Note that
		// the transport is an implementation detail: it doesn't leak out of
		// this function. Nice!

		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		service := addsvcgrpcclient.New(conn, tracer, logger)
		endpoint := makeEndpoint(service)

		// Notice that the addsvc gRPC client converts the connection to a
		// complete addsvc, and we just throw away everything except the method
		// we're interested in. A smarter factory would mux multiple methods
		// over the same connection. But that would require more work to manage
		// the returned io.Closer, e.g. reference counting. Since this is for
		// the purposes of demonstration, we'll just keep it simple.

		return endpoint, conn, nil
	}
}

func stringsvcFactory(ctx context.Context, method, path string) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}
		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		tgt.Path = path

		// Since stringsvc doesn't have any kind of package we can import, or
		// any formal spec, we are forced to just assert where the endpoints
		// live, and write our own code to encode and decode requests and
		// responses. Ideally, if you write the service, you will want to
		// provide stronger guarantees to your clients.

		var (
			enc httptransport.EncodeRequestFunc
			dec httptransport.DecodeResponseFunc
		)
		switch path {
		case "/uppercase":
			enc, dec = encodeJSONRequest, decodeUppercaseResponse
		case "/count":
			enc, dec = encodeJSONRequest, decodeCountResponse
		default:
			return nil, nil, fmt.Errorf("unknown stringsvc path %q", path)
		}

		return httptransport.NewClient(method, tgt, enc, dec).Endpoint(), nil, nil
	}
}

func encodeJSONRequest(_ context.Context, req *http.Request, request interface{}) error {
	// Both uppercase and count requests are encoded in the same way:
	// simple JSON serialization to the request body.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	return nil
}

func encodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// I've just copied these functions from stringsvc3/transport.go, inlining the
// struct definitions.

func decodeUppercaseResponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	var response struct {
		V   string `json:"v"`
		Err string `json:"err,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func decodeCountResponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	var response struct {
		V int `json:"v"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func decodeUppercaseRequest(ctx context.Context, req *http.Request) (interface{}, error) {
	var request struct {
		S string `json:"s"`
	}
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCountRequest(ctx context.Context, req *http.Request) (interface{}, error) {
	var request struct {
		S string `json:"s"`
	}
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}
