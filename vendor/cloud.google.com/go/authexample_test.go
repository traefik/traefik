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

package cloud_test

import (
	"cloud.google.com/go/datastore"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

func Example_applicationDefaultCredentials() {
	// Google Application Default Credentials is the recommended way to authorize
	// and authenticate clients.
	//
	// See the following link on how to create and obtain Application Default Credentials:
	// https://developers.google.com/identity/protocols/application-default-credentials.
	client, err := datastore.NewClient(context.Background(), "project-id")
	if err != nil {
		// TODO: handle error.
	}
	_ = client // Use the client.
}

func Example_serviceAccountFile() {
	// Use a JSON key file associated with a Google service account to
	// authenticate and authorize. Service Account keys can be created and
	// downloaded from https://console.developers.google.com/permissions/serviceaccounts.
	//
	// Note: This example uses the datastore client, but the same steps apply to
	// the other client libraries underneath this package.
	client, err := datastore.NewClient(context.Background(),
		"project-id", option.WithServiceAccountFile("/path/to/service-account-key.json"))
	if err != nil {
		// TODO: handle error.
	}
	_ = client // Use the client.
}
