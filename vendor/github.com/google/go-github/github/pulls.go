// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

// PullRequestsService handles communication with the pull request related
// methods of the GitHub API.
//
// GitHub API docs: https://developer.github.com/v3/pulls/
type PullRequestsService service

// PullRequest represents a GitHub pull request on a repository.
type PullRequest struct {
	ID                  *int       `json:"id,omitempty"`
	Number              *int       `json:"number,omitempty"`
	State               *string    `json:"state,omitempty"`
	Title               *string    `json:"title,omitempty"`
	Body                *string    `json:"body,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
	ClosedAt            *time.Time `json:"closed_at,omitempty"`
	MergedAt            *time.Time `json:"merged_at,omitempty"`
	User                *User      `json:"user,omitempty"`
	Merged              *bool      `json:"merged,omitempty"`
	Mergeable           *bool      `json:"mergeable,omitempty"`
	MergedBy            *User      `json:"merged_by,omitempty"`
	MergeCommitSHA      *string    `json:"merge_commit_sha,omitempty"`
	Comments            *int       `json:"comments,omitempty"`
	Commits             *int       `json:"commits,omitempty"`
	Additions           *int       `json:"additions,omitempty"`
	Deletions           *int       `json:"deletions,omitempty"`
	ChangedFiles        *int       `json:"changed_files,omitempty"`
	URL                 *string    `json:"url,omitempty"`
	HTMLURL             *string    `json:"html_url,omitempty"`
	IssueURL            *string    `json:"issue_url,omitempty"`
	StatusesURL         *string    `json:"statuses_url,omitempty"`
	DiffURL             *string    `json:"diff_url,omitempty"`
	PatchURL            *string    `json:"patch_url,omitempty"`
	ReviewCommentsURL   *string    `json:"review_comments_url,omitempty"`
	ReviewCommentURL    *string    `json:"review_comment_url,omitempty"`
	Assignee            *User      `json:"assignee,omitempty"`
	Assignees           []*User    `json:"assignees,omitempty"`
	Milestone           *Milestone `json:"milestone,omitempty"`
	MaintainerCanModify *bool      `json:"maintainer_can_modify,omitempty"`

	Head *PullRequestBranch `json:"head,omitempty"`
	Base *PullRequestBranch `json:"base,omitempty"`
}

func (p PullRequest) String() string {
	return Stringify(p)
}

// PullRequestBranch represents a base or head branch in a GitHub pull request.
type PullRequestBranch struct {
	Label *string     `json:"label,omitempty"`
	Ref   *string     `json:"ref,omitempty"`
	SHA   *string     `json:"sha,omitempty"`
	Repo  *Repository `json:"repo,omitempty"`
	User  *User       `json:"user,omitempty"`
}

// PullRequestListOptions specifies the optional parameters to the
// PullRequestsService.List method.
type PullRequestListOptions struct {
	// State filters pull requests based on their state. Possible values are:
	// open, closed. Default is "open".
	State string `url:"state,omitempty"`

	// Head filters pull requests by head user and branch name in the format of:
	// "user:ref-name".
	Head string `url:"head,omitempty"`

	// Base filters pull requests by base branch name.
	Base string `url:"base,omitempty"`

	// Sort specifies how to sort pull requests. Possible values are: created,
	// updated, popularity, long-running. Default is "created".
	Sort string `url:"sort,omitempty"`

	// Direction in which to sort pull requests. Possible values are: asc, desc.
	// If Sort is "created" or not specified, Default is "desc", otherwise Default
	// is "asc"
	Direction string `url:"direction,omitempty"`

	ListOptions
}

// List the pull requests for the specified repository.
//
// GitHub API docs: https://developer.github.com/v3/pulls/#list-pull-requests
func (s *PullRequestsService) List(ctx context.Context, owner string, repo string, opt *PullRequestListOptions) ([]*PullRequest, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var pulls []*PullRequest
	resp, err := s.client.Do(ctx, req, &pulls)
	if err != nil {
		return nil, resp, err
	}

	return pulls, resp, nil
}

// Get a single pull request.
//
// GitHub API docs: https://developer.github.com/v3/pulls/#get-a-single-pull-request
func (s *PullRequestsService) Get(ctx context.Context, owner string, repo string, number int) (*PullRequest, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d", owner, repo, number)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	pull := new(PullRequest)
	resp, err := s.client.Do(ctx, req, pull)
	if err != nil {
		return nil, resp, err
	}

	return pull, resp, nil
}

// GetRaw gets raw (diff or patch) format of a pull request.
func (s *PullRequestsService) GetRaw(ctx context.Context, owner string, repo string, number int, opt RawOptions) (string, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d", owner, repo, number)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return "", nil, err
	}

	switch opt.Type {
	case Diff:
		req.Header.Set("Accept", mediaTypeV3Diff)
	case Patch:
		req.Header.Set("Accept", mediaTypeV3Patch)
	default:
		return "", nil, fmt.Errorf("unsupported raw type %d", opt.Type)
	}

	ret := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, ret)
	if err != nil {
		return "", resp, err
	}

	return ret.String(), resp, nil
}

// NewPullRequest represents a new pull request to be created.
type NewPullRequest struct {
	Title               *string `json:"title,omitempty"`
	Head                *string `json:"head,omitempty"`
	Base                *string `json:"base,omitempty"`
	Body                *string `json:"body,omitempty"`
	Issue               *int    `json:"issue,omitempty"`
	MaintainerCanModify *bool   `json:"maintainer_can_modify,omitempty"`
}

// Create a new pull request on the specified repository.
//
// GitHub API docs: https://developer.github.com/v3/pulls/#create-a-pull-request
func (s *PullRequestsService) Create(ctx context.Context, owner string, repo string, pull *NewPullRequest) (*PullRequest, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls", owner, repo)
	req, err := s.client.NewRequest("POST", u, pull)
	if err != nil {
		return nil, nil, err
	}

	p := new(PullRequest)
	resp, err := s.client.Do(ctx, req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

type pullRequestUpdate struct {
	Title               *string `json:"title,omitempty"`
	Body                *string `json:"body,omitempty"`
	State               *string `json:"state,omitempty"`
	Base                *string `json:"base,omitempty"`
	MaintainerCanModify *bool   `json:"maintainer_can_modify,omitempty"`
}

// Edit a pull request.
// pull must not be nil.
//
// The following fields are editable: Title, Body, State, Base.Ref and MaintainerCanModify.
// Base.Ref updates the base branch of the pull request.
//
// GitHub API docs: https://developer.github.com/v3/pulls/#update-a-pull-request
func (s *PullRequestsService) Edit(ctx context.Context, owner string, repo string, number int, pull *PullRequest) (*PullRequest, *Response, error) {
	if pull == nil {
		return nil, nil, fmt.Errorf("pull must be provided")
	}

	u := fmt.Sprintf("repos/%v/%v/pulls/%d", owner, repo, number)

	update := &pullRequestUpdate{
		Title:               pull.Title,
		Body:                pull.Body,
		State:               pull.State,
		MaintainerCanModify: pull.MaintainerCanModify,
	}
	if pull.Base != nil {
		update.Base = pull.Base.Ref
	}

	req, err := s.client.NewRequest("PATCH", u, update)
	if err != nil {
		return nil, nil, err
	}

	p := new(PullRequest)
	resp, err := s.client.Do(ctx, req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ListCommits lists the commits in a pull request.
//
// GitHub API docs: https://developer.github.com/v3/pulls/#list-commits-on-a-pull-request
func (s *PullRequestsService) ListCommits(ctx context.Context, owner string, repo string, number int, opt *ListOptions) ([]*RepositoryCommit, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/commits", owner, repo, number)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var commits []*RepositoryCommit
	resp, err := s.client.Do(ctx, req, &commits)
	if err != nil {
		return nil, resp, err
	}

	return commits, resp, nil
}

// ListFiles lists the files in a pull request.
//
// GitHub API docs: https://developer.github.com/v3/pulls/#list-pull-requests-files
func (s *PullRequestsService) ListFiles(ctx context.Context, owner string, repo string, number int, opt *ListOptions) ([]*CommitFile, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/files", owner, repo, number)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var commitFiles []*CommitFile
	resp, err := s.client.Do(ctx, req, &commitFiles)
	if err != nil {
		return nil, resp, err
	}

	return commitFiles, resp, nil
}

// IsMerged checks if a pull request has been merged.
//
// GitHub API docs: https://developer.github.com/v3/pulls/#get-if-a-pull-request-has-been-merged
func (s *PullRequestsService) IsMerged(ctx context.Context, owner string, repo string, number int) (bool, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/merge", owner, repo, number)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return false, nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	merged, err := parseBoolResponse(err)
	return merged, resp, err
}

// PullRequestMergeResult represents the result of merging a pull request.
type PullRequestMergeResult struct {
	SHA     *string `json:"sha,omitempty"`
	Merged  *bool   `json:"merged,omitempty"`
	Message *string `json:"message,omitempty"`
}

// PullRequestOptions lets you define how a pull request will be merged.
type PullRequestOptions struct {
	CommitTitle string // Extra detail to append to automatic commit message. (Optional.)
	SHA         string // SHA that pull request head must match to allow merge. (Optional.)

	// The merge method to use. Possible values include: "merge", "squash", and "rebase" with the default being merge. (Optional.)
	MergeMethod string
}

type pullRequestMergeRequest struct {
	CommitMessage string `json:"commit_message"`
	CommitTitle   string `json:"commit_title,omitempty"`
	MergeMethod   string `json:"merge_method,omitempty"`
	SHA           string `json:"sha,omitempty"`
}

// Merge a pull request (Merge Button™).
// commitMessage is the title for the automatic commit message.
//
// GitHub API docs: https://developer.github.com/v3/pulls/#merge-a-pull-request-merge-buttontrade
func (s *PullRequestsService) Merge(ctx context.Context, owner string, repo string, number int, commitMessage string, options *PullRequestOptions) (*PullRequestMergeResult, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/merge", owner, repo, number)

	pullRequestBody := &pullRequestMergeRequest{CommitMessage: commitMessage}
	if options != nil {
		pullRequestBody.CommitTitle = options.CommitTitle
		pullRequestBody.MergeMethod = options.MergeMethod
		pullRequestBody.SHA = options.SHA
	}
	req, err := s.client.NewRequest("PUT", u, pullRequestBody)
	if err != nil {
		return nil, nil, err
	}

	// TODO: This header will be unnecessary when the API is no longer in preview.
	req.Header.Set("Accept", mediaTypeSquashPreview)

	mergeResult := new(PullRequestMergeResult)
	resp, err := s.client.Do(ctx, req, mergeResult)
	if err != nil {
		return nil, resp, err
	}

	return mergeResult, resp, nil
}
