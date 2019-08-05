package egoscale

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// UserAgent is the "User-Agent" HTTP request header added to outgoing HTTP requests.
var UserAgent = fmt.Sprintf("egoscale/%s (%s; %s/%s)",
	Version,
	runtime.Version(),
	runtime.GOOS,
	runtime.GOARCH)

// Taggable represents a resource to which tags can be attached
//
// This is a helper to fill the resourcetype of a CreateTags call
type Taggable interface {
	// ResourceType is the name of the Taggable type
	ResourceType() string
}

// Deletable represents an Interface that can be "Delete" by the client
type Deletable interface {
	// Delete removes the given resource(s) or throws
	Delete(context context.Context, client *Client) error
}

// Listable represents an Interface that can be "List" by the client
type Listable interface {
	// ListRequest builds the list command
	ListRequest() (ListCommand, error)
}

// Client represents the API client
type Client struct {
	// HTTPClient holds the HTTP client
	HTTPClient *http.Client
	// Endpoint is the HTTP URL
	Endpoint string
	// APIKey is the API identifier
	APIKey string
	// apisecret is the API secret, hence non exposed
	apiSecret string
	// PageSize represents the default size for a paginated result
	PageSize int
	// Timeout represents the default timeout for the async requests
	Timeout time.Duration
	// Expiration representation how long a signed payload may be used
	Expiration time.Duration
	// RetryStrategy represents the waiting strategy for polling the async requests
	RetryStrategy RetryStrategyFunc
	// Logger contains any log, plug your own
	Logger *log.Logger
}

// RetryStrategyFunc represents a how much time to wait between two calls to the API
type RetryStrategyFunc func(int64) time.Duration

// IterateItemFunc represents the callback to iterate a list of results, if false stops
type IterateItemFunc func(interface{}, error) bool

// WaitAsyncJobResultFunc represents the callback to wait a results of an async request, if false stops
type WaitAsyncJobResultFunc func(*AsyncJobResult, error) bool

// NewClient creates an API client with default timeout (60)
//
// Timeout is set to both the HTTP client and the client itself.
func NewClient(endpoint, apiKey, apiSecret string) *Client {
	timeout := 60 * time.Second
	expiration := 10 * time.Minute

	httpClient := &http.Client{
		Transport: http.DefaultTransport,
	}

	client := &Client{
		HTTPClient:    httpClient,
		Endpoint:      endpoint,
		APIKey:        apiKey,
		apiSecret:     apiSecret,
		PageSize:      50,
		Timeout:       timeout,
		Expiration:    expiration,
		RetryStrategy: MonotonicRetryStrategyFunc(2),
		Logger:        log.New(ioutil.Discard, "", 0),
	}

	if prefix, ok := os.LookupEnv("EXOSCALE_TRACE"); ok {
		client.Logger = log.New(os.Stderr, prefix, log.LstdFlags)
		client.TraceOn()
	}

	return client
}

// Get populates the given resource or fails
func (client *Client) Get(ls Listable) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.GetWithContext(ctx, ls)
}

// GetWithContext populates the given resource or fails
func (client *Client) GetWithContext(ctx context.Context, ls Listable) (interface{}, error) {
	gs, err := client.ListWithContext(ctx, ls)
	if err != nil {
		return nil, err
	}

	count := len(gs)
	if count != 1 {
		req, err := ls.ListRequest()
		if err != nil {
			return nil, err
		}
		params, err := client.Payload(req)
		if err != nil {
			return nil, err
		}

		// removing sensitive/useless informations
		params.Del("expires")
		params.Del("response")
		params.Del("signature")
		params.Del("signatureversion")

		// formatting the query string nicely
		payload := params.Encode()
		payload = strings.Replace(payload, "&", ", ", -1)

		if count == 0 {
			return nil, &ErrorResponse{
				CSErrorCode: ServerAPIException,
				ErrorCode:   ParamError,
				ErrorText:   fmt.Sprintf("not found, query: %s", payload),
			}
		}
		return nil, fmt.Errorf("more than one element found: %s", payload)
	}

	return gs[0], nil
}

// Delete removes the given resource of fails
func (client *Client) Delete(g Deletable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.DeleteWithContext(ctx, g)
}

// DeleteWithContext removes the given resource of fails
func (client *Client) DeleteWithContext(ctx context.Context, g Deletable) error {
	return g.Delete(ctx, client)
}

// List lists the given resource (and paginate till the end)
func (client *Client) List(g Listable) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.ListWithContext(ctx, g)
}

// ListWithContext lists the given resources (and paginate till the end)
func (client *Client) ListWithContext(ctx context.Context, g Listable) (s []interface{}, err error) {
	s = make([]interface{}, 0)

	defer func() {
		if e := recover(); e != nil {
			if g == nil || reflect.ValueOf(g).IsNil() {
				err = fmt.Errorf("g Listable shouldn't be nil, got %#v", g)
				return
			}

			panic(e)
		}
	}()

	req, e := g.ListRequest()
	if e != nil {
		err = e
		return
	}
	client.PaginateWithContext(ctx, req, func(item interface{}, e error) bool {
		if item != nil {
			s = append(s, item)
			return true
		}
		err = e
		return false
	})

	return
}

