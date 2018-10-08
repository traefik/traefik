// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/snapcore/snapd/jsonutil"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/snap"
)

var validKey = regexp.MustCompile("^(?:[a-z0-9]+-?)*[a-z](?:-?[a-z0-9])*$")

func ParseKey(key string) (subkeys []string, err error) {
	if key == "" {
		return []string{}, nil
	}
	subkeys = strings.Split(key, ".")
	for _, subkey := range subkeys {
		if !validKey.MatchString(subkey) {
			return nil, fmt.Errorf("invalid option name: %q", subkey)
		}
	}
	return subkeys, nil
}

func PatchConfig(snapName string, subkeys []string, pos int, config interface{}, value *json.RawMessage) (interface{}, error) {

	switch config := config.(type) {
	case nil:
		// Missing update map. Create and nest final value under it.
		configm := make(map[string]interface{})
		_, err := PatchConfig(snapName, subkeys, pos, configm, value)
		if err != nil {
			return nil, err
		}
		return configm, nil

	case *json.RawMessage:
		// Raw replaces pristine on commit. Unpack, update, and repack.
		var configm map[string]interface{}

		if err := jsonutil.DecodeWithNumber(bytes.NewReader(*config), &configm); err != nil {
			return nil, fmt.Errorf("snap %q option %q is not a map", snapName, strings.Join(subkeys[:pos], "."))
		}
		_, err := PatchConfig(snapName, subkeys, pos, configm, value)
		if err != nil {
			return nil, err
		}
		return jsonRaw(configm), nil

	case map[string]interface{}:
		// Update map to apply against pristine on commit.
		if pos+1 == len(subkeys) {
			config[subkeys[pos]] = value
			return config, nil
		} else {
			result, err := PatchConfig(snapName, subkeys, pos+1, config[subkeys[pos]], value)
			if err != nil {
				return nil, err
			}
			config[subkeys[pos]] = result
			return config, nil
		}
	}
	panic(fmt.Errorf("internal error: unexpected configuration type %T", config))
}

// Get unmarshals into result the value of the provided snap's configuration key.
// If the key does not exist, an error of type *NoOptionError is returned.
// The provided key may be formed as a dotted key path through nested maps.
// For example, the "a.b.c" key describes the {a: {b: {c: value}}} map.
func GetFromChange(snapName string, subkeys []string, pos int, config map[string]interface{}, result interface{}) error {
	// special case - get root document
	if len(subkeys) == 0 {
		if config == nil {
			return &NoOptionError{SnapName: snapName, Key: ""}
		}
		raw := jsonRaw(config)

		if err := jsonutil.DecodeWithNumber(bytes.NewReader(*raw), &result); err != nil {
			return fmt.Errorf("internal error: cannot unmarshal snap %q root document: %s", snapName, err)
		}
		return nil
	}
	value, ok := config[subkeys[pos]]
	if !ok {
		return &NoOptionError{SnapName: snapName, Key: strings.Join(subkeys[:pos+1], ".")}
	}

	if pos+1 == len(subkeys) {
		raw, ok := value.(*json.RawMessage)
		if !ok {
			raw = jsonRaw(value)
		}
		if err := jsonutil.DecodeWithNumber(bytes.NewReader(*raw), &result); err != nil {
			key := strings.Join(subkeys, ".")
			return fmt.Errorf("internal error: cannot unmarshal snap %q option %q into %T: %s, json: %s", snapName, key, result, err, *raw)
		}
		return nil
	}

	configm, ok := value.(map[string]interface{})
	if !ok {
		raw, ok := value.(*json.RawMessage)
		if !ok {
			raw = jsonRaw(value)
		}
		if err := jsonutil.DecodeWithNumber(bytes.NewReader(*raw), &configm); err != nil {
			return fmt.Errorf("snap %q option %q is not a map", snapName, strings.Join(subkeys[:pos+1], "."))
		}
	}
	return GetFromChange(snapName, subkeys, pos+1, configm, result)
}

