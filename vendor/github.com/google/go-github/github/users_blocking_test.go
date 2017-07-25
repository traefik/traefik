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

func TestUsersService_ListBlockedUsers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/blocks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeBlockUsersPreview)
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{
			"login": "octocat"
		}]`)
	})

	opt := &ListOptions{Page: 2}
	blockedUsers, _, err := client.Users.ListBlockedUsers(context.Background(), opt)
	if err != nil {
		t.Errorf("Users.ListBlockedUsers returned error: %v", err)
	}

	want := []*User{{Login: String("octocat")}}
	if !reflect.DeepEqual(blockedUsers, want) {
		t.Errorf("Users.ListBlockedUsers returned %+v, want %+v", blockedUsers, want)
	}
}

func TestUsersService_IsBlocked(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/blocks/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeBlockUsersPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	isBlocked, _, err := client.Users.IsBlocked(context.Background(), "u")
	if err != nil {
		t.Errorf("Users.IsBlocked returned error: %v", err)
	}
	if want := true; isBlocked != want {
		t.Errorf("Users.IsBlocked returned %+v, want %+v", isBlocked, want)
	}
}

func TestUsersService_BlockUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/blocks/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		testHeader(t, r, "Accept", mediaTypeBlockUsersPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Users.BlockUser(context.Background(), "u")
	if err != nil {
		t.Errorf("Users.BlockUser returned error: %v", err)
	}
}

func TestUsersService_UnblockUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/blocks/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeBlockUsersPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Users.UnblockUser(context.Background(), "u")
	if err != nil {
		t.Errorf("Users.UnblockUser returned error: %v", err)
	}
}
