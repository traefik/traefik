// Copyright 2017, Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AUTO-GENERATED CODE. DO NOT EDIT.

package database

import (
	"math"
	"time"

	"cloud.google.com/go/internal/version"
	"cloud.google.com/go/longrunning"
	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
	longrunningpb "google.golang.org/genproto/googleapis/longrunning"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	databaseAdminInstancePathTemplate = gax.MustCompilePathTemplate("projects/{project}/instances/{instance}")
	databaseAdminDatabasePathTemplate = gax.MustCompilePathTemplate("projects/{project}/instances/{instance}/databases/{database}")
)

// DatabaseAdminCallOptions contains the retry settings for each method of DatabaseAdminClient.
type DatabaseAdminCallOptions struct {
	ListDatabases      []gax.CallOption
	CreateDatabase     []gax.CallOption
	GetDatabase        []gax.CallOption
	UpdateDatabaseDdl  []gax.CallOption
	DropDatabase       []gax.CallOption
	GetDatabaseDdl     []gax.CallOption
	SetIamPolicy       []gax.CallOption
	GetIamPolicy       []gax.CallOption
	TestIamPermissions []gax.CallOption
}

func defaultDatabaseAdminClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("spanner.googleapis.com:443"),
		option.WithScopes(
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/spanner.admin",
		),
	}
}

func defaultDatabaseAdminCallOptions() *DatabaseAdminCallOptions {
	retry := map[[2]string][]gax.CallOption{
		{"default", "idempotent"}: {
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    1000 * time.Millisecond,
					Max:        32000 * time.Millisecond,
					Multiplier: 1.3,
				})
			}),
		},
		{"default", "non_idempotent"}: {
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    1000 * time.Millisecond,
					Max:        32000 * time.Millisecond,
					Multiplier: 1.3,
				})
			}),
		},
	}
	return &DatabaseAdminCallOptions{
		ListDatabases:      retry[[2]string{"default", "idempotent"}],
		CreateDatabase:     retry[[2]string{"default", "non_idempotent"}],
		GetDatabase:        retry[[2]string{"default", "idempotent"}],
		UpdateDatabaseDdl:  retry[[2]string{"default", "idempotent"}],
		DropDatabase:       retry[[2]string{"default", "idempotent"}],
		GetDatabaseDdl:     retry[[2]string{"default", "idempotent"}],
		SetIamPolicy:       retry[[2]string{"default", "non_idempotent"}],
		GetIamPolicy:       retry[[2]string{"default", "idempotent"}],
		TestIamPermissions: retry[[2]string{"default", "non_idempotent"}],
	}
}

// DatabaseAdminClient is a client for interacting with Cloud Spanner Database Admin API.
type DatabaseAdminClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	databaseAdminClient databasepb.DatabaseAdminClient

	// The call options for this service.
	CallOptions *DatabaseAdminCallOptions

	// The metadata to be sent with each request.
	xGoogHeader string
}

// NewDatabaseAdminClient creates a new database admin client.
//
// Cloud Spanner Database Admin API
//
// The Cloud Spanner Database Admin API can be used to create, drop, and
// list databases. It also enables updating the schema of pre-existing
// databases.
func NewDatabaseAdminClient(ctx context.Context, opts ...option.ClientOption) (*DatabaseAdminClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultDatabaseAdminClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &DatabaseAdminClient{
		conn:        conn,
		CallOptions: defaultDatabaseAdminCallOptions(),

		databaseAdminClient: databasepb.NewDatabaseAdminClient(conn),
	}
	c.SetGoogleClientInfo()
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *DatabaseAdminClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *DatabaseAdminClient) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *DatabaseAdminClient) SetGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", version.Go()}, keyval...)
	kv = append(kv, "gapic", version.Repo, "gax", gax.Version, "grpc", "")
	c.xGoogHeader = gax.XGoogHeader(kv...)
}

