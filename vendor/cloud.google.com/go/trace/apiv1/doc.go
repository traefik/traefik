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

// Package trace is an experimental, auto-generated package for the
// trace API.
//
// Send and retrieve trace data from Stackdriver Trace. Data is generated and
// available by default for all App Engine applications. Data from other
// applications can be written to Stackdriver Trace for display, reporting,
// and analysis.
//
// Use the client at cloud.google.com/go/trace in preference to this.
package trace // import "cloud.google.com/go/trace/apiv1"

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func insertXGoog(ctx context.Context, val string) context.Context {
	md, _ := metadata.FromContext(ctx)
	md = md.Copy()
	md["x-goog-api-client"] = []string{val}
	return metadata.NewContext(ctx, md)
}
