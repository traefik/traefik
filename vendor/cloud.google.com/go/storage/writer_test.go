// Copyright 2014 Google Inc. All Rights Reserved.
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

package storage

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/api/option"
)

type fakeTransport struct {
	gotReq *http.Request
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.gotReq = req
	return nil, fmt.Errorf("error handling request")
}

func TestErrorOnObjectsInsertCall(t *testing.T) {
	ctx := context.Background()
	hc := &http.Client{Transport: &fakeTransport{}}
	client, err := NewClient(ctx, option.WithHTTPClient(hc))
	if err != nil {
		t.Fatalf("error when creating client: %v", err)
	}
	wc := client.Bucket("bucketname").Object("filename1").NewWriter(ctx)
	wc.ContentType = "text/plain"

	// We can't check that the Write fails, since it depends on the write to the
	// underling fakeTransport failing which is racy.
	wc.Write([]byte("hello world"))

	// Close must always return an error though since it waits for the transport to
	// have closed.
	if err := wc.Close(); err == nil {
		t.Errorf("expected error on close, got nil")
	}
}

func TestEncryption(t *testing.T) {
	ctx := context.Background()
	ft := &fakeTransport{}
	hc := &http.Client{Transport: ft}
	client, err := NewClient(ctx, option.WithHTTPClient(hc))
	if err != nil {
		t.Fatalf("error when creating client: %v", err)
	}
	obj := client.Bucket("bucketname").Object("filename1")
	key := []byte("secret-key-that-is-32-bytes-long")
	wc := obj.Key(key).NewWriter(ctx)
	// TODO(jba): use something other than fakeTransport, which always returns error.
	wc.Write([]byte("hello world"))
	wc.Close()
	if got, want := ft.gotReq.Header.Get("x-goog-encryption-algorithm"), "AES256"; got != want {
		t.Errorf("algorithm: got %q, want %q", got, want)
	}
	gotKey, err := base64.StdEncoding.DecodeString(ft.gotReq.Header.Get("x-goog-encryption-key"))
	if err != nil {
		t.Fatalf("decoding key: %v", err)
	}
	if !reflect.DeepEqual(gotKey, key) {
		t.Errorf("key: got %v, want %v", gotKey, key)
	}
	wantHash := sha256.Sum256(key)
	gotHash, err := base64.StdEncoding.DecodeString(ft.gotReq.Header.Get("x-goog-encryption-key-sha256"))
	if err != nil {
		t.Fatalf("decoding hash: %v", err)
	}
	if !reflect.DeepEqual(gotHash, wantHash[:]) { // wantHash is an array
		t.Errorf("hash: got\n%v, want\n%v", gotHash, wantHash)
	}
}
