package provider

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func makeEcsInstance(containerDef *ecs.ContainerDefinition) ecsInstance {
	container := &ecs.Container{
		Name:            containerDef.Name,
		NetworkBindings: make([]*ecs.NetworkBinding, len(containerDef.PortMappings)),
	}

	for i, pm := range containerDef.PortMappings {
		container.NetworkBindings[i] = &ecs.NetworkBinding{
			HostPort:      pm.HostPort,
			ContainerPort: pm.ContainerPort,
			Protocol:      pm.Protocol,
			BindIP:        aws.String("0.0.0.0"),
		}
	}

	return ecsInstance{
		Name: "foo-http",
		ID:   "123456789abc",
		task: &ecs.Task{
			Containers: []*ecs.Container{container},
		},
		taskDefinition: &ecs.TaskDefinition{
			ContainerDefinitions: []*ecs.ContainerDefinition{containerDef},
		},
		container:           container,
		containerDefinition: containerDef,
		machine: &ec2.Instance{
			PrivateIpAddress: aws.String("10.0.0.0"),
		},
	}
}

func simpleEcsInstance(labels map[string]*string) ecsInstance {
	return makeEcsInstance(&ecs.ContainerDefinition{
		Name: aws.String("http"),
		PortMappings: []*ecs.PortMapping{{
			HostPort:      aws.Int64(80),
			ContainerPort: aws.Int64(80),
			Protocol:      aws.String("tcp"),
		}},
		DockerLabels: labels,
	})
}

func TestEcsProtocol(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "http",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: "https",
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.protocol": aws.String("https"),
			}),
		},
	}

	for _, c := range cases {
		value := c.instanceInfo.Protocol()
		if value != c.expected {
			t.Fatalf("Should have been %s, got %s", c.expected, value)
		}
	}
}

func TestEcsHost(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "10.0.0.0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
	}

	for _, c := range cases {
		value := c.instanceInfo.Host()
		if value != c.expected {
			t.Fatalf("Should have been %s, got %s", c.expected, value)
		}
	}
}

func TestEcsPort(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "80",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
	}

	for _, c := range cases {
		value := c.instanceInfo.Port()
		if value != c.expected {
			t.Fatalf("Should have been %s, got %s", c.expected, value)
		}
	}
}

func TestEcsWeight(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: "10",
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.weight": aws.String("10"),
			}),
		},
	}

	for _, c := range cases {
		value := c.instanceInfo.Weight()
		if value != c.expected {
			t.Fatalf("Should have been %s, got %s", c.expected, value)
		}
	}
}

func TestEcsPassHostHeader(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "true",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: "false",
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.frontend.passHostHeader": aws.String("false"),
			}),
		},
	}

	for _, c := range cases {
		value := c.instanceInfo.PassHostHeader()
		if value != c.expected {
			t.Fatalf("Should have been %s, got %s", c.expected, value)
		}
	}
}

func TestEcsPriority(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: "10",
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.frontend.priority": aws.String("10"),
			}),
		},
	}

	for _, c := range cases {
		value := c.instanceInfo.Priority()
		if value != c.expected {
			t.Fatalf("Should have been %s, got %s", c.expected, value)
		}
	}
}

func TestEcsEntryPoints(t *testing.T) {
	cases := []struct {
		expected     []string
		instanceInfo ecsInstance
	}{
		{
			expected:     []string{},
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: []string{"http"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.frontend.entryPoints": aws.String("http"),
			}),
		},
		{
			expected: []string{"http", "https"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.frontend.entryPoints": aws.String("http,https"),
			}),
		},
	}

	for _, c := range cases {
		value := c.instanceInfo.EntryPoints()
		if !reflect.DeepEqual(value, c.expected) {
			t.Fatalf("Should have been %s, got %s", c.expected, value)
		}
	}
}
