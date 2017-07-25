package service

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/idresolver"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/stringid"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type logsOptions struct {
	noResolve  bool
	noTrunc    bool
	noTaskIDs  bool
	follow     bool
	since      string
	timestamps bool
	tail       string
	details    bool
	raw        bool

	target string
}

func newLogsCommand(dockerCli *command.DockerCli) *cobra.Command {
	var opts logsOptions

	cmd := &cobra.Command{
		Use:   "logs [OPTIONS] SERVICE|TASK",
		Short: "Fetch the logs of a service or task",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.target = args[0]
			return runLogs(dockerCli, &opts)
		},
		Tags: map[string]string{"version": "1.29"},
	}

	flags := cmd.Flags()
	// options specific to service logs
	flags.BoolVar(&opts.noResolve, "no-resolve", false, "Do not map IDs to Names in output")
	flags.BoolVar(&opts.noTrunc, "no-trunc", false, "Do not truncate output")
	flags.BoolVar(&opts.raw, "raw", false, "Do not neatly format logs")
	flags.SetAnnotation("raw", "version", []string{"1.30"})
	flags.BoolVar(&opts.noTaskIDs, "no-task-ids", false, "Do not include task IDs in output")
	// options identical to container logs
	flags.BoolVarP(&opts.follow, "follow", "f", false, "Follow log output")
	flags.StringVar(&opts.since, "since", "", "Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)")
	flags.BoolVarP(&opts.timestamps, "timestamps", "t", false, "Show timestamps")
	flags.BoolVar(&opts.details, "details", false, "Show extra details provided to logs")
	flags.SetAnnotation("details", "version", []string{"1.30"})
	flags.StringVar(&opts.tail, "tail", "all", "Number of lines to show from the end of the logs")
	return cmd
}

func runLogs(dockerCli *command.DockerCli, opts *logsOptions) error {
	ctx := context.Background()

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      opts.since,
		Timestamps: opts.timestamps,
		Follow:     opts.follow,
		Tail:       opts.tail,
		// get the details if we request it OR if we're not doing raw mode
		// (we need them for the context to pretty print)
		Details: opts.details || !opts.raw,
	}

	cli := dockerCli.Client()

	var (
		maxLength    = 1
		responseBody io.ReadCloser
		tty          bool
		// logfunc is used to delay the call to logs so that we can do some
		// processing before we actually get the logs
		logfunc func(context.Context, string, types.ContainerLogsOptions) (io.ReadCloser, error)
	)

	service, _, err := cli.ServiceInspectWithRaw(ctx, opts.target, types.ServiceInspectOptions{})
	if err != nil {
		// if it's any error other than service not found, it's Real
		if !client.IsErrServiceNotFound(err) {
			return err
		}
		task, _, err := cli.TaskInspectWithRaw(ctx, opts.target)
		if err != nil {
			if client.IsErrTaskNotFound(err) {
				// if the task isn't found, rewrite the error to be clear
				// that we looked for services AND tasks and found none
				err = fmt.Errorf("no such task or service: %v", opts.target)
			}
			return err
		}

		tty = task.Spec.ContainerSpec.TTY
		maxLength = getMaxLength(task.Slot)

		// use the TaskLogs api function
		logfunc = cli.TaskLogs
	} else {
		// use ServiceLogs api function
		logfunc = cli.ServiceLogs
		tty = service.Spec.TaskTemplate.ContainerSpec.TTY
		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			// if replicas are initialized, figure out if we need to pad them
			replicas := *service.Spec.Mode.Replicated.Replicas
			maxLength = getMaxLength(int(replicas))
		}
	}

	// we can't prettify tty logs. tell the user that this is the case.
	// this is why we assign the logs function to a variable and delay calling
	// it. we want to check this before we make the call and checking twice in
	// each branch is even sloppier than this CLI disaster already is
	if tty && !opts.raw {
		return errors.New("tty service logs only supported with --raw")
	}

	// now get the logs
	responseBody, err = logfunc(ctx, opts.target, options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	// tty logs get straight copied. they're not muxed with stdcopy
	if tty {
		_, err = io.Copy(dockerCli.Out(), responseBody)
		return err
	}

	// otherwise, logs are multiplexed. if we're doing pretty printing, also
	// create a task formatter.
	var stdout, stderr io.Writer
	stdout = dockerCli.Out()
	stderr = dockerCli.Err()
	if !opts.raw {
		taskFormatter := newTaskFormatter(cli, opts, maxLength)

		stdout = &logWriter{ctx: ctx, opts: opts, f: taskFormatter, w: stdout}
		stderr = &logWriter{ctx: ctx, opts: opts, f: taskFormatter, w: stderr}
	}

	_, err = stdcopy.StdCopy(stdout, stderr, responseBody)
	return err
}

// getMaxLength gets the maximum length of the number in base 10
func getMaxLength(i int) int {
	return len(strconv.FormatInt(int64(i), 10))
}

type taskFormatter struct {
	client  client.APIClient
	opts    *logsOptions
	padding int

	r *idresolver.IDResolver
	// cache saves a pre-cooked logContext formatted string based on a
	// logcontext object, so we don't have to resolve names every time
	cache map[logContext]string
}

