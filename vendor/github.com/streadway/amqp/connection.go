// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxChannelMax = (2 << 15) - 1

	defaultHeartbeat         = 10 * time.Second
	defaultConnectionTimeout = 30 * time.Second
	defaultProduct           = "https://github.com/streadway/amqp"
	defaultVersion           = "β"
	defaultChannelMax        = maxChannelMax
	defaultLocale            = "en_US"
)

// Config is used in DialConfig and Open to specify the desired tuning
// parameters used during a connection open handshake.  The negotiated tuning
// will be stored in the returned connection's Config field.
type Config struct {
	// The SASL mechanisms to try in the client request, and the successful
	// mechanism used on the Connection object.
	// If SASL is nil, PlainAuth from the URL is used.
	SASL []Authentication

	// Vhost specifies the namespace of permissions, exchanges, queues and
	// bindings on the server.  Dial sets this to the path parsed from the URL.
	Vhost string

	ChannelMax int           // 0 max channels means 2^16 - 1
	FrameSize  int           // 0 max bytes means unlimited
	Heartbeat  time.Duration // less than 1s uses the server's interval

	// TLSClientConfig specifies the client configuration of the TLS connection
	// when establishing a tls transport.
	// If the URL uses an amqps scheme, then an empty tls.Config with the
	// ServerName from the URL is used.
	TLSClientConfig *tls.Config

	// Properties is table of properties that the client advertises to the server.
	// This is an optional setting - if the application does not set this,
	// the underlying library will use a generic set of client properties.
	Properties Table

	// Connection locale that we expect to always be en_US
	// Even though servers must return it as per the AMQP 0-9-1 spec,
	// we are not aware of it being used other than to satisfy the spec requirements
	Locale string

	// Dial returns a net.Conn prepared for a TLS handshake with TSLClientConfig,
	// then an AMQP connection handshake.
	// If Dial is nil, net.DialTimeout with a 30s connection and 30s deadline is
	// used during TLS and AMQP handshaking.
	Dial func(network, addr string) (net.Conn, error)
}

// Connection manages the serialization and deserialization of frames from IO
// and dispatches the frames to the appropriate channel.  All RPC methods and
// asyncronous Publishing, Delivery, Ack, Nack and Return messages are
// multiplexed on this channel.  There must always be active receivers for
// every asynchronous message on this connection.
type Connection struct {
	destructor sync.Once  // shutdown once
	sendM      sync.Mutex // conn writer mutex
	m          sync.Mutex // struct field mutex

	conn io.ReadWriteCloser

	rpc       chan message
	writer    *writer
	sends     chan time.Time     // timestamps of each frame sent
	deadlines chan readDeadliner // heartbeater updates read deadlines

	allocator *allocator // id generator valid after openTune
	channels  map[uint16]*Channel

	noNotify bool // true when we will never notify again
	closes   []chan *Error
	blocks   []chan Blocking

	errors chan *Error

	Config Config // The negotiated Config after connection.open

	Major      int      // Server's major version
	Minor      int      // Server's minor version
	Properties Table    // Server properties
	Locales    []string // Server locales

	closed int32 // Will be 1 if the connection is closed, 0 otherwise. Should only be accessed as atomic
}

type readDeadliner interface {
	SetReadDeadline(time.Time) error
}

// defaultDial establishes a connection when config.Dial is not provided
func defaultDial(network, addr string) (net.Conn, error) {
	conn, err := net.DialTimeout(network, addr, defaultConnectionTimeout)
	if err != nil {
		return nil, err
	}

	// Heartbeating hasn't started yet, don't stall forever on a dead server.
	// A deadline is set for TLS and AMQP handshaking. After AMQP is established,
	// the deadline is cleared in openComplete.
	if err := conn.SetDeadline(time.Now().Add(defaultConnectionTimeout)); err != nil {
		return nil, err
	}

	return conn, nil
}

