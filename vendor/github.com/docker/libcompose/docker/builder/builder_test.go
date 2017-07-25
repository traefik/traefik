package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/pkg/errors"
	"strings"
)

type daemonClient struct {
	client.Client
	contextDir string
	imageName  string
	changes    int
	message    jsonmessage.JSONMessage
}

func (c *daemonClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	if c.imageName != "" {
		if len(options.Tags) != 1 || options.Tags[0] != c.imageName {
			return types.ImageBuildResponse{}, fmt.Errorf("expected image %q, got %v", c.imageName, options.Tags)
		}
	}
	if c.contextDir != "" {
		tmp, err := ioutil.TempDir("", "image-build-test")
		if err != nil {
			return types.ImageBuildResponse{}, err
		}
		if err := archive.Untar(context, tmp, nil); err != nil {
			return types.ImageBuildResponse{}, err
		}
		changes, err := archive.ChangesDirs(tmp, c.contextDir)
		if err != nil {
			return types.ImageBuildResponse{}, err
		}
		if len(changes) != c.changes {
			return types.ImageBuildResponse{}, fmt.Errorf("expected %d changes, got %v", c.changes, changes)
		}
		b, err := json.Marshal(c.message)
		if err != nil {
			return types.ImageBuildResponse{}, err
		}
		return types.ImageBuildResponse{
			Body: ioutil.NopCloser(bytes.NewReader(b)),
		}, nil
	}
	return types.ImageBuildResponse{
		Body: ioutil.NopCloser(strings.NewReader("{}")),
	}, errors.New("Engine no longer exists")
}

func TestBuildInvalidContextDirectoryOrDockerfile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		contextDirectory string
		dockerfile       string
		expected         string
	}{
		{
			contextDirectory: "",
			dockerfile:       "",
			expected:         "Cannot locate Dockerfile: Dockerfile",
		},
		{
			contextDirectory: "",
			dockerfile:       "test.Dockerfile",
			expected:         "Cannot locate Dockerfile: test.Dockerfile",
		},
		{
			contextDirectory: "/I/DONT/EXISTS",
			dockerfile:       "",
			expected:         "Cannot locate Dockerfile: Dockerfile",
		},
		{
			contextDirectory: "/I/DONT/EXISTS",
			dockerfile:       "test.Dockerfile",
			expected:         "Cannot locate Dockerfile: /I/DONT/EXISTS/test.Dockerfile",
		},
		{
			contextDirectory: tmpDir,
			dockerfile:       "test.Dockerfile",
			expected:         "Cannot locate Dockerfile: " + filepath.Join(tmpDir, "test.Dockerfile"),
		},
	}
	for _, c := range testCases {
		builder := &DaemonBuilder{
			ContextDirectory: c.contextDirectory,
			Dockerfile:       c.dockerfile,
		}
		err := builder.Build(context.Background(), "image")
		if err == nil || err.Error() != c.expected {
			t.Fatalf("expected an error %q, got %s", c.expected, err)
		}

	}
}

func TestBuildWithClientBuildError(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := ioutil.WriteFile(filepath.Join(tmpDir, DefaultDockerfileName), []byte("FROM busybox"), 0700); err != nil {
		t.Fatal(err)
	}

	imageName := "image"
	client := &daemonClient{}
	builder := &DaemonBuilder{
		ContextDirectory: tmpDir,
		Client:           client,
	}

	err = builder.Build(context.Background(), imageName)
	if err == nil || err.Error() != "Engine no longer exists" {
		t.Fatalf("expected an 'Engine no longer exists', got %s", err)
	}
}

func TestBuildWithDefaultDockerfile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := ioutil.WriteFile(filepath.Join(tmpDir, DefaultDockerfileName), []byte("FROM busybox"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "afile"), []byte("another file"), 0700); err != nil {
		t.Fatal(err)
	}

	imageName := "image"
	client := &daemonClient{
		contextDir: tmpDir,
		imageName:  imageName,
	}
	builder := &DaemonBuilder{
		ContextDirectory: tmpDir,
		Client:           client,
	}

	err = builder.Build(context.Background(), imageName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildWithDefaultLowercaseDockerfile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "dockerfile"), []byte("FROM busybox"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "afile"), []byte("another file"), 0700); err != nil {
		t.Fatal(err)
	}

	imageName := "image"
	client := &daemonClient{
		contextDir: tmpDir,
		imageName:  imageName,
	}
	builder := &DaemonBuilder{
		ContextDirectory: tmpDir,
		Client:           client,
	}

	err = builder.Build(context.Background(), imageName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildWithSpecificDockerfile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "test.Dockerfile"), []byte("FROM busybox"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "afile"), []byte("another file"), 0700); err != nil {
		t.Fatal(err)
	}

	imageName := "image"
	client := &daemonClient{
		contextDir: tmpDir,
		imageName:  imageName,
	}
	builder := &DaemonBuilder{
		ContextDirectory: tmpDir,
		Dockerfile:       "test.Dockerfile",
		Client:           client,
	}

	err = builder.Build(context.Background(), imageName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildWithDockerignoreNothing(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "test.Dockerfile"), []byte("FROM busybox"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "afile"), []byte("another file"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, ".dockerignore"), []byte(""), 0700); err != nil {
		t.Fatal(err)
	}

	imageName := "image"
	client := &daemonClient{
		contextDir: tmpDir,
		imageName:  imageName,
	}
	builder := &DaemonBuilder{
		ContextDirectory: tmpDir,
		Dockerfile:       "test.Dockerfile",
		Client:           client,
	}

	err = builder.Build(context.Background(), imageName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildWithDockerignoreDockerfileAndItself(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "test.Dockerfile"), []byte("FROM busybox"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "afile"), []byte("another file"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, ".dockerignore"), []byte("Dockerfile\n.dockerignore"), 0700); err != nil {
		t.Fatal(err)
	}

	imageName := "image"
	client := &daemonClient{
		contextDir: tmpDir,
		imageName:  imageName,
	}
	builder := &DaemonBuilder{
		ContextDirectory: tmpDir,
		Dockerfile:       "test.Dockerfile",
		Client:           client,
	}

	err = builder.Build(context.Background(), imageName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildWithDockerignoreAfile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "test.Dockerfile"), []byte("FROM busybox"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "afile"), []byte("another file"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, ".dockerignore"), []byte("afile"), 0700); err != nil {
		t.Fatal(err)
	}

	imageName := "image"
	client := &daemonClient{
		contextDir: tmpDir,
		imageName:  imageName,
		changes:    1,
	}
	builder := &DaemonBuilder{
		ContextDirectory: tmpDir,
		Dockerfile:       "test.Dockerfile",
		Client:           client,
	}

	err = builder.Build(context.Background(), imageName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildWithErrorJSONMessage(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "daemonbuilder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := ioutil.WriteFile(filepath.Join(tmpDir, DefaultDockerfileName), []byte("FROM busybox"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "afile"), []byte("another file"), 0700); err != nil {
		t.Fatal(err)
	}

	imageName := "image"
	client := &daemonClient{
		contextDir: tmpDir,
		imageName:  imageName,
		message: jsonmessage.JSONMessage{
			Error: &jsonmessage.JSONError{
				Code:    0,
				Message: "error",
			},
		},
	}
	builder := &DaemonBuilder{
		ContextDirectory: tmpDir,
		Client:           client,
	}

	err = builder.Build(context.Background(), imageName)
	expectedError := "Status: error, Code: 1"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("expected an error about %q, got %s", expectedError, err)
	}
}
