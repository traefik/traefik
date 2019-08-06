package cli

// ResourceLoader is a configuration resource loader.
type ResourceLoader interface {
	// Load populates cmd.Configuration, optionally using args to do so.
	Load(args []string, cmd *Command) (bool, error)
}