// Dial accepts a string in the AMQP URI format and returns a new Connection
// over TCP using PlainAuth.  Defaults to a server heartbeat interval of 10
// seconds and sets the handshake deadline to 30 seconds. After handshake,
// deadlines are cleared.
//
// Dial uses the zero value of tls.Config when it encounters an amqps://
// scheme.  It is equivalent to calling DialTLS(amqp, nil).
func Dial(url string) (*Connection, error) {
	return DialConfig(url, Config{
		Heartbeat: defaultHeartbeat,
		Locale:    defaultLocale,
	})
}

// DialTLS accepts a string in the AMQP URI format and returns a new Connection
// over TCP using PlainAuth.  Defaults to a server heartbeat interval of 10
// seconds and sets the initial read deadline to 30 seconds.
//
// DialTLS uses the provided tls.Config when encountering an amqps:// scheme.
func DialTLS(url string, amqps *tls.Config) (*Connection, error) {
	return DialConfig(url, Config{
		Heartbeat:       defaultHeartbeat,
		TLSClientConfig: amqps,
		Locale:          defaultLocale,
	})
}

// DialConfig accepts a string in the AMQP URI format and a configuration for
// the transport and connection setup, returning a new Connection.  Defaults to
// a server heartbeat interval of 10 seconds and sets the initial read deadline
// to 30 seconds.
func DialConfig(url string, config Config) (*Connection, error) {
	var err error
	var conn net.Conn

	uri, err := ParseURI(url)
	if err != nil {
		return nil, err
	}

	if config.SASL == nil {
		config.SASL = []Authentication{uri.PlainAuth()}
	}

	if config.Vhost == "" {
		config.Vhost = uri.Vhost
	}

	addr := net.JoinHostPort(uri.Host, strconv.FormatInt(int64(uri.Port), 10))

	dialer := config.Dial
	if dialer == nil {
		dialer = defaultDial
	}

	conn, err = dialer("tcp", addr)
	if err != nil {
		return nil, err
	}

	if uri.Scheme == "amqps" {
		if config.TLSClientConfig == nil {
			config.TLSClientConfig = new(tls.Config)
		}

		// If ServerName has not been specified in TLSClientConfig,
		// set it to the URI host used for this connection.
		if config.TLSClientConfig.ServerName == "" {
			config.TLSClientConfig.ServerName = uri.Host
		}

		client := tls.Client(conn, config.TLSClientConfig)
		if err := client.Handshake(); err != nil {
			conn.Close()
			return nil, err
		}

		conn = client
	}

	return Open(conn, config)
}

/*
Open accepts an already established connection, or other io.ReadWriteCloser as
a transport.  Use this method if you have established a TLS connection or wish
to use your own custom transport.

*/
func Open(conn io.ReadWriteCloser, config Config) (*Connection, error) {
	c := &Connection{
		conn:      conn,
		writer:    &writer{bufio.NewWriter(conn)},
		channels:  make(map[uint16]*Channel),
		rpc:       make(chan message),
		sends:     make(chan time.Time),
		errors:    make(chan *Error, 1),
		deadlines: make(chan readDeadliner, 1),
	}
	go c.reader(conn)
	return c, c.open(config)
}

/*
LocalAddr returns the local TCP peer address, or ":0" (the zero value of net.TCPAddr)
as a fallback default value if the underlying transport does not support LocalAddr().
*/
func (c *Connection) LocalAddr() net.Addr {
	if conn, ok := c.conn.(interface {
		LocalAddr() net.Addr
	}); ok {
		return conn.LocalAddr()
	}
	return &net.TCPAddr{}
}

// ConnectionState returns basic TLS details of the underlying transport.
// Returns a zero value when the underlying connection does not implement
// ConnectionState() tls.ConnectionState.
func (c *Connection) ConnectionState() tls.ConnectionState {
	if conn, ok := c.conn.(interface {
		ConnectionState() tls.ConnectionState
	}); ok {
		return conn.ConnectionState()
	}
	return tls.ConnectionState{}
}

