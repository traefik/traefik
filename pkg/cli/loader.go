package cli

// ResourceLoader is a configuration resource loader.
type ResourceLoader interface {
	// Load populates cmd.Configuration, optionally using args to do so.
	Load(args []string, cmd *Command) (bool, error)
}

type filenameGetter interface {
	GetFilename() string
}

// GetConfigFile returns the configuration file corresponding to the first configuration file loader found in ResourceLoader, if any.
func GetConfigFile(loaders []ResourceLoader) string {
	for _, loader := range loaders {
		if v, ok := loader.(filenameGetter); ok {
			return v.GetFilename()
		}
	}
	return ""
}
