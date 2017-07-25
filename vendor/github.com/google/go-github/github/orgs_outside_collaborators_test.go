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

func TestOrganizationsService_ListOutsideCollaborators(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/outside_collaborators", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"filter": "2fa_disabled",
			"page":   "2",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOutsideCollaboratorsOptions{
		Filter:      "2fa_disabled",
		ListOptions: ListOptions{Page: 2},
	}
	members, _, err := client.Organizations.ListOutsideCollaborators(context.Background(), "o", opt)
	if err != nil {
		t.Errorf("Organizations.ListOutsideCollaborators returned error: %v", err)
	}

	want := []*User{{ID: Int(1)}}
	if !reflect.DeepEqual(members, want) {
		t.Errorf("Organizations.ListOutsideCollaborators returned %+v, want %+v", members, want)
	}
}

func TestOrganizationsService_ListOutsideCollaborators_invalidOrg(t *testing.T) {
	_, _, err := client.Organizations.ListOutsideCollaborators(context.Background(), "%", nil)
	testURLParseError(t, err)
}