/*
NotifyClose registers a listener for close events either initiated by an error
accompaning a connection.close method or by a normal shutdown.

On normal shutdowns, the chan will be closed.

To reconnect after a transport or protocol error, register a listener here and
re-run your setup process.

*/
func (c *Connection) NotifyClose(receiver chan *Error) chan *Error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.noNotify {
		close(receiver)
	} else {
		c.closes = append(c.closes, receiver)
	}

	return receiver
}

/*
NotifyBlocked registers a listener for RabbitMQ specific TCP flow control
method extensions connection.blocked and connection.unblocked.  Flow control is
active with a reason when Blocking.Blocked is true.  When a Connection is
blocked, all methods will block across all connections until server resources
become free again.

This optional extension is supported by the server when the
"connection.blocked" server capability key is true.

*/
func (c *Connection) NotifyBlocked(receiver chan Blocking) chan Blocking {
	c.m.Lock()
	defer c.m.Unlock()

	if c.noNotify {
		close(receiver)
	} else {
		c.blocks = append(c.blocks, receiver)
	}

	return receiver
}

/*
Close requests and waits for the response to close the AMQP connection.

It's advisable to use this message when publishing to ensure all kernel buffers
have been flushed on the server and client before exiting.

An error indicates that server may not have received this request to close but
the connection should be treated as closed regardless.

After returning from this call, all resources associated with this connection,
including the underlying io, Channels, Notify listeners and Channel consumers
will also be closed.
*/
func (c *Connection) Close() error {
	if c.isClosed() {
		return ErrClosed
	}

	defer c.shutdown(nil)
	return c.call(
		&connectionClose{
			ReplyCode: replySuccess,
			ReplyText: "kthxbai",
		},
		&connectionCloseOk{},
	)
}

func (c *Connection) closeWith(err *Error) error {
	if c.isClosed() {
		return ErrClosed
	}

	defer c.shutdown(err)
	return c.call(
		&connectionClose{
			ReplyCode: uint16(err.Code),
			ReplyText: err.Reason,
		},
		&connectionCloseOk{},
	)
}

func (c *Connection) isClosed() bool {
	return (atomic.LoadInt32(&c.closed) == 1)
}

func (c *Connection) send(f frame) error {
	if c.isClosed() {
		return ErrClosed
	}

	c.sendM.Lock()
	err := c.writer.WriteFrame(f)
	c.sendM.Unlock()

	if err != nil {
		// shutdown could be re-entrant from signaling notify chans
		go c.shutdown(&Error{
			Code:   FrameError,
			Reason: err.Error(),
		})
	} else {
		// Broadcast we sent a frame, reducing heartbeats, only
		// if there is something that can receive - like a non-reentrant
		// call or if the heartbeater isn't running
		select {
		case c.sends <- time.Now():
		default:
		}
	}

	return err
}

func (c *Connection) shutdown(err *Error) {
	atomic.StoreInt32(&c.closed, 1)

	c.destructor.Do(func() {
		c.m.Lock()
		defer c.m.Unlock()

		if err != nil {
			for _, c := range c.closes {
				c <- err
			}
		}

		if err != nil {
			c.errors <- err
		}
		// Shutdown handler goroutine can still receive the result.
		close(c.errors)

		for _, c := range c.closes {
			close(c)
		}

		for _, c := range c.blocks {
			close(c)
		}

		// Shutdown the channel, but do not use closeChannel() as it calls
		// releaseChannel() which requires the connection lock.
		//
		// Ranging over c.channels and calling releaseChannel() that mutates
		// c.channels is racy - see commit 6063341 for an example.
		for _, ch := range c.channels {
			ch.shutdown(err)
		}

		c.conn.Close()

		c.channels = map[uint16]*Channel{}
		c.allocator = newAllocator(1, c.Config.ChannelMax)
		c.noNotify = true
	})
}

