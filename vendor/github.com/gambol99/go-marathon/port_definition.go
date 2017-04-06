/*
Copyright 2016 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

// PortDefinition is a definition of a port that should be considered
// part of a resource. Port definitions are necessary when you are
// using HOST networking and no port mappings are specified.
type PortDefinition struct {
	Port     *int               `json:"port,omitempty"`
	Protocol string             `json:"protocol,omitempty"`
	Name     string             `json:"name,omitempty"`
	Labels   *map[string]string `json:"labels,omitempty"`
}

// SetPort sets the given port for the PortDefinition
func (p PortDefinition) SetPort(port int) PortDefinition {
	p.Port = &port
	return p
}

// AddLabel adds a label to the PortDefinition
//		name: the name of the label
//		value: value for this label
func (p PortDefinition) AddLabel(name, value string) PortDefinition {
	if p.Labels == nil {
		p.EmptyLabels()
	}
	(*p.Labels)[name] = value

	return p
}

// EmptyLabels explicitly empties the labels -- use this if you need to empty
// the labels of a PortDefinition that already has labels set
// (setting labels to nill will keep the current value)
func (p *PortDefinition) EmptyLabels() *PortDefinition {
	p.Labels = &map[string]string{}

	return p
}