// AsyncListWithContext lists the given resources (and paginate till the end)
//
//
//	// NB: goroutine may leak if not read until the end. Create a proper context!
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	outChan, errChan := client.AsyncListWithContext(ctx, new(egoscale.VirtualMachine))
//
//	for {
//		select {
//		case i, ok := <- outChan:
//			if ok {
//				vm := i.(egoscale.VirtualMachine)
//				// ...
//			} else {
//				outChan = nil
//			}
//		case err, ok := <- errChan:
//			if ok {
//				// do something
//			}
//			// Once an error has been received, you can expect the channels to be closed.
//			errChan = nil
//		}
//		if errChan == nil && outChan == nil {
//			break
//		}
//	}
//
func (client *Client) AsyncListWithContext(ctx context.Context, g Listable) (<-chan interface{}, <-chan error) {
	outChan := make(chan interface{}, client.PageSize)
	errChan := make(chan error)

	go func() {
		defer close(outChan)
		defer close(errChan)

		req, err := g.ListRequest()
		if err != nil {
			errChan <- err
			return
		}
		client.PaginateWithContext(ctx, req, func(item interface{}, e error) bool {
			if item != nil {
				outChan <- item
				return true
			}
			errChan <- e
			return false
		})
	}()

	return outChan, errChan
}

// Paginate runs the ListCommand and paginates
func (client *Client) Paginate(g Listable, callback IterateItemFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	client.PaginateWithContext(ctx, g, callback)
}

// PaginateWithContext runs the ListCommand as long as the ctx is valid
func (client *Client) PaginateWithContext(ctx context.Context, g Listable, callback IterateItemFunc) {
	req, err := g.ListRequest()
	if err != nil {
		callback(nil, err)
		return
	}

	pageSize := client.PageSize

	page := 1

	for {
		req.SetPage(page)
		req.SetPageSize(pageSize)
		resp, err := client.RequestWithContext(ctx, req)
		if err != nil {
			// in case of 431, the response is knowingly empty
			if errResponse, ok := err.(*ErrorResponse); ok && page == 1 && errResponse.ErrorCode == ParamError {
				break
			}

			callback(nil, err)
			break
		}

		size := 0
		didErr := false
		req.Each(resp, func(element interface{}, err error) bool {
			// If the context was cancelled, kill it in flight
			if e := ctx.Err(); e != nil {
				element = nil
				err = e
			}

			if callback(element, err) {
				size++
				return true
			}

			didErr = true
			return false
		})

		if size < pageSize || didErr {
			break
		}

		page++
	}
}

// APIName returns the name of the given command
func (client *Client) APIName(command Command) string {
	// This is due to a limitation of Go<=1.7
	_, ok := command.(*AuthorizeSecurityGroupEgress)
	_, okPtr := command.(AuthorizeSecurityGroupEgress)
	if ok || okPtr {
		return "authorizeSecurityGroupEgress"
	}

	info, err := info(command)
	if err != nil {
		panic(err)
	}
	return info.Name
}

// APIDescription returns the description of the given command
func (client *Client) APIDescription(command Command) string {
	info, err := info(command)
	if err != nil {
		return "*missing description*"
	}
	return info.Description
}

// Response returns the response structure of the given command
func (client *Client) Response(command Command) interface{} {
	switch c := command.(type) {
	case AsyncCommand:
		return c.AsyncResponse()
	default:
		return command.Response()
	}
}

// TraceOn activates the HTTP tracer
func (client *Client) TraceOn() {
	if _, ok := client.HTTPClient.Transport.(*traceTransport); !ok {
		client.HTTPClient.Transport = &traceTransport{
			transport: client.HTTPClient.Transport,
			logger:    client.Logger,
		}
	}
}

// TraceOff deactivates the HTTP tracer
func (client *Client) TraceOff() {
	if rt, ok := client.HTTPClient.Transport.(*traceTransport); ok {
		client.HTTPClient.Transport = rt.transport
	}
}

// traceTransport  contains the original HTTP transport to enable it to be reverted
type traceTransport struct {
	transport http.RoundTripper
	logger    *log.Logger
}

// RoundTrip executes a single HTTP transaction
func (t *traceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if dump, err := httputil.DumpRequest(req, true); err == nil {
		t.logger.Printf("%s", dump)
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if dump, err := httputil.DumpResponse(resp, true); err == nil {
		t.logger.Printf("%s", dump)
	}

	return resp, nil
}

// MonotonicRetryStrategyFunc returns a function that waits for n seconds for each iteration
func MonotonicRetryStrategyFunc(seconds int) RetryStrategyFunc {
	return func(iteration int64) time.Duration {
		return time.Duration(seconds) * time.Second
	}
}

// FibonacciRetryStrategy waits for an increasing amount of time following the Fibonacci sequence
func FibonacciRetryStrategy(iteration int64) time.Duration {
	var a, b, i, tmp int64
	a = 0
	b = 1
	for i = 0; i < iteration; i++ {
		tmp = a + b
		a = b
		b = tmp
	}
	return time.Duration(a) * time.Second
}
