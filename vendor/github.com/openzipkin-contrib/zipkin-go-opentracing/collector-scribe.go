package zipkintracer

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/scribe"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

const defaultScribeCategory = "zipkin"

// defaultScribeBatchInterval in seconds
const defaultScribeBatchInterval = 1

const defaultScribeBatchSize = 100

const defaultScribeMaxBacklog = 1000

// ScribeCollector implements Collector by forwarding spans to a Scribe
// service, in batches.
type ScribeCollector struct {
	logger        Logger
	category      string
	factory       func() (scribe.Scribe, error)
	client        scribe.Scribe
	batchInterval time.Duration
	batchSize     int
	maxBacklog    int
	batch         []*scribe.LogEntry
	spanc         chan *zipkincore.Span
	quit          chan struct{}
	shutdown      chan error
	sendMutex     *sync.Mutex
	batchMutex    *sync.Mutex
}

// ScribeOption sets a parameter for the StdlibAdapter.
type ScribeOption func(s *ScribeCollector)

// ScribeLogger sets the logger used to report errors in the collection
// process. By default, a no-op logger is used, i.e. no errors are logged
// anywhere. It's important to set this option in a production service.
func ScribeLogger(logger Logger) ScribeOption {
	return func(s *ScribeCollector) { s.logger = logger }
}

// ScribeBatchSize sets the maximum batch size, after which a collect will be
// triggered. The default batch size is 100 traces.
func ScribeBatchSize(n int) ScribeOption {
	return func(s *ScribeCollector) { s.batchSize = n }
}

// ScribeMaxBacklog sets the maximum backlog size,
// when batch size reaches this threshold, spans from the
// beginning of the batch will be disposed
func ScribeMaxBacklog(n int) ScribeOption {
	return func(c *ScribeCollector) { c.maxBacklog = n }
}

// ScribeBatchInterval sets the maximum duration we will buffer traces before
// emitting them to the collector. The default batch interval is 1 second.
func ScribeBatchInterval(d time.Duration) ScribeOption {
	return func(s *ScribeCollector) { s.batchInterval = d }
}

// ScribeCategory sets the Scribe category used to transmit the spans.
func ScribeCategory(category string) ScribeOption {
	return func(s *ScribeCollector) { s.category = category }
}

// NewScribeCollector returns a new Scribe-backed Collector. addr should be a
// TCP endpoint of the form "host:port". timeout is passed to the Thrift dial
// function NewTSocketFromAddrTimeout. batchSize and batchInterval control the
// maximum size and interval of a batch of spans; as soon as either limit is
// reached, the batch is sent. The logger is used to log errors, such as batch
// send failures; users should provide an appropriate context, if desired.
func NewScribeCollector(addr string, timeout time.Duration, options ...ScribeOption) (Collector, error) {
	factory := scribeClientFactory(addr, timeout)
	client, err := factory()
	if err != nil {
		return nil, err
	}
	c := &ScribeCollector{
		logger:        NewNopLogger(),
		category:      defaultScribeCategory,
		factory:       factory,
		client:        client,
		batchInterval: defaultScribeBatchInterval * time.Second,
		batchSize:     defaultScribeBatchSize,
		maxBacklog:    defaultScribeMaxBacklog,
		batch:         []*scribe.LogEntry{},
		spanc:         make(chan *zipkincore.Span),
		quit:          make(chan struct{}),
		shutdown:      make(chan error, 1),
		sendMutex:     &sync.Mutex{},
		batchMutex:    &sync.Mutex{},
	}

	for _, option := range options {
		option(c)
	}

	go c.loop()
	return c, nil
}

// Collect implements Collector.
func (c *ScribeCollector) Collect(s *zipkincore.Span) error {
	select {
	case c.spanc <- s:
		// Accepted.
	case <-c.quit:
		// Collector concurrently closed.
	}
	return nil
}

// Close implements Collector.
func (c *ScribeCollector) Close() error {
	close(c.quit)
	return <-c.shutdown
}

func scribeSerialize(s *zipkincore.Span) string {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	if err := s.Write(p); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(t.Buffer.Bytes())
}

func (c *ScribeCollector) loop() {
	var (
		nextSend = time.Now().Add(c.batchInterval)
		ticker   = time.NewTicker(c.batchInterval / 10)
		tickc    = ticker.C
	)
	defer ticker.Stop()

	for {
		select {
		case span := <-c.spanc:
			currentBatchSize := c.append(span)
			if currentBatchSize >= c.batchSize {
				nextSend = time.Now().Add(c.batchInterval)
				go c.send()
			}
		case <-tickc:
			if time.Now().After(nextSend) {
				nextSend = time.Now().Add(c.batchInterval)
				go c.send()
			}
		case <-c.quit:
			c.shutdown <- c.send()
			return
		}
	}
}

func (c *ScribeCollector) append(span *zipkincore.Span) (newBatchSize int) {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	c.batch = append(c.batch, &scribe.LogEntry{
		Category: c.category,
		Message:  scribeSerialize(span),
	})
	if len(c.batch) > c.maxBacklog {
		dispose := len(c.batch) - c.maxBacklog
		c.logger.Log("Backlog too long, disposing spans.", "count", dispose)
		c.batch = c.batch[dispose:]
	}
	newBatchSize = len(c.batch)
	return
}

func (c *ScribeCollector) send() error {
	// in order to prevent sending the same batch twice
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	// Select all current spans in the batch to be sent
	c.batchMutex.Lock()
	sendBatch := c.batch[:]
	c.batchMutex.Unlock()

	// Do not send an empty batch
	if len(sendBatch) == 0 {
		return nil
	}

	if c.client == nil {
		var err error
		if c.client, err = c.factory(); err != nil {
			_ = c.logger.Log("err", fmt.Sprintf("during reconnect: %v", err))
			return err
		}
	}
	if rc, err := c.client.Log(context.Background(), sendBatch); err != nil {
		c.client = nil
		_ = c.logger.Log("err", fmt.Sprintf("during Log: %v", err))
		return err
	} else if rc != scribe.ResultCode_OK {
		// probably transient error; don't reset client
		_ = c.logger.Log("err", fmt.Sprintf("remote returned %s", rc))
	}

	// Remove sent spans from the batch
	c.batchMutex.Lock()
	c.batch = c.batch[len(sendBatch):]
	c.batchMutex.Unlock()

	return nil
}

func scribeClientFactory(addr string, timeout time.Duration) func() (scribe.Scribe, error) {
	return func() (scribe.Scribe, error) {
		a, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return nil, err
		}
		socket := thrift.NewTSocketFromAddrTimeout(a, timeout)
		transport := thrift.NewTFramedTransport(socket)
		if err := transport.Open(); err != nil {
			_ = socket.Close()
			return nil, err
		}
		proto := thrift.NewTBinaryProtocolTransport(transport)
		client := scribe.NewScribeClientProtocol(transport, proto, proto)
		return client, nil
	}
}
