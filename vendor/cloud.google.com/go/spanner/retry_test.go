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
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	edpb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// Test if runRetryable loop deals with various errors correctly.
func TestRetry(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	responses := []error{
		grpc.Errorf(codes.Internal, "transport is closing"),
		grpc.Errorf(codes.Unknown, "unexpected EOF"),
		grpc.Errorf(codes.Internal, "stream terminated by RST_STREAM with error code: 2"),
		grpc.Errorf(codes.Unavailable, "service is currently unavailable"),
		errRetry(fmt.Errorf("just retry it")),
	}
	err := runRetryable(context.Background(), func(ct context.Context) error {
		var r error
		if len(responses) > 0 {
			r = responses[0]
			responses = responses[1:]
		}
		return r
	})
	if err != nil {
		t.Errorf("runRetryable should be able to survive all retryable errors, but it returns %v", err)
	}
	// Unretryable errors
	injErr := errors.New("this is unretryable")
	err = runRetryable(context.Background(), func(ct context.Context) error {
		return injErr
	})
	if wantErr := toSpannerError(injErr); !reflect.DeepEqual(err, wantErr) {
		t.Errorf("runRetryable returns error %v, want %v", err, wantErr)
	}
	// Timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	retryErr := errRetry(fmt.Errorf("still retrying"))
	err = runRetryable(ctx, func(ct context.Context) error {
		// Expect to trigger timeout in retryable runner after 10 executions.
		<-time.After(100 * time.Millisecond)
		// Let retryable runner to retry so that timeout will eventually happen.
		return retryErr
	})
	if wantErr := errContextCanceled(retryErr); !reflect.DeepEqual(err, wantErr) {
		t.Errorf("runRetryable returns error: %v, want error: %v", err, wantErr)
	}
	// Cancellation
	ctx, cancel = context.WithCancel(context.Background())
	retries := 3
	retryErr = errRetry(fmt.Errorf("retry before cancel"))
	err = runRetryable(ctx, func(ct context.Context) error {
		retries--
		if retries == 0 {
			cancel()
		}
		return retryErr
	})
	if wantErr := errContextCanceled(retryErr); !reflect.DeepEqual(err, wantErr) || retries != 0 {
		t.Errorf("<err, retries>=<%v, %v>, want <%v, %v>", err, retries, wantErr, 0)
	}
}

func TestRetryInfo(t *testing.T) {
	b, _ := proto.Marshal(&edpb.RetryInfo{
		RetryDelay: ptypes.DurationProto(time.Second),
	})
	trailers := map[string]string{
		retryInfoKey: string(b),
	}
	gotDelay, ok := extractRetryDelay(errRetry(toSpannerErrorWithMetadata(grpc.Errorf(codes.Aborted, ""), metadata.New(trailers))))
	if !ok || !reflect.DeepEqual(time.Second, gotDelay) {
		t.Errorf("<ok, retryDelay> = <%t, %v>, want <true, %v>", ok, gotDelay, time.Second)
	}
}
