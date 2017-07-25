/*
Copyright 2017 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spanner

import (
	"fmt"
	"regexp"
	"runtime"
	"sync/atomic"
	"time"

	"cloud.google.com/go/internal/version"

	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	sppb "google.golang.org/genproto/googleapis/spanner/v1"
)

const (
	prodAddr = "spanner.googleapis.com:443"

	// resourcePrefixHeader is the name of the metadata header used to indicate
	// the resource being operated on.
	resourcePrefixHeader = "google-cloud-resource-prefix"
	// apiClientHeader is the name of the metadata header used to indicate client
	// information.
	apiClientHeader = "x-goog-api-client"
)

const (
	// Scope is the scope for Cloud Spanner Data API.
	Scope = "https://www.googleapis.com/auth/spanner.data"

	// AdminScope is the scope for Cloud Spanner Admin APIs.
	AdminScope = "https://www.googleapis.com/auth/spanner.admin"
)

var (
	validDBPattern  = regexp.MustCompile("^projects/[^/]+/instances/[^/]+/databases/[^/]+$")
	clientUserAgent = fmt.Sprintf("cloudspanner go/%s", runtime.Version())
)

func validDatabaseName(db string) error {
	if matched := validDBPattern.MatchString(db); !matched {
		return fmt.Errorf("database name %q should conform to pattern %q",
			db, validDBPattern.String())
	}
	return nil
}

// Client is a client for reading and writing data to a Cloud Spanner database.  A
// client is safe to use concurrently, except for its Close method.
type Client struct {
	// rr must be accessed through atomic operations.
	rr       uint32
	conns    []*grpc.ClientConn
	clients  []sppb.SpannerClient
	database string
	// Metadata to be sent with each request.
	md           metadata.MD
	idleSessions *sessionPool
}

// ClientConfig has configurations for the client.
type ClientConfig struct {
	// NumChannels is the number of GRPC channels.
	NumChannels int
	co          []option.ClientOption
	// SessionPoolConfig is the configuration for session pool.
	SessionPoolConfig
}

// errDial returns error for dialing to Cloud Spanner.
func errDial(ci int, err error) error {
	e := toSpannerError(err).(*Error)
	e.decorate(fmt.Sprintf("dialing fails for channel[%v]", ci))
	return e
}

func contextWithMetadata(ctx context.Context, md metadata.MD) context.Context {
	existing, ok := metadata.FromContext(ctx)
	if ok {
		md = metadata.Join(existing, md)
	}
	return metadata.NewContext(ctx, md)
}

// NewClient creates a client to a database. A valid database name has the
// form projects/PROJECT_ID/instances/INSTANCE_ID/databases/DATABASE_ID. It uses a default
// configuration.
func NewClient(ctx context.Context, database string, opts ...option.ClientOption) (*Client, error) {
	return NewClientWithConfig(ctx, database, ClientConfig{}, opts...)
}

// NewClientWithConfig creates a client to a database. A valid database name has the
// form projects/PROJECT_ID/instances/INSTANCE_ID/databases/DATABASE_ID.
func NewClientWithConfig(ctx context.Context, database string, config ClientConfig, opts ...option.ClientOption) (*Client, error) {
	// Validate database path.
	if err := validDatabaseName(database); err != nil {
		return nil, err
	}
	c := &Client{
		database: database,
		md: metadata.Pairs(
			resourcePrefixHeader, database,
			apiClientHeader, clientUserAgent,
			"x-goog-api-client", fmt.Sprintf("gl-go/%s gccl/%s grpc/", version.Go(), version.Repo)),
	}
	allOpts := []option.ClientOption{option.WithEndpoint(prodAddr), option.WithScopes(Scope), option.WithUserAgent(clientUserAgent)}
	allOpts = append(allOpts, opts...)
	// Prepare gRPC channels.
	if config.NumChannels == 0 {
		config.NumChannels = 4
	}
	for i := 0; i < config.NumChannels; i++ {
		conn, err := transport.DialGRPC(ctx, allOpts...)
		if err != nil {
			return nil, errDial(i, err)
		}
		c.conns = append(c.conns, conn)
		c.clients = append(c.clients, sppb.NewSpannerClient(conn))
	}
	// Prepare session pool.
	config.SessionPoolConfig.getRPCClient = func() (sppb.SpannerClient, error) {
		// TODO: support more loadbalancing options.
		return c.rrNext(), nil
	}
	sp, err := newSessionPool(database, config.SessionPoolConfig, c.md)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.idleSessions = sp
	return c, nil
}

// rrNext returns the next available Cloud Spanner RPC client in a round-robin manner.
func (c *Client) rrNext() sppb.SpannerClient {
	return c.clients[atomic.AddUint32(&c.rr, 1)%uint32(len(c.clients))]
}

// Close closes the client.
func (c *Client) Close() {
	if c.idleSessions != nil {
		c.idleSessions.close()
	}
	for _, conn := range c.conns {
		conn.Close()
	}
}

// Single provides a read-only snapshot transaction optimized for the case
// where only a single read or query is needed.  This is more efficient than
// using ReadOnlyTransaction() for a single read or query.
//
// Single will use a strong TimestampBound by default. Use
// ReadOnlyTransaction.WithTimestampBound to specify a different
// TimestampBound. A non-strong bound can be used to reduce latency, or
// "time-travel" to prior versions of the database, see the documentation of
// TimestampBound for details.
func (c *Client) Single() *ReadOnlyTransaction {
	t := &ReadOnlyTransaction{singleUse: true, sp: c.idleSessions}
	t.txReadOnly.txReadEnv = t
	return t
}

// ReadOnlyTransaction returns a ReadOnlyTransaction that can be used for
// multiple reads from the database.  You must call Close() when the
// ReadOnlyTransaction is no longer needed to release resources on the server.
//
// ReadOnlyTransaction will use a strong TimestampBound by default.  Use
// ReadOnlyTransaction.WithTimestampBound to specify a different
// TimestampBound.  A non-strong bound can be used to reduce latency, or
// "time-travel" to prior versions of the database, see the documentation of
// TimestampBound for details.
func (c *Client) ReadOnlyTransaction() *ReadOnlyTransaction {
	t := &ReadOnlyTransaction{
		singleUse:       false,
		sp:              c.idleSessions,
		txReadyOrClosed: make(chan struct{}),
	}
	t.txReadOnly.txReadEnv = t
	return t
}

// ReadWriteTransaction executes a read-write transaction, with retries as
// necessary.
//
// The function f will be called one or more times. It must not maintain
// any state between calls.
//
// If the transaction cannot be committed or if f returns an IsAborted error,
// ReadWriteTransaction will call f again. It will continue to call f until the
// transaction can be committed or the Context times out or is cancelled.  If f
// returns an error other than IsAborted, ReadWriteTransaction will abort the
// transaction and return the error.
//
// To limit the number of retries, set a deadline on the Context rather than
// using a fixed limit on the number of attempts. ReadWriteTransaction will
// retry as needed until that deadline is met.
func (c *Client) ReadWriteTransaction(ctx context.Context, f func(t *ReadWriteTransaction) error) (time.Time, error) {
	var (
		ts time.Time
		sh *sessionHandle
	)
	err := runRetryable(ctx, func(ctx context.Context) error {
		var (
			err error
			t   *ReadWriteTransaction
		)
		if sh == nil || sh.getID() == "" || sh.getClient() == nil {
			// Session handle hasn't been allocated or has been destroyed.
			sh, err = c.idleSessions.takeWriteSession(ctx)
			if err != nil {
				// If session retrieval fails, just fail the transaction.
				return err
			}
			t = &ReadWriteTransaction{
				sh: sh,
				tx: sh.getTransactionID(),
			}
		} else {
			t = &ReadWriteTransaction{
				sh: sh,
			}
		}
		t.txReadOnly.txReadEnv = t
		if err = t.begin(ctx); err != nil {
			// Mask error from begin operation as retryable error.
			return errRetry(err)
		}
		ts, err = t.runInTransaction(ctx, f)
		if err != nil {
			return err
		}
		return nil
	})
	if sh != nil {
		sh.recycle()
	}
	return ts, err
}

// applyOption controls the behavior of Client.Apply.
type applyOption struct {
	// If atLeastOnce == true, Client.Apply will execute the mutations on Cloud Spanner at least once.
	atLeastOnce bool
}

// An ApplyOption is an optional argument to Apply.
type ApplyOption func(*applyOption)

// ApplyAtLeastOnce returns an ApplyOption that removes replay protection.
//
// With this option, Apply may attempt to apply mutations more than once; if
// the mutations are not idempotent, this may lead to a failure being reported
// when the mutation was applied more than once. For example, an insert may
// fail with ALREADY_EXISTS even though the row did not exist before Apply was
// called. For this reason, most users of the library will prefer not to use
// this option.  However, ApplyAtLeastOnce requires only a single RPC, whereas
// Apply's default replay protection may require an additional RPC.  So this
// option may be appropriate for latency sensitive and/or high throughput blind
// writing.
func ApplyAtLeastOnce() ApplyOption {
	return func(ao *applyOption) {
		ao.atLeastOnce = true
	}
}

// Apply applies a list of mutations atomically to the database.
func (c *Client) Apply(ctx context.Context, ms []*Mutation, opts ...ApplyOption) (time.Time, error) {
	ao := &applyOption{}
	for _, opt := range opts {
		opt(ao)
	}
	if !ao.atLeastOnce {
		return c.ReadWriteTransaction(ctx, func(t *ReadWriteTransaction) error {
			t.BufferWrite(ms)
			return nil
		})
	}
	t := &writeOnlyTransaction{c.idleSessions}
	return t.applyAtLeastOnce(ctx, ms...)
}
