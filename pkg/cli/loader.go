package cli

// ResourceLoader a configuration resource loader.
type ResourceLoader interface {
	Load(args []string, cmd *Command) (bool, error)
}

type filenameGetter interface {
	GetFilename() string
}

// GetConfigFile returns the configuration file if any.
func GetConfigFile(loaders []ResourceLoader) string {
	for _, loader := range loaders {
		if v, ok := loader.(filenameGetter); ok {
			return v.GetFilename()
		}
	}
	return ""
}
