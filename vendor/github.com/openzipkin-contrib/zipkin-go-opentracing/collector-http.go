package zipkintracer

import (
	"bytes"
	"net/http"
	"time"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

// Default timeout for http request in seconds
const defaultHTTPTimeout = time.Second * 5

// defaultBatchInterval in seconds
const defaultHTTPBatchInterval = 1

const defaultHTTPBatchSize = 100

const defaultHTTPMaxBacklog = 1000

// HTTPCollector implements Collector by forwarding spans to a http server.
type HTTPCollector struct {
	logger        Logger
	url           string
	client        *http.Client
	batchInterval time.Duration
	batchSize     int
	maxBacklog    int
	spanc         chan *zipkincore.Span
	quit          chan struct{}
	shutdown      chan error
	reqCallback   RequestCallback
}

// RequestCallback receives the initialized request from the Collector before
// sending it over the wire. This allows one to plug in additional headers or
// do other customization.
type RequestCallback func(*http.Request)

// HTTPOption sets a parameter for the HttpCollector
type HTTPOption func(c *HTTPCollector)

// HTTPLogger sets the logger used to report errors in the collection
// process. By default, a no-op logger is used, i.e. no errors are logged
// anywhere. It's important to set this option in a production service.
func HTTPLogger(logger Logger) HTTPOption {
	return func(c *HTTPCollector) { c.logger = logger }
}

// HTTPTimeout sets maximum timeout for http request.
func HTTPTimeout(duration time.Duration) HTTPOption {
	return func(c *HTTPCollector) { c.client.Timeout = duration }
}

// HTTPBatchSize sets the maximum batch size, after which a collect will be
// triggered. The default batch size is 100 traces.
func HTTPBatchSize(n int) HTTPOption {
	return func(c *HTTPCollector) { c.batchSize = n }
}

// HTTPMaxBacklog sets the maximum backlog size,
// when batch size reaches this threshold, spans from the
// beginning of the batch will be disposed
func HTTPMaxBacklog(n int) HTTPOption {
	return func(c *HTTPCollector) { c.maxBacklog = n }
}

// HTTPBatchInterval sets the maximum duration we will buffer traces before
// emitting them to the collector. The default batch interval is 1 second.
func HTTPBatchInterval(d time.Duration) HTTPOption {
	return func(c *HTTPCollector) { c.batchInterval = d }
}

// HTTPClient sets a custom http client to use.
func HTTPClient(client *http.Client) HTTPOption {
	return func(c *HTTPCollector) { c.client = client }
}

// HTTPRequestCallback registers a callback function to adjust the collector
// *http.Request before it sends the request to Zipkin.
func HTTPRequestCallback(rc RequestCallback) HTTPOption {
	return func(c *HTTPCollector) { c.reqCallback = rc }
}

// NewHTTPCollector returns a new HTTP-backend Collector. url should be a http
// url for handle post request. timeout is passed to http client. queueSize control
// the maximum size of buffer of async queue. The logger is used to log errors,
// such as send failures;
func NewHTTPCollector(url string, options ...HTTPOption) (Collector, error) {
	c := &HTTPCollector{
		logger:        NewNopLogger(),
		url:           url,
		client:        &http.Client{Timeout: defaultHTTPTimeout},
		batchInterval: defaultHTTPBatchInterval * time.Second,
		batchSize:     defaultHTTPBatchSize,
		maxBacklog:    defaultHTTPMaxBacklog,
		quit:          make(chan struct{}, 1),
		shutdown:      make(chan error, 1),
	}

	for _, option := range options {
		option(c)
	}

	// spanc can immediately accept maxBacklog spans and everything else is dropped.
	c.spanc = make(chan *zipkincore.Span, c.maxBacklog)

	go c.loop()
	return c, nil
}

// Collect implements Collector.
// attempts a non blocking send on the channel.
func (c *HTTPCollector) Collect(s *zipkincore.Span) error {
	select {
	case c.spanc <- s:
		// Accepted.
	case <-c.quit:
		// Collector concurrently closed.
	default:
		c.logger.Log("msg", "queue full, disposing spans.", "size", len(c.spanc))
	}
	return nil
}

// Close implements Collector.
func (c *HTTPCollector) Close() error {
	close(c.quit)
	return <-c.shutdown
}

func httpSerialize(spans []*zipkincore.Span) *bytes.Buffer {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	if err := p.WriteListBegin(thrift.STRUCT, len(spans)); err != nil {
		panic(err)
	}
	for _, s := range spans {
		if err := s.Write(p); err != nil {
			panic(err)
		}
	}
	if err := p.WriteListEnd(); err != nil {
		panic(err)
	}
	return t.Buffer
}

func (c *HTTPCollector) loop() {
	var (
		nextSend = time.Now().Add(c.batchInterval)
		ticker   = time.NewTicker(c.batchInterval / 10)
		tickc    = ticker.C
	)
	defer ticker.Stop()

	// The following loop is single threaded
	// allocate enough space so we don't have to reallocate.
	batch := make([]*zipkincore.Span, 0, c.batchSize)

	for {
		select {
		case span := <-c.spanc:
			batch = append(batch, span)
			if len(batch) == c.batchSize {
				c.send(batch)
				batch = batch[0:0]
				nextSend = time.Now().Add(c.batchInterval)
			}
		case <-tickc:
			if time.Now().After(nextSend) {
				if len(batch) > 0 {
					c.send(batch)
					batch = batch[0:0]
				}
				nextSend = time.Now().Add(c.batchInterval)
			}
		case <-c.quit:
			c.shutdown <- c.send(batch)
			return
		}
	}
}

func (c *HTTPCollector) send(sendBatch []*zipkincore.Span) error {
	req, err := http.NewRequest(
		"POST",
		c.url,
		httpSerialize(sendBatch))
	if err != nil {
		c.logger.Log("err", err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/x-thrift")
	if c.reqCallback != nil {
		c.reqCallback(req)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Log("err", err.Error())
		return err
	}
	resp.Body.Close()
	// non 2xx code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Log("err", "HTTP POST span failed", "code", resp.Status)
	}
	return nil
}
