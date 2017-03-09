// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// These event types are shared between the Events API and used as Webhook payloads.

package github

// CommitCommentEvent is triggered when a commit comment is created.
// The Webhook event name is "commit_comment".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#commitcommentevent
type CommitCommentEvent struct {
	Comment *RepositoryComment `json:"comment,omitempty"`

	// The following fields are only populated by Webhook events.
	Action *string     `json:"action,omitempty"`
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}

// CreateEvent represents a created repository, branch, or tag.
// The Webhook event name is "create".
//
// Note: webhooks will not receive this event for created repositories.
// Additionally, webhooks will not receive this event for tags if more
// than three tags are pushed at once.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#createevent
type CreateEvent struct {
	Ref *string `json:"ref,omitempty"`
	// RefType is the object that was created. Possible values are: "repository", "branch", "tag".
	RefType      *string `json:"ref_type,omitempty"`
	MasterBranch *string `json:"master_branch,omitempty"`
	Description  *string `json:"description,omitempty"`

	// The following fields are only populated by Webhook events.
	PusherType *string     `json:"pusher_type,omitempty"`
	Repo       *Repository `json:"repository,omitempty"`
	Sender     *User       `json:"sender,omitempty"`
}

// DeleteEvent represents a deleted branch or tag.
// The Webhook event name is "delete".
//
// Note: webhooks will not receive this event for tags if more than three tags
// are deleted at once.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#deleteevent
type DeleteEvent struct {
	Ref *string `json:"ref,omitempty"`
	// RefType is the object that was deleted. Possible values are: "branch", "tag".
	RefType *string `json:"ref_type,omitempty"`

	// The following fields are only populated by Webhook events.
	PusherType *string     `json:"pusher_type,omitempty"`
	Repo       *Repository `json:"repository,omitempty"`
	Sender     *User       `json:"sender,omitempty"`
}

// DeploymentEvent represents a deployment.
// The Webhook event name is "deployment".
//
// Events of this type are not visible in timelines, they are only used to trigger hooks.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#deploymentevent
type DeploymentEvent struct {
	Deployment *Deployment `json:"deployment,omitempty"`
	Repo       *Repository `json:"repository,omitempty"`

	// The following fields are only populated by Webhook events.
	Sender *User `json:"sender,omitempty"`
}

// DeploymentStatusEvent represents a deployment status.
// The Webhook event name is "deployment_status".
//
// Events of this type are not visible in timelines, they are only used to trigger hooks.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#deploymentstatusevent
type DeploymentStatusEvent struct {
	Deployment       *Deployment       `json:"deployment,omitempty"`
	DeploymentStatus *DeploymentStatus `json:"deployment_status,omitempty"`
	Repo             *Repository       `json:"repository,omitempty"`

	// The following fields are only populated by Webhook events.
	Sender *User `json:"sender,omitempty"`
}

// ForkEvent is triggered when a user forks a repository.
// The Webhook event name is "fork".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#forkevent
type ForkEvent struct {
	// Forkee is the created repository.
	Forkee *Repository `json:"forkee,omitempty"`

	// The following fields are only populated by Webhook events.
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}

// Page represents a single Wiki page.
type Page struct {
	PageName *string `json:"page_name,omitempty"`
	Title    *string `json:"title,omitempty"`
	Summary  *string `json:"summary,omitempty"`
	Action   *string `json:"action,omitempty"`
	SHA      *string `json:"sha,omitempty"`
	HTMLURL  *string `json:"html_url,omitempty"`
}

// GollumEvent is triggered when a Wiki page is created or updated.
// The Webhook event name is "gollum".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#gollumevent
type GollumEvent struct {
	Pages []*Page `json:"pages,omitempty"`

	// The following fields are only populated by Webhook events.
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}

// IssueActivityEvent represents the payload delivered by Issue webhook.
//
// Deprecated: Use IssuesEvent instead.
type IssueActivityEvent struct {
	Action *string `json:"action,omitempty"`
	Issue  *Issue  `json:"issue,omitempty"`

	// The following fields are only populated by Webhook events.
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}

