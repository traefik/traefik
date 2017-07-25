// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestRepositoriesService_ListCommits(t *testing.T) {
	setup()
	defer teardown()

	// given
	mux.HandleFunc("/repos/o/r/commits", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r,
			values{
				"sha":    "s",
				"path":   "p",
				"author": "a",
				"since":  "2013-08-01T00:00:00Z",
				"until":  "2013-09-03T00:00:00Z",
			})
		fmt.Fprintf(w, `[{"sha": "s"}]`)
	})

	opt := &CommitsListOptions{
		SHA:    "s",
		Path:   "p",
		Author: "a",
		Since:  time.Date(2013, time.August, 1, 0, 0, 0, 0, time.UTC),
		Until:  time.Date(2013, time.September, 3, 0, 0, 0, 0, time.UTC),
	}
	commits, _, err := client.Repositories.ListCommits(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Repositories.ListCommits returned error: %v", err)
	}

	want := []*RepositoryCommit{{SHA: String("s")}}
	if !reflect.DeepEqual(commits, want) {
		t.Errorf("Repositories.ListCommits returned %+v, want %+v", commits, want)
	}
}

func TestRepositoriesService_GetCommit(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/commits/s", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeGitSigningPreview)
		fmt.Fprintf(w, `{
		  "sha": "s",
		  "commit": { "message": "m" },
		  "author": { "login": "l" },
		  "committer": { "login": "l" },
		  "parents": [ { "sha": "s" } ],
		  "stats": { "additions": 104, "deletions": 4, "total": 108 },
		  "files": [
		    {
		      "filename": "f",
		      "additions": 10,
		      "deletions": 2,
		      "changes": 12,
		      "status": "s",
		      "patch": "p",
		      "blob_url": "b",
		      "raw_url": "r",
		      "contents_url": "c"
		    }
		  ]
		}`)
	})

	commit, _, err := client.Repositories.GetCommit(context.Background(), "o", "r", "s")
	if err != nil {
		t.Errorf("Repositories.GetCommit returned error: %v", err)
	}

	want := &RepositoryCommit{
		SHA: String("s"),
		Commit: &Commit{
			Message: String("m"),
		},
		Author: &User{
			Login: String("l"),
		},
		Committer: &User{
			Login: String("l"),
		},
		Parents: []Commit{
			{
				SHA: String("s"),
			},
		},
		Stats: &CommitStats{
			Additions: Int(104),
			Deletions: Int(4),
			Total:     Int(108),
		},
		Files: []CommitFile{
			{
				Filename:    String("f"),
				Additions:   Int(10),
				Deletions:   Int(2),
				Changes:     Int(12),
				Status:      String("s"),
				Patch:       String("p"),
				BlobURL:     String("b"),
				RawURL:      String("r"),
				ContentsURL: String("c"),
			},
		},
	}
	if !reflect.DeepEqual(commit, want) {
		t.Errorf("Repositories.GetCommit returned \n%+v, want \n%+v", commit, want)
	}
}

func TestRepositoriesService_GetCommitSHA1(t *testing.T) {
	setup()
	defer teardown()
	const sha1 = "01234abcde"

	mux.HandleFunc("/repos/o/r/commits/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeV3SHA)

		fmt.Fprintf(w, sha1)
	})

	got, _, err := client.Repositories.GetCommitSHA1(context.Background(), "o", "r", "master", "")
	if err != nil {
		t.Errorf("Repositories.GetCommitSHA1 returned error: %v", err)
	}

	want := sha1
	if got != want {
		t.Errorf("Repositories.GetCommitSHA1 = %v, want %v", got, want)
	}

	mux.HandleFunc("/repos/o/r/commits/tag", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeV3SHA)
		testHeader(t, r, "If-None-Match", `"`+sha1+`"`)

		w.WriteHeader(http.StatusNotModified)
	})

	got, _, err = client.Repositories.GetCommitSHA1(context.Background(), "o", "r", "tag", sha1)
	if err == nil {
		t.Errorf("Expected HTTP 304 response")
	}

	want = ""
	if got != want {
		t.Errorf("Repositories.GetCommitSHA1 = %v, want %v", got, want)
	}
}

func TestRepositoriesService_CompareCommits(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/compare/b...h", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{
		  "base_commit": {
		    "sha": "s",
		    "commit": {
		      "author": { "name": "n" },
		      "committer": { "name": "n" },
		      "message": "m",
		      "tree": { "sha": "t" }
		    },
		    "author": { "login": "l" },
		    "committer": { "login": "l" },
		    "parents": [ { "sha": "s" } ]
		  },
		  "status": "s",
		  "ahead_by": 1,
		  "behind_by": 2,
		  "total_commits": 1,
		  "commits": [
		    {
		      "sha": "s",
		      "commit": { "author": { "name": "n" } },
		      "author": { "login": "l" },
		      "committer": { "login": "l" },
		      "parents": [ { "sha": "s" } ]
		    }
		  ],
		  "files": [ { "filename": "f" } ],
		  "html_url":      "https://github.com/o/r/compare/b...h",
		  "permalink_url": "https://github.com/o/r/compare/o:bbcd538c8e72b8c175046e27cc8f907076331401...o:0328041d1152db8ae77652d1618a02e57f745f17",
		  "diff_url":      "https://github.com/o/r/compare/b...h.diff",
		  "patch_url":     "https://github.com/o/r/compare/b...h.patch",
		  "url":           "https://api.github.com/repos/o/r/compare/b...h"
		}`)
	})

	got, _, err := client.Repositories.CompareCommits(context.Background(), "o", "r", "b", "h")
	if err != nil {
		t.Errorf("Repositories.CompareCommits returned error: %v", err)
	}

	want := &CommitsComparison{
		BaseCommit: &RepositoryCommit{
			SHA: String("s"),
			Commit: &Commit{
				Author:    &CommitAuthor{Name: String("n")},
				Committer: &CommitAuthor{Name: String("n")},
				Message:   String("m"),
				Tree:      &Tree{SHA: String("t")},
			},
			Author:    &User{Login: String("l")},
			Committer: &User{Login: String("l")},
			Parents: []Commit{
				{
					SHA: String("s"),
				},
			},
		},
		Status:       String("s"),
		AheadBy:      Int(1),
		BehindBy:     Int(2),
		TotalCommits: Int(1),
		Commits: []RepositoryCommit{
			{
				SHA: String("s"),
				Commit: &Commit{
					Author: &CommitAuthor{Name: String("n")},
				},
				Author:    &User{Login: String("l")},
				Committer: &User{Login: String("l")},
				Parents: []Commit{
					{
						SHA: String("s"),
					},
				},
			},
		},
		Files: []CommitFile{
			{
				Filename: String("f"),
			},
		},
		HTMLURL:      String("https://github.com/o/r/compare/b...h"),
		PermalinkURL: String("https://github.com/o/r/compare/o:bbcd538c8e72b8c175046e27cc8f907076331401...o:0328041d1152db8ae77652d1618a02e57f745f17"),
		DiffURL:      String("https://github.com/o/r/compare/b...h.diff"),
		PatchURL:     String("https://github.com/o/r/compare/b...h.patch"),
		URL:          String("https://api.github.com/repos/o/r/compare/b...h"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Repositories.CompareCommits returned \n%+v, want \n%+v", got, want)
	}
}