// DatabaseAdminInstancePath returns the path for the instance resource.
func DatabaseAdminInstancePath(project, instance string) string {
	path, err := databaseAdminInstancePathTemplate.Render(map[string]string{
		"project":  project,
		"instance": instance,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// DatabaseAdminDatabasePath returns the path for the database resource.
func DatabaseAdminDatabasePath(project, instance, database string) string {
	path, err := databaseAdminDatabasePathTemplate.Render(map[string]string{
		"project":  project,
		"instance": instance,
		"database": database,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// ListDatabases lists Cloud Spanner databases.
func (c *DatabaseAdminClient) ListDatabases(ctx context.Context, req *databasepb.ListDatabasesRequest) *DatabaseIterator {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	it := &DatabaseIterator{}
	it.InternalFetch = func(pageSize int, pageToken string) ([]*databasepb.Database, string, error) {
		var resp *databasepb.ListDatabasesResponse
		req.PageToken = pageToken
		if pageSize > math.MaxInt32 {
			req.PageSize = math.MaxInt32
		} else {
			req.PageSize = int32(pageSize)
		}
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			resp, err = c.databaseAdminClient.ListDatabases(ctx, req)
			return err
		}, c.CallOptions.ListDatabases...)
		if err != nil {
			return nil, "", err
		}
		return resp.Databases, resp.NextPageToken, nil
	}
	fetch := func(pageSize int, pageToken string) (string, error) {
		items, nextPageToken, err := it.InternalFetch(pageSize, pageToken)
		if err != nil {
			return "", err
		}
		it.items = append(it.items, items...)
		return nextPageToken, nil
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(fetch, it.bufLen, it.takeBuf)
	return it
}

// CreateDatabase creates a new Cloud Spanner database and starts to prepare it for serving.
// The returned [long-running operation][google.longrunning.Operation] will
// have a name of the format `<database_name>/operations/<operation_id>` and
// can be used to track preparation of the database. The
// [metadata][google.longrunning.Operation.metadata] field type is
// [CreateDatabaseMetadata][google.spanner.admin.database.v1.CreateDatabaseMetadata]. The
// [response][google.longrunning.Operation.response] field type is
// [Database][google.spanner.admin.database.v1.Database], if successful.
func (c *DatabaseAdminClient) CreateDatabase(ctx context.Context, req *databasepb.CreateDatabaseRequest) (*DatabaseOperation, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *longrunningpb.Operation
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.databaseAdminClient.CreateDatabase(ctx, req)
		return err
	}, c.CallOptions.CreateDatabase...)
	if err != nil {
		return nil, err
	}
	return &DatabaseOperation{
		lro: longrunning.InternalNewOperation(c.Connection(), resp),
	}, nil
}

// GetDatabase gets the state of a Cloud Spanner database.
func (c *DatabaseAdminClient) GetDatabase(ctx context.Context, req *databasepb.GetDatabaseRequest) (*databasepb.Database, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *databasepb.Database
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.databaseAdminClient.GetDatabase(ctx, req)
		return err
	}, c.CallOptions.GetDatabase...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateDatabaseDdl updates the schema of a Cloud Spanner database by
// creating/altering/dropping tables, columns, indexes, etc. The returned
// [long-running operation][google.longrunning.Operation] will have a name of
// the format `<database_name>/operations/<operation_id>` and can be used to
// track execution of the schema change(s). The
// [metadata][google.longrunning.Operation.metadata] field type is
// [UpdateDatabaseDdlMetadata][google.spanner.admin.database.v1.UpdateDatabaseDdlMetadata].  The operation has no response.
func (c *DatabaseAdminClient) UpdateDatabaseDdl(ctx context.Context, req *databasepb.UpdateDatabaseDdlRequest) (*EmptyOperation, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *longrunningpb.Operation
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.databaseAdminClient.UpdateDatabaseDdl(ctx, req)
		return err
	}, c.CallOptions.UpdateDatabaseDdl...)
	if err != nil {
		return nil, err
	}
	return &EmptyOperation{
		lro: longrunning.InternalNewOperation(c.Connection(), resp),
	}, nil
}

// DropDatabase drops (aka deletes) a Cloud Spanner database.
func (c *DatabaseAdminClient) DropDatabase(ctx context.Context, req *databasepb.DropDatabaseRequest) error {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.databaseAdminClient.DropDatabase(ctx, req)
		return err
	}, c.CallOptions.DropDatabase...)
	return err
}

// GetDatabaseDdl returns the schema of a Cloud Spanner database as a list of formatted
// DDL statements. This method does not show pending schema updates, those may
// be queried using the [Operations][google.longrunning.Operations] API.
func (c *DatabaseAdminClient) GetDatabaseDdl(ctx context.Context, req *databasepb.GetDatabaseDdlRequest) (*databasepb.GetDatabaseDdlResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *databasepb.GetDatabaseDdlResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.databaseAdminClient.GetDatabaseDdl(ctx, req)
		return err
	}, c.CallOptions.GetDatabaseDdl...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SetIamPolicy sets the access control policy on a database resource. Replaces any
// existing policy.
//
// Authorization requires `spanner.databases.setIamPolicy` permission on
// [resource][google.iam.v1.SetIamPolicyRequest.resource].
func (c *DatabaseAdminClient) SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *iampb.Policy
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.databaseAdminClient.SetIamPolicy(ctx, req)
		return err
	}, c.CallOptions.SetIamPolicy...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetIamPolicy gets the access control policy for a database resource. Returns an empty
// policy if a database exists but does not have a policy set.
//
// Authorization requires `spanner.databases.getIamPolicy` permission on
// [resource][google.iam.v1.GetIamPolicyRequest.resource].
func (c *DatabaseAdminClient) GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest) (*iampb.Policy, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *iampb.Policy
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.databaseAdminClient.GetIamPolicy(ctx, req)
		return err
	}, c.CallOptions.GetIamPolicy...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// TestIamPermissions returns permissions that the caller has on the specified database resource.
