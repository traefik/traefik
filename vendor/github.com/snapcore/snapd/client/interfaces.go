// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package client

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
)

// Plug represents the potential of a given snap to connect to a slot.
type Plug struct {
	Snap        string                 `json:"snap"`
	Name        string                 `json:"plug"`
	Interface   string                 `json:"interface,omitempty"`
	Attrs       map[string]interface{} `json:"attrs,omitempty"`
	Apps        []string               `json:"apps,omitempty"`
	Label       string                 `json:"label,omitempty"`
	Connections []SlotRef              `json:"connections,omitempty"`
}

// PlugRef is a reference to a plug.
type PlugRef struct {
	Snap string `json:"snap"`
	Name string `json:"plug"`
}

// Slot represents a capacity offered by a snap.
type Slot struct {
	Snap        string                 `json:"snap"`
	Name        string                 `json:"slot"`
	Interface   string                 `json:"interface,omitempty"`
	Attrs       map[string]interface{} `json:"attrs,omitempty"`
	Apps        []string               `json:"apps,omitempty"`
	Label       string                 `json:"label,omitempty"`
	Connections []PlugRef              `json:"connections,omitempty"`
}

// SlotRef is a reference to a slot.
type SlotRef struct {
	Snap string `json:"snap"`
	Name string `json:"slot"`
}

// Connections contains information about all plugs, slots and their connections
type Connections struct {
	Plugs []Plug `json:"plugs"`
	Slots []Slot `json:"slots"`
}

// Interface holds information about a given interface and its instances.
type Interface struct {
	Name    string `json:"name,omitempty"`
	Summary string `json:"summary,omitempty"`
	DocURL  string `json:"doc-url,omitempty"`
	Plugs   []Plug `json:"plugs,omitempty"`
	Slots   []Slot `json:"slots,omitempty"`
}

// InterfaceAction represents an action performed on the interface system.
type InterfaceAction struct {
	Action string `json:"action"`
	Plugs  []Plug `json:"plugs,omitempty"`
	Slots  []Slot `json:"slots,omitempty"`
}

// Connections returns all plugs, slots and their connections.
func (client *Client) Connections() (Connections, error) {
	var conns Connections
	_, err := client.doSync("GET", "/v2/interfaces", nil, nil, nil, &conns)
	return conns, err
}

// InterfaceOptions represents opt-in elements include in responses.
type InterfaceOptions struct {
	Names     []string
	Doc       bool
	Plugs     bool
	Slots     bool
	Connected bool
}

func (client *Client) Interfaces(opts *InterfaceOptions) ([]*Interface, error) {
	query := url.Values{}
	if opts != nil && len(opts.Names) > 0 {
		query.Set("names", strings.Join(opts.Names, ",")) // Return just those specific interfaces.
	}
	if opts != nil {
		if opts.Doc {
			query.Set("doc", "true") // Return documentation of each selected interface.
		}
		if opts.Plugs {
			query.Set("plugs", "true") // Return plugs of each selected interface.
		}
		if opts.Slots {
			query.Set("slots", "true") // Return slots of each selected interface.
		}
	}
	// NOTE: Presence of "select" triggers the use of the new response format.
	if opts != nil && opts.Connected {
		query.Set("select", "connected") // Return just the connected interfaces.
	} else {
		query.Set("select", "all") // Return all interfaces.
	}
	var interfaces []*Interface
	_, err := client.doSync("GET", "/v2/interfaces", query, nil, nil, &interfaces)

	return interfaces, err
}

// performInterfaceAction performs a single action on the interface system.
func (client *Client) performInterfaceAction(sa *InterfaceAction) (changeID string, err error) {
	b, err := json.Marshal(sa)
	if err != nil {
		return "", err
	}
	return client.doAsync("POST", "/v2/interfaces", nil, nil, bytes.NewReader(b))
}

// Connect establishes a connection between a plug and a slot.
// The plug and the slot must have the same interface.
func (client *Client) Connect(plugSnapName, plugName, slotSnapName, slotName string) (changeID string, err error) {
	return client.performInterfaceAction(&InterfaceAction{
		Action: "connect",
		Plugs:  []Plug{{Snap: plugSnapName, Name: plugName}},
		Slots:  []Slot{{Snap: slotSnapName, Name: slotName}},
	})
}

// Disconnect breaks the connection between a plug and a slot.
func (client *Client) Disconnect(plugSnapName, plugName, slotSnapName, slotName string) (changeID string, err error) {
	return client.performInterfaceAction(&InterfaceAction{
		Action: "disconnect",
		Plugs:  []Plug{{Snap: plugSnapName, Name: plugName}},
		Slots:  []Slot{{Snap: slotSnapName, Name: slotName}},
	})
}
