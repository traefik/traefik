// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Repository contents API methods.
// GitHub API docs: https://developer.github.com/v3/repos/contents/

package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// RepositoryContent represents a file or directory in a github repository.
type RepositoryContent struct {
	Type     *string `json:"type,omitempty"`
	Encoding *string `json:"encoding,omitempty"`
	Size     *int    `json:"size,omitempty"`
	Name     *string `json:"name,omitempty"`
	Path     *string `json:"path,omitempty"`
	// Content contains the actual file content, which may be encoded.
	// Callers should call GetContent which will decode the content if
	// necessary.
	Content     *string `json:"content,omitempty"`
	SHA         *string `json:"sha,omitempty"`
	URL         *string `json:"url,omitempty"`
	GitURL      *string `json:"git_url,omitempty"`
	HTMLURL     *string `json:"html_url,omitempty"`
	DownloadURL *string `json:"download_url,omitempty"`
}

// RepositoryContentResponse holds the parsed response from CreateFile, UpdateFile, and DeleteFile.
type RepositoryContentResponse struct {
	Content *RepositoryContent `json:"content,omitempty"`
	Commit  `json:"commit,omitempty"`
}

// RepositoryContentFileOptions specifies optional parameters for CreateFile, UpdateFile, and DeleteFile.
type RepositoryContentFileOptions struct {
	Message   *string       `json:"message,omitempty"`
	Content   []byte        `json:"content,omitempty"` // unencoded
	SHA       *string       `json:"sha,omitempty"`
	Branch    *string       `json:"branch,omitempty"`
	Author    *CommitAuthor `json:"author,omitempty"`
	Committer *CommitAuthor `json:"committer,omitempty"`
}

// RepositoryContentGetOptions represents an optional ref parameter, which can be a SHA,
// branch, or tag
type RepositoryContentGetOptions struct {
	Ref string `url:"ref,omitempty"`
}

// String converts RepositoryContent to a string. It's primarily for testing.
func (r RepositoryContent) String() string {
	return Stringify(r)
}

// GetContent returns the content of r, decoding it if necessary.
func (r *RepositoryContent) GetContent() (string, error) {
	var encoding string
	if r.Encoding != nil {
		encoding = *r.Encoding
	}

	switch encoding {
	case "base64":
		c, err := base64.StdEncoding.DecodeString(*r.Content)
		return string(c), err
	case "":
		if r.Content == nil {
			return "", nil
		}
		return *r.Content, nil
	default:
		return "", fmt.Errorf("unsupported content encoding: %v", encoding)
	}
}

// GetReadme gets the Readme file for the repository.
//
// GitHub API docs: https://developer.github.com/v3/repos/contents/#get-the-readme
func (s *RepositoriesService) GetReadme(ctx context.Context, owner, repo string, opt *RepositoryContentGetOptions) (*RepositoryContent, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/readme", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}
	readme := new(RepositoryContent)
	resp, err := s.client.Do(ctx, req, readme)
	if err != nil {
		return nil, resp, err
	}
	return readme, resp, nil
}

// DownloadContents returns an io.ReadCloser that reads the contents of the
// specified file. This function will work with files of any size, as opposed
// to GetContents which is limited to 1 Mb files. It is the caller's
// responsibility to close the ReadCloser.
func (s *RepositoriesService) DownloadContents(ctx context.Context, owner, repo, filepath string, opt *RepositoryContentGetOptions) (io.ReadCloser, error) {
	dir := path.Dir(filepath)
	filename := path.Base(filepath)
	_, dirContents, _, err := s.GetContents(ctx, owner, repo, dir, opt)
	if err != nil {
		return nil, err
	}
	for _, contents := range dirContents {
		if *contents.Name == filename {
			if contents.DownloadURL == nil || *contents.DownloadURL == "" {
				return nil, fmt.Errorf("No download link found for %s", filepath)
			}
			resp, err := s.client.client.Get(*contents.DownloadURL)
			if err != nil {
				return nil, err
			}
			return resp.Body, nil
		}
	}
	return nil, fmt.Errorf("No file named %s found in %s", filename, dir)
}

