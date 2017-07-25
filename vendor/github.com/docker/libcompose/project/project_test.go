package project

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project/options"
	"github.com/docker/libcompose/yaml"
	"github.com/stretchr/testify/assert"
)

type TestServiceFactory struct {
	Counts map[string]int
}

type TestService struct {
	factory *TestServiceFactory
	name    string
	config  *config.ServiceConfig
	EmptyService
	Count int
}

func (t *TestService) Config() *config.ServiceConfig {
	return t.config
}

func (t *TestService) Name() string {
	return t.name
}

func (t *TestService) Run(ctx context.Context, commandParts []string, opts options.Run) (int, error) {
	return 0, nil
}

func (t *TestService) Create(ctx context.Context, options options.Create) error {
	key := t.name + ".create"
	t.factory.Counts[key] = t.factory.Counts[key] + 1
	return nil
}

func (t *TestService) DependentServices() []ServiceRelationship {
	return nil
}

func (t *TestServiceFactory) Create(project *Project, name string, serviceConfig *config.ServiceConfig) (Service, error) {
	return &TestService{
		factory: t,
		config:  serviceConfig,
		name:    name,
	}, nil
}

func TestTwoCall(t *testing.T) {
	factory := &TestServiceFactory{
		Counts: map[string]int{},
	}

	p := NewProject(&Context{
		ServiceFactory: factory,
	}, nil, nil)
	p.ServiceConfigs = config.NewServiceConfigs()
	p.ServiceConfigs.Add("foo", &config.ServiceConfig{})

	if err := p.Create(context.Background(), options.Create{}, "foo"); err != nil {
		t.Fatal(err)
	}

	if err := p.Create(context.Background(), options.Create{}, "foo"); err != nil {
		t.Fatal(err)
	}

	if factory.Counts["foo.create"] != 2 {
		t.Fatal("Failed to create twice")
	}
}

func TestGetServiceConfig(t *testing.T) {

	p := NewProject(&Context{}, nil, nil)
	p.ServiceConfigs = config.NewServiceConfigs()
	fooService := &config.ServiceConfig{}
	p.ServiceConfigs.Add("foo", fooService)

	config, ok := p.GetServiceConfig("foo")
	if !ok {
		t.Fatal("Foo service not found")
	}

	if config != fooService {
		t.Fatal("Incorrect Service Config returned")
	}

	config, ok = p.GetServiceConfig("unknown")
	if ok {
		t.Fatal("Found service incorrectly")
	}

	if config != nil {
		t.Fatal("Incorrect Service Config returned")
	}
}

func TestParseWithBadContent(t *testing.T) {
	p := NewProject(&Context{
		ComposeBytes: [][]byte{
			[]byte("garbage"),
		},
	}, nil, nil)

	err := p.Parse()
	if err == nil {
		t.Fatal("Should have failed parse")
	}

	if !strings.Contains(err.Error(), "cannot unmarshal !!str `garbage` into config.Config") {
		t.Fatalf("Should have failed parse: %#v", err)
	}
}

func TestParseWithGoodContent(t *testing.T) {
	p := NewProject(&Context{
		ComposeBytes: [][]byte{
			[]byte("not-garbage:\n  image: foo"),
		},
	}, nil, nil)

	err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseWithDefaultEnvironmentLookup(t *testing.T) {
	p := NewProject(&Context{
		ComposeBytes: [][]byte{
			[]byte("not-garbage:\n  image: foo:${version}"),
		},
	}, nil, nil)

	err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
}

type TestEnvironmentLookup struct {
}

func (t *TestEnvironmentLookup) Lookup(key string, config *config.ServiceConfig) []string {
	return []string{fmt.Sprintf("%s=X", key)}
}

func TestEnvironmentResolve(t *testing.T) {
	factory := &TestServiceFactory{
		Counts: map[string]int{},
	}

	p := NewProject(&Context{
		ServiceFactory:    factory,
		EnvironmentLookup: &TestEnvironmentLookup{},
	}, nil, nil)
	p.ServiceConfigs = config.NewServiceConfigs()
	p.ServiceConfigs.Add("foo", &config.ServiceConfig{
		Environment: yaml.MaporEqualSlice([]string{
			"A",
			"A=",
			"A=B",
		}),
	})

	service, err := p.CreateService("foo")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(service.Config().Environment, yaml.MaporEqualSlice{"A=X", "A=", "A=B"}) {
		t.Fatal("Invalid environment", service.Config().Environment)
	}
}

func TestParseWithMultipleComposeFiles(t *testing.T) {
	configOne := []byte(`
  multiple:
    image: tianon/true
    ports:
      - 8000`)

	configTwo := []byte(`
  multiple:
    image: busybox
    container_name: multi
    ports:
      - 9000`)

	configThree := []byte(`
  multiple:
    image: busybox
    mem_limit: "40m"
    memswap_limit: 40000000
    ports:
      - 10000`)

	p := NewProject(&Context{
		ComposeBytes: [][]byte{configOne, configTwo},
	}, nil, nil)

	err := p.Parse()

	assert.Nil(t, err)

	multipleConfig, _ := p.ServiceConfigs.Get("multiple")
	assert.Equal(t, "busybox", multipleConfig.Image)
	assert.Equal(t, "multi", multipleConfig.ContainerName)
	assert.Equal(t, []string{"8000", "9000"}, multipleConfig.Ports)

	p = NewProject(&Context{
		ComposeBytes: [][]byte{configTwo, configOne},
	}, nil, nil)

	err = p.Parse()

	assert.Nil(t, err)

	multipleConfig, _ = p.ServiceConfigs.Get("multiple")
	assert.Equal(t, "tianon/true", multipleConfig.Image)
	assert.Equal(t, "multi", multipleConfig.ContainerName)
	assert.Equal(t, []string{"9000", "8000"}, multipleConfig.Ports)

	p = NewProject(&Context{
		ComposeBytes: [][]byte{configOne, configTwo, configThree},
	}, nil, nil)

	err = p.Parse()

	assert.Nil(t, err)

	multipleConfig, _ = p.ServiceConfigs.Get("multiple")
	assert.Equal(t, "busybox", multipleConfig.Image)
	assert.Equal(t, "multi", multipleConfig.ContainerName)
	assert.Equal(t, []string{"8000", "9000", "10000"}, multipleConfig.Ports)
	assert.Equal(t, yaml.MemStringorInt(41943040), multipleConfig.MemLimit)
	assert.Equal(t, yaml.MemStringorInt(40000000), multipleConfig.MemSwapLimit)
}
