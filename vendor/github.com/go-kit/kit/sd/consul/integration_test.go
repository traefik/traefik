// +build integration

package consul

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	stdconsul "github.com/hashicorp/consul/api"
)

func TestIntegration(t *testing.T) {
	// Connect to Consul.
	// docker run -p 8500:8500 progrium/consul -server -bootstrap
	consulAddr := os.Getenv("CONSUL_ADDRESS")
	if consulAddr == "" {
		t.Fatal("CONSUL_ADDRESS is not set")
	}
	stdClient, err := stdconsul.NewClient(&stdconsul.Config{
		Address: consulAddr,
	})
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(stdClient)
	logger := log.NewLogfmtLogger(os.Stderr)

	// Produce a fake service registration.
	r := &stdconsul.AgentServiceRegistration{
		ID:                "my-service-ID",
		Name:              "my-service-name",
		Tags:              []string{"alpha", "beta"},
		Port:              12345,
		Address:           "my-address",
		EnableTagOverride: false,
		// skipping check(s)
	}

	// Build a subscriber on r.Name + r.Tags.
	factory := func(instance string) (endpoint.Endpoint, io.Closer, error) {
		t.Logf("factory invoked for %q", instance)
		return endpoint.Nop, nil, nil
	}
	subscriber := NewSubscriber(
		client,
		factory,
		log.NewContext(logger).With("component", "subscriber"),
		r.Name,
		r.Tags,
		true,
	)

	time.Sleep(time.Second)

	// Before we publish, we should have no endpoints.
	endpoints, err := subscriber.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Build a registrar for r.
	registrar := NewRegistrar(client, r, log.NewContext(logger).With("component", "registrar"))
	registrar.Register()
	defer registrar.Deregister()

	time.Sleep(time.Second)

	// Now we should have one active endpoints.
	endpoints, err = subscriber.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
