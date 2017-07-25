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

package admin_test

import (
	"cloud.google.com/go/iam/admin/apiv1"
	"golang.org/x/net/context"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

func ExampleNewIamClient() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use client.
	_ = c
}

func ExampleIamClient_ListServiceAccounts() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.ListServiceAccountsRequest{
	// TODO: Fill request struct fields.
	}
	it := c.ListServiceAccounts(ctx, req)
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

func ExampleIamClient_GetServiceAccount() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.GetServiceAccountRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.GetServiceAccount(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleIamClient_CreateServiceAccount() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.CreateServiceAccountRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.CreateServiceAccount(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleIamClient_UpdateServiceAccount() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.ServiceAccount{
	// TODO: Fill request struct fields.
	}
	resp, err := c.UpdateServiceAccount(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleIamClient_DeleteServiceAccount() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.DeleteServiceAccountRequest{
	// TODO: Fill request struct fields.
	}
	err = c.DeleteServiceAccount(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleIamClient_ListServiceAccountKeys() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.ListServiceAccountKeysRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.ListServiceAccountKeys(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleIamClient_GetServiceAccountKey() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.GetServiceAccountKeyRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.GetServiceAccountKey(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleIamClient_CreateServiceAccountKey() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.CreateServiceAccountKeyRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.CreateServiceAccountKey(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleIamClient_DeleteServiceAccountKey() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.DeleteServiceAccountKeyRequest{
	// TODO: Fill request struct fields.
	}
	err = c.DeleteServiceAccountKey(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleIamClient_SignBlob() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.SignBlobRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.SignBlob(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}

func ExampleIamClient_TestIamPermissions() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
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

func ExampleIamClient_QueryGrantableRoles() {
	ctx := context.Background()
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &adminpb.QueryGrantableRolesRequest{
	// TODO: Fill request struct fields.
	}
	resp, err := c.QueryGrantableRoles(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use resp.
	_ = resp
}