// GetSnapConfig retrieves the raw configuration of a given snap.
func GetSnapConfig(st *state.State, snapName string) (*json.RawMessage, error) {
	var config map[string]*json.RawMessage
	err := st.Get("config", &config)
	if err == state.ErrNoState {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	snapcfg, ok := config[snapName]
	if !ok {
		return nil, nil
	}
	return snapcfg, nil
}

// SetSnapConfig replaces the configuration of a given snap.
func SetSnapConfig(st *state.State, snapName string, snapcfg *json.RawMessage) error {
	var config map[string]*json.RawMessage
	err := st.Get("config", &config)
	isNil := snapcfg == nil || len(*snapcfg) == 0
	if err == state.ErrNoState {
		if isNil {
			// bail out early
			return nil
		}
		config = make(map[string]*json.RawMessage, 1)
	} else if err != nil {
		return err
	}
	if isNil {
		delete(config, snapName)
	} else {
		config[snapName] = snapcfg
	}
	st.Set("config", config)
	return nil
}

// SaveRevisionConfig makes a copy of config -> snapSnape configuration into the versioned config.
// It doesn't do anything if there is no configuration for given snap in the state.
// The caller is responsible for locking the state.
func SaveRevisionConfig(st *state.State, snapName string, rev snap.Revision) error {
	var config map[string]*json.RawMessage                    // snap => configuration
	var revisionConfig map[string]map[string]*json.RawMessage // snap => revision => configuration

	// Get current configuration of the snap from state
	err := st.Get("config", &config)
	if err == state.ErrNoState {
		return nil
	} else if err != nil {
		return fmt.Errorf("internal error: cannot unmarshal configuration: %v", err)
	}
	snapcfg, ok := config[snapName]
	if !ok {
		return nil
	}

	err = st.Get("revision-config", &revisionConfig)
	if err == state.ErrNoState {
		revisionConfig = make(map[string]map[string]*json.RawMessage)
	} else if err != nil {
		return err
	}
	cfgs := revisionConfig[snapName]
	if cfgs == nil {
		cfgs = make(map[string]*json.RawMessage)
	}
	cfgs[rev.String()] = snapcfg
	revisionConfig[snapName] = cfgs
	st.Set("revision-config", revisionConfig)
	return nil
}

// RestoreRevisionConfig restores a given revision of snap configuration into config -> snapName.
// If no configuration exists for given revision it does nothing (no error).
// The caller is responsible for locking the state.
func RestoreRevisionConfig(st *state.State, snapName string, rev snap.Revision) error {
	var config map[string]*json.RawMessage                    // snap => configuration
	var revisionConfig map[string]map[string]*json.RawMessage // snap => revision => configuration

	err := st.Get("revision-config", &revisionConfig)
	if err == state.ErrNoState {
		return nil
	} else if err != nil {
		return fmt.Errorf("internal error: cannot unmarshal revision-config: %v", err)
	}

	err = st.Get("config", &config)
	if err == state.ErrNoState {
		config = make(map[string]*json.RawMessage)
	} else if err != nil {
		return fmt.Errorf("internal error: cannot unmarshal configuration: %v", err)
	}

	if cfg, ok := revisionConfig[snapName]; ok {
		if revCfg, ok := cfg[rev.String()]; ok {
			config[snapName] = revCfg
			st.Set("config", config)
		}
	}

	return nil
}

// DiscardRevisionConfig removes configuration snapshot of given snap/revision.
// If no configuration exists for given revision it does nothing (no error).
// The caller is responsible for locking the state.
func DiscardRevisionConfig(st *state.State, snapName string, rev snap.Revision) error {
	var revisionConfig map[string]map[string]*json.RawMessage // snap => revision => configuration
	err := st.Get("revision-config", &revisionConfig)
	if err == state.ErrNoState {
		return nil
	} else if err != nil {
		return fmt.Errorf("internal error: cannot unmarshal revision-config: %v", err)
	}

	if revCfgs, ok := revisionConfig[snapName]; ok {
		delete(revCfgs, rev.String())
		if len(revCfgs) == 0 {
			delete(revisionConfig, snapName)
		} else {
			revisionConfig[snapName] = revCfgs
		}
		st.Set("revision-config", revisionConfig)
	}
	return nil
}

// DeleteSnapConfig removed configuration of given snap from the state.
func DeleteSnapConfig(st *state.State, snapName string) error {
	var config map[string]map[string]*json.RawMessage // snap => key => value

	err := st.Get("config", &config)
	if err == state.ErrNoState {
		return nil
	} else if err != nil {
		return fmt.Errorf("internal error: cannot unmarshal configuration: %v", err)
	}
	if _, ok := config[snapName]; ok {
		delete(config, snapName)
		st.Set("config", config)
	}
	return nil
}
