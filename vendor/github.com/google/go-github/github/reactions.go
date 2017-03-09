// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import "fmt"

// ReactionsService provides access to the reactions-related functions in the
// GitHub API.
//
// GitHub API docs: https://developer.github.com/v3/reactions/
type ReactionsService service

// Reaction represents a GitHub reaction.
type Reaction struct {
	// ID is the Reaction ID.
	ID   *int  `json:"id,omitempty"`
	User *User `json:"user,omitempty"`
	// Content is the type of reaction.
	// Possible values are:
	//     "+1", "-1", "laugh", "confused", "heart", "hooray".
	Content *string `json:"content,omitempty"`
}

// Reactions represents a summary of GitHub reactions.
type Reactions struct {
	TotalCount *int    `json:"total_count,omitempty"`
	PlusOne    *int    `json:"+1,omitempty"`
	MinusOne   *int    `json:"-1,omitempty"`
	Laugh      *int    `json:"laugh,omitempty"`
	Confused   *int    `json:"confused,omitempty"`
	Heart      *int    `json:"heart,omitempty"`
	Hooray     *int    `json:"hooray,omitempty"`
	URL        *string `json:"url,omitempty"`
}

func (r Reaction) String() string {
	return Stringify(r)
}

// ListCommentReactions lists the reactions for a commit comment.
//
// GitHub API docs: https://developer.github.com/v3/reactions/#list-reactions-for-a-commit-comment
func (s *ReactionsService) ListCommentReactions(owner, repo string, id int, opt *ListOptions) ([]*Reaction, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/comments/%v/reactions", owner, repo, id)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	var m []*Reaction
	resp, err := s.client.Do(req, &m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// CreateCommentReaction creates a reaction for a commit comment.
// Note that if you have already created a reaction of type content, the
// previously created reaction will be returned with Status: 200 OK.
//
// GitHub API docs: https://developer.github.com/v3/reactions/#create-reaction-for-a-commit-comment
func (s ReactionsService) CreateCommentReaction(owner, repo string, id int, content string) (*Reaction, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/comments/%v/reactions", owner, repo, id)

	body := &Reaction{Content: String(content)}
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	m := &Reaction{}
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// ListIssueReactions lists the reactions for an issue.
//
// GitHub API docs: https://developer.github.com/v3/reactions/#list-reactions-for-an-issue
func (s *ReactionsService) ListIssueReactions(owner, repo string, number int, opt *ListOptions) ([]*Reaction, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/%v/reactions", owner, repo, number)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	var m []*Reaction
	resp, err := s.client.Do(req, &m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// CreateIssueReaction creates a reaction for an issue.
// Note that if you have already created a reaction of type content, the
// previously created reaction will be returned with Status: 200 OK.
//
// GitHub API docs: https://developer.github.com/v3/reactions/#create-reaction-for-an-issue
func (s ReactionsService) CreateIssueReaction(owner, repo string, number int, content string) (*Reaction, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/%v/reactions", owner, repo, number)

	body := &Reaction{Content: String(content)}
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	m := &Reaction{}
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// ListIssueCommentReactions lists the reactions for an issue comment.
//
// GitHub API docs: https://developer.github.com/v3/reactions/#list-reactions-for-an-issue-comment
func (s *ReactionsService) ListIssueCommentReactions(owner, repo string, id int, opt *ListOptions) ([]*Reaction, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/comments/%v/reactions", owner, repo, id)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	var m []*Reaction
	resp, err := s.client.Do(req, &m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// CreateIssueCommentReaction creates a reaction for an issue comment.
// Note that if you have already created a reaction of type content, the
// previously created reaction will be returned with Status: 200 OK.
//
// GitHub API docs: https://developer.github.com/v3/reactions/#create-reaction-for-an-issue-comment
func (s ReactionsService) CreateIssueCommentReaction(owner, repo string, id int, content string) (*Reaction, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/comments/%v/reactions", owner, repo, id)

	body := &Reaction{Content: String(content)}
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	m := &Reaction{}
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// ListPullRequestCommentReactions lists the reactions for a pull request review comment.
//
// GitHub API docs: https://developer.github.com/v3/reactions/#list-reactions-for-an-issue-comment
func (s *ReactionsService) ListPullRequestCommentReactions(owner, repo string, id int, opt *ListOptions) ([]*Reaction, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/comments/%v/reactions", owner, repo, id)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	var m []*Reaction
	resp, err := s.client.Do(req, &m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// CreatePullRequestCommentReaction creates a reaction for a pull request review comment.
// Note that if you have already created a reaction of type content, the
// previously created reaction will be returned with Status: 200 OK.
//
// GitHub API docs: https://developer.github.com/v3/reactions/#create-reaction-for-an-issue-comment
func (s ReactionsService) CreatePullRequestCommentReaction(owner, repo string, id int, content string) (*Reaction, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/comments/%v/reactions", owner, repo, id)

	body := &Reaction{Content: String(content)}
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	m := &Reaction{}
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// DeleteReaction deletes a reaction.
//
// GitHub API docs: https://developer.github.com/v3/reaction/reactions/#delete-a-reaction-archive
func (s *ReactionsService) DeleteReaction(id int) (*Response, error) {
	u := fmt.Sprintf("reactions/%v", id)

	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeReactionsPreview)

	return s.client.Do(req, nil)
}
