// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gensupport

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestContentSniffing(t *testing.T) {
	type testCase struct {
		data     []byte // the data to read from the Reader
		finalErr error  // error to return after data has been read

		wantContentType       string
		wantContentTypeResult bool
	}

	for _, tc := range []testCase{
		{
			data:                  []byte{0, 0, 0, 0},
			finalErr:              nil,
			wantContentType:       "application/octet-stream",
			wantContentTypeResult: true,
		},
		{
			data:                  []byte(""),
			finalErr:              nil,
			wantContentType:       "text/plain; charset=utf-8",
			wantContentTypeResult: true,
		},
		{
			data:                  []byte(""),
			finalErr:              io.ErrUnexpectedEOF,
			wantContentType:       "text/plain; charset=utf-8",
			wantContentTypeResult: false,
		},
		{
			data:                  []byte("abc"),
			finalErr:              nil,
			wantContentType:       "text/plain; charset=utf-8",
			wantContentTypeResult: true,
		},
		{
			data:                  []byte("abc"),
			finalErr:              io.ErrUnexpectedEOF,
			wantContentType:       "text/plain; charset=utf-8",
			wantContentTypeResult: false,
		},
		// The following examples contain more bytes than are buffered for sniffing.
		{
			data:                  bytes.Repeat([]byte("a"), 513),
			finalErr:              nil,
			wantContentType:       "text/plain; charset=utf-8",
			wantContentTypeResult: true,
		},
		{
			data:                  bytes.Repeat([]byte("a"), 513),
			finalErr:              io.ErrUnexpectedEOF,
			wantContentType:       "text/plain; charset=utf-8",
			wantContentTypeResult: true, // true because error is after first 512 bytes.
		},
	} {
		er := &errReader{buf: tc.data, err: tc.finalErr}

		sct := newContentSniffer(er)

		// Even if was an error during the first 512 bytes, we should still be able to read those bytes.
		buf, err := ioutil.ReadAll(sct)

		if !reflect.DeepEqual(buf, tc.data) {
			t.Fatalf("Failed reading buffer: got: %q; want:%q", buf, tc.data)
		}

		if err != tc.finalErr {
			t.Fatalf("Reading buffer error: got: %v; want: %v", err, tc.finalErr)
		}

		ct, ok := sct.ContentType()
		if ok != tc.wantContentTypeResult {
			t.Fatalf("Content type result got: %v; want: %v", ok, tc.wantContentTypeResult)
		}
		if ok && ct != tc.wantContentType {
			t.Fatalf("Content type got: %q; want: %q", ct, tc.wantContentType)
		}
	}
}

type staticContentTyper struct {
	io.Reader
}

func (sct staticContentTyper) ContentType() string {
	return "static content type"
}

func TestDetermineContentType(t *testing.T) {
	data := []byte("abc")
	rdr := func() io.Reader {
		return bytes.NewBuffer(data)
	}

	type testCase struct {
		r                  io.Reader
		explicitConentType string
		wantContentType    string
	}

	for _, tc := range []testCase{
		{
			r:               rdr(),
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			r:               staticContentTyper{rdr()},
			wantContentType: "static content type",
		},
		{
			r:                  staticContentTyper{rdr()},
			explicitConentType: "explicit",
			wantContentType:    "explicit",
		},
	} {
		r, ctype := DetermineContentType(tc.r, tc.explicitConentType)
		got, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatalf("Failed reading buffer: %v", err)
		}
		if !reflect.DeepEqual(got, data) {
			t.Fatalf("Failed reading buffer: got: %q; want:%q", got, data)
		}

		if ctype != tc.wantContentType {
			t.Fatalf("Content type got: %q; want: %q", ctype, tc.wantContentType)
		}
	}
}