//
// Attempting this RPC on a non-existent Cloud Spanner database will result in
// a NOT_FOUND error if the user has `spanner.databases.list` permission on
// the containing Cloud Spanner instance. Otherwise returns an empty set of
// permissions.
func (c *DatabaseAdminClient) TestIamPermissions(ctx context.Context, req *iampb.TestIamPermissionsRequest) (*iampb.TestIamPermissionsResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *iampb.TestIamPermissionsResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.databaseAdminClient.TestIamPermissions(ctx, req)
		return err
	}, c.CallOptions.TestIamPermissions...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DatabaseIterator manages a stream of *databasepb.Database.
type DatabaseIterator struct {
	items    []*databasepb.Database
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*databasepb.Database, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *DatabaseIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *DatabaseIterator) Next() (*databasepb.Database, error) {
	var item *databasepb.Database
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *DatabaseIterator) bufLen() int {
	return len(it.items)
}

func (it *DatabaseIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// DatabaseOperation manages a long-running operation yielding databasepb.Database.
type DatabaseOperation struct {
	lro *longrunning.Operation
}

// DatabaseOperation returns a new DatabaseOperation from a given name.
// The name must be that of a previously created DatabaseOperation, possibly from a different process.
func (c *DatabaseAdminClient) DatabaseOperation(name string) *DatabaseOperation {
	return &DatabaseOperation{
		lro: longrunning.InternalNewOperation(c.Connection(), &longrunningpb.Operation{Name: name}),
	}
}

// Wait blocks until the long-running operation is completed, returning the response and any errors encountered.
//
// See documentation of Poll for error-handling information.
func (op *DatabaseOperation) Wait(ctx context.Context) (*databasepb.Database, error) {
	var resp databasepb.Database
	if err := op.lro.Wait(ctx, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Poll fetches the latest state of the long-running operation.
//
// Poll also fetches the latest metadata, which can be retrieved by Metadata.
//
// If Poll fails, the error is returned and op is unmodified. If Poll succeeds and
// the operation has completed with failure, the error is returned and op.Done will return true.
// If Poll succeeds and the operation has completed successfully,
// op.Done will return true, and the response of the operation is returned.
// If Poll succeeds and the operation has not completed, the returned response and error are both nil.
func (op *DatabaseOperation) Poll(ctx context.Context) (*databasepb.Database, error) {
	var resp databasepb.Database
	if err := op.lro.Poll(ctx, &resp); err != nil {
		return nil, err
	}
	if !op.Done() {
		return nil, nil
	}
	return &resp, nil
}

// Metadata returns metadata associated with the long-running operation.
// Metadata itself does not contact the server, but Poll does.
// To get the latest metadata, call this method after a successful call to Poll.
// If the metadata is not available, the returned metadata and error are both nil.
func (op *DatabaseOperation) Metadata() (*databasepb.CreateDatabaseMetadata, error) {
	var meta databasepb.CreateDatabaseMetadata
	if err := op.lro.Metadata(&meta); err == longrunning.ErrNoMetadata {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &meta, nil
}

// Done reports whether the long-running operation has completed.
func (op *DatabaseOperation) Done() bool {
	return op.lro.Done()
}

// Name returns the name of the long-running operation.
// The name is assigned by the server and is unique within the service from which the operation is created.
func (op *DatabaseOperation) Name() string {
	return op.lro.Name()
}

// EmptyOperation manages a long-running operation with no result.
type EmptyOperation struct {
	lro *longrunning.Operation
}

// EmptyOperation returns a new EmptyOperation from a given name.
// The name must be that of a previously created EmptyOperation, possibly from a different process.
func (c *DatabaseAdminClient) EmptyOperation(name string) *EmptyOperation {
	return &EmptyOperation{
		lro: longrunning.InternalNewOperation(c.Connection(), &longrunningpb.Operation{Name: name}),
	}
}

// Wait blocks until the long-running operation is completed, returning any error encountered.
//
// See documentation of Poll for error-handling information.
func (op *EmptyOperation) Wait(ctx context.Context) error {
	return op.lro.Wait(ctx, nil)
}

// Poll fetches the latest state of the long-running operation.
//
// Poll also fetches the latest metadata, which can be retrieved by Metadata.
//
// If Poll fails, the error is returned and op is unmodified. If Poll succeeds and
// the operation has completed with failure, the error is returned and op.Done will return true.
// If Poll succeeds and the operation has completed successfully, op.Done will return true.
func (op *EmptyOperation) Poll(ctx context.Context) error {
	return op.lro.Poll(ctx, nil)
}

// Metadata returns metadata associated with the long-running operation.
// Metadata itself does not contact the server, but Poll does.
// To get the latest metadata, call this method after a successful call to Poll.
// If the metadata is not available, the returned metadata and error are both nil.
func (op *EmptyOperation) Metadata() (*databasepb.UpdateDatabaseDdlMetadata, error) {
	var meta databasepb.UpdateDatabaseDdlMetadata
	if err := op.lro.Metadata(&meta); err == longrunning.ErrNoMetadata {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &meta, nil
}

// Done reports whether the long-running operation has completed.
func (op *EmptyOperation) Done() bool {
	return op.lro.Done()
}

// Name returns the name of the long-running operation.
// The name is assigned by the server and is unique within the service from which the operation is created.
func (op *EmptyOperation) Name() string {
	return op.lro.Name()
}
