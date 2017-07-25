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

// Package errorreporting is an experimental, auto-generated package for the
// errorreporting API.
//
// Stackdriver Error Reporting groups and counts similar errors from cloud
// services. The Stackdriver Error Reporting API provides a way to report new
// errors and read access to error groups and their associated errors.
package errorreporting // import "cloud.google.com/go/errorreporting/apiv1beta1"

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
