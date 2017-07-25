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

package database_test

import (
	"cloud.google.com/go/spanner/admin/database/apiv1"
	"golang.org/x/net/context"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
)

func ExampleNewDatabaseAdminClient() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use client.
	_ = c
}

func ExampleDatabaseAdminClient_ListDatabases() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &databasepb.ListDatabasesRequest{
	// TODO: Fill request struct fields.
	}
	it := c.ListDatabases(ctx, req)
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

func ExampleDatabaseAdminClient_CreateDatabase() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &databasepb.CreateDatabaseRequest{
	// TODO: Fill request struct fields.
	}
	op, err := c.CreateDatabase(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleDatabaseAdminClient_GetDatabase() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &databasepb.GetDatabaseRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.GetDatabase(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleDatabaseAdminClient_UpdateDatabaseDdl() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &databasepb.UpdateDatabaseDdlRequest{
	// TODO: Fill request struct fields.
	}
	op, err := c.UpdateDatabaseDdl(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}

	err = op.Wait(ctx)
	// TODO: Handle error.
}

func ExampleDatabaseAdminClient_DropDatabase() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &databasepb.DropDatabaseRequest{
	// TODO: Fill request struct fields.
	}
	err = c.DropDatabase(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleDatabaseAdminClient_GetDatabaseDdl() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &databasepb.GetDatabaseDdlRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.GetDatabaseDdl(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleDatabaseAdminClient_SetIamPolicy() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &iampb.SetIamPolicyRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.SetIamPolicy(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleDatabaseAdminClient_GetIamPolicy() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &iampb.GetIamPolicyRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.GetIamPolicy(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleDatabaseAdminClient_TestIamPermissions() {
	ctx := context.Background()
	c, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &iampb.TestIamPermissionsRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.TestIamPermissions(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}
