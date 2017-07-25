// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestRepositoriesService_ListReleases(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOptions{Page: 2}
	releases, _, err := client.Repositories.ListReleases(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Repositories.ListReleases returned error: %v", err)
	}
	want := []*RepositoryRelease{{ID: Int(1)}}
	if !reflect.DeepEqual(releases, want) {
		t.Errorf("Repositories.ListReleases returned %+v, want %+v", releases, want)
	}
}

func TestRepositoriesService_GetRelease(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1,"author":{"login":"l"}}`)
	})

	release, resp, err := client.Repositories.GetRelease(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Repositories.GetRelease returned error: %v\n%v", err, resp.Body)
	}

	want := &RepositoryRelease{ID: Int(1), Author: &User{Login: String("l")}}
	if !reflect.DeepEqual(release, want) {
		t.Errorf("Repositories.GetRelease returned %+v, want %+v", release, want)
	}
}

func TestRepositoriesService_GetLatestRelease(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":3}`)
	})

	release, resp, err := client.Repositories.GetLatestRelease(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.GetLatestRelease returned error: %v\n%v", err, resp.Body)
	}

	want := &RepositoryRelease{ID: Int(3)}
	if !reflect.DeepEqual(release, want) {
		t.Errorf("Repositories.GetLatestRelease returned %+v, want %+v", release, want)
	}
}

func TestRepositoriesService_GetReleaseByTag(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/tags/foo", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":13}`)
	})

	release, resp, err := client.Repositories.GetReleaseByTag(context.Background(), "o", "r", "foo")
	if err != nil {
		t.Errorf("Repositories.GetReleaseByTag returned error: %v\n%v", err, resp.Body)
	}

	want := &RepositoryRelease{ID: Int(13)}
	if !reflect.DeepEqual(release, want) {
		t.Errorf("Repositories.GetReleaseByTag returned %+v, want %+v", release, want)
	}
}

func TestRepositoriesService_CreateRelease(t *testing.T) {
	setup()
	defer teardown()

	input := &RepositoryRelease{Name: String("v1.0")}

	mux.HandleFunc("/repos/o/r/releases", func(w http.ResponseWriter, r *http.Request) {
		v := new(RepositoryRelease)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		fmt.Fprint(w, `{"id":1}`)
	})

	release, _, err := client.Repositories.CreateRelease(context.Background(), "o", "r", input)
	if err != nil {
		t.Errorf("Repositories.CreateRelease returned error: %v", err)
	}

	want := &RepositoryRelease{ID: Int(1)}
	if !reflect.DeepEqual(release, want) {
		t.Errorf("Repositories.CreateRelease returned %+v, want %+v", release, want)
	}
}

func TestRepositoriesService_EditRelease(t *testing.T) {
	setup()
	defer teardown()

	input := &RepositoryRelease{Name: String("n")}

	mux.HandleFunc("/repos/o/r/releases/1", func(w http.ResponseWriter, r *http.Request) {
		v := new(RepositoryRelease)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		fmt.Fprint(w, `{"id":1}`)
	})

	release, _, err := client.Repositories.EditRelease(context.Background(), "o", "r", 1, input)
	if err != nil {
		t.Errorf("Repositories.EditRelease returned error: %v", err)
	}
	want := &RepositoryRelease{ID: Int(1)}
	if !reflect.DeepEqual(release, want) {
		t.Errorf("Repositories.EditRelease returned = %+v, want %+v", release, want)
	}
}

func TestRepositoriesService_DeleteRelease(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Repositories.DeleteRelease(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Repositories.DeleteRelease returned error: %v", err)
	}
}

func TestRepositoriesService_ListReleaseAssets(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/1/assets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOptions{Page: 2}
	assets, _, err := client.Repositories.ListReleaseAssets(context.Background(), "o", "r", 1, opt)
	if err != nil {
		t.Errorf("Repositories.ListReleaseAssets returned error: %v", err)
	}
	want := []*ReleaseAsset{{ID: Int(1)}}
	if !reflect.DeepEqual(assets, want) {
		t.Errorf("Repositories.ListReleaseAssets returned %+v, want %+v", assets, want)
	}
}

func TestRepositoriesService_GetReleaseAsset(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/assets/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})

	asset, _, err := client.Repositories.GetReleaseAsset(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Repositories.GetReleaseAsset returned error: %v", err)
	}
	want := &ReleaseAsset{ID: Int(1)}
	if !reflect.DeepEqual(asset, want) {
		t.Errorf("Repositories.GetReleaseAsset returned %+v, want %+v", asset, want)
	}
}

func TestRepositoriesService_DownloadReleaseAsset_Stream(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/assets/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", defaultMediaType)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=hello-world.txt")
		fmt.Fprint(w, "Hello World")
	})

	reader, _, err := client.Repositories.DownloadReleaseAsset(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Repositories.DownloadReleaseAsset returned error: %v", err)
	}
	want := []byte("Hello World")
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Errorf("Repositories.DownloadReleaseAsset returned bad reader: %v", err)
	}
	if !bytes.Equal(want, content) {
		t.Errorf("Repositories.DownloadReleaseAsset returned %+v, want %+v", content, want)
	}
}

func TestRepositoriesService_DownloadReleaseAsset_Redirect(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/assets/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", defaultMediaType)
		http.Redirect(w, r, "/yo", http.StatusFound)
	})

	_, got, err := client.Repositories.DownloadReleaseAsset(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Repositories.DownloadReleaseAsset returned error: %v", err)
	}
	want := "/yo"
	if !strings.HasSuffix(got, want) {
		t.Errorf("Repositories.DownloadReleaseAsset returned %+v, want %+v", got, want)
	}
}

func TestRepositoriesService_DownloadReleaseAsset_APIError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/assets/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", defaultMediaType)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Not Found","documentation_url":"https://developer.github.com/v3"}`)
	})

	resp, loc, err := client.Repositories.DownloadReleaseAsset(context.Background(), "o", "r", 1)
	if err == nil {
		t.Error("Repositories.DownloadReleaseAsset did not return an error")
	}

	if resp != nil {
		resp.Close()
		t.Error("Repositories.DownloadReleaseAsset returned stream, want nil")
	}

	if loc != "" {
		t.Errorf(`Repositories.DownloadReleaseAsset returned "%s", want empty ""`, loc)
	}
}

