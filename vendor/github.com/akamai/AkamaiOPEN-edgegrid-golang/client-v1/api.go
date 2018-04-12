package client

import (
	"encoding/json"
)

// Resource is the "base" type for all API resources
type Resource struct {
	Complete chan bool `json:"-"`
}

// Init initializes the Complete channel, if it is necessary
// need to create a resource specific Init(), make sure to
// initialize the channel.
func (resource *Resource) Init() {
	resource.Complete = make(chan bool, 1)
}

// PostUnmarshalJSON is a default implementation of the
// PostUnmarshalJSON hook that simply calls Init() and
// sends true to the Complete channel. This is overridden
// in many resources, in particular those that represent
// collections, and have to initialize sub-resources also.
func (resource *Resource) PostUnmarshalJSON() error {
	resource.Init()
	resource.Complete <- true
	return nil
}

// GetJSON returns the raw (indented) JSON (as []bytes)
func (resource *Resource) GetJSON() ([]byte, error) {
	return json.MarshalIndent(resource, "", "    ")
}

// JSONBody is a generic struct for temporary JSON unmarshalling.
type JSONBody map[string]interface{}
