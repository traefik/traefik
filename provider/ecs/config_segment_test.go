package ecs

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSegmentBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc              string
		instanceInfo      ecsInstance
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc: "when container 2 container-ports",
			instanceInfo: ecsInstance{
				Name:                "foo-http",
				ID:                  "123456789abc",
				containerDefinition: &ecs.ContainerDefinition{},
				machine: &machine{
					state:     ec2.InstanceStateNameRunning,
					privateIP: "10.0.0.1",
					ports: []portMapping{
						{hostPort: 12354, containerPort: 8000}, // service1
						{hostPort: 45678, containerPort: 8001}, // service2
					},
				},
				TraefikLabels: map[string]string{
					"traefik.service1.port": "8000",
					"traefik.service2.port": "8001",
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-foo-http-service1": {
					Backend:        "backend-foo-http-service1",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-foo-http-service1": {
							Rule: "Host:foo-http.",
						},
					},
				},
				"frontend-foo-http-service2": {
					Backend:        "backend-foo-http-service2",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-foo-http-service2": {
							Rule: "Host:foo-http.",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-http-service1": {
					Servers: map[string]types.Server{
						"server-foo-http-123456789abc-service1": {
							URL:    "http://10.0.0.1:12354",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
				"backend-foo-http-service2": {
					Servers: map[string]types.Server{
						"server-foo-http-123456789abc-service2": {
							URL:    "http://10.0.0.1:45678",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{ExposedByDefault: true}

			var instances []ecsInstance
			instances = append(instances, test.instanceInfo)

			actualConfig, _ := p.buildConfiguration(instances)

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}
