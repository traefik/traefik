package builders

import (
	"time"

	"github.com/docker/docker/api/types/swarm"
)

// Config creates a config with default values.
// Any number of config builder functions can be passed to augment it.
func Config(builders ...func(config *swarm.Config)) *swarm.Config {
	config := &swarm.Config{}

	for _, builder := range builders {
		builder(config)
	}

	return config
}

// ConfigLabels sets the config's labels
func ConfigLabels(labels map[string]string) func(config *swarm.Config) {
	return func(config *swarm.Config) {
		config.Spec.Labels = labels
	}
}

// ConfigName sets the config's name
func ConfigName(name string) func(config *swarm.Config) {
	return func(config *swarm.Config) {
		config.Spec.Name = name
	}
}

// ConfigID sets the config's ID
func ConfigID(ID string) func(config *swarm.Config) {
	return func(config *swarm.Config) {
		config.ID = ID
	}
}

// ConfigVersion sets the version for the config
func ConfigVersion(v swarm.Version) func(*swarm.Config) {
	return func(config *swarm.Config) {
		config.Version = v
	}
}

// ConfigCreatedAt sets the creation time for the config
func ConfigCreatedAt(t time.Time) func(*swarm.Config) {
	return func(config *swarm.Config) {
		config.CreatedAt = t
	}
}

// ConfigUpdatedAt sets the update time for the config
func ConfigUpdatedAt(t time.Time) func(*swarm.Config) {
	return func(config *swarm.Config) {
		config.UpdatedAt = t
	}
}

// ConfigData sets the config payload.
func ConfigData(data []byte) func(*swarm.Config) {
	return func(config *swarm.Config) {
		config.Spec.Data = data
	}
}
