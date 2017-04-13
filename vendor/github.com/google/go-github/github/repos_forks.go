// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// RepositoryListForksOptions specifies the optional parameters to the
// RepositoriesService.ListForks method.
type RepositoryListForksOptions struct {
	// How to sort the forks list. Possible values are: newest, oldest,
	// watchers. Default is "newest".
	Sort string `url:"sort,omitempty"`

	ListOptions
}

// ListForks lists the forks of the specified repository.
//
// GitHub API docs: https://developer.github.com/v3/repos/forks/#list-forks
func (s *RepositoriesService) ListForks(ctx context.Context, owner, repo string, opt *RepositoryListForksOptions) ([]*Repository, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/forks", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var repos []*Repository
	resp, err := s.client.Do(ctx, req, &repos)
	if err != nil {
		return nil, resp, err
	}

	return repos, resp, nil
}

// RepositoryCreateForkOptions specifies the optional parameters to the
// RepositoriesService.CreateFork method.
type RepositoryCreateForkOptions struct {
	// The organization to fork the repository into.
	Organization string `url:"organization,omitempty"`
}

// CreateFork creates a fork of the specified repository.
//
// This method might return an *AcceptedError and a status code of
// 202. This is because this is the status that GitHub returns to signify that
// it is now computing creating the fork in a background task.
// A follow up request, after a delay of a second or so, should result
// in a successful request.
//
// GitHub API docs: https://developer.github.com/v3/repos/forks/#create-a-fork
func (s *RepositoriesService) CreateFork(ctx context.Context, owner, repo string, opt *RepositoryCreateForkOptions) (*Repository, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/forks", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("POST", u, nil)
	if err != nil {
		return nil, nil, err
	}

	fork := new(Repository)
	resp, err := s.client.Do(ctx, req, fork)
	if err != nil {
		return nil, resp, err
	}

	return fork, resp, nil
}