// EditChange represents the changes when an issue, pull request, or comment has
// been edited.
type EditChange struct {
	Title *struct {
		From *string `json:"from,omitempty"`
	} `json:"title,omitempty"`
	Body *struct {
		From *string `json:"from,omitempty"`
	} `json:"body,omitempty"`
}

// IntegrationInstallationEvent is triggered when an integration is created or deleted.
// The Webhook event name is "integration_installation".
//
// GitHub docs: https://developer.github.com/early-access/integrations/webhooks/#integrationinstallationevent
type IntegrationInstallationEvent struct {
	// The action that was performed. Possible values for an "integration_installation"
	// event are: "created", "deleted".
	Action       *string       `json:"action,omitempty"`
	Installation *Installation `json:"installation,omitempty"`
	Sender       *User         `json:"sender,omitempty"`
}

// IntegrationInstallationRepositoriesEvent is triggered when an integration repository
// is added or removed. The Webhook event name is "integration_installation_repositories".
//
// GitHub docs: https://developer.github.com/early-access/integrations/webhooks/#integrationinstallationrepositoriesevent
type IntegrationInstallationRepositoriesEvent struct {
	// The action that was performed. Possible values for an "integration_installation_repositories"
	// event are: "added", "removed".
	Action              *string       `json:"action,omitempty"`
	Installation        *Installation `json:"installation,omitempty"`
	RepositoriesAdded   []*Repository `json:"repositories_added,omitempty"`
	RepositoriesRemoved []*Repository `json:"repositories_removed,omitempty"`
	Sender              *User         `json:"sender,omitempty"`
}

// IssueCommentEvent is triggered when an issue comment is created on an issue
// or pull request.
// The Webhook event name is "issue_comment".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#issuecommentevent
type IssueCommentEvent struct {
	// Action is the action that was performed on the comment.
	// Possible values are: "created", "edited", "deleted".
	Action  *string       `json:"action,omitempty"`
	Issue   *Issue        `json:"issue,omitempty"`
	Comment *IssueComment `json:"comment,omitempty"`

	// The following fields are only populated by Webhook events.
	Changes *EditChange `json:"changes,omitempty"`
	Repo    *Repository `json:"repository,omitempty"`
	Sender  *User       `json:"sender,omitempty"`
}

// IssuesEvent is triggered when an issue is assigned, unassigned, labeled,
// unlabeled, opened, closed, or reopened.
// The Webhook event name is "issues".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#issuesevent
type IssuesEvent struct {
	// Action is the action that was performed. Possible values are: "assigned",
	// "unassigned", "labeled", "unlabeled", "opened", "closed", "reopened", "edited".
	Action   *string `json:"action,omitempty"`
	Issue    *Issue  `json:"issue,omitempty"`
	Assignee *User   `json:"assignee,omitempty"`
	Label    *Label  `json:"label,omitempty"`

	// The following fields are only populated by Webhook events.
	Changes *EditChange `json:"changes,omitempty"`
	Repo    *Repository `json:"repository,omitempty"`
	Sender  *User       `json:"sender,omitempty"`
}

// LabelEvent is triggered when a repository's label is created, edited, or deleted.
// The Webhook event name is "label"
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#labelevent
type LabelEvent struct {
	// Action is the action that was performed. Possible values are:
	// "created", "edited", "deleted"
	Action *string `json:"action,omitempty"`
	Label  *Label  `json:"label,omitempty"`

	// The following fields are only populated by Webhook events.
	Changes *EditChange   `json:"changes,omitempty"`
	Repo    *Repository   `json:"repository,omitempty"`
	Org     *Organization `json:"organization,omitempty"`
}

// MemberEvent is triggered when a user is added as a collaborator to a repository.
// The Webhook event name is "member".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#memberevent
type MemberEvent struct {
	// Action is the action that was performed. Possible value is: "added".
	Action *string `json:"action,omitempty"`
	Member *User   `json:"member,omitempty"`

	// The following fields are only populated by Webhook events.
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}

