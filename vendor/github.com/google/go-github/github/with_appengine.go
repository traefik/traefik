// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build appengine

// This file provides glue for making github work on App Engine.
// In order to get the entire github package to compile with
// Go 1.6, you will need to rewrite all the import "context" lines.
// Fortunately, this is easy with "gofmt":
//
//     gofmt -w -r '"context" -> "golang.org/x/net/context"' *.go

package github

import (
	"context"
	"net/http"

	"google.golang.org/appengine"
)

func withContext(ctx context.Context, req *http.Request) (context.Context, *http.Request) {
	return appengine.WithContext(ctx, req), req
}
