/*
 *  Copyright 2018 Expedia Group.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 *     you may not use this file except in compliance with the License.
 *     You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *     Unless required by applicable law or agreed to in writing, software
 *     distributed under the License is distributed on an "AS IS" BASIS,
 *     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *     See the License for the specific language governing permissions and
 *     limitations under the License.
 *
 */

package haystack

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

/*RemoteClient remote client*/
type RemoteClient interface {
	Send(span *Span)
	Close() error
	SetLogger(logger Logger)
}

/*GrpcClient grpc client*/
type GrpcClient struct {
	conn    *grpc.ClientConn
	client  SpanAgentClient
	timeout time.Duration
	logger  Logger
}

/*NewGrpcClient returns a new grpc client*/
func NewGrpcClient(host string, port int, timeout time.Duration) *GrpcClient {
	targetHost := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(targetHost, grpc.WithInsecure())

	if err != nil {
		panic(fmt.Sprintf("fail to connect to agent with error: %v", err))
	}

	return &GrpcClient{
		conn:    conn,
		client:  NewSpanAgentClient(conn),
		timeout: timeout,
	}
}

/*Send a proto span to grpc server*/
func (c *GrpcClient) Send(span *Span) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	result, err := c.client.Dispatch(ctx, span)

	if err != nil {
		c.logger.Error("Fail to dispatch to haystack-agent with error %v", err)
	} else if result.GetCode() != DispatchResult_SUCCESS {
		c.logger.Error(fmt.Sprintf("Fail to dispatch to haystack-agent with error code: %d, message :%s", result.GetCode(), result.GetErrorMessage()))
	} else {
		c.logger.Debug(fmt.Sprintf("span [%v] has been successfully dispatched to haystack", span))
	}
}

/*Close the grpc client*/
func (c *GrpcClient) Close() error {
	return c.conn.Close()
}

/*SetLogger sets the logger*/
func (c *GrpcClient) SetLogger(logger Logger) {
	c.logger = logger
}

/*HTTPClient a http client*/
type HTTPClient struct {
	url     string
	headers map[string]string
	client  *http.Client
	logger  Logger
}

/*NewHTTPClient returns a new http client*/
func NewHTTPClient(url string, headers map[string]string, timeout time.Duration) *HTTPClient {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &HTTPClient{
		url:     url,
		headers: headers,
		client:  httpClient,
	}
}

/*Send a proto span to http server*/
func (c *HTTPClient) Send(span *Span) {
	serializedBytes, marshalErr := proto.Marshal(span)

	if marshalErr != nil {
		c.logger.Error("Fail to serialize the span to proto bytes, error=%v", marshalErr)
		return
	}

	postRequest, requestErr := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(serializedBytes))
	if requestErr != nil {
		c.logger.Error("Fail to create request for posting span to haystack server, error=%v", requestErr)
		return
	}

	if c.headers != nil {
		for k, v := range c.headers {
			postRequest.Header.Add(k, v)
		}
	}

	resp, err := c.client.Do(postRequest)

	if err != nil {
		c.logger.Error("Fail to dispatch to haystack http server, error=%v", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			/* do nothing */
		}
	}()

	respBytes, respErr := ioutil.ReadAll(resp.Body)
	if respErr != nil {
		c.logger.Error("Fail to read the http response from haystack server, error=%v", respErr)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Error("Fail to dispatch the span to haystack http server with statusCode=%d , payload=%s", resp.StatusCode, string(respBytes))
	} else {
		c.logger.Debug(fmt.Sprintf("span [%v] has been successfully dispatched to haystack, response=%s", span, string(respBytes)))
	}
}

/*Close the http client*/
func (c *HTTPClient) Close() error {
	return nil
}

/*SetLogger sets the logger*/
func (c *HTTPClient) SetLogger(logger Logger) {
	c.logger = logger
}
