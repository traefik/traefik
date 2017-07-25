package service

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/labels"
	"github.com/pkg/errors"
)

func TestSingleNamer(t *testing.T) {
	expectedName := "myName"
	expectedNumber := 1
	namer := NewSingleNamer("myName")
	for i := 0; i < 10; i++ {
		name, number := namer.Next()
		if name != expectedName {
			t.Fatalf("expected %s, got %s", expectedName, name)
		}
		if number != expectedNumber {
			t.Fatalf("expected %d, got %d", expectedNumber, number)
		}
	}
}

type NamerClient struct {
	client.Client
	err                  error
	expectedLabelFilters []string
	containers           []types.Container
}

func (client *NamerClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	if len(client.expectedLabelFilters) > 1 {
		labelFilters := options.Filters.Get("label")
		if len(labelFilters) != len(client.expectedLabelFilters) {
			return []types.Container{}, fmt.Errorf("expected filters %v, got %v", client.expectedLabelFilters, labelFilters)
		}
		for _, expectedLabelFilter := range client.expectedLabelFilters {
			found := false
			for _, labelFilter := range labelFilters {
				if labelFilter == expectedLabelFilter {
					found = true
					break
				}
			}
			if !found {
				return []types.Container{}, fmt.Errorf("expected to find filter %s, did not in %v", expectedLabelFilter, labelFilters)
			}
		}
	}
	return client.containers, client.err
}

func TestDefaultNamerClientError(t *testing.T) {
	client := &NamerClient{
		err: errors.New("Engine no longer exists"),
	}
	_, err := NewNamer(context.Background(), client, "project", "service", false)
	if err == nil || err.Error() != "Engine no longer exists" {
		t.Fatalf("expected an error 'Engine no longer exists', got %s", err)
	}
}

func TestDefaultNamerLabelNotANumber(t *testing.T) {
	client := &NamerClient{
		containers: []types.Container{
			{
				Labels: map[string]string{
					labels.ONEOFF.Str(): "IAmAString",
				},
			},
		},
	}
	_, err := NewNamer(context.Background(), client, "project", "service", false)
	if err == nil {
		t.Fatal("expected an error, got nothing")
	}
}

func TestDefaultNamer(t *testing.T) {
	cases := []struct {
		projectName    string
		serviceName    string
		oneOff         bool
		containers     []types.Container
		expectedLabels []string
		expectedName   string
		expectedNumber int
	}{
		{
			projectName: "",
			serviceName: "",
			oneOff:      false,
			containers:  []types.Container{},
			expectedLabels: []string{
				fmt.Sprintf("%s=", labels.PROJECT.Str()),
				fmt.Sprintf("%s=", labels.SERVICE.Str()),
				fmt.Sprintf("%s=False", labels.ONEOFF.Str()),
			},
			expectedName:   "__1",
			expectedNumber: 1,
		},
		{
			projectName: "project",
			serviceName: "service",
			oneOff:      false,
			containers:  []types.Container{},
			expectedLabels: []string{
				fmt.Sprintf("%s=project", labels.PROJECT.Str()),
				fmt.Sprintf("%s=service", labels.SERVICE.Str()),
				fmt.Sprintf("%s=False", labels.ONEOFF.Str()),
			},
			expectedName:   "project_service_1",
			expectedNumber: 1,
		},
		{
			projectName: "project",
			serviceName: "service",
			oneOff:      false,
			containers: []types.Container{
				{
					Labels: map[string]string{
						labels.NUMBER.Str(): "1",
					},
				},
			},
			expectedLabels: []string{
				fmt.Sprintf("%s=project", labels.PROJECT.Str()),
				fmt.Sprintf("%s=service", labels.SERVICE.Str()),
				fmt.Sprintf("%s=False", labels.ONEOFF.Str()),
			},
			expectedName:   "project_service_2",
			expectedNumber: 2,
		},
		{
			projectName: "project",
			serviceName: "anotherservice",
			oneOff:      false,
			containers: []types.Container{
				{
					Labels: map[string]string{
						labels.NUMBER.Str(): "10",
					},
				},
			},
			expectedLabels: []string{
				fmt.Sprintf("%s=project", labels.PROJECT.Str()),
				fmt.Sprintf("%s=anotherservice", labels.SERVICE.Str()),
				fmt.Sprintf("%s=False", labels.ONEOFF.Str()),
			},
			expectedName:   "project_anotherservice_11",
			expectedNumber: 11,
		},
	}

	for _, c := range cases {
		client := &NamerClient{
			expectedLabelFilters: c.expectedLabels,
			containers:           c.containers,
		}
		namer, err := NewNamer(context.Background(), client, c.projectName, c.serviceName, c.oneOff)
		if err != nil {
			t.Error(err)
		}
		name, number := namer.Next()
		if name != c.expectedName {
			t.Errorf("Expected %s, got %s for %v", c.expectedName, name, c)
		}
		if number != c.expectedNumber {
			t.Errorf("Expected %d, got %d for %v", c.expectedNumber, number, c)
		}
		_, number = namer.Next()
		if number != c.expectedNumber+1 {
			t.Errorf("Expected a 2nd call to increment numbre to %d, got %d for %v", c.expectedNumber+1, number, c)
		}
	}
}
