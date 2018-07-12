package docker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/containous/traefik/provider/docker/mocks"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	dockertypes "github.com/docker/docker/api/types"
	eventtypes "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
)

type testEventCallback struct {
	callCount int
	msgEvents []eventtypes.Message
}

func (tec *testEventCallback) callback(msgEvent eventtypes.Message, configurationChan chan<- types.ConfigMessage) error {
	tec.callCount++
	tec.msgEvents = append(tec.msgEvents, msgEvent)
	return nil
}

func TestTickerListener(t *testing.T) {
	mainCtx := context.Background()
	ctx, cancel := context.WithCancel(mainCtx)

	l := &tickerListener{
		ticker: time.NewTicker(500 * time.Millisecond),
	}

	c := &testEventCallback{
		callCount: 0,
		msgEvents: []eventtypes.Message{},
	}

	configurationChan := make(chan types.ConfigMessage)

	var listenRetVal error
	go func() {
		pool := safe.NewPool(ctx)
		listenRetVal = l.listen(pool, configurationChan, c.callback)
	}()

	time.Sleep(1200 * time.Millisecond)
	cancel()

	if listenRetVal != nil {
		t.Fatal("expected", nil, "got", listenRetVal)
	}

	if c.callCount != 2 {
		t.Fatal("expected", 2, "got", c.callCount)
	}
}

func TestStreamerListenerSuccessfulReturn(t *testing.T) {
	mainCtx := context.Background()
	ctx, cancel := context.WithCancel(mainCtx)
	pool := safe.NewPool(ctx)

	dockerClient := &mocks.APIClient{}

	eventsMsgChan := make(chan eventtypes.Message)
	eventsErrChan := make(chan error)
	eventsMsgChanReadOnly := func() <-chan eventtypes.Message {
		return eventsMsgChan
	}()
	eventsErrChanReadOnly := func() <-chan error {
		return eventsErrChan
	}()

	dockerClient.On(
		"Events",
		pool.Ctx(),
		dockertypes.EventsOptions{
			Filters: filters.NewArgs(
				filters.Arg("scope", "swarm"),
				filters.Arg("type", "service"),
			),
		},
	).Once().Return(eventsMsgChanReadOnly, eventsErrChanReadOnly)

	l := &streamerListener{
		dockerClient: dockerClient,
	}

	c := &testEventCallback{
		callCount: 0,
		msgEvents: []eventtypes.Message{},
	}

	configurationChan := make(chan types.ConfigMessage)

	var listenRetVal error
	go func() {
		listenRetVal = l.listen(pool, configurationChan, c.callback)
	}()

	msgEvents := []eventtypes.Message{
		{
			ID: "ASDF",
		},
		{
			ID: "QWERTY",
		},
	}

	for _, event := range msgEvents {
		eventsMsgChan <- event

		time.Sleep(100 * time.Millisecond)
	}

	cancel()

	if listenRetVal != nil {
		t.Fatal("expected", nil, "got", listenRetVal)
	}

	if c.callCount != 2 {
		t.Fatal("expected", 2, "got", c.callCount)
	}

	for i, event := range msgEvents {
		if event.ID != c.msgEvents[i].ID {
			t.Fatal("expected", event.ID, "got", c.msgEvents[i].ID)
		}
	}

	dockerClient.AssertExpectations(t)
	dockerClient.AssertNumberOfCalls(t, "Events", 1)
}

func TestStreamerListenerErrorReturn(t *testing.T) {
	mainCtx := context.Background()
	ctx, cancel := context.WithCancel(mainCtx)
	defer cancel()
	pool := safe.NewPool(ctx)

	dockerClient := &mocks.APIClient{}

	eventsMsgChan := make(chan eventtypes.Message)
	eventsErrChan := make(chan error)
	eventsMsgChanReadOnly := func() <-chan eventtypes.Message {
		return eventsMsgChan
	}()
	eventsErrChanReadOnly := func() <-chan error {
		return eventsErrChan
	}()

	dockerClient.On(
		"Events",
		pool.Ctx(),
		dockertypes.EventsOptions{
			Filters: filters.NewArgs(
				filters.Arg("scope", "swarm"),
				filters.Arg("type", "service"),
			),
		},
	).Once().Return(eventsMsgChanReadOnly, eventsErrChanReadOnly)

	l := &streamerListener{
		dockerClient: dockerClient,
	}

	c := &testEventCallback{
		callCount: 0,
		msgEvents: []eventtypes.Message{},
	}

	configurationChan := make(chan types.ConfigMessage)

	listenRetValChan := make(chan error)
	go func() {
		listenRetValChan <- l.listen(pool, configurationChan, c.callback)
	}()

	errEvent := fmt.Errorf("All your error are belong to us")
	eventsErrChan <- errEvent

	listenRetVal := <-listenRetValChan

	if listenRetVal == nil || listenRetVal != errEvent {
		t.Fatal("expected", errEvent, "got", listenRetVal)
	}

	dockerClient.AssertExpectations(t)
	dockerClient.AssertNumberOfCalls(t, "Events", 1)
}

func TestSwarmEventsCapabilitiesReturnTrue(t *testing.T) {
	ctx := context.Background()

	dockerClient := &mocks.APIClient{}

	serverVersion := dockertypes.Version{
		APIVersion: "1.29",
	}

	dockerClient.On(
		"ServerVersion",
		ctx,
	).Once().Return(serverVersion, nil)

	cap, err := swarmEventsCapabilities(ctx, dockerClient)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	if !cap {
		t.Fatal("expected", true, "got", cap)
	}

	dockerClient.AssertExpectations(t)
	dockerClient.AssertNumberOfCalls(t, "ServerVersion", 1)
}

func TestSwarmEventsCapabilitiesReturnFalse(t *testing.T) {
	ctx := context.Background()

	dockerClient := &mocks.APIClient{}

	serverVersion := dockertypes.Version{
		APIVersion: "1.28",
	}

	dockerClient.On(
		"ServerVersion",
		ctx,
	).Once().Return(serverVersion, nil)

	cap, err := swarmEventsCapabilities(ctx, dockerClient)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	if cap {
		t.Fatal("expected", false, "got", cap)
	}

	dockerClient.AssertExpectations(t)
	dockerClient.AssertNumberOfCalls(t, "ServerVersion", 1)
}

func TestSwarmEventsCapabilitiesReturnError(t *testing.T) {
	ctx := context.Background()

	dockerClient := &mocks.APIClient{}

	serverVersion := dockertypes.Version{}

	serverVersionErr := fmt.Errorf("All your error are belong to us")
	dockerClient.On(
		"ServerVersion",
		ctx,
	).Once().Return(serverVersion, serverVersionErr)

	cap, err := swarmEventsCapabilities(ctx, dockerClient)
	if err == nil {
		t.Fatal("expected", serverVersionErr, "got", nil)
	}

	if cap {
		t.Fatal("expected", false, "got", cap)
	}

	dockerClient.AssertExpectations(t)
	dockerClient.AssertNumberOfCalls(t, "ServerVersion", 1)
}
