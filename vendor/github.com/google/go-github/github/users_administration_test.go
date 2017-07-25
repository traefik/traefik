// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"net/http"
	"testing"
)

func TestUsersService_PromoteSiteAdmin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users/u/site_admin", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Users.PromoteSiteAdmin(context.Background(), "u")
	if err != nil {
		t.Errorf("Users.PromoteSiteAdmin returned error: %v", err)
	}
}

func TestUsersService_DemoteSiteAdmin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users/u/site_admin", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Users.DemoteSiteAdmin(context.Background(), "u")
	if err != nil {
		t.Errorf("Users.DemoteSiteAdmin returned error: %v", err)
	}
}

func TestUsersService_Suspend(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users/u/suspended", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Users.Suspend(context.Background(), "u")
	if err != nil {
		t.Errorf("Users.Suspend returned error: %v", err)
	}
}

func TestUsersService_Unsuspend(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users/u/suspended", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Users.Unsuspend(context.Background(), "u")
	if err != nil {
		t.Errorf("Users.Unsuspend returned error: %v", err)
	}
}
