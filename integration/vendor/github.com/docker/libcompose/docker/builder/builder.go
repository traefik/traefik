package builder

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

// DefaultDockerfileName is the default name of a Dockerfile
const DefaultDockerfileName = "Dockerfile"

// Builder defines methods to provide a docker builder. This makes libcompose
// not tied up to the docker daemon builder.
type Builder interface {
	Build(imageName string) error
}

// DaemonBuilder is the daemon "docker build" Builder implementation.
type DaemonBuilder struct {
	Client           client.ImageAPIClient
	ContextDirectory string
	Dockerfile       string
	AuthConfigs      map[string]types.AuthConfig
	NoCache          bool
	ForceRemove      bool
	Pull             bool
	BuildArgs        map[string]string
}

// Build implements Builder. It consumes the docker build API endpoint and sends
// a tar of the specified service build context.
func (d *DaemonBuilder) Build(ctx context.Context, imageName string) error {
	buildCtx, err := createTar(d.ContextDirectory, d.Dockerfile)
	if err != nil {
		return err
	}
	defer buildCtx.Close()

	var progBuff io.Writer = os.Stdout
	var buildBuff io.Writer = os.Stdout

	// Setup an upload progress bar
	progressOutput := streamformatter.NewStreamFormatter().NewProgressOutput(progBuff, true)

	var body io.Reader = progress.NewProgressReader(buildCtx, progressOutput, 0, "", "Sending build context to Docker daemon")

	logrus.Infof("Building %s...", imageName)

	outFd, isTerminalOut := term.GetFdInfo(os.Stdout)

	response, err := d.Client.ImageBuild(ctx, body, types.ImageBuildOptions{
		Tags:        []string{imageName},
		NoCache:     d.NoCache,
		Remove:      true,
		ForceRemove: d.ForceRemove,
		PullParent:  d.Pull,
		Dockerfile:  d.Dockerfile,
		AuthConfigs: d.AuthConfigs,
		BuildArgs:   d.BuildArgs,
	})
	if err != nil {
		return err
	}

	err = jsonmessage.DisplayJSONMessagesStream(response.Body, buildBuff, outFd, isTerminalOut, nil)
	if err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			// If no error code is set, default to 1
			if jerr.Code == 0 {
				jerr.Code = 1
			}
			return fmt.Errorf("Status: %s, Code: %d", jerr.Message, jerr.Code)
		}
	}
	return err
}

// CreateTar create a build context tar for the specified project and service name.
func createTar(contextDirectory, dockerfile string) (io.ReadCloser, error) {
	// This code was ripped off from docker/api/client/build.go
	dockerfileName := filepath.Join(contextDirectory, dockerfile)

	absContextDirectory, err := filepath.Abs(contextDirectory)
	if err != nil {
		return nil, err
	}

	filename := dockerfileName

	if dockerfile == "" {
		// No -f/--file was specified so use the default
		dockerfileName = DefaultDockerfileName
		filename = filepath.Join(absContextDirectory, dockerfileName)

		// Just to be nice ;-) look for 'dockerfile' too but only
		// use it if we found it, otherwise ignore this check
		if _, err = os.Lstat(filename); os.IsNotExist(err) {
			tmpFN := path.Join(absContextDirectory, strings.ToLower(dockerfileName))
			if _, err = os.Lstat(tmpFN); err == nil {
				dockerfileName = strings.ToLower(dockerfileName)
				filename = tmpFN
			}
		}
	}

	origDockerfile := dockerfileName // used for error msg
	if filename, err = filepath.Abs(filename); err != nil {
		return nil, err
	}

	// Now reset the dockerfileName to be relative to the build context
	dockerfileName, err = filepath.Rel(absContextDirectory, filename)
	if err != nil {
		return nil, err
	}

	// And canonicalize dockerfile name to a platform-independent one
	dockerfileName, err = archive.CanonicalTarNameForPath(dockerfileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot canonicalize dockerfile path %s: %v", dockerfileName, err)
	}

	if _, err = os.Lstat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cannot locate Dockerfile: %s", origDockerfile)
	}
	var includes = []string{"."}
	var excludes []string

	dockerIgnorePath := path.Join(contextDirectory, ".dockerignore")
	dockerIgnore, err := os.Open(dockerIgnorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		logrus.Warnf("Error while reading .dockerignore (%s) : %s", dockerIgnorePath, err.Error())
		excludes = make([]string, 0)
	} else {
		excludes, err = dockerignore.ReadAll(dockerIgnore)
		if err != nil {
			return nil, err
		}
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed.  The deamon will remove them for us, if needed, after it
	// parses the Dockerfile.
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(dockerfileName, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", dockerfileName)
	}

	if err := builder.ValidateContextDirectory(contextDirectory, excludes); err != nil {
		return nil, fmt.Errorf("Error checking context is accessible: '%s'. Please check permissions and try again.", err)
	}

	options := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	return archive.TarWithOptions(contextDirectory, options)
}