// All methods sent to the connection channel should be synchronous so we
// can handle them directly without a framing component
func (c *Connection) demux(f frame) {
	if f.channel() == 0 {
		c.dispatch0(f)
	} else {
		c.dispatchN(f)
	}
}

func (c *Connection) dispatch0(f frame) {
	switch mf := f.(type) {
	case *methodFrame:
		switch m := mf.Method.(type) {
		case *connectionClose:
			// Send immediately as shutdown will close our side of the writer.
			c.send(&methodFrame{
				ChannelId: 0,
				Method:    &connectionCloseOk{},
			})

			c.shutdown(newError(m.ReplyCode, m.ReplyText))
		case *connectionBlocked:
			for _, c := range c.blocks {
				c <- Blocking{Active: true, Reason: m.Reason}
			}
		case *connectionUnblocked:
			for _, c := range c.blocks {
				c <- Blocking{Active: false}
			}
		default:
			c.rpc <- m
		}
	case *heartbeatFrame:
		// kthx - all reads reset our deadline.  so we can drop this
	default:
		// lolwat - channel0 only responds to methods and heartbeats
		c.closeWith(ErrUnexpectedFrame)
	}
}

func (c *Connection) dispatchN(f frame) {
	c.m.Lock()
	channel := c.channels[f.channel()]
	c.m.Unlock()

	if channel != nil {
		channel.recv(channel, f)
	} else {
		c.dispatchClosed(f)
	}
}

// section 2.3.7: "When a peer decides to close a channel or connection, it
// sends a Close method.  The receiving peer MUST respond to a Close with a
// Close-Ok, and then both parties can close their channel or connection.  Note
// that if peers ignore Close, deadlock can happen when both peers send Close
// at the same time."
//
// When we don't have a channel, so we must respond with close-ok on a close
// method.  This can happen between a channel exception on an asynchronous
// method like basic.publish and a synchronous close with channel.close.
// In that case, we'll get both a channel.close and channel.close-ok in any
// order.
func (c *Connection) dispatchClosed(f frame) {
	// Only consider method frames, drop content/header frames
	if mf, ok := f.(*methodFrame); ok {
		switch mf.Method.(type) {
		case *channelClose:
			c.send(&methodFrame{
				ChannelId: f.channel(),
				Method:    &channelCloseOk{},
			})
		case *channelCloseOk:
			// we are already closed, so do nothing
		default:
			// unexpected method on closed channel
			c.closeWith(ErrClosed)
		}
	}
}

// Reads each frame off the IO and hand off to the connection object that
// will demux the streams and dispatch to one of the opened channels or
// handle on channel 0 (the connection channel).
func (c *Connection) reader(r io.Reader) {
	buf := bufio.NewReader(r)
	frames := &reader{buf}
	conn, haveDeadliner := r.(readDeadliner)

	for {
		frame, err := frames.ReadFrame()

		if err != nil {
			c.shutdown(&Error{Code: FrameError, Reason: err.Error()})
			return
		}

		c.demux(frame)

		if haveDeadliner {
			c.deadlines <- conn
		}
	}
}

// Ensures that at least one frame is being sent at the tuned interval with a
// jitter tolerance of 1s
func (c *Connection) heartbeater(interval time.Duration, done chan *Error) {
	const maxServerHeartbeatsInFlight = 3

	var sendTicks <-chan time.Time
	if interval > 0 {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		sendTicks = ticker.C
	}

	lastSent := time.Now()

	for {
		select {
		case at, stillSending := <-c.sends:
			// When actively sending, depend on sent frames to reset server timer
			if stillSending {
				lastSent = at
			} else {
				return
			}

		case at := <-sendTicks:
			// When idle, fill the space with a heartbeat frame
			if at.Sub(lastSent) > interval-time.Second {
				if err := c.send(&heartbeatFrame{}); err != nil {
					// send heartbeats even after close/closeOk so we
					// tick until the connection starts erroring
					return
				}
			}

		case conn := <-c.deadlines:
			// When reading, reset our side of the deadline, if we've negotiated one with
			// a deadline that covers at least 2 server heartbeats
			if interval > 0 {
				conn.SetReadDeadline(time.Now().Add(maxServerHeartbeatsInFlight * interval))
			}

		case <-done:
			return
		}
	}
}

