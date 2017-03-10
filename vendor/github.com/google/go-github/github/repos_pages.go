// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import "fmt"

// Pages represents a GitHub Pages site configuration.
type Pages struct {
	URL       *string `json:"url,omitempty"`
	Status    *string `json:"status,omitempty"`
	CNAME     *string `json:"cname,omitempty"`
	Custom404 *bool   `json:"custom_404,omitempty"`
	HTMLURL   *string `json:"html_url,omitempty"`
}

// PagesError represents a build error for a GitHub Pages site.
type PagesError struct {
	Message *string `json:"message,omitempty"`
}

// PagesBuild represents the build information for a GitHub Pages site.
type PagesBuild struct {
	URL       *string     `json:"url,omitempty"`
	Status    *string     `json:"status,omitempty"`
	Error     *PagesError `json:"error,omitempty"`
	Pusher    *User       `json:"pusher,omitempty"`
	Commit    *string     `json:"commit,omitempty"`
	Duration  *int        `json:"duration,omitempty"`
	CreatedAt *Timestamp  `json:"created_at,omitempty"`
	UpdatedAt *Timestamp  `json:"updated_at,omitempty"`
}

// GetPagesInfo fetches information about a GitHub Pages site.
//
// GitHub API docs: https://developer.github.com/v3/repos/pages/#get-information-about-a-pages-site
func (s *RepositoriesService) GetPagesInfo(owner, repo string) (*Pages, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pages", owner, repo)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypePagesPreview)

	site := new(Pages)
	resp, err := s.client.Do(req, site)
	if err != nil {
		return nil, resp, err
	}

	return site, resp, err
}

// ListPagesBuilds lists the builds for a GitHub Pages site.
//
// GitHub API docs: https://developer.github.com/v3/repos/pages/#list-pages-builds
func (s *RepositoriesService) ListPagesBuilds(owner, repo string) ([]*PagesBuild, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pages/builds", owner, repo)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var pages []*PagesBuild
	resp, err := s.client.Do(req, &pages)
	if err != nil {
		return nil, resp, err
	}

	return pages, resp, err
}

// GetLatestPagesBuild fetches the latest build information for a GitHub pages site.
//
// GitHub API docs: https://developer.github.com/v3/repos/pages/#list-latest-pages-build
func (s *RepositoriesService) GetLatestPagesBuild(owner, repo string) (*PagesBuild, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pages/builds/latest", owner, repo)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	build := new(PagesBuild)
	resp, err := s.client.Do(req, build)
	if err != nil {
		return nil, resp, err
	}

	return build, resp, err
}

// GetPageBuild fetches the specific build information for a GitHub pages site.
//
// GitHub API docs: https://developer.github.com/v3/repos/pages/#list-a-specific-pages-build
func (s *RepositoriesService) GetPageBuild(owner, repo string, id int) (*PagesBuild, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pages/builds/%v", owner, repo, id)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	build := new(PagesBuild)
	resp, err := s.client.Do(req, build)
	if err != nil {
		return nil, resp, err
	}

	return build, resp, err
}

// RequestPageBuild requests a build of a GitHub Pages site without needing to push new commit.
//
// GitHub API docs: https://developer.github.com/v3/repos/pages/#request-a-page-build
func (s *RepositoriesService) RequestPageBuild(owner, repo string) (*PagesBuild, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pages/builds", owner, repo)
	req, err := s.client.NewRequest("POST", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypePagesPreview)

	build := new(PagesBuild)
	resp, err := s.client.Do(req, build)
	if err != nil {
		return nil, resp, err
	}

	return build, resp, err
}
