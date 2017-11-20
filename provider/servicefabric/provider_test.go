package servicefabric

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	sfsdk "github.com/jjcollinge/servicefabric"
)

func TestUpdateConfig(t *testing.T) {
	apps := &sfsdk.ApplicationItemsPage{
		ContinuationToken: nil,
		Items: []sfsdk.ApplicationItem{
			{
				HealthState: "Ok",
				ID:          "TestApplication",
				Name:        "fabric:/TestApplication",
				Parameters: []*struct {
					Key   string `json:"Key"`
					Value string `json:"Value"`
				}{
					{"TraefikPublish", "fabric:/TestApplication/TestService"},
				},
				Status:      "Ready",
				TypeName:    "TestApplicationType",
				TypeVersion: "1.0.0",
			},
		},
	}
	services := &sfsdk.ServiceItemsPage{
		ContinuationToken: nil,
		Items: []sfsdk.ServiceItem{
			{
				HasPersistedState: true,
				HealthState:       "Ok",
				ID:                "TestApplication/TestService",
				IsServiceGroup:    false,
				ManifestVersion:   "1.0.0",
				Name:              "fabric:/TestApplication/TestService",
				ServiceKind:       "Stateless",
				ServiceStatus:     "Active",
				TypeName:          "TestServiceType",
			},
		},
	}
	partitions := &sfsdk.PartitionItemsPage{
		ContinuationToken: nil,
		Items: []sfsdk.PartitionItem{
			{
				CurrentConfigurationEpoch: struct {
					ConfigurationVersion string `json:"ConfigurationVersion"`
					DataLossVersion      string `json:"DataLossVersion"`
				}{
					ConfigurationVersion: "12884901891",
					DataLossVersion:      "131496928071680379",
				},
				HealthState:       "Ok",
				MinReplicaSetSize: 1,
				PartitionInformation: struct {
					HighKey              string `json:"HighKey"`
					ID                   string `json:"Id"`
					LowKey               string `json:"LowKey"`
					ServicePartitionKind string `json:"ServicePartitionKind"`
				}{
					HighKey:              "9223372036854775807",
					ID:                   "bce46a8c-b62d-4996-89dc-7ffc00a96902",
					LowKey:               "-9223372036854775808",
					ServicePartitionKind: "Int64Range",
				},
				PartitionStatus:      "Ready",
				ServiceKind:          "Stateless",
				TargetReplicaSetSize: 1,
			},
		},
	}
	instances := &sfsdk.InstanceItemsPage{
		ContinuationToken: nil,
		Items: []sfsdk.InstanceItem{
			{
				ReplicaItemBase: &sfsdk.ReplicaItemBase{
					Address:                      "{\"Endpoints\":{\"\":\"http:\\/\\/localhost:8081\"}}",
					HealthState:                  "Ok",
					LastInBuildDurationInSeconds: "3",
					NodeName:                     "_Node_0",
					ReplicaStatus:                "Ready",
					ServiceKind:                  "Stateless",
				},
				ID: "131497042182378182",
			},
		},
	}

	labels := map[string]string{
		"expose":                      "",
		"frontend.rule.default":       "Path: /",
		"backend.loadbalancer.method": "wrr",
		"backend.circuitbreaker":      "NetworkErrorRatio() > 0.5",
	}

	client := &clientMock{
		applications: apps,
		services:     services,
		partitions:   partitions,
		replicas:     nil,
		instances:    instances,
		labels:       labels,
	}
	expected := types.ConfigMessage{
		ProviderName: "servicefabric",
		Configuration: &types.Configuration{
			Frontends: map[string]*types.Frontend{
				"fabric:/TestApplication/TestService": {
					EntryPoints: []string{},
					Backend:     "fabric:/TestApplication/TestService",
					Routes: map[string]types.Route{
						"frontend.rule.default": {
							Rule: "Path: /",
						},
					},
				},
			},
			Backends: map[string]*types.Backend{
				"fabric:/TestApplication/TestService": {
					LoadBalancer: &types.LoadBalancer{
						Method: "wrr",
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					Servers: map[string]types.Server{
						"131497042182378182": {
							URL:    "http://localhost:8081",
							Weight: 1,
						},
					},
				},
			},
		},
	}

	provider := Provider{}
	configurationChan := make(chan types.ConfigMessage)
	ctx := context.Background()
	pool := safe.NewPool(ctx)
	defer pool.Stop()
	provider.updateConfig(configurationChan, pool, client, time.Millisecond*100)

	timeout := make(chan string, 1)
	go func() {
		time.Sleep(time.Second * 2)
		timeout <- "Timeout triggered"
	}()

	select {
	case actual := <-configurationChan:
		err := compareConfigurations(actual, expected)
		if err != nil {
			res, _ := json.Marshal(actual)
			t.Log(string(res))
			t.Error("actual != expected")
		}
	case <-timeout:
		t.Error("Provider failed to return configuration")
	}
}

func TestIsPrimary(t *testing.T) {
	provider := Provider{}
	replica := &sfsdk.ReplicaItem{
		ReplicaItemBase: &sfsdk.ReplicaItemBase{
			Address:                      "{\"Endpoints\":{\"\":\"localhost:30001+bce46a8c-b62d-4996-89dc-7ffc00a96902-131496928082309293\"}}",
			HealthState:                  "Ok",
			LastInBuildDurationInSeconds: "1",
			NodeName:                     "_Node_0",
			ReplicaRole:                  "Primary",
			ReplicaStatus:                "Ready",
			ServiceKind:                  "Stateful",
		},
		ID: "131496928082309293",
	}
	isPrimary := provider.isPrimary(replica)
	if !isPrimary {
		t.Error("Failed to identify replica as primary")
	}
}

func TestIsPrimaryWhenSecondary(t *testing.T) {
	provider := Provider{}
	replica := &sfsdk.ReplicaItem{
		ReplicaItemBase: &sfsdk.ReplicaItemBase{
			Address:                      "{\"Endpoints\":{\"\":\"localhost:30001+bce46a8c-b62d-4996-89dc-7ffc00a96902-131496928082309293\"}}",
			HealthState:                  "Ok",
			LastInBuildDurationInSeconds: "1",
			NodeName:                     "_Node_0",
			ReplicaRole:                  "Secondary",
			ReplicaStatus:                "Ready",
			ServiceKind:                  "Stateful",
		},
		ID: "131496928082309293",
	}
	isPrimary := provider.isPrimary(replica)
	if isPrimary {
		t.Error("Incorrectly identified replica as primary")
	}
}

func TestIsHealthy(t *testing.T) {
	replica := &sfsdk.ReplicaItem{
		ReplicaItemBase: &sfsdk.ReplicaItemBase{
			Address:                      "{\"Endpoints\":{\"\":\"localhost:30001+bce46a8c-b62d-4996-89dc-7ffc00a96902-131496928082309293\"}}",
			HealthState:                  "Ok",
			LastInBuildDurationInSeconds: "1",
			NodeName:                     "_Node_0",
			ReplicaRole:                  "Primary",
			ReplicaStatus:                "Ready",
			ServiceKind:                  "Stateful",
		},
		ID: "131496928082309293",
	}
	isHealthy := isHealthy(*replica.ReplicaItemBase)
	if !isHealthy {
		t.Error("Failed to identify replica as healthy")
	}
}

func TestIsHealthyWhenError(t *testing.T) {
	replica := &sfsdk.ReplicaItem{
		ReplicaItemBase: &sfsdk.ReplicaItemBase{
			Address:                      "{\"Endpoints\":{\"\":\"localhost:30001+bce46a8c-b62d-4996-89dc-7ffc00a96902-131496928082309293\"}}",
			HealthState:                  "Error",
			LastInBuildDurationInSeconds: "1",
			NodeName:                     "_Node_0",
			ReplicaRole:                  "Primary",
			ReplicaStatus:                "Error",
			ServiceKind:                  "Stateful",
		},
		ID: "131496928082309293",
	}
	isHealthy := isHealthy(*replica.ReplicaItemBase)
	if isHealthy {
		t.Error("Incorrectly identified replica as healthy")
	}
}

func compareConfigurations(actual, expected types.ConfigMessage) error {
	if actual.ProviderName == expected.ProviderName {
		if len(actual.Configuration.Frontends) == len(expected.Configuration.Frontends) {
			if len(actual.Configuration.Backends) == len(expected.Configuration.Backends) {
				actualFrontends, err := json.Marshal(actual.Configuration.Frontends)
				if err != nil {
					return err
				}
				actualFrontendsStr := string(actualFrontends)
				expectedFrontends, err := json.Marshal(expected.Configuration.Frontends)
				if err != nil {
					return err
				}
				expectedFrontendsStr := string(expectedFrontends)

				if actualFrontendsStr != expectedFrontendsStr {
					return errors.New("Backend configuration differs from expected configuration")
				}

				actualBackends, err := json.Marshal(actual.Configuration.Backends)
				if err != nil {
					return err
				}
				actualBackendsStr := string(actualBackends)
				expectedBackends, err := json.Marshal(expected.Configuration.Backends)
				if err != nil {
					return err
				}
				expectedBackendsStr := string(expectedBackends)

				if actualBackendsStr != expectedBackendsStr {
					return err
				}
				return nil
			}
		}
	}
	return errors.New("Provider name differs from expected")
}