// Convenience method to inspect the Connection.Properties["capabilities"]
// Table for server identified capabilities like "basic.ack" or
// "confirm.select".
func (c *Connection) isCapable(featureName string) bool {
	capabilities, _ := c.Properties["capabilities"].(Table)
	hasFeature, _ := capabilities[featureName].(bool)
	return hasFeature
}

// allocateChannel records but does not open a new channel with a unique id.
// This method is the initial part of the channel lifecycle and paired with
// releaseChannel
func (c *Connection) allocateChannel() (*Channel, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.isClosed() {
		return nil, ErrClosed
	}

	id, ok := c.allocator.next()
	if !ok {
		return nil, ErrChannelMax
	}

	ch := newChannel(c, uint16(id))
	c.channels[uint16(id)] = ch

	return ch, nil
}

// releaseChannel removes a channel from the registry as the final part of the
// channel lifecycle
func (c *Connection) releaseChannel(id uint16) {
	c.m.Lock()
	defer c.m.Unlock()

	delete(c.channels, id)
	c.allocator.release(int(id))
}

// openChannel allocates and opens a channel, must be paired with closeChannel
func (c *Connection) openChannel() (*Channel, error) {
	ch, err := c.allocateChannel()
	if err != nil {
		return nil, err
	}

	if err := ch.open(); err != nil {
		c.releaseChannel(ch.id)
		return nil, err
	}
	return ch, nil
}

// closeChannel releases and initiates a shutdown of the channel.  All channel
// closures should be initiated here for proper channel lifecycle management on
// this connection.
func (c *Connection) closeChannel(ch *Channel, e *Error) {
	ch.shutdown(e)
	c.releaseChannel(ch.id)
}

/*
Channel opens a unique, concurrent server channel to process the bulk of AMQP
messages.  Any error from methods on this receiver will render the receiver
invalid and a new Channel should be opened.

*/
func (c *Connection) Channel() (*Channel, error) {
	return c.openChannel()
}

func (c *Connection) call(req message, res ...message) error {
	// Special case for when the protocol header frame is sent insted of a
	// request method
	if req != nil {
		if err := c.send(&methodFrame{ChannelId: 0, Method: req}); err != nil {
			return err
		}
	}

	select {
	case err, ok := <-c.errors:
		if !ok {
			return ErrClosed
		}
		return err

	case msg := <-c.rpc:
		// Try to match one of the result types
		for _, try := range res {
			if reflect.TypeOf(msg) == reflect.TypeOf(try) {
				// *res = *msg
				vres := reflect.ValueOf(try).Elem()
				vmsg := reflect.ValueOf(msg).Elem()
				vres.Set(vmsg)
				return nil
			}
		}
		return ErrCommandInvalid
	}
	// unreachable
}

//    Connection          = open-Connection *use-Connection close-Connection
//    open-Connection     = C:protocol-header
//                          S:START C:START-OK
//                          *challenge
//                          S:TUNE C:TUNE-OK
//                          C:OPEN S:OPEN-OK
//    challenge           = S:SECURE C:SECURE-OK
//    use-Connection      = *channel
//    close-Connection    = C:CLOSE S:CLOSE-OK
//                        / S:CLOSE C:CLOSE-OK
func (c *Connection) open(config Config) error {
	if err := c.send(&protocolHeader{}); err != nil {
		return err
	}

	return c.openStart(config)
}

