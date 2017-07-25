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

package monitoring

import (
	"math"
	"time"

	"cloud.google.com/go/internal/version"
	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	metricProjectPathTemplate                     = gax.MustCompilePathTemplate("projects/{project}")
	metricMetricDescriptorPathTemplate            = gax.MustCompilePathTemplate("projects/{project}/metricDescriptors/{metric_descriptor=**}")
	metricMonitoredResourceDescriptorPathTemplate = gax.MustCompilePathTemplate("projects/{project}/monitoredResourceDescriptors/{monitored_resource_descriptor}")
)

// MetricCallOptions contains the retry settings for each method of MetricClient.
type MetricCallOptions struct {
	ListMonitoredResourceDescriptors []gax.CallOption
	GetMonitoredResourceDescriptor   []gax.CallOption
	ListMetricDescriptors            []gax.CallOption
	GetMetricDescriptor              []gax.CallOption
	CreateMetricDescriptor           []gax.CallOption
	DeleteMetricDescriptor           []gax.CallOption
	ListTimeSeries                   []gax.CallOption
	CreateTimeSeries                 []gax.CallOption
}

func defaultMetricClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("monitoring.googleapis.com:443"),
		option.WithScopes(
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/monitoring",
			"https://www.googleapis.com/auth/monitoring.read",
			"https://www.googleapis.com/auth/monitoring.write",
		),
	}
}

func defaultMetricCallOptions() *MetricCallOptions {
	retry := map[[2]string][]gax.CallOption{
		{"default", "idempotent"}: {
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.3,
				})
			}),
		},
		{"default", "non_idempotent"}: {
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.3,
				})
			}),
		},
	}
	return &MetricCallOptions{
		ListMonitoredResourceDescriptors: retry[[2]string{"default", "idempotent"}],
		GetMonitoredResourceDescriptor:   retry[[2]string{"default", "idempotent"}],
		ListMetricDescriptors:            retry[[2]string{"default", "idempotent"}],
		GetMetricDescriptor:              retry[[2]string{"default", "idempotent"}],
		CreateMetricDescriptor:           retry[[2]string{"default", "non_idempotent"}],
		DeleteMetricDescriptor:           retry[[2]string{"default", "idempotent"}],
		ListTimeSeries:                   retry[[2]string{"default", "idempotent"}],
		CreateTimeSeries:                 retry[[2]string{"default", "non_idempotent"}],
	}
}

// MetricClient is a client for interacting with Stackdriver Monitoring API.
type MetricClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	metricClient monitoringpb.MetricServiceClient

	// The call options for this service.
	CallOptions *MetricCallOptions

	// The metadata to be sent with each request.
	xGoogHeader string
}

// NewMetricClient creates a new metric service client.
//
// Manages metric descriptors, monitored resource descriptors, and
// time series data.
func NewMetricClient(ctx context.Context, opts ...option.ClientOption) (*MetricClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultMetricClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &MetricClient{
		conn:        conn,
		CallOptions: defaultMetricCallOptions(),

		metricClient: monitoringpb.NewMetricServiceClient(conn),
	}
	c.SetGoogleClientInfo()
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *MetricClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *MetricClient) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *MetricClient) SetGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", version.Go()}, keyval...)
	kv = append(kv, "gapic", version.Repo, "gax", gax.Version, "grpc", "")
	c.xGoogHeader = gax.XGoogHeader(kv...)
}

