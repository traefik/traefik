package command

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/command/base"
)

const (
	// lockKillGracePeriod is how long we allow a child between
	// a SIGTERM and a SIGKILL. This is to let the child cleanup
	// any necessary state. We have to balance this with the risk
	// of a split-brain where multiple children may be acting as if
	// they hold a lock. This value is currently based on the default
	// lock-delay value of 15 seconds. This only affects locks and not
	// semaphores.
	lockKillGracePeriod = 5 * time.Second

	// defaultMonitorRetry is the number of 500 errors we will tolerate
	// before declaring the lock gone.
	defaultMonitorRetry = 3

	// defaultMonitorRetryTime is the amount of time to wait between
	// retries.
	defaultMonitorRetryTime = 1 * time.Second
)

// LockCommand is a Command implementation that is used to setup
// a "lock" which manages lock acquisition and invokes a sub-process
type LockCommand struct {
	base.Command

	ShutdownCh <-chan struct{}

	child     *os.Process
	childLock sync.Mutex

	verbose bool
}

func (c *LockCommand) Help() string {
	helpText := `
Usage: consul lock [options] prefix child...

  Acquires a lock or semaphore at a given path, and invokes a child process
  when successful. The child process can assume the lock is held while it
  executes. If the lock is lost or communication is disrupted the child
  process will be sent a SIGTERM signal and given time to gracefully exit.
  After the grace period expires the process will be hard terminated.

  For Consul agents on Windows, the child process is always hard terminated
  with a SIGKILL, since Windows has no POSIX compatible notion for SIGTERM.

  When -n=1, only a single lock holder or leader exists providing mutual
  exclusion. Setting a higher value switches to a semaphore allowing multiple
  holders to coordinate.

  The prefix provided must have write privileges.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *LockCommand) Run(args []string) int {
	var lu *LockUnlock
	return c.run(args, &lu)
}

func (c *LockCommand) run(args []string, lu **LockUnlock) int {
	var childDone chan struct{}
	var limit int
	var monitorRetry int
	var name string
	var passStdin bool
	var timeout time.Duration

	f := c.Command.NewFlagSet(c)
	f.IntVar(&limit, "n", 1,
		"Optional limit on the number of concurrent lock holders. The underlying "+
			"implementation switches from a lock to a semaphore when the value is "+
			"greater than 1. The default value is 1.")
	f.IntVar(&monitorRetry, "monitor-retry", defaultMonitorRetry,
		"Number of times to retry if Consul returns a 500 error while monitoring "+
			"the lock. This allows riding out brief periods of unavailability "+
			"without causing leader elections, but increases the amount of time "+
			"required to detect a lost lock in some cases. The default value is 3, "+
			"with a 1s wait between retries. Set this value to 0 to disable retires.")
	f.StringVar(&name, "name", "",
		"Optional name to associate with the lock session. It not provided, one "+
			"is generated based on the provided child command.")
	f.BoolVar(&passStdin, "pass-stdin", false,
		"Pass stdin to the child process.")
	f.DurationVar(&timeout, "timeout", 0,
		"Maximum amount of time to wait to acquire the lock, specified as a "+
			"timestamp like \"1s\" or \"3h\". The default value is 0.")
	f.BoolVar(&c.verbose, "verbose", false,
		"Enable verbose (debugging) output.")

	// Deprecations
	f.DurationVar(&timeout, "try", 0,
		"DEPRECATED. Use -timeout instead.")

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	// Check the limit
	if limit <= 0 {
		c.Ui.Error(fmt.Sprintf("Lock holder limit must be positive"))
		return 1
	}

	// Verify the prefix and child are provided
	extra := f.Args()
	if len(extra) < 2 {
		c.Ui.Error("Key prefix and child command must be specified")
		return 1
	}
	prefix := extra[0]
	prefix = strings.TrimPrefix(prefix, "/")
	script := strings.Join(extra[1:], " ")

	if timeout < 0 {
		c.Ui.Error("Timeout must be positive")
		return 1
	}

	// Calculate a session name if none provided
	if name == "" {
		name = fmt.Sprintf("Consul lock for '%s' at '%s'", script, prefix)
	}

	// Calculate oneshot
	oneshot := timeout > 0

	// Check the retry parameter
	if monitorRetry < 0 {
		c.Ui.Error("Number for 'monitor-retry' must be >= 0")
		return 1
	}

	// Create and test the HTTP client
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}
	_, err = client.Agent().NodeName()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
		return 1
	}

	// Setup the lock or semaphore
	if limit == 1 {
		*lu, err = c.setupLock(client, prefix, name, oneshot, timeout, monitorRetry)
	} else {
		*lu, err = c.setupSemaphore(client, limit, prefix, name, oneshot, timeout, monitorRetry)
	}
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Lock setup failed: %s", err))
		return 1
	}

	// Attempt the acquisition
	if c.verbose {
		c.Ui.Info("Attempting lock acquisition")
	}
	lockCh, err := (*lu).lockFn(c.ShutdownCh)
	if lockCh == nil {
		if err == nil {
			c.Ui.Error("Shutdown triggered or timeout during lock acquisition")
		} else {
			c.Ui.Error(fmt.Sprintf("Lock acquisition failed: %s", err))
		}
		return 1
	}

	// Check if we were shutdown but managed to still acquire the lock
	select {
	case <-c.ShutdownCh:
		c.Ui.Error("Shutdown triggered during lock acquisition")
		goto RELEASE
	default:
	}

	// Start the child process
	childDone = make(chan struct{})
	go func() {
		if err := c.startChild(script, childDone, passStdin); err != nil {
			c.Ui.Error(fmt.Sprintf("%s", err))
		}
	}()

	// Monitor for shutdown, child termination, or lock loss
	select {
	case <-c.ShutdownCh:
		if c.verbose {
			c.Ui.Info("Shutdown triggered, killing child")
		}
	case <-lockCh:
		if c.verbose {
			c.Ui.Info("Lock lost, killing child")
		}
	case <-childDone:
		if c.verbose {
			c.Ui.Info("Child terminated, releasing lock")
		}
		goto RELEASE
	}

	// Prevent starting a new child.  The lock is never released
	// after this point.
	c.childLock.Lock()
	// Kill any existing child
	if err := c.killChild(childDone); err != nil {
		c.Ui.Error(fmt.Sprintf("%s", err))
	}

RELEASE:
	// Release the lock before termination
	if err := (*lu).unlockFn(); err != nil {
		c.Ui.Error(fmt.Sprintf("Lock release failed: %s", err))
		return 1
	}

	// Cleanup the lock if no longer in use
	if err := (*lu).cleanupFn(); err != nil {
		if err != (*lu).inUseErr {
			c.Ui.Error(fmt.Sprintf("Lock cleanup failed: %s", err))
			return 1
		} else if c.verbose {
			c.Ui.Info("Cleanup aborted, lock in use")
		}
	} else if c.verbose {
		c.Ui.Info("Cleanup succeeded")
	}
	return 0
}

// setupLock is used to setup a new Lock given the API client, the key prefix to
// operate on, and an optional session name. If oneshot is true then we will set
// up for a single attempt at acquisition, using the given wait time. The retry
// parameter sets how many 500 errors the lock monitor will tolerate before
// giving up the lock.
func (c *LockCommand) setupLock(client *api.Client, prefix, name string,
	oneshot bool, wait time.Duration, retry int) (*LockUnlock, error) {
	// Use the DefaultSemaphoreKey extension, this way if a lock and
	// semaphore are both used at the same prefix, we will get a conflict
	// which we can report to the user.
	key := path.Join(prefix, api.DefaultSemaphoreKey)
	if c.verbose {
		c.Ui.Info(fmt.Sprintf("Setting up lock at path: %s", key))
	}
	opts := api.LockOptions{
		Key:              key,
		SessionName:      name,
		MonitorRetries:   retry,
		MonitorRetryTime: defaultMonitorRetryTime,
	}
	if oneshot {
		opts.LockTryOnce = true
		opts.LockWaitTime = wait
	}
	l, err := client.LockOpts(&opts)
	if err != nil {
		return nil, err
	}
	lu := &LockUnlock{
		lockFn:    l.Lock,
		unlockFn:  l.Unlock,
		cleanupFn: l.Destroy,
		inUseErr:  api.ErrLockInUse,
		rawOpts:   &opts,
	}
	return lu, nil
}

// setupSemaphore is used to setup a new Semaphore given the API client, key
// prefix, session name, and slot holder limit. If oneshot is true then we will
// set up for a single attempt at acquisition, using the given wait time. The
// retry parameter sets how many 500 errors the lock monitor will tolerate
// before giving up the semaphore.
func (c *LockCommand) setupSemaphore(client *api.Client, limit int, prefix, name string,
	oneshot bool, wait time.Duration, retry int) (*LockUnlock, error) {
	if c.verbose {
		c.Ui.Info(fmt.Sprintf("Setting up semaphore (limit %d) at prefix: %s", limit, prefix))
	}
	opts := api.SemaphoreOptions{
		Prefix:           prefix,
		Limit:            limit,
		SessionName:      name,
		MonitorRetries:   retry,
		MonitorRetryTime: defaultMonitorRetryTime,
	}
	if oneshot {
		opts.SemaphoreTryOnce = true
		opts.SemaphoreWaitTime = wait
	}
	s, err := client.SemaphoreOpts(&opts)
	if err != nil {
		return nil, err
	}
	lu := &LockUnlock{
		lockFn:    s.Acquire,
		unlockFn:  s.Release,
		cleanupFn: s.Destroy,
		inUseErr:  api.ErrSemaphoreInUse,
		rawOpts:   &opts,
	}
	return lu, nil
}

// startChild is a long running routine used to start and
// wait for the child process to exit.
func (c *LockCommand) startChild(script string, doneCh chan struct{}, passStdin bool) error {
	defer close(doneCh)
	if c.verbose {
		c.Ui.Info(fmt.Sprintf("Starting handler '%s'", script))
	}
	// Create the command
	cmd, err := agent.ExecScript(script)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error executing handler: %s", err))
		return err
	}

	// Setup the command streams
	cmd.Env = append(os.Environ(),
		"CONSUL_LOCK_HELD=true",
	)
	if passStdin {
		if c.verbose {
			c.Ui.Info("Stdin passed to handler process")
		}
		cmd.Stdin = os.Stdin
	} else {
		cmd.Stdin = nil
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the child process
	c.childLock.Lock()
	if err := cmd.Start(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error starting handler: %s", err))
		c.childLock.Unlock()
		return err
	}

	// Setup the child info
	c.child = cmd.Process
	c.childLock.Unlock()

	// Wait for the child process
	if err := cmd.Wait(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error running handler: %s", err))
		return err
	}
	return nil
}

// killChild is used to forcefully kill the child, first using SIGTERM
// to allow for a graceful cleanup and then using SIGKILL for a hard
// termination.
// On Windows, the child is always hard terminated with a SIGKILL, even
// on the first attempt.
func (c *LockCommand) killChild(childDone chan struct{}) error {
	// Get the child process
	child := c.child

	// If there is no child process (failed to start), we can quit early
	if child == nil {
		if c.verbose {
			c.Ui.Info("No child process to kill")
		}
		return nil
	}

	// Attempt termination first
	if c.verbose {
		c.Ui.Info(fmt.Sprintf("Terminating child pid %d", child.Pid))
	}
	if err := signalPid(child.Pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("Failed to terminate %d: %v", child.Pid, err)
	}

	// Wait for termination, or until a timeout
	select {
	case <-childDone:
		if c.verbose {
			c.Ui.Info("Child terminated")
		}
		return nil
	case <-time.After(lockKillGracePeriod):
		if c.verbose {
			c.Ui.Info(fmt.Sprintf("Child did not exit after grace period of %v",
				lockKillGracePeriod))
		}
	}

	// Send a final SIGKILL
	if c.verbose {
		c.Ui.Info(fmt.Sprintf("Killing child pid %d", child.Pid))
	}
	if err := signalPid(child.Pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("Failed to kill %d: %v", child.Pid, err)
	}
	return nil
}

func (c *LockCommand) Synopsis() string {
	return "Execute a command holding a lock"
}

// LockUnlock is used to abstract over the differences between
// a lock and a semaphore.
type LockUnlock struct {
	lockFn    func(<-chan struct{}) (<-chan struct{}, error)
	unlockFn  func() error
	cleanupFn func() error
	inUseErr  error
	rawOpts   interface{}
}