func TestRepositoriesService_EditReleaseAsset(t *testing.T) {
	setup()
	defer teardown()

	input := &ReleaseAsset{Name: String("n")}

	mux.HandleFunc("/repos/o/r/releases/assets/1", func(w http.ResponseWriter, r *http.Request) {
		v := new(ReleaseAsset)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		fmt.Fprint(w, `{"id":1}`)
	})

	asset, _, err := client.Repositories.EditReleaseAsset(context.Background(), "o", "r", 1, input)
	if err != nil {
		t.Errorf("Repositories.EditReleaseAsset returned error: %v", err)
	}
	want := &ReleaseAsset{ID: Int(1)}
	if !reflect.DeepEqual(asset, want) {
		t.Errorf("Repositories.EditReleaseAsset returned = %+v, want %+v", asset, want)
	}
}

func TestRepositoriesService_DeleteReleaseAsset(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/assets/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Repositories.DeleteReleaseAsset(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Repositories.DeleteReleaseAsset returned error: %v", err)
	}
}

func TestRepositoriesService_UploadReleaseAsset(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases/1/assets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", "text/plain; charset=utf-8")
		testHeader(t, r, "Content-Length", "12")
		testFormValues(t, r, values{"name": "n"})
		testBody(t, r, "Upload me !\n")

		fmt.Fprintf(w, `{"id":1}`)
	})

	file, dir, err := openTestFile("upload.txt", "Upload me !\n")
	if err != nil {
		t.Fatalf("Unable to create temp file: %v", err)
	}
	defer os.RemoveAll(dir)

	opt := &UploadOptions{Name: "n"}
	asset, _, err := client.Repositories.UploadReleaseAsset(context.Background(), "o", "r", 1, opt, file)
	if err != nil {
		t.Errorf("Repositories.UploadReleaseAssert returned error: %v", err)
	}
	want := &ReleaseAsset{ID: Int(1)}
	if !reflect.DeepEqual(asset, want) {
		t.Errorf("Repositories.UploadReleaseAssert returned %+v, want %+v", asset, want)
	}
}
