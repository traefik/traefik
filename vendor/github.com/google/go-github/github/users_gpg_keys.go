// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"fmt"
	"time"
)

// GPGKey represents a GitHub user's public GPG key used to verify GPG signed commits and tags.
//
// https://developer.github.com/changes/2016-04-04-git-signing-api-preview/
type GPGKey struct {
	ID                *int       `json:"id,omitempty"`
	PrimaryKeyID      *int       `json:"primary_key_id,omitempty"`
	KeyID             *string    `json:"key_id,omitempty"`
	PublicKey         *string    `json:"public_key,omitempty"`
	Emails            []GPGEmail `json:"emails,omitempty"`
	Subkeys           []GPGKey   `json:"subkeys,omitempty"`
	CanSign           *bool      `json:"can_sign,omitempty"`
	CanEncryptComms   *bool      `json:"can_encrypt_comms,omitempty"`
	CanEncryptStorage *bool      `json:"can_encrypt_storage,omitempty"`
	CanCertify        *bool      `json:"can_certify,omitempty"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// String stringifies a GPGKey.
func (k GPGKey) String() string {
	return Stringify(k)
}

// GPGEmail represents an email address associated to a GPG key.
type GPGEmail struct {
	Email    *string `json:"email,omitempty"`
	Verified *bool   `json:"verified,omitempty"`
}

// ListGPGKeys lists the current user's GPG keys. It requires authentication
// via Basic Auth or via OAuth with at least read:gpg_key scope.
//
// GitHub API docs: https://developer.github.com/v3/users/gpg_keys/#list-your-gpg-keys
func (s *UsersService) ListGPGKeys() ([]*GPGKey, *Response, error) {
	req, err := s.client.NewRequest("GET", "user/gpg_keys", nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeGitSigningPreview)

	var keys []*GPGKey
	resp, err := s.client.Do(req, &keys)
	if err != nil {
		return nil, resp, err
	}

	return keys, resp, err
}

// GetGPGKey gets extended details for a single GPG key. It requires authentication
// via Basic Auth or via OAuth with at least read:gpg_key scope.
//
// GitHub API docs: https://developer.github.com/v3/users/gpg_keys/#get-a-single-gpg-key
func (s *UsersService) GetGPGKey(id int) (*GPGKey, *Response, error) {
	u := fmt.Sprintf("user/gpg_keys/%v", id)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeGitSigningPreview)

	key := &GPGKey{}
	resp, err := s.client.Do(req, key)
	if err != nil {
		return nil, resp, err
	}

	return key, resp, err
}

// CreateGPGKey creates a GPG key. It requires authenticatation via Basic Auth
// or OAuth with at least write:gpg_key scope.
//
// GitHub API docs: https://developer.github.com/v3/users/gpg_keys/#create-a-gpg-key
func (s *UsersService) CreateGPGKey(armoredPublicKey string) (*GPGKey, *Response, error) {
	gpgKey := &struct {
		ArmoredPublicKey string `json:"armored_public_key"`
	}{ArmoredPublicKey: armoredPublicKey}
	req, err := s.client.NewRequest("POST", "user/gpg_keys", gpgKey)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeGitSigningPreview)

	key := &GPGKey{}
	resp, err := s.client.Do(req, key)
	if err != nil {
		return nil, resp, err
	}

	return key, resp, err
}

// DeleteGPGKey deletes a GPG key. It requires authentication via Basic Auth or
// via OAuth with at least admin:gpg_key scope.
//
// GitHub API docs: https://developer.github.com/v3/users/gpg_keys/#delete-a-gpg-key
func (s *UsersService) DeleteGPGKey(id int) (*Response, error) {
	u := fmt.Sprintf("user/gpg_keys/%v", id)
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeGitSigningPreview)

	return s.client.Do(req, nil)
}