// MetricProjectPath returns the path for the project resource.
func MetricProjectPath(project string) string {
	path, err := metricProjectPathTemplate.Render(map[string]string{
		"project": project,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// MetricMetricDescriptorPath returns the path for the metric descriptor resource.
func MetricMetricDescriptorPath(project, metricDescriptor string) string {
	path, err := metricMetricDescriptorPathTemplate.Render(map[string]string{
		"project":           project,
		"metric_descriptor": metricDescriptor,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// MetricMonitoredResourceDescriptorPath returns the path for the monitored resource descriptor resource.
func MetricMonitoredResourceDescriptorPath(project, monitoredResourceDescriptor string) string {
	path, err := metricMonitoredResourceDescriptorPathTemplate.Render(map[string]string{
		"project":                       project,
		"monitored_resource_descriptor": monitoredResourceDescriptor,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// ListMonitoredResourceDescriptors lists monitored resource descriptors that match a filter. This method does not require a Stackdriver account.
func (c *MetricClient) ListMonitoredResourceDescriptors(ctx context.Context, req *monitoringpb.ListMonitoredResourceDescriptorsRequest) *MonitoredResourceDescriptorIterator {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	it := &MonitoredResourceDescriptorIterator{}
	it.InternalFetch = func(pageSize int, pageToken string) ([]*monitoredrespb.MonitoredResourceDescriptor, string, error) {
		var resp *monitoringpb.ListMonitoredResourceDescriptorsResponse
		req.PageToken = pageToken
		if pageSize > math.MaxInt32 {
			req.PageSize = math.MaxInt32
		} else {
			req.PageSize = int32(pageSize)
		}
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			resp, err = c.metricClient.ListMonitoredResourceDescriptors(ctx, req)
			return err
		}, c.CallOptions.ListMonitoredResourceDescriptors...)
		if err != nil {
			return nil, "", err
		}
		return resp.ResourceDescriptors, resp.NextPageToken, nil
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

// GetMonitoredResourceDescriptor gets a single monitored resource descriptor. This method does not require a Stackdriver account.
func (c *MetricClient) GetMonitoredResourceDescriptor(ctx context.Context, req *monitoringpb.GetMonitoredResourceDescriptorRequest) (*monitoredrespb.MonitoredResourceDescriptor, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *monitoredrespb.MonitoredResourceDescriptor
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.metricClient.GetMonitoredResourceDescriptor(ctx, req)
		return err
	}, c.CallOptions.GetMonitoredResourceDescriptor...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListMetricDescriptors lists metric descriptors that match a filter. This method does not require a Stackdriver account.
func (c *MetricClient) ListMetricDescriptors(ctx context.Context, req *monitoringpb.ListMetricDescriptorsRequest) *MetricDescriptorIterator {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	it := &MetricDescriptorIterator{}
	it.InternalFetch = func(pageSize int, pageToken string) ([]*metricpb.MetricDescriptor, string, error) {
		var resp *monitoringpb.ListMetricDescriptorsResponse
		req.PageToken = pageToken
		if pageSize > math.MaxInt32 {
			req.PageSize = math.MaxInt32
		} else {
			req.PageSize = int32(pageSize)
		}
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			resp, err = c.metricClient.ListMetricDescriptors(ctx, req)
			return err
		}, c.CallOptions.ListMetricDescriptors...)
		if err != nil {
			return nil, "", err
		}
		return resp.MetricDescriptors, resp.NextPageToken, nil
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

// GetMetricDescriptor gets a single metric descriptor. This method does not require a Stackdriver account.
func (c *MetricClient) GetMetricDescriptor(ctx context.Context, req *monitoringpb.GetMetricDescriptorRequest) (*metricpb.MetricDescriptor, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *metricpb.MetricDescriptor
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.metricClient.GetMetricDescriptor(ctx, req)
		return err
	}, c.CallOptions.GetMetricDescriptor...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateMetricDescriptor creates a new metric descriptor.
// User-created metric descriptors define
// [custom metrics](/monitoring/custom-metrics).
func (c *MetricClient) CreateMetricDescriptor(ctx context.Context, req *monitoringpb.CreateMetricDescriptorRequest) (*metricpb.MetricDescriptor, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *metricpb.MetricDescriptor
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.metricClient.CreateMetricDescriptor(ctx, req)
		return err
	}, c.CallOptions.CreateMetricDescriptor...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteMetricDescriptor deletes a metric descriptor. Only user-created
// [custom metrics](/monitoring/custom-metrics) can be deleted.
func (c *MetricClient) DeleteMetricDescriptor(ctx context.Context, req *monitoringpb.DeleteMetricDescriptorRequest) error {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.metricClient.DeleteMetricDescriptor(ctx, req)
		return err
	}, c.CallOptions.DeleteMetricDescriptor...)
	return err
}

// ListTimeSeries lists time series that match a filter. This method does not require a Stackdriver account.
func (c *MetricClient) ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) *TimeSeriesIterator {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	it := &TimeSeriesIterator{}
	it.InternalFetch = func(pageSize int, pageToken string) ([]*monitoringpb.TimeSeries, string, error) {
		var resp *monitoringpb.ListTimeSeriesResponse
		req.PageToken = pageToken
		if pageSize > math.MaxInt32 {
			req.PageSize = math.MaxInt32
		} else {
			req.PageSize = int32(pageSize)
		}
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			resp, err = c.metricClient.ListTimeSeries(ctx, req)
			return err
		}, c.CallOptions.ListTimeSeries...)
		if err != nil {
			return nil, "", err
		}
		return resp.TimeSeries, resp.NextPageToken, nil
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

// CreateTimeSeries creates or adds data to one or more time series.
// The response is empty if all time series in the request were written.
// If any time series could not be written, a corresponding failure message is
// included in the error response.
func (c *MetricClient) CreateTimeSeries(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.metricClient.CreateTimeSeries(ctx, req)
		return err
	}, c.CallOptions.CreateTimeSeries...)
	return err
}

// MetricDescriptorIterator manages a stream of *metricpb.MetricDescriptor.
type MetricDescriptorIterator struct {
	items    []*metricpb.MetricDescriptor
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*metricpb.MetricDescriptor, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *MetricDescriptorIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *MetricDescriptorIterator) Next() (*metricpb.MetricDescriptor, error) {
	var item *metricpb.MetricDescriptor
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *MetricDescriptorIterator) bufLen() int {
	return len(it.items)
}

func (it *MetricDescriptorIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// MonitoredResourceDescriptorIterator manages a stream of *monitoredrespb.MonitoredResourceDescriptor.
type MonitoredResourceDescriptorIterator struct {
	items    []*monitoredrespb.MonitoredResourceDescriptor
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoredrespb.MonitoredResourceDescriptor, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *MonitoredResourceDescriptorIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *MonitoredResourceDescriptorIterator) Next() (*monitoredrespb.MonitoredResourceDescriptor, error) {
	var item *monitoredrespb.MonitoredResourceDescriptor
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *MonitoredResourceDescriptorIterator) bufLen() int {
	return len(it.items)
}

func (it *MonitoredResourceDescriptorIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// TimeSeriesIterator manages a stream of *monitoringpb.TimeSeries.
type TimeSeriesIterator struct {
	items    []*monitoringpb.TimeSeries
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.TimeSeries, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *TimeSeriesIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *TimeSeriesIterator) Next() (*monitoringpb.TimeSeries, error) {
	var item *monitoringpb.TimeSeries
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *TimeSeriesIterator) bufLen() int {
	return len(it.items)
}

func (it *TimeSeriesIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}
