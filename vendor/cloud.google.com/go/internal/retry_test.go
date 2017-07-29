// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/net/context"

	gax "github.com/googleapis/gax-go"
)

func TestRetry(t *testing.T) {
	ctx := context.Background()
	// Without a context deadline, retry will run until the function
	// says not to retry any more.
	n := 0
	endRetry := errors.New("end retry")
	err := retry(ctx, gax.Backoff{},
		func() (bool, error) {
			n++
			if n < 10 {
				return false, nil
			}
			return true, endRetry
		},
		func(context.Context, time.Duration) error { return nil })
	if got, want := err, endRetry; got != want {
		t.Errorf("got %v, want %v", err, endRetry)
	}
	if n != 10 {
		t.Errorf("n: got %d, want %d", n, 10)
	}

	// If the context has a deadline, sleep will return an error
	// and end the function.
	n = 0
	err = retry(ctx, gax.Backoff{},
		func() (bool, error) { return false, nil },
		func(context.Context, time.Duration) error {
			n++
			if n < 10 {
				return nil
			}
			return context.DeadlineExceeded
		})
	if err == nil {
		t.Error("got nil, want error")
	}
}
