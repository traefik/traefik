package image

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/image/build"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/pkg/progress"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/filesync"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

const clientSessionRemote = "client-session"

func isSessionSupported(dockerCli command.Cli) bool {
	return dockerCli.ServerInfo().HasExperimental && versions.GreaterThanOrEqualTo(dockerCli.Client().ClientVersion(), "1.31")
}

func trySession(dockerCli command.Cli, contextDir string) (*session.Session, error) {
	var s *session.Session
	if isSessionSupported(dockerCli) {
		sharedKey, err := getBuildSharedKey(contextDir)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get build shared key")
		}
		s, err = session.NewSession(filepath.Base(contextDir), sharedKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create session")
		}
	}
	return s, nil
}

func addDirToSession(session *session.Session, contextDir string, progressOutput progress.Output, done chan error) error {
	excludes, err := build.ReadDockerignore(contextDir)
	if err != nil {
		return err
	}

	p := &sizeProgress{out: progressOutput, action: "Streaming build context to Docker daemon"}

	workdirProvider := filesync.NewFSSyncProvider([]filesync.SyncedDir{
		{Dir: contextDir, Excludes: excludes},
	})
	session.Allow(workdirProvider)

	// this will be replaced on parallel build jobs. keep the current
	// progressbar for now
	if snpc, ok := workdirProvider.(interface {
		SetNextProgressCallback(func(int, bool), chan error)
	}); ok {
		snpc.SetNextProgressCallback(p.update, done)
	}

	return nil
}

type sizeProgress struct {
	out     progress.Output
	action  string
	limiter *rate.Limiter
}

func (sp *sizeProgress) update(size int, last bool) {
	if sp.limiter == nil {
		sp.limiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 1)
	}
	if last || sp.limiter.Allow() {
		sp.out.WriteProgress(progress.Progress{Action: sp.action, Current: int64(size), LastUpdate: last})
	}
}

type bufferedWriter struct {
	done chan error
	io.Writer
	buf     *bytes.Buffer
	flushed chan struct{}
	mu      sync.Mutex
}

func newBufferedWriter(done chan error, w io.Writer) *bufferedWriter {
	bw := &bufferedWriter{done: done, Writer: w, buf: new(bytes.Buffer), flushed: make(chan struct{})}
	go func() {
		<-done
		bw.flushBuffer()
	}()
	return bw
}

func (bw *bufferedWriter) Write(dt []byte) (int, error) {
	select {
	case <-bw.done:
		bw.flushBuffer()
		return bw.Writer.Write(dt)
	default:
		return bw.buf.Write(dt)
	}
}

func (bw *bufferedWriter) flushBuffer() {
	bw.mu.Lock()
	select {
	case <-bw.flushed:
	default:
		bw.Writer.Write(bw.buf.Bytes())
		close(bw.flushed)
	}
	bw.mu.Unlock()
}

func getBuildSharedKey(dir string) (string, error) {
	// build session is hash of build dir with node based randomness
	s := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", tryNodeIdentifier(), dir)))
	return hex.EncodeToString(s[:]), nil
}

func tryNodeIdentifier() string {
	out := cliconfig.Dir() // return config dir as default on permission error
	if err := os.MkdirAll(cliconfig.Dir(), 0700); err == nil {
		sessionFile := filepath.Join(cliconfig.Dir(), ".buildNodeID")
		if _, err := os.Lstat(sessionFile); err != nil {
			if os.IsNotExist(err) { // create a new file with stored randomness
				b := make([]byte, 32)
				if _, err := rand.Read(b); err != nil {
					return out
				}
				if err := ioutil.WriteFile(sessionFile, []byte(hex.EncodeToString(b)), 0600); err != nil {
					return out
				}
			}
		}

		dt, err := ioutil.ReadFile(sessionFile)
		if err == nil {
			return string(dt)
		}
	}
	return out
}
