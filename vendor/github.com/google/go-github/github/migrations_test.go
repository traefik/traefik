// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestMigrationService_StartMigration(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/migrations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeMigrationsPreview)

		w.WriteHeader(http.StatusCreated)
		w.Write(migrationJSON)
	})

	opt := &MigrationOptions{
		LockRepositories:   true,
		ExcludeAttachments: false,
	}
	got, _, err := client.Migrations.StartMigration(context.Background(), "o", []string{"r"}, opt)
	if err != nil {
		t.Errorf("StartMigration returned error: %v", err)
	}
	if want := wantMigration; !reflect.DeepEqual(got, want) {
		t.Errorf("StartMigration = %+v, want %+v", got, want)
	}
}

func TestMigrationService_ListMigrations(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/migrations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeMigrationsPreview)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("[%s]", migrationJSON)))
	})

	got, _, err := client.Migrations.ListMigrations(context.Background(), "o")
	if err != nil {
		t.Errorf("ListMigrations returned error: %v", err)
	}
	if want := []*Migration{wantMigration}; !reflect.DeepEqual(got, want) {
		t.Errorf("ListMigrations = %+v, want %+v", got, want)
	}
}

func TestMigrationService_MigrationStatus(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/migrations/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeMigrationsPreview)

		w.WriteHeader(http.StatusOK)
		w.Write(migrationJSON)
	})

	got, _, err := client.Migrations.MigrationStatus(context.Background(), "o", 1)
	if err != nil {
		t.Errorf("MigrationStatus returned error: %v", err)
	}
	if want := wantMigration; !reflect.DeepEqual(got, want) {
		t.Errorf("MigrationStatus = %+v, want %+v", got, want)
	}
}

func TestMigrationService_MigrationArchiveURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/migrations/1/archive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeMigrationsPreview)

		http.Redirect(w, r, "/yo", http.StatusFound)
	})
	mux.HandleFunc("/yo", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("0123456789abcdef"))
	})

	got, err := client.Migrations.MigrationArchiveURL(context.Background(), "o", 1)
	if err != nil {
		t.Errorf("MigrationStatus returned error: %v", err)
	}
	if want := "/yo"; !strings.HasSuffix(got, want) {
		t.Errorf("MigrationArchiveURL = %+v, want %+v", got, want)
	}
}

func TestMigrationService_DeleteMigration(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/migrations/1/archive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeMigrationsPreview)

		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Migrations.DeleteMigration(context.Background(), "o", 1); err != nil {
		t.Errorf("DeleteMigration returned error: %v", err)
	}
}

func TestMigrationService_UnlockRepo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/migrations/1/repos/r/lock", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeMigrationsPreview)

		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Migrations.UnlockRepo(context.Background(), "o", 1, "r"); err != nil {
		t.Errorf("UnlockRepo returned error: %v", err)
	}
}

var migrationJSON = []byte(`{
  "id": 79,
  "guid": "0b989ba4-242f-11e5-81e1-c7b6966d2516",
  "state": "pending",
  "lock_repositories": true,
  "exclude_attachments": false,
  "url": "https://api.github.com/orgs/octo-org/migrations/79",
  "created_at": "2015-07-06T15:33:38-07:00",
  "updated_at": "2015-07-06T15:33:38-07:00",
  "repositories": [
    {
      "id": 1296269,
      "name": "Hello-World",
      "full_name": "octocat/Hello-World",
      "description": "This your first repo!"
    }
  ]
}`)

var wantMigration = &Migration{
	ID:                 Int(79),
	GUID:               String("0b989ba4-242f-11e5-81e1-c7b6966d2516"),
	State:              String("pending"),
	LockRepositories:   Bool(true),
	ExcludeAttachments: Bool(false),
	URL:                String("https://api.github.com/orgs/octo-org/migrations/79"),
	CreatedAt:          String("2015-07-06T15:33:38-07:00"),
	UpdatedAt:          String("2015-07-06T15:33:38-07:00"),
	Repositories: []*Repository{
		{
			ID:          Int(1296269),
			Name:        String("Hello-World"),
			FullName:    String("octocat/Hello-World"),
			Description: String("This your first repo!"),
		},
	},
}