func (c *Connection) openStart(config Config) error {
	start := &connectionStart{}

	if err := c.call(nil, start); err != nil {
		return err
	}

	c.Major = int(start.VersionMajor)
	c.Minor = int(start.VersionMinor)
	c.Properties = Table(start.ServerProperties)
	c.Locales = strings.Split(start.Locales, " ")

	// eventually support challenge/response here by also responding to
	// connectionSecure.
	auth, ok := pickSASLMechanism(config.SASL, strings.Split(start.Mechanisms, " "))
	if !ok {
		return ErrSASL
	}

	// Save this mechanism off as the one we chose
	c.Config.SASL = []Authentication{auth}

	// Set the connection locale to client locale
	c.Config.Locale = config.Locale

	return c.openTune(config, auth)
}

func (c *Connection) openTune(config Config, auth Authentication) error {
	if len(config.Properties) == 0 {
		config.Properties = Table{
			"product": defaultProduct,
			"version": defaultVersion,
		}
	}

	config.Properties["capabilities"] = Table{
		"connection.blocked":     true,
		"consumer_cancel_notify": true,
	}

	ok := &connectionStartOk{
		ClientProperties: config.Properties,
		Mechanism:        auth.Mechanism(),
		Response:         auth.Response(),
		Locale:           config.Locale,
	}
	tune := &connectionTune{}

	if err := c.call(ok, tune); err != nil {
		// per spec, a connection can only be closed when it has been opened
		// so at this point, we know it's an auth error, but the socket
		// was closed instead.  Return a meaningful error.
		return ErrCredentials
	}

	// When the server and client both use default 0, then the max channel is
	// only limited by uint16.
	c.Config.ChannelMax = pick(config.ChannelMax, int(tune.ChannelMax))
	if c.Config.ChannelMax == 0 {
		c.Config.ChannelMax = defaultChannelMax
	}
	c.Config.ChannelMax = min(c.Config.ChannelMax, maxChannelMax)

	// Frame size includes headers and end byte (len(payload)+8), even if
	// this is less than FrameMinSize, use what the server sends because the
	// alternative is to stop the handshake here.
	c.Config.FrameSize = pick(config.FrameSize, int(tune.FrameMax))

	// Save this off for resetDeadline()
	c.Config.Heartbeat = time.Second * time.Duration(pick(
		int(config.Heartbeat/time.Second),
		int(tune.Heartbeat)))

	// "The client should start sending heartbeats after receiving a
	// Connection.Tune method"
	go c.heartbeater(c.Config.Heartbeat, c.NotifyClose(make(chan *Error, 1)))

	if err := c.send(&methodFrame{
		ChannelId: 0,
		Method: &connectionTuneOk{
			ChannelMax: uint16(c.Config.ChannelMax),
			FrameMax:   uint32(c.Config.FrameSize),
			Heartbeat:  uint16(c.Config.Heartbeat / time.Second),
		},
	}); err != nil {
		return err
	}

	return c.openVhost(config)
}

func (c *Connection) openVhost(config Config) error {
	req := &connectionOpen{VirtualHost: config.Vhost}
	res := &connectionOpenOk{}

	if err := c.call(req, res); err != nil {
		// Cannot be closed yet, but we know it's a vhost problem
		return ErrVhost
	}

	c.Config.Vhost = config.Vhost

	return c.openComplete()
}

// openComplete performs any final Connection initialization dependent on the
// connection handshake and clears any state needed for TLS and AMQP handshaking.
func (c *Connection) openComplete() error {
	// We clear the deadlines and let the heartbeater reset the read deadline if requested.
	// RabbitMQ uses TCP flow control at this point for pushback so Writes can
	// intentionally block.
	if deadliner, ok := c.conn.(interface {
		SetDeadline(time.Time) error
	}); ok {
		_ = deadliner.SetDeadline(time.Time{})
	}

	c.allocator = newAllocator(1, c.Config.ChannelMax)
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func pick(client, server int) int {
	if client == 0 || server == 0 {
		return max(client, server)
	}
	return min(client, server)
}
