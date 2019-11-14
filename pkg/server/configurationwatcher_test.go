package server

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/safe"
	th "github.com/containous/traefik/v2/pkg/testhelpers"
	"github.com/containous/traefik/v2/pkg/tls"
	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	messages []dynamic.Message
	wait     time.Duration
}

func (p *mockProvider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	for _, message := range p.messages {
		configurationChan <- message

		wait := p.wait
		if wait == 0 {
			wait = 20 * time.Millisecond
		}

		fmt.Println("wait", wait, time.Now().Nanosecond())
		time.Sleep(wait)
	}

	return nil
}

func (p *mockProvider) Init() error {
	panic("implement me")
}

func TestNewConfigurationWatcher(t *testing.T) {
	routinesPool := safe.NewPool(context.Background())
	pvd := &mockProvider{
		messages: []dynamic.Message{{
			ProviderName: "mock",
			Configuration: &dynamic.Configuration{
				HTTP: th.BuildConfiguration(
					th.WithRouters(
						th.WithRouter("test",
							th.WithEntryPoints("e"),
							th.WithServiceName("scv"))),
				),
			},
		}},
	}

	watcher := NewConfigurationWatcher(routinesPool, pvd, time.Second)

	run := make(chan struct{})

	watcher.AddListener(func(conf dynamic.Configuration) {
		expected := dynamic.Configuration{
			HTTP: th.BuildConfiguration(
				th.WithRouters(
					th.WithRouter("test@mock",
						th.WithEntryPoints("e"),
						th.WithServiceName("scv"))),
				th.WithMiddlewares(),
				th.WithLoadBalancerServices(),
			),
			TCP: &dynamic.TCPConfiguration{
				Routers:  map[string]*dynamic.TCPRouter{},
				Services: map[string]*dynamic.TCPService{},
			},
			TLS: &dynamic.TLSConfiguration{
				Options: map[string]tls.Options{
					"default": {},
				},
				Stores: map[string]tls.Store{},
			},
		}

		assert.Equal(t, expected, conf)
		close(run)
	})

	watcher.Start()
	<-run
}

func TestListenProvidersThrottleProviderConfigReload(t *testing.T) {
	routinesPool := safe.NewPool(context.Background())

	pvd := &mockProvider{
		wait: 10 * time.Millisecond,
	}

	for i := 0; i < 5; i++ {
		pvd.messages = append(pvd.messages, dynamic.Message{
			ProviderName: "mock",
			Configuration: &dynamic.Configuration{
				HTTP: th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo"+strconv.Itoa(i))),
					th.WithLoadBalancerServices(th.WithService("bar")),
				),
			},
		})
	}

	watcher := NewConfigurationWatcher(routinesPool, pvd, 30*time.Millisecond)

	publishedConfigCount := 0
	watcher.AddListener(func(_ dynamic.Configuration) {
		publishedConfigCount++
	})

	watcher.Start()
	defer watcher.Stop()

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)

	// after 50 milliseconds 5 new configs were published
	// with a throttle duration of 30 milliseconds this means, we should have received 3 new configs
	assert.Equal(t, 3, publishedConfigCount, "times configs were published")
}

func TestListenProvidersSkipsEmptyConfigs(t *testing.T) {
	routinesPool := safe.NewPool(context.Background())
	pvd := &mockProvider{
		messages: []dynamic.Message{{ProviderName: "mock"}},
	}

	watcher := NewConfigurationWatcher(routinesPool, pvd, time.Second)
	watcher.AddListener(func(_ dynamic.Configuration) {
		t.Error("An empty configuration was published but it should not")
	})
	watcher.Start()
	defer watcher.Stop()

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)
}

func TestListenProvidersSkipsSameConfigurationForProvider(t *testing.T) {
	routinesPool := safe.NewPool(context.Background())
	message := dynamic.Message{
		ProviderName: "mock",
		Configuration: &dynamic.Configuration{
			HTTP: th.BuildConfiguration(
				th.WithRouters(th.WithRouter("foo")),
				th.WithLoadBalancerServices(th.WithService("bar")),
			),
		},
	}
	pvd := &mockProvider{
		messages: []dynamic.Message{message, message},
	}

	watcher := NewConfigurationWatcher(routinesPool, pvd, 0)

	alreadyCalled := false
	watcher.AddListener(func(_ dynamic.Configuration) {
		if alreadyCalled {
			t.Error("Same configuration should not be published multiple times")
		}
		alreadyCalled = true
	})

	watcher.Start()
	defer watcher.Stop()

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)
}

func TestListenProvidersPublishesConfigForEachProvider(t *testing.T) {
	routinesPool := safe.NewPool(context.Background())

	configuration := &dynamic.Configuration{
		HTTP: th.BuildConfiguration(
			th.WithRouters(th.WithRouter("foo")),
			th.WithLoadBalancerServices(th.WithService("bar")),
		),
	}

	pvd := &mockProvider{
		messages: []dynamic.Message{
			{ProviderName: "mock", Configuration: configuration},
			{ProviderName: "mock2", Configuration: configuration},
		},
	}

	watcher := NewConfigurationWatcher(routinesPool, pvd, 0)

	var publishedProviderConfig dynamic.Configuration

	watcher.AddListener(func(conf dynamic.Configuration) {
		publishedProviderConfig = conf
	})

	watcher.Start()
	defer watcher.Stop()

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)

	expected := dynamic.Configuration{
		HTTP: th.BuildConfiguration(
			th.WithRouters(th.WithRouter("foo@mock"), th.WithRouter("foo@mock2")),
			th.WithLoadBalancerServices(th.WithService("bar@mock"), th.WithService("bar@mock2")),
			th.WithMiddlewares(),
		),
		TCP: &dynamic.TCPConfiguration{
			Routers:  map[string]*dynamic.TCPRouter{},
			Services: map[string]*dynamic.TCPService{},
		},
		TLS: &dynamic.TLSConfiguration{
			Options: map[string]tls.Options{
				"default": {},
			},
			Stores: map[string]tls.Store{},
		},
	}

	assert.Equal(t, expected, publishedProviderConfig)
}
