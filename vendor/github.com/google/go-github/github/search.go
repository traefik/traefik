// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"

	qs "github.com/google/go-querystring/query"
)

// SearchService provides access to the search related functions
// in the GitHub API.
//
// GitHub API docs: https://developer.github.com/v3/search/
type SearchService service

// SearchOptions specifies optional parameters to the SearchService methods.
type SearchOptions struct {
	// How to sort the search results. Possible values are:
	//   - for repositories: stars, fork, updated
	//   - for commits: author-date, committer-date
	//   - for code: indexed
	//   - for issues: comments, created, updated
	//   - for users: followers, repositories, joined
	//
	// Default is to sort by best match.
	Sort string `url:"sort,omitempty"`

	// Sort order if sort parameter is provided. Possible values are: asc,
	// desc. Default is desc.
	Order string `url:"order,omitempty"`

	// Whether to retrieve text match metadata with a query
	TextMatch bool `url:"-"`

	ListOptions
}

// RepositoriesSearchResult represents the result of a repositories search.
type RepositoriesSearchResult struct {
	Total             *int         `json:"total_count,omitempty"`
	IncompleteResults *bool        `json:"incomplete_results,omitempty"`
	Repositories      []Repository `json:"items,omitempty"`
}

// Repositories searches repositories via various criteria.
//
// GitHub API docs: https://developer.github.com/v3/search/#search-repositories
func (s *SearchService) Repositories(ctx context.Context, query string, opt *SearchOptions) (*RepositoriesSearchResult, *Response, error) {
	result := new(RepositoriesSearchResult)
	resp, err := s.search(ctx, "repositories", query, opt, result)
	return result, resp, err
}

// CommitsSearchResult represents the result of a commits search.
type CommitsSearchResult struct {
	Total             *int            `json:"total_count,omitempty"`
	IncompleteResults *bool           `json:"incomplete_results,omitempty"`
	Commits           []*CommitResult `json:"items,omitempty"`
}

// CommitResult represents a commit object as returned in commit search endpoint response.
type CommitResult struct {
	SHA         *string   `json:"sha,omitempty"`
	Commit      *Commit   `json:"commit,omitempty"`
	Author      *User     `json:"author,omitempty"`
	Committer   *User     `json:"committer,omitempty"`
	Parents     []*Commit `json:"parents,omitempty"`
	HTMLURL     *string   `json:"html_url,omitempty"`
	URL         *string   `json:"url,omitempty"`
	CommentsURL *string   `json:"comments_url,omitempty"`

	Repository *Repository `json:"repository,omitempty"`
	Score      *float64    `json:"score,omitempty"`
}

// Commits searches commits via various criteria.
//
// GitHub API docs: https://developer.github.com/v3/search/#search-commits
func (s *SearchService) Commits(ctx context.Context, query string, opt *SearchOptions) (*CommitsSearchResult, *Response, error) {
	result := new(CommitsSearchResult)
	resp, err := s.search(ctx, "commits", query, opt, result)
	return result, resp, err
}

// IssuesSearchResult represents the result of an issues search.
type IssuesSearchResult struct {
	Total             *int    `json:"total_count,omitempty"`
	IncompleteResults *bool   `json:"incomplete_results,omitempty"`
	Issues            []Issue `json:"items,omitempty"`
}

// Issues searches issues via various criteria.
//
// GitHub API docs: https://developer.github.com/v3/search/#search-issues
func (s *SearchService) Issues(ctx context.Context, query string, opt *SearchOptions) (*IssuesSearchResult, *Response, error) {
	result := new(IssuesSearchResult)
	resp, err := s.search(ctx, "issues", query, opt, result)
	return result, resp, err
}

// UsersSearchResult represents the result of a users search.
type UsersSearchResult struct {
	Total             *int   `json:"total_count,omitempty"`
	IncompleteResults *bool  `json:"incomplete_results,omitempty"`
	Users             []User `json:"items,omitempty"`
}

// Users searches users via various criteria.
//
// GitHub API docs: https://developer.github.com/v3/search/#search-users
func (s *SearchService) Users(ctx context.Context, query string, opt *SearchOptions) (*UsersSearchResult, *Response, error) {
	result := new(UsersSearchResult)
	resp, err := s.search(ctx, "users", query, opt, result)
	return result, resp, err
}

// Match represents a single text match.
type Match struct {
	Text    *string `json:"text,omitempty"`
	Indices []int   `json:"indices,omitempty"`
}

// TextMatch represents a text match for a SearchResult
type TextMatch struct {
	ObjectURL  *string `json:"object_url,omitempty"`
	ObjectType *string `json:"object_type,omitempty"`
	Property   *string `json:"property,omitempty"`
	Fragment   *string `json:"fragment,omitempty"`
	Matches    []Match `json:"matches,omitempty"`
}

func (tm TextMatch) String() string {
	return Stringify(tm)
}

// CodeSearchResult represents the result of a code search.
type CodeSearchResult struct {
	Total             *int         `json:"total_count,omitempty"`
	IncompleteResults *bool        `json:"incomplete_results,omitempty"`
	CodeResults       []CodeResult `json:"items,omitempty"`
}

// CodeResult represents a single search result.
type CodeResult struct {
	Name        *string     `json:"name,omitempty"`
	Path        *string     `json:"path,omitempty"`
	SHA         *string     `json:"sha,omitempty"`
	HTMLURL     *string     `json:"html_url,omitempty"`
	Repository  *Repository `json:"repository,omitempty"`
	TextMatches []TextMatch `json:"text_matches,omitempty"`
}

func (c CodeResult) String() string {
	return Stringify(c)
}

// Code searches code via various criteria.
//
// GitHub API docs: https://developer.github.com/v3/search/#search-code
func (s *SearchService) Code(ctx context.Context, query string, opt *SearchOptions) (*CodeSearchResult, *Response, error) {
	result := new(CodeSearchResult)
	resp, err := s.search(ctx, "code", query, opt, result)
	return result, resp, err
}

// Helper function that executes search queries against different
// GitHub search types (repositories, commits, code, issues, users)
func (s *SearchService) search(ctx context.Context, searchType string, query string, opt *SearchOptions, result interface{}) (*Response, error) {
	params, err := qs.Values(opt)
	if err != nil {
		return nil, err
	}
	params.Set("q", query)
	u := fmt.Sprintf("search/%s?%s", searchType, params.Encode())

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	switch {
	case searchType == "commits":
		// Accept header for search commits preview endpoint
		// TODO: remove custom Accept header when this API fully launches.
		req.Header.Set("Accept", mediaTypeCommitSearchPreview)
	case opt != nil && opt.TextMatch:
		// Accept header defaults to "application/vnd.github.v3+json"
		// We change it here to fetch back text-match metadata
		req.Header.Set("Accept", "application/vnd.github.v3.text-match+json")
	}

	return s.client.Do(ctx, req, result)
}
