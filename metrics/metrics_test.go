package metrics

import "testing"

func TestNewVoidRegistry(t *testing.T) {
	registry := NewVoidRegistry()

	registry.ConfigReloadsCounter().With("some", "value").Add(1)
	registry.ConfigReloadFailuresCounter().With("some", "value").Add(1)
	registry.LastConfigReloadSuccessGauge().With("some", "value").Set(1)
	registry.EntrypointReqsCounter().With("some", "value").Add(1)
	registry.EntrypointReqDurationHistogram().With("some", "value").Observe(1)
	registry.BackendReqsCounter().With("some", "value").Add(1)
	registry.BackendReqDurationHistogram().With("some", "value").Observe(1)
	registry.BackendRetriesCounter().With("some", "value").Add(1)
}
