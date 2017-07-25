// Copyright 2016 The go-github AUTHORS. All rights reserved.
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

func TestRepositoriesService_ListInvitations(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/invitations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeRepositoryInvitationsPreview)
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprintf(w, `[{"id":1}, {"id":2}]`)
	})

	opt := &ListOptions{Page: 2}
	got, _, err := client.Repositories.ListInvitations(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Repositories.ListInvitations returned error: %v", err)
	}

	want := []*RepositoryInvitation{{ID: Int(1)}, {ID: Int(2)}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Repositories.ListInvitations = %+v, want %+v", got, want)
	}
}

func TestRepositoriesService_DeleteInvitation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/invitations/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeRepositoryInvitationsPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Repositories.DeleteInvitation(context.Background(), "o", "r", 2)
	if err != nil {
		t.Errorf("Repositories.DeleteInvitation returned error: %v", err)
	}
}

func TestRepositoriesService_UpdateInvitation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/invitations/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		testHeader(t, r, "Accept", mediaTypeRepositoryInvitationsPreview)
		fmt.Fprintf(w, `{"id":1}`)
	})

	got, _, err := client.Repositories.UpdateInvitation(context.Background(), "o", "r", 2, "write")
	if err != nil {
		t.Errorf("Repositories.UpdateInvitation returned error: %v", err)
	}

	want := &RepositoryInvitation{ID: Int(1)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Repositories.UpdateInvitation = %+v, want %+v", got, want)
	}
}