// GetContents can return either the metadata and content of a single file
// (when path references a file) or the metadata of all the files and/or
// subdirectories of a directory (when path references a directory). To make it
// easy to distinguish between both result types and to mimic the API as much
// as possible, both result types will be returned but only one will contain a
// value and the other will be nil.
//
// GitHub API docs: https://developer.github.com/v3/repos/contents/#get-contents
func (s *RepositoriesService) GetContents(ctx context.Context, owner, repo, path string, opt *RepositoryContentGetOptions) (fileContent *RepositoryContent, directoryContent []*RepositoryContent, resp *Response, err error) {
	escapedPath := (&url.URL{Path: path}).String()
	u := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, escapedPath)
	u, err = addOptions(u, opt)
	if err != nil {
		return nil, nil, nil, err
	}
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	var rawJSON json.RawMessage
	resp, err = s.client.Do(ctx, req, &rawJSON)
	if err != nil {
		return nil, nil, resp, err
	}
	fileUnmarshalError := json.Unmarshal(rawJSON, &fileContent)
	if fileUnmarshalError == nil {
		return fileContent, nil, resp, nil
	}
	directoryUnmarshalError := json.Unmarshal(rawJSON, &directoryContent)
	if directoryUnmarshalError == nil {
		return nil, directoryContent, resp, nil
	}
	return nil, nil, resp, fmt.Errorf("unmarshalling failed for both file and directory content: %s and %s", fileUnmarshalError, directoryUnmarshalError)
}

// CreateFile creates a new file in a repository at the given path and returns
// the commit and file metadata.
//
// GitHub API docs: https://developer.github.com/v3/repos/contents/#create-a-file
func (s *RepositoriesService) CreateFile(ctx context.Context, owner, repo, path string, opt *RepositoryContentFileOptions) (*RepositoryContentResponse, *Response, error) {
	u := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path)
	req, err := s.client.NewRequest("PUT", u, opt)
	if err != nil {
		return nil, nil, err
	}
	createResponse := new(RepositoryContentResponse)
	resp, err := s.client.Do(ctx, req, createResponse)
	if err != nil {
		return nil, resp, err
	}
	return createResponse, resp, nil
}

// UpdateFile updates a file in a repository at the given path and returns the
// commit and file metadata. Requires the blob SHA of the file being updated.
//
// GitHub API docs: https://developer.github.com/v3/repos/contents/#update-a-file
func (s *RepositoriesService) UpdateFile(ctx context.Context, owner, repo, path string, opt *RepositoryContentFileOptions) (*RepositoryContentResponse, *Response, error) {
	u := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path)
	req, err := s.client.NewRequest("PUT", u, opt)
	if err != nil {
		return nil, nil, err
	}
	updateResponse := new(RepositoryContentResponse)
	resp, err := s.client.Do(ctx, req, updateResponse)
	if err != nil {
		return nil, resp, err
	}
	return updateResponse, resp, nil
}

// DeleteFile deletes a file from a repository and returns the commit.
// Requires the blob SHA of the file to be deleted.
//
// GitHub API docs: https://developer.github.com/v3/repos/contents/#delete-a-file
func (s *RepositoriesService) DeleteFile(ctx context.Context, owner, repo, path string, opt *RepositoryContentFileOptions) (*RepositoryContentResponse, *Response, error) {
	u := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path)
	req, err := s.client.NewRequest("DELETE", u, opt)
	if err != nil {
		return nil, nil, err
	}
	deleteResponse := new(RepositoryContentResponse)
	resp, err := s.client.Do(ctx, req, deleteResponse)
	if err != nil {
		return nil, resp, err
	}
	return deleteResponse, resp, nil
}

// archiveFormat is used to define the archive type when calling GetArchiveLink.
type archiveFormat string

const (
	// Tarball specifies an archive in gzipped tar format.
	Tarball archiveFormat = "tarball"

	// Zipball specifies an archive in zip format.
	Zipball archiveFormat = "zipball"
)

// GetArchiveLink returns an URL to download a tarball or zipball archive for a
// repository. The archiveFormat can be specified by either the github.Tarball
// or github.Zipball constant.
//
// GitHub API docs: https://developer.github.com/v3/repos/contents/#get-archive-link
func (s *RepositoriesService) GetArchiveLink(ctx context.Context, owner, repo string, archiveformat archiveFormat, opt *RepositoryContentGetOptions) (*url.URL, *Response, error) {
	u := fmt.Sprintf("repos/%s/%s/%s", owner, repo, archiveformat)
	if opt != nil && opt.Ref != "" {
		u += fmt.Sprintf("/%s", opt.Ref)
	}
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}
	var resp *http.Response
	// Use http.DefaultTransport if no custom Transport is configured
	ctx, req = withContext(ctx, req)
	if s.client.client.Transport == nil {
		resp, err = http.DefaultTransport.RoundTrip(req)
	} else {
		resp, err = s.client.client.Transport.RoundTrip(req)
	}
	if err != nil {
		return nil, nil, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		return nil, newResponse(resp), fmt.Errorf("unexpected status code: %s", resp.Status)
	}
	parsedURL, err := url.Parse(resp.Header.Get("Location"))
	return parsedURL, newResponse(resp), err
}
