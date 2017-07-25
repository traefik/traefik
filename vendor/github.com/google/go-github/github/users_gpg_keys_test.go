// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestUsersService_ListGPGKeys_authenticatedUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/gpg_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeGitSigningPreview)
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"id":1,"primary_key_id":2}]`)
	})

	opt := &ListOptions{Page: 2}
	keys, _, err := client.Users.ListGPGKeys(context.Background(), "", opt)
	if err != nil {
		t.Errorf("Users.ListGPGKeys returned error: %v", err)
	}

	want := []*GPGKey{{ID: Int(1), PrimaryKeyID: Int(2)}}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("Users.ListGPGKeys = %+v, want %+v", keys, want)
	}
}

func TestUsersService_ListGPGKeys_specifiedUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users/u/gpg_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeGitSigningPreview)
		fmt.Fprint(w, `[{"id":1,"primary_key_id":2}]`)
	})

	keys, _, err := client.Users.ListGPGKeys(context.Background(), "u", nil)
	if err != nil {
		t.Errorf("Users.ListGPGKeys returned error: %v", err)
	}

	want := []*GPGKey{{ID: Int(1), PrimaryKeyID: Int(2)}}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("Users.ListGPGKeys = %+v, want %+v", keys, want)
	}
}

func TestUsersService_ListGPGKeys_invalidUser(t *testing.T) {
	_, _, err := client.Users.ListGPGKeys(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestUsersService_GetGPGKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/gpg_keys/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeGitSigningPreview)
		fmt.Fprint(w, `{"id":1}`)
	})

	key, _, err := client.Users.GetGPGKey(context.Background(), 1)
	if err != nil {
		t.Errorf("Users.GetGPGKey returned error: %v", err)
	}

	want := &GPGKey{ID: Int(1)}
	if !reflect.DeepEqual(key, want) {
		t.Errorf("Users.GetGPGKey = %+v, want %+v", key, want)
	}
}

func TestUsersService_CreateGPGKey(t *testing.T) {
	setup()
	defer teardown()

	input := `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Comment: GPGTools - https://gpgtools.org

mQINBFcEd9kBEACo54TDbGhKlXKWMvJgecEUKPPcv7XdnpKdGb3LRw5MvFwT0V0f
...
=tqfb
-----END PGP PUBLIC KEY BLOCK-----`

	mux.HandleFunc("/user/gpg_keys", func(w http.ResponseWriter, r *http.Request) {
		var gpgKey struct {
			ArmoredPublicKey *string `json:"armored_public_key,omitempty"`
		}
		json.NewDecoder(r.Body).Decode(&gpgKey)

		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeGitSigningPreview)
		if gpgKey.ArmoredPublicKey == nil || *gpgKey.ArmoredPublicKey != input {
			t.Errorf("gpgKey = %+v, want %q", gpgKey, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	gpgKey, _, err := client.Users.CreateGPGKey(context.Background(), input)
	if err != nil {
		t.Errorf("Users.GetGPGKey returned error: %v", err)
	}

	want := &GPGKey{ID: Int(1)}
	if !reflect.DeepEqual(gpgKey, want) {
		t.Errorf("Users.GetGPGKey = %+v, want %+v", gpgKey, want)
	}
}

func TestUsersService_DeleteGPGKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/gpg_keys/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeGitSigningPreview)
	})

	_, err := client.Users.DeleteGPGKey(context.Background(), 1)
	if err != nil {
		t.Errorf("Users.DeleteGPGKey returned error: %v", err)
	}
}
