package service

import (
	"fmt"
	"strconv"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/labels"
)

const format = "%s_%s_%d"

// Namer defines method to provide container name.
type Namer interface {
	Next() (string, int)
}

type defaultNamer struct {
	project       string
	service       string
	oneOff        bool
	currentNumber int
}

type singleNamer struct {
	name string
}

// NewSingleNamer returns a namer that only allows a single name.
func NewSingleNamer(name string) Namer {
	return &singleNamer{name}
}

// NewNamer returns a namer that returns names based on the specified project and
// service name and an inner counter, e.g. project_service_1, project_service_2…
func NewNamer(ctx context.Context, client client.ContainerAPIClient, project, service string, oneOff bool) (Namer, error) {
	namer := &defaultNamer{
		project: project,
		service: service,
		oneOff:  oneOff,
	}

	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("%s=%s", labels.PROJECT.Str(), project))
	filter.Add("label", fmt.Sprintf("%s=%s", labels.SERVICE.Str(), service))
	if oneOff {
		filter.Add("label", fmt.Sprintf("%s=%s", labels.ONEOFF.Str(), "True"))
	} else {
		filter.Add("label", fmt.Sprintf("%s=%s", labels.ONEOFF.Str(), "False"))
	}

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}

	maxNumber := 0
	for _, container := range containers {
		number, err := strconv.Atoi(container.Labels[labels.NUMBER.Str()])
		if err != nil {
			return nil, err
		}
		if number > maxNumber {
			maxNumber = number
		}
	}
	namer.currentNumber = maxNumber + 1

	return namer, nil
}

func (i *defaultNamer) Next() (string, int) {
	service := i.service
	if i.oneOff {
		service = i.service + "_run"
	}
	name := fmt.Sprintf(format, i.project, service, i.currentNumber)
	number := i.currentNumber
	i.currentNumber = i.currentNumber + 1
	return name, number
}

func (s *singleNamer) Next() (string, int) {
	return s.name, 1
}
