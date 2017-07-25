// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github_test

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

func ExampleClient_Markdown() {
	client := github.NewClient(nil)

	input := "# heading #\n\nLink to issue #1"
	opt := &github.MarkdownOptions{Mode: "gfm", Context: "google/go-github"}

	output, _, err := client.Markdown(context.Background(), input, opt)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(output)
}

func ExampleRepositoriesService_GetReadme() {
	client := github.NewClient(nil)

	readme, _, err := client.Repositories.GetReadme(context.Background(), "google", "go-github", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	content, err := readme.GetContent()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("google/go-github README:\n%v\n", content)
}

func ExampleRepositoriesService_List() {
	client := github.NewClient(nil)

	user := "willnorris"
	opt := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc"}

	repos, _, err := client.Repositories.List(context.Background(), user, opt)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Recently updated repositories by %q: %v", user, github.Stringify(repos))
}

func ExampleUsersService_ListAll() {
	client := github.NewClient(nil)
	opts := &github.UserListOptions{}
	for {
		users, _, err := client.Users.ListAll(context.Background(), opts)
		if err != nil {
			log.Fatalf("error listing users: %v", err)
		}
		if len(users) == 0 {
			break
		}
		opts.Since = *users[len(users)-1].ID
		// Process users...
	}
}
