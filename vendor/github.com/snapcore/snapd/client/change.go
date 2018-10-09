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
	"fmt"
	"net/url"
	"time"
)

// A Change is a modification to the system state.
type Change struct {
	ID      string  `json:"id"`
	Kind    string  `json:"kind"`
	Summary string  `json:"summary"`
	Status  string  `json:"status"`
	Tasks   []*Task `json:"tasks,omitempty"`
	Ready   bool    `json:"ready"`
	Err     string  `json:"err,omitempty"`

	SpawnTime time.Time `json:"spawn-time,omitempty"`
	ReadyTime time.Time `json:"ready-time,omitempty"`

	data map[string]*json.RawMessage
}

var ErrNoData = fmt.Errorf("data entry not found")

// Get unmarshals into value the kind-specific data with the provided key.
func (c *Change) Get(key string, value interface{}) error {
	raw := c.data[key]
	if raw == nil {
		return ErrNoData
	}
	return json.Unmarshal([]byte(*raw), value)
}

// A Task is an operation done to change the system's state.
type Task struct {
	ID       string       `json:"id"`
	Kind     string       `json:"kind"`
	Summary  string       `json:"summary"`
	Status   string       `json:"status"`
	Log      []string     `json:"log,omitempty"`
	Progress TaskProgress `json:"progress"`

	SpawnTime time.Time `json:"spawn-time,omitempty"`
	ReadyTime time.Time `json:"ready-time,omitempty"`
}

type TaskProgress struct {
	Label string `json:"label"`
	Done  int    `json:"done"`
	Total int    `json:"total"`
}

type changeAndData struct {
	Change
	Data map[string]*json.RawMessage `json:"data"`
}

// Change fetches information about a Change given its ID.
func (client *Client) Change(id string) (*Change, error) {
	var chgd changeAndData
	_, err := client.doSync("GET", "/v2/changes/"+id, nil, nil, nil, &chgd)
	if err != nil {
		return nil, err
	}

	chgd.Change.data = chgd.Data
	return &chgd.Change, nil
}

// Abort attempts to abort a change that is in not yet ready.
func (client *Client) Abort(id string) (*Change, error) {
	var postData struct {
		Action string `json:"action"`
	}
	postData.Action = "abort"

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(postData); err != nil {
		return nil, err
	}

	var chg Change
	if _, err := client.doSync("POST", "/v2/changes/"+id, nil, nil, &body, &chg); err != nil {
		return nil, err
	}

	return &chg, nil
}

type ChangeSelector uint8

func (c ChangeSelector) String() string {
	switch c {
	case ChangesInProgress:
		return "in-progress"
	case ChangesReady:
		return "ready"
	case ChangesAll:
		return "all"
	}

	panic(fmt.Sprintf("unknown ChangeSelector %d", c))
}

const (
	ChangesInProgress ChangeSelector = 1 << iota
	ChangesReady
	ChangesAll = ChangesReady | ChangesInProgress
)

type ChangesOptions struct {
	SnapName string // if empty, no filtering by name is done
	Selector ChangeSelector
}

func (client *Client) Changes(opts *ChangesOptions) ([]*Change, error) {
	query := url.Values{}
	if opts != nil {
		if opts.Selector != 0 {
			query.Set("select", opts.Selector.String())
		}
		if opts.SnapName != "" {
			query.Set("for", opts.SnapName)
		}
	}

	var chgds []changeAndData
	_, err := client.doSync("GET", "/v2/changes", query, nil, nil, &chgds)
	if err != nil {
		return nil, err
	}

	var chgs []*Change
	for i := range chgds {
		chgd := &chgds[i]
		chgd.Change.data = chgd.Data
		chgs = append(chgs, &chgd.Change)
	}

	return chgs, err
}
