// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

// Installation represents a GitHub integration installation.
type Installation struct {
	ID              *int    `json:"id,omitempty"`
	Account         *User   `json:"account,omitempty"`
	AccessTokensURL *string `json:"access_tokens_url,omitempty"`
	RepositoriesURL *string `json:"repositories_url,omitempty"`
}

func (i Installation) String() string {
	return Stringify(i)
}

// ListRepos lists the repositories that the current installation has access to.
//
// GitHub API docs: https://developer.github.com/v3/integrations/installations/#list-repositories
func (s *IntegrationsService) ListRepos(opt *ListOptions) ([]*Repository, *Response, error) {
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
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r.Repositories, resp, err
}
