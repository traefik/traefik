package cony

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
)

const (
	noRun = iota
	run
)

var (
	// ErrNoConnection is an indicator that currently there is no connection
	// available
	ErrNoConnection = errors.New("No connection available")
)

// ClientOpt is a Client's functional option type
type ClientOpt func(*Client)

// Client is a Main AMQP client wrapper
type Client struct {
	addr         string
	declarations []Declaration
	consumers    map[*Consumer]struct{}
	publishers   map[*Publisher]struct{}
	errs         chan error
	blocking     chan amqp.Blocking
	run          int32        // bool
	conn         atomic.Value //*amqp.Connection
	bo           Backoffer
	attempt      int32
	l            sync.Mutex
	config       amqp.Config
}

// Declare used to declare queues/exchanges/bindings.
// Declaration is saved and will be re-run every time Client gets connection
func (c *Client) Declare(d []Declaration) {
	c.l.Lock()
	defer c.l.Unlock()
	c.declarations = append(c.declarations, d...)
	if ch, err := c.channel(); err == nil {
		for _, declare := range d {
			declare(ch)
		}
	}
}

// Consume used to declare consumers
func (c *Client) Consume(cons *Consumer) {
	c.l.Lock()
	defer c.l.Unlock()
	c.consumers[cons] = struct{}{}
	if ch, err := c.channel(); err == nil {
		go cons.serve(c, ch)
	}
}

func (c *Client) deleteConsumer(cons *Consumer) {
	c.l.Lock()
	defer c.l.Unlock()
	delete(c.consumers, cons)
}

// Publish used to declare publishers
func (c *Client) Publish(pub *Publisher) {
	c.l.Lock()
	defer c.l.Unlock()
	c.publishers[pub] = struct{}{}
	if ch, err := c.channel(); err == nil {
		go pub.serve(c, ch)
	}
}

func (c *Client) deletePublisher(pub *Publisher) {
	c.l.Lock()
	defer c.l.Unlock()
	delete(c.publishers, pub)
}

// Errors returns AMQP connection level errors. Default buffer size is 100.
// Messages will be dropped in case if receiver can't keep up
func (c *Client) Errors() <-chan error {
	return c.errs
}

// Blocking notifies the server's TCP flow control of the Connection. Default
// buffer size is 10. Messages will be dropped in case if receiver can't keep up
func (c *Client) Blocking() <-chan amqp.Blocking {
	return c.blocking
}

// Close shutdown the client
func (c *Client) Close() {
	atomic.StoreInt32(&c.run, noRun) // c.run = false
	conn, _ := c.conn.Load().(*amqp.Connection)
	if conn != nil {
		conn.Close()
	}
	c.conn.Store((*amqp.Connection)(nil))
}

// Loop should be run as condition for `for` with receiving from (*Client).Errors()
//
// It will manage AMQP connection, run queue and exchange declarations, consumers.
// Will start to return false once (*Client).Close() called.
func (c *Client) Loop() bool {
	var (
		err error
	)

	if atomic.LoadInt32(&c.run) == noRun {
		return false
	}

	conn, _ := c.conn.Load().(*amqp.Connection)

	if conn != nil {
		return true
	}

	if c.bo != nil {
		time.Sleep(c.bo.Backoff(int(c.attempt)))
		atomic.AddInt32(&c.attempt, 1)
	}

	// set default Heartbeat to 10 seconds like in original amqp.Dial
	if c.config.Heartbeat == 0 {
		c.config.Heartbeat = 10 * time.Second
	}

	conn, err = amqp.DialConfig(c.addr, c.config)

	if c.reportErr(err) {
		return true
	}
	c.conn.Store(conn)

	atomic.StoreInt32(&c.attempt, 0)

	// guard conn
	go func() {
		chanErr := make(chan *amqp.Error)
		chanBlocking := make(chan amqp.Blocking)
		conn.NotifyClose(chanErr)
		conn.NotifyBlocked(chanBlocking)

		// loop for blocking/deblocking
		for {
			select {
			case err1 := <-chanErr:
				c.reportErr(err1)

				if conn1 := c.conn.Load().(*amqp.Connection); conn1 != nil {
					c.conn.Store((*amqp.Connection)(nil))
					conn1.Close()
				}
				// return from routine to launch reconnect process
				return
			case blocking := <-chanBlocking:
				select {
				case c.blocking <- blocking:
				default:
				}
			}
		}

	}()

	ch, err := conn.Channel()
	if c.reportErr(err) {
		return true
	}

	for _, declare := range c.declarations {
		c.reportErr(declare(ch))
	}

	for cons := range c.consumers {
		ch1, err := c.channel()
		if err == nil {
			go cons.serve(c, ch1)
		}
	}

	for pub := range c.publishers {
		ch1, err := c.channel()
		if err == nil {
			go pub.serve(c, ch1)
		}
	}

	return true
}

func (c *Client) reportErr(err error) bool {
	if err != nil {
		select {
		case c.errs <- err:
		default:
		}
		return true
	}
	return false
}

func (c *Client) channel() (*amqp.Channel, error) {
	conn, err := c.connection()
	if err != nil {
		return nil, err
	}

	return conn.Channel()
}

func (c *Client) connection() (*amqp.Connection, error) {
	conn, _ := c.conn.Load().(*amqp.Connection)
	if conn == nil {
		return nil, ErrNoConnection
	}

	return conn, nil
}

// NewClient initializes new Client
func NewClient(opts ...ClientOpt) *Client {
	c := &Client{
		run:          run,
		declarations: make([]Declaration, 0),
		consumers:    make(map[*Consumer]struct{}),
		publishers:   make(map[*Publisher]struct{}),
		errs:         make(chan error, 100),
		blocking:     make(chan amqp.Blocking, 10),
	}

	for _, o := range opts {
		o(c)
	}
	return c
}

// URL is a functional option, used in `NewClient` constructor
// default URL is amqp://guest:guest@localhost/
func URL(addr string) ClientOpt {
	return func(c *Client) {
		if addr == "" {
			addr = "amqp://guest:guest@localhost/"
		}
		c.addr = addr
	}
}

// Backoff is a functional option, used to define backoff policy, used in
// `NewClient` constructor
func Backoff(bo Backoffer) ClientOpt {
	return func(c *Client) {
		c.bo = bo
	}
}

// ErrorsChan is a functional option, used to initialize error reporting channel
// in client code, maintaining control over buffer size. Default buffer size is
// 100. Messages will be dropped in case if receiver can't keep up, used in
// `NewClient` constructor
func ErrorsChan(errChan chan error) ClientOpt {
	return func(c *Client) {
		c.errs = errChan
	}
}

// BlockingChan is a functional option, used to initialize blocking reporting
// channel in client code, maintaining control over buffering, used in
// `NewClient` constructor
func BlockingChan(blockingChan chan amqp.Blocking) ClientOpt {
	return func(c *Client) {
		c.blocking = blockingChan
	}
}

// Config is a functional option, used to setup extended amqp configuration
func Config(config amqp.Config) ClientOpt {
	return func(c *Client) {
		c.config = config
	}
}
