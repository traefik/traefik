package server

import "github.com/traefik/traefik/v3/pkg/network"

// entryPointSources returns the sources the networks extension gives the named entryPoint,
// which is none without an extension.
func entryPointSources(networks network.Extension, name string) network.Sources {
	if networks == nil {
		return network.Sources{}
	}

	return networks.Sources(name)
}
