// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import "fmt"

// ListProjects lists the projects for a repo.
//
// GitHub API docs: https://developer.github.com/v3/projects/#list-repository-projects
func (s *RepositoriesService) ListProjects(owner, repo string, opt *ListOptions) ([]*Project, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/projects", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeProjectsPreview)

	projects := []*Project{}
	resp, err := s.client.Do(req, &projects)
	if err != nil {
		return nil, resp, err
	}

	return projects, resp, err
}

// CreateProject creates a GitHub Project for the specified repository.
//
// GitHub API docs: https://developer.github.com/v3/projects/#create-a-repository-project
func (s *RepositoriesService) CreateProject(owner, repo string, opt *ProjectOptions) (*Project, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/projects", owner, repo)
	req, err := s.client.NewRequest("POST", u, opt)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeProjectsPreview)

	project := &Project{}
	resp, err := s.client.Do(req, project)
	if err != nil {
		return nil, resp, err
	}

	return project, resp, err
}
