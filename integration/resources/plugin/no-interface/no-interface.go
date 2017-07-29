package main

// NoInterface is a plugin that does not implement any interface
type NoInterface struct {
}

// Load loads the plugin instance
func Load() interface{} {
	return &NoInterface{}
}