// MembershipEvent is triggered when a user is added or removed from a team.
// The Webhook event name is "membership".
//
// Events of this type are not visible in timelines, they are only used to
// trigger organization webhooks.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#membershipevent
type MembershipEvent struct {
	// Action is the action that was performed. Possible values are: "added", "removed".
	Action *string `json:"action,omitempty"`
	// Scope is the scope of the membership. Possible value is: "team".
	Scope  *string `json:"scope,omitempty"`
	Member *User   `json:"member,omitempty"`
	Team   *Team   `json:"team,omitempty"`

	// The following fields are only populated by Webhook events.
	Org    *Organization `json:"organization,omitempty"`
	Sender *User         `json:"sender,omitempty"`
}

// MilestoneEvent is triggered when a milestone is created, closed, opened, edited, or deleted.
// The Webhook event name is "milestone".
//
// Github docs: https://developer.github.com/v3/activity/events/types/#milestoneevent
type MilestoneEvent struct {
	// Action is the action that was performed. Possible values are:
	// "created", "closed", "opened", "edited", "deleted"
	Action    *string    `json:"action,omitempty"`
	Milestone *Milestone `json:"milestone,omitempty"`

	// The following fields are only populated by Webhook events.
	Changes *EditChange   `json:"changes,omitempty"`
	Repo    *Repository   `json:"repository,omitempty"`
	Sender  *User         `json:"sender,omitempty"`
	Org     *Organization `json:"organization,omitempty"`
}

// PageBuildEvent represents an attempted build of a GitHub Pages site, whether
// successful or not.
// The Webhook event name is "page_build".
//
// This event is triggered on push to a GitHub Pages enabled branch (gh-pages
// for project pages, master for user and organization pages).
//
// Events of this type are not visible in timelines, they are only used to trigger hooks.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#pagebuildevent
type PageBuildEvent struct {
	Build *PagesBuild `json:"build,omitempty"`

	// The following fields are only populated by Webhook events.
	ID     *int        `json:"id,omitempty"`
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}

// PingEvent is triggered when a Webhook is added to GitHub.
//
// GitHub docs: https://developer.github.com/webhooks/#ping-event
type PingEvent struct {
	// Random string of GitHub zen.
	Zen *string `json:"zen,omitempty"`
	// The ID of the webhook that triggered the ping.
	HookID *int `json:"hook_id,omitempty"`
	// The webhook configuration.
	Hook *Hook `json:"hook,omitempty"`
}

// PublicEvent is triggered when a private repository is open sourced.
// According to GitHub: "Without a doubt: the best GitHub event."
// The Webhook event name is "public".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#publicevent
type PublicEvent struct {
	// The following fields are only populated by Webhook events.
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}

// PullRequestEvent is triggered when a pull request is assigned, unassigned,
// labeled, unlabeled, opened, closed, reopened, or synchronized.
// The Webhook event name is "pull_request".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#pullrequestevent
type PullRequestEvent struct {
	// Action is the action that was performed. Possible values are: "assigned",
	// "unassigned", "labeled", "unlabeled", "opened", "closed", or "reopened",
	// "synchronize", "edited". If the action is "closed" and the merged key is false,
	// the pull request was closed with unmerged commits. If the action is "closed"
	// and the merged key is true, the pull request was merged.
	Action      *string      `json:"action,omitempty"`
	Number      *int         `json:"number,omitempty"`
	PullRequest *PullRequest `json:"pull_request,omitempty"`

	// The following fields are only populated by Webhook events.
	Changes *EditChange `json:"changes,omitempty"`
	Repo    *Repository `json:"repository,omitempty"`
	Sender  *User       `json:"sender,omitempty"`
}

