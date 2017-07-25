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

package monitoring_test

import (
	"cloud.google.com/go/monitoring/apiv3"
	"golang.org/x/net/context"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func ExampleNewGroupClient() {
	ctx := context.Background()
	c, err := monitoring.NewGroupClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use client.
	_ = c
}

func ExampleGroupClient_ListGroups() {
	ctx := context.Background()
	c, err := monitoring.NewGroupClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &monitoringpb.ListGroupsRequest{
	// TODO: Fill request struct fields.
	}
	it := c.ListGroups(ctx, req)
	for {
		resp, err := it.Next()
		if err != nil {
			// TODO: Handle error.
			break
		}
		// TODO: Use resp.
		_ = resp
	}
}

func ExampleGroupClient_GetGroup() {
	ctx := context.Background()
	c, err := monitoring.NewGroupClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &monitoringpb.GetGroupRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.GetGroup(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleGroupClient_CreateGroup() {
	ctx := context.Background()
	c, err := monitoring.NewGroupClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &monitoringpb.CreateGroupRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.CreateGroup(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleGroupClient_UpdateGroup() {
	ctx := context.Background()
	c, err := monitoring.NewGroupClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &monitoringpb.UpdateGroupRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.UpdateGroup(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleGroupClient_DeleteGroup() {
	ctx := context.Background()
	c, err := monitoring.NewGroupClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &monitoringpb.DeleteGroupRequest{
	// TODO: Fill request struct fields.
	}
	err = c.DeleteGroup(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleGroupClient_ListGroupMembers() {
	ctx := context.Background()
	c, err := monitoring.NewGroupClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &monitoringpb.ListGroupMembersRequest{
	// TODO: Fill request struct fields.
	}
	it := c.ListGroupMembers(ctx, req)
	for {
		resp, err := it.Next()
		if err != nil {
			// TODO: Handle error.
			break
		}
		// TODO: Use resp.
		_ = resp
	}
}
