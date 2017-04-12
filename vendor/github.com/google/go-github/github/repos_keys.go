// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// The Key type is defined in users_keys.go

// ListKeys lists the deploy keys for a repository.
//
// GitHub API docs: https://developer.github.com/v3/repos/keys/#list
func (s *RepositoriesService) ListKeys(ctx context.Context, owner string, repo string, opt *ListOptions) ([]*Key, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/keys", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var keys []*Key
	resp, err := s.client.Do(ctx, req, &keys)
	if err != nil {
		return nil, resp, err
	}

	return keys, resp, nil
}

// GetKey fetches a single deploy key.
//
// GitHub API docs: https://developer.github.com/v3/repos/keys/#get
func (s *RepositoriesService) GetKey(ctx context.Context, owner string, repo string, id int) (*Key, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/keys/%v", owner, repo, id)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	key := new(Key)
	resp, err := s.client.Do(ctx, req, key)
	if err != nil {
		return nil, resp, err
	}

	return key, resp, nil
}

// CreateKey adds a deploy key for a repository.
//
// GitHub API docs: https://developer.github.com/v3/repos/keys/#create
func (s *RepositoriesService) CreateKey(ctx context.Context, owner string, repo string, key *Key) (*Key, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/keys", owner, repo)

	req, err := s.client.NewRequest("POST", u, key)
	if err != nil {
		return nil, nil, err
	}

	k := new(Key)
	resp, err := s.client.Do(ctx, req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// EditKey edits a deploy key.
//
// GitHub API docs: https://developer.github.com/v3/repos/keys/#edit
func (s *RepositoriesService) EditKey(ctx context.Context, owner string, repo string, id int, key *Key) (*Key, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/keys/%v", owner, repo, id)

	req, err := s.client.NewRequest("PATCH", u, key)
	if err != nil {
		return nil, nil, err
	}

	k := new(Key)
	resp, err := s.client.Do(ctx, req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// DeleteKey deletes a deploy key.
//
// GitHub API docs: https://developer.github.com/v3/repos/keys/#delete
func (s *RepositoriesService) DeleteKey(ctx context.Context, owner string, repo string, id int) (*Response, error) {
	u := fmt.Sprintf("repos/%v/%v/keys/%v", owner, repo, id)

	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}
