package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	"github.com/urfave/cli"
)

// ProjectEvents listen for real-time events of containers.
func ProjectEvents(p project.APIProject, c *cli.Context) error {
	evts, err := p.Events(context.Background(), c.Args()...)
	if err != nil {
		return err
	}
	var printfn func(events.ContainerEvent)

	if c.Bool("json") {
		printfn = printJSON
	} else {
		printfn = printStd
	}
	for event := range evts {
		printfn(event)
	}
	return nil
}

func printStd(event events.ContainerEvent) {
	output := os.Stdout
	fmt.Fprintf(output, "%s ", event.Time.Format("2006-01-02 15:04:05.999999999"))
	fmt.Fprintf(output, "%s %s %s", event.Type, event.Event, event.ID)
	attrs := []string{}
	for attr, value := range event.Attributes {
		attrs = append(attrs, fmt.Sprintf("%s=%s", attr, value))
	}

	fmt.Fprintf(output, " (%s)", strings.Join(attrs, ", "))
	fmt.Fprint(output, "\n")
}

func printJSON(event events.ContainerEvent) {
	json, err := json.Marshal(event)
	if err != nil {
		logrus.Warn(err)
	}
	output := os.Stdout
	fmt.Fprintf(output, "%s", json)
}
