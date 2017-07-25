// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestOrganizationsService_ListBlockedUsers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/blocks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeBlockUsersPreview)
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{
			"login": "octocat"
		}]`)
	})

	opt := &ListOptions{Page: 2}
	blockedUsers, _, err := client.Organizations.ListBlockedUsers(context.Background(), "o", opt)
	if err != nil {
		t.Errorf("Organizations.ListBlockedUsers returned error: %v", err)
	}

	want := []*User{{Login: String("octocat")}}
	if !reflect.DeepEqual(blockedUsers, want) {
		t.Errorf("Organizations.ListBlockedUsers returned %+v, want %+v", blockedUsers, want)
	}
}

func TestOrganizationsService_IsBlocked(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/blocks/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeBlockUsersPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	isBlocked, _, err := client.Organizations.IsBlocked(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.IsBlocked returned error: %v", err)
	}
	if want := true; isBlocked != want {
		t.Errorf("Organizations.IsBlocked returned %+v, want %+v", isBlocked, want)
	}
}

func TestOrganizationsService_BlockUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/blocks/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeBlockUsersPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Organizations.BlockUser(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.BlockUser returned error: %v", err)
	}
}

func TestOrganizationsService_UnblockUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/blocks/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeBlockUsersPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Organizations.UnblockUser(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.UnblockUser returned error: %v", err)
	}
}
