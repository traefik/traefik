package docker

// ContainerConfig holds container libkermit configuration possibilities
type ContainerConfig struct {
	Name       string
	Cmd        []string
	Entrypoint []string
	Labels     map[string]string
}