// PullRequestReviewEvent is triggered when a review is submitted on a pull
// request.
// The Webhook event name is "pull_request_review".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#pullrequestreviewevent
type PullRequestReviewEvent struct {
	// Action is always "submitted".
	Action      *string            `json:"action,omitempty"`
	Review      *PullRequestReview `json:"review,omitempty"`
	PullRequest *PullRequest       `json:"pull_request,omitempty"`

	// The following fields are only populated by Webhook events.
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`

	// The following field is only present when the webhook is triggered on
	// a repository belonging to an organization.
	Organization *Organization `json:"organization,omitempty"`
}

// PullRequestReviewCommentEvent is triggered when a comment is created on a
// portion of the unified diff of a pull request.
// The Webhook event name is "pull_request_review_comment".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#pullrequestreviewcommentevent
type PullRequestReviewCommentEvent struct {
	// Action is the action that was performed on the comment.
	// Possible values are: "created", "edited", "deleted".
	Action      *string             `json:"action,omitempty"`
	PullRequest *PullRequest        `json:"pull_request,omitempty"`
	Comment     *PullRequestComment `json:"comment,omitempty"`

	// The following fields are only populated by Webhook events.
	Changes *EditChange `json:"changes,omitempty"`
	Repo    *Repository `json:"repository,omitempty"`
	Sender  *User       `json:"sender,omitempty"`
}

// PushEvent represents a git push to a GitHub repository.
//
// GitHub API docs: http://developer.github.com/v3/activity/events/types/#pushevent
type PushEvent struct {
	PushID       *int                 `json:"push_id,omitempty"`
	Head         *string              `json:"head,omitempty"`
	Ref          *string              `json:"ref,omitempty"`
	Size         *int                 `json:"size,omitempty"`
	Commits      []PushEventCommit    `json:"commits,omitempty"`
	Repo         *PushEventRepository `json:"repository,omitempty"`
	Before       *string              `json:"before,omitempty"`
	DistinctSize *int                 `json:"distinct_size,omitempty"`

	// The following fields are only populated by Webhook events.
	After      *string          `json:"after,omitempty"`
	Created    *bool            `json:"created,omitempty"`
	Deleted    *bool            `json:"deleted,omitempty"`
	Forced     *bool            `json:"forced,omitempty"`
	BaseRef    *string          `json:"base_ref,omitempty"`
	Compare    *string          `json:"compare,omitempty"`
	HeadCommit *PushEventCommit `json:"head_commit,omitempty"`
	Pusher     *User            `json:"pusher,omitempty"`
	Sender     *User            `json:"sender,omitempty"`
}

func (p PushEvent) String() string {
	return Stringify(p)
}

// PushEventCommit represents a git commit in a GitHub PushEvent.
type PushEventCommit struct {
	Message  *string       `json:"message,omitempty"`
	Author   *CommitAuthor `json:"author,omitempty"`
	URL      *string       `json:"url,omitempty"`
	Distinct *bool         `json:"distinct,omitempty"`

	// The following fields are only populated by Events API.
	SHA *string `json:"sha,omitempty"`

	// The following fields are only populated by Webhook events.
	ID        *string       `json:"id,omitempty"`
	TreeID    *string       `json:"tree_id,omitempty"`
	Timestamp *Timestamp    `json:"timestamp,omitempty"`
	Committer *CommitAuthor `json:"committer,omitempty"`
	Added     []string      `json:"added,omitempty"`
	Removed   []string      `json:"removed,omitempty"`
	Modified  []string      `json:"modified,omitempty"`
}

func (p PushEventCommit) String() string {
	return Stringify(p)
}

// PushEventRepository represents the repo object in a PushEvent payload
type PushEventRepository struct {
	ID              *int                `json:"id,omitempty"`
	Name            *string             `json:"name,omitempty"`
	FullName        *string             `json:"full_name,omitempty"`
	Owner           *PushEventRepoOwner `json:"owner,omitempty"`
	Private         *bool               `json:"private,omitempty"`
	Description     *string             `json:"description,omitempty"`
	Fork            *bool               `json:"fork,omitempty"`
	CreatedAt       *Timestamp          `json:"created_at,omitempty"`
	PushedAt        *Timestamp          `json:"pushed_at,omitempty"`
	UpdatedAt       *Timestamp          `json:"updated_at,omitempty"`
	Homepage        *string             `json:"homepage,omitempty"`
	Size            *int                `json:"size,omitempty"`
	StargazersCount *int                `json:"stargazers_count,omitempty"`
	WatchersCount   *int                `json:"watchers_count,omitempty"`
	Language        *string             `json:"language,omitempty"`
	HasIssues       *bool               `json:"has_issues,omitempty"`
	HasDownloads    *bool               `json:"has_downloads,omitempty"`
	HasWiki         *bool               `json:"has_wiki,omitempty"`
	HasPages        *bool               `json:"has_pages,omitempty"`
	ForksCount      *int                `json:"forks_count,omitempty"`
	OpenIssuesCount *int                `json:"open_issues_count,omitempty"`
	DefaultBranch   *string             `json:"default_branch,omitempty"`
	MasterBranch    *string             `json:"master_branch,omitempty"`
	Organization    *string             `json:"organization,omitempty"`

	// The following fields are only populated by Webhook events.
	URL     *string `json:"url,omitempty"`
	HTMLURL *string `json:"html_url,omitempty"`
}

// PushEventRepoOwner is a basic reporesntation of user/org in a PushEvent payload
type PushEventRepoOwner struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}

// ReleaseEvent is triggered when a release is published.
// The Webhook event name is "release".
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#releaseevent
type ReleaseEvent struct {
	// Action is the action that was performed. Possible value is: "published".
	Action  *string            `json:"action,omitempty"`
	Release *RepositoryRelease `json:"release,omitempty"`

	// The following fields are only populated by Webhook events.
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}

// RepositoryEvent is triggered when a repository is created.
// The Webhook event name is "repository".
//
// Events of this type are not visible in timelines, they are only used to
// trigger organization webhooks.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#repositoryevent
type RepositoryEvent struct {
	// Action is the action that was performed. Possible values are: "created", "deleted",
	// "publicized", "privatized".
	Action *string     `json:"action,omitempty"`
	Repo   *Repository `json:"repository,omitempty"`

	// The following fields are only populated by Webhook events.
	Org    *Organization `json:"organization,omitempty"`
	Sender *User         `json:"sender,omitempty"`
}

// StatusEvent is triggered when the status of a Git commit changes.
// The Webhook event name is "status".
//
// Events of this type are not visible in timelines, they are only used to
// trigger hooks.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#statusevent
type StatusEvent struct {
	SHA *string `json:"sha,omitempty"`
	// State is the new state. Possible values are: "pending", "success", "failure", "error".
	State       *string   `json:"state,omitempty"`
	Description *string   `json:"description,omitempty"`
	TargetURL   *string   `json:"target_url,omitempty"`
	Branches    []*Branch `json:"branches,omitempty"`

	// The following fields are only populated by Webhook events.
	ID        *int              `json:"id,omitempty"`
	Name      *string           `json:"name,omitempty"`
	Context   *string           `json:"context,omitempty"`
	Commit    *RepositoryCommit `json:"commit,omitempty"`
	CreatedAt *Timestamp        `json:"created_at,omitempty"`
	UpdatedAt *Timestamp        `json:"updated_at,omitempty"`
	Repo      *Repository       `json:"repository,omitempty"`
	Sender    *User             `json:"sender,omitempty"`
}

// TeamAddEvent is triggered when a repository is added to a team.
// The Webhook event name is "team_add".
//
// Events of this type are not visible in timelines. These events are only used
// to trigger hooks.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#teamaddevent
type TeamAddEvent struct {
	Team *Team       `json:"team,omitempty"`
	Repo *Repository `json:"repository,omitempty"`

	// The following fields are only populated by Webhook events.
	Org    *Organization `json:"organization,omitempty"`
	Sender *User         `json:"sender,omitempty"`
}

// WatchEvent is related to starring a repository, not watching. See this API
// blog post for an explanation: https://developer.github.com/changes/2012-09-05-watcher-api/
//
// The event’s actor is the user who starred a repository, and the event’s
// repository is the repository that was starred.
//
// GitHub docs: https://developer.github.com/v3/activity/events/types/#watchevent
type WatchEvent struct {
	// Action is the action that was performed. Possible value is: "started".
	Action *string `json:"action,omitempty"`

	// The following fields are only populated by Webhook events.
	Repo   *Repository `json:"repository,omitempty"`
	Sender *User       `json:"sender,omitempty"`
}