func newTaskFormatter(client client.APIClient, opts *logsOptions, padding int) *taskFormatter {
	return &taskFormatter{
		client:  client,
		opts:    opts,
		padding: padding,
		r:       idresolver.New(client, opts.noResolve),
		cache:   make(map[logContext]string),
	}
}

func (f *taskFormatter) format(ctx context.Context, logCtx logContext) (string, error) {
	if cached, ok := f.cache[logCtx]; ok {
		return cached, nil
	}

	nodeName, err := f.r.Resolve(ctx, swarm.Node{}, logCtx.nodeID)
	if err != nil {
		return "", err
	}

	serviceName, err := f.r.Resolve(ctx, swarm.Service{}, logCtx.serviceID)
	if err != nil {
		return "", err
	}

	task, _, err := f.client.TaskInspectWithRaw(ctx, logCtx.taskID)
	if err != nil {
		return "", err
	}

	taskName := fmt.Sprintf("%s.%d", serviceName, task.Slot)
	if !f.opts.noTaskIDs {
		if f.opts.noTrunc {
			taskName += fmt.Sprintf(".%s", task.ID)
		} else {
			taskName += fmt.Sprintf(".%s", stringid.TruncateID(task.ID))
		}
	}

	paddingCount := f.padding - getMaxLength(task.Slot)
	padding := ""
	if paddingCount > 0 {
		padding = strings.Repeat(" ", paddingCount)
	}
	formatted := taskName + "@" + nodeName + padding
	f.cache[logCtx] = formatted
	return formatted, nil
}

type logWriter struct {
	ctx  context.Context
	opts *logsOptions
	f    *taskFormatter
	w    io.Writer
}

func (lw *logWriter) Write(buf []byte) (int, error) {
	// this works but ONLY because stdcopy calls write a whole line at a time.
	// if this ends up horribly broken or panics, check to see if stdcopy has
	// reneged on that asssumption. (@god forgive me)
	// also this only works because the logs format is, like, barely parsable.
	// if something changes in the logs format, this is gonna break

	// there should always be at least 2 parts: details and message. if there
	// is no timestamp, details will be first (index 0) when we split on
	// spaces. if there is a timestamp, details will be 2nd (`index 1)
	detailsIndex := 0
	numParts := 2
	if lw.opts.timestamps {
		detailsIndex++
		numParts++
	}

	// break up the log line into parts.
	parts := bytes.SplitN(buf, []byte(" "), numParts)
	if len(parts) != numParts {
		return 0, errors.Errorf("invalid context in log message: %v", string(buf))
	}
	// parse the details out
	details, err := client.ParseLogDetails(string(parts[detailsIndex]))
	if err != nil {
		return 0, err
	}
	// and then create a context from the details
	// this removes the context-specific details from the details map, so we
	// can more easily print the details later
	logCtx, err := lw.parseContext(details)
	if err != nil {
		return 0, err
	}

	output := []byte{}
	// if we included timestamps, add them to the front
	if lw.opts.timestamps {
		output = append(output, parts[0]...)
		output = append(output, ' ')
	}
	// add the context, nice and formatted
	formatted, err := lw.f.format(lw.ctx, logCtx)
	if err != nil {
		return 0, err
	}
	output = append(output, []byte(formatted+"    | ")...)
	// if the user asked for details, add them to be log message
	if lw.opts.details {
		// ugh i hate this it's basically a dupe of api/server/httputils/write_log_stream.go:stringAttrs()
		// ok but we're gonna do it a bit different

		// there are optimizations that can be made here. for starters, i'd
		// suggest caching the details keys. then, we can maybe draw maps and
		// slices from a pool to avoid alloc overhead on them. idk if it's
		// worth the time yet.

		// first we need a slice
		d := make([]string, 0, len(details))
		// then let's add all the pairs
		for k := range details {
			d = append(d, k+"="+details[k])
		}
		// then sort em
		sort.Strings(d)
		// then join and append
		output = append(output, []byte(strings.Join(d, ","))...)
		output = append(output, ' ')
	}

	// add the log message itself, finally
	output = append(output, parts[detailsIndex+1]...)

	_, err = lw.w.Write(output)
	if err != nil {
		return 0, err
	}

	return len(buf), nil
}

// parseContext returns a log context and REMOVES the context from the details map
func (lw *logWriter) parseContext(details map[string]string) (logContext, error) {
	nodeID, ok := details["com.docker.swarm.node.id"]
	if !ok {
		return logContext{}, errors.Errorf("missing node id in details: %v", details)
	}
	delete(details, "com.docker.swarm.node.id")

	serviceID, ok := details["com.docker.swarm.service.id"]
	if !ok {
		return logContext{}, errors.Errorf("missing service id in details: %v", details)
	}
	delete(details, "com.docker.swarm.service.id")

	taskID, ok := details["com.docker.swarm.task.id"]
	if !ok {
		return logContext{}, errors.Errorf("missing task id in details: %s", details)
	}
	delete(details, "com.docker.swarm.task.id")

	return logContext{
		nodeID:    nodeID,
		serviceID: serviceID,
		taskID:    taskID,
	}, nil
}

type logContext struct {
	nodeID    string
	serviceID string
	taskID    string
}
