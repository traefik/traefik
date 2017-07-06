// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import "context"

// Installation represents a GitHub Apps installation.
type Installation struct {
	ID              *int    `json:"id,omitempty"`
	Account         *User   `json:"account,omitempty"`
	AccessTokensURL *string `json:"access_tokens_url,omitempty"`
	RepositoriesURL *string `json:"repositories_url,omitempty"`
	HTMLURL         *string `json:"html_url,omitempty"`
}

func (i Installation) String() string {
	return Stringify(i)
}

// ListRepos lists the repositories that are accessible to the authenticated installation.
//
// GitHub API docs: https://developer.github.com/v3/apps/installations/#list-repositories
func (s *AppsService) ListRepos(ctx context.Context, opt *ListOptions) ([]*Repository, *Response, error) {
	u, err := addOptions("installation/repositories", opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeIntegrationPreview)

	var r struct {
		Repositories []*Repository `json:"repositories"`
	}
	resp, err := s.client.Do(ctx, req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r.Repositories, resp, nil
}
