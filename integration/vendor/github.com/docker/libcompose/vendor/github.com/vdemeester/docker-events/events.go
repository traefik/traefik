package events

import (
	"encoding/json"
	"io"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	eventtypes "github.com/docker/engine-api/types/events"
)

// Monitor subscribes to the docker events api using engine api and will execute the
// specified function on each message.
// It will pass the specified options to the underline method (i.e Events).
func Monitor(ctx context.Context, cli client.SystemAPIClient, options types.EventsOptions, fun func(m eventtypes.Message)) chan error {
	handler := NewHandler(func(_ eventtypes.Message) string {
		// Let's return always the same thing to not filter at all
		return ""
	})
	handler.Handle("", fun)

	return MonitorWithHandler(ctx, cli, options, handler)
}

// MonitorWithHandler subscribes to the docker events api using engine api and will pass the message
// to the specified Handler, that will take care of it.
// It will pass the specified options to the underline method (i.e Events).
func MonitorWithHandler(ctx context.Context, cli client.SystemAPIClient, options types.EventsOptions, handler *Handler) chan error {
	eventChan := make(chan eventtypes.Message)
	errChan := make(chan error)
	started := make(chan struct{})

	go handler.Watch(eventChan)
	go monitorEvents(ctx, cli, options, started, eventChan, errChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				// close(eventChan)
				errChan <- nil
			}
		}
	}()

	<-started
	return errChan
}

func monitorEvents(ctx context.Context, cli client.SystemAPIClient, options types.EventsOptions, started chan struct{}, eventChan chan eventtypes.Message, errChan chan error) {
	body, err := cli.Events(ctx, options)
	// Whether we successfully subscribed to events or not, we can now
	// unblock the main goroutine.
	close(started)
	if err != nil {
		errChan <- err
		return
	}
	defer body.Close()

	if err := decodeEvents(body, func(event eventtypes.Message, err error) error {
		if err != nil {
			return err
		}
		eventChan <- event
		return nil
	}); err != nil {
		errChan <- err
		return
	}
}

type eventProcessor func(event eventtypes.Message, err error) error

func decodeEvents(input io.Reader, ep eventProcessor) error {
	dec := json.NewDecoder(input)
	for {
		var event eventtypes.Message
		err := dec.Decode(&event)
		if err != nil && err == io.EOF {
			break
		}

		if procErr := ep(event, err); procErr != nil {
			return procErr
		}
	}
	return nil
}
