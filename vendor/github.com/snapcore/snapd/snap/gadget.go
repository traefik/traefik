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

package snap

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type GadgetInfo struct {
	Volumes map[string]GadgetVolume `yaml:"volumes,omitempty"`

	// Default configuration for snaps (snap-id => key => value).
	Defaults map[string]map[string]interface{} `yaml:"defaults,omitempty"`

	Connections []GadgetConnection `yaml:"connections"`
}

type GadgetVolume struct {
	Schema     string            `yaml:"schema"`
	Bootloader string            `yaml:"bootloader"`
	ID         string            `yaml:"id"`
	Structure  []VolumeStructure `yaml:"structure"`
}

// TODO Offsets and sizes are strings to support unit suffixes.
// Is that a good idea? *2^N or *10^N? We'll probably want a richer
// type when we actually handle these.

type VolumeStructure struct {
	Label       string          `yaml:"filesystem-label"`
	Offset      string          `yaml:"offset"`
	OffsetWrite string          `yaml:"offset-write"`
	Size        string          `yaml:"size"`
	Type        string          `yaml:"type"`
	ID          string          `yaml:"id"`
	Filesystem  string          `yaml:"filesystem"`
	Content     []VolumeContent `yaml:"content"`
}

type VolumeContent struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`

	Image       string `yaml:"image"`
	Offset      string `yaml:"offset"`
	OffsetWrite string `yaml:"offset-write"`
	Size        string `yaml:"size"`

	Unpack bool `yaml:"unpack"`
}

// GadgetConnect describes an interface connection requested by the gadget
// between seeded snaps. The syntax is of a mapping like:
//
//  plug: (<plug-snap-id>|system):plug
//  [slot: (<slot-snap-id>|system):slot]
//
// "system" indicates a system plug or slot.
// Fully omitting the slot part indicates a system slot with the same name
// as the plug.
type GadgetConnection struct {
	Plug GadgetConnectionPlug `yaml:"plug"`
	Slot GadgetConnectionSlot `yaml:"slot"`
}

type GadgetConnectionPlug struct {
	SnapID string
	Plug   string
}

func (gcplug *GadgetConnectionPlug) Empty() bool {
	return gcplug.SnapID == "" && gcplug.Plug == ""
}

func (gcplug *GadgetConnectionPlug) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	snapID, name, err := parseSnapIDColonName(s)
	if err != nil {
		return fmt.Errorf("in gadget connection plug: %v", err)
	}
	gcplug.SnapID = snapID
	gcplug.Plug = name
	return nil
}

type GadgetConnectionSlot struct {
	SnapID string
	Slot   string
}

func (gcslot *GadgetConnectionSlot) Empty() bool {
	return gcslot.SnapID == "" && gcslot.Slot == ""
}

func (gcslot *GadgetConnectionSlot) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	snapID, name, err := parseSnapIDColonName(s)
	if err != nil {
		return fmt.Errorf("in gadget connection slot: %v", err)
	}
	gcslot.SnapID = snapID
	gcslot.Slot = name
	return nil
}

func parseSnapIDColonName(s string) (snapID, name string, err error) {
	parts := strings.Split(s, ":")
	if len(parts) == 2 {
		snapID = parts[0]
		name = parts[1]
	}
	if snapID == "" || name == "" {
		return "", "", fmt.Errorf(`expected "(<snap-id>|system):name" not %q`, s)
	}
	return snapID, name, nil
}

func systemOrSnapID(s string) bool {
	if s != "system" && len(s) != validSnapIDLength {
		return false
	}
	return true
}

// ReadGadgetInfo reads the gadget specific metadata from gadget.yaml
// in the snap. classic set to true means classic rules apply,
// i.e. content/presence of gadget.yaml is fully optional.
func ReadGadgetInfo(info *Info, classic bool) (*GadgetInfo, error) {
	const errorFormat = "cannot read gadget snap details: %s"

	if info.Type != TypeGadget {
		return nil, fmt.Errorf(errorFormat, "not a gadget snap")
	}

	var gi GadgetInfo

	gadgetYamlFn := filepath.Join(info.MountDir(), "meta", "gadget.yaml")
	gmeta, err := ioutil.ReadFile(gadgetYamlFn)
	if classic && os.IsNotExist(err) {
		// gadget.yaml is optional for classic gadgets
		return &gi, nil
	}
	if err != nil {
		return nil, fmt.Errorf(errorFormat, err)
	}

	if err := yaml.Unmarshal(gmeta, &gi); err != nil {
		return nil, fmt.Errorf(errorFormat, err)
	}

	for k, v := range gi.Defaults {
		if !systemOrSnapID(k) {
			return nil, fmt.Errorf(`default stanza not keyed by "system" or snap-id: %s`, k)
		}
		dflt, err := normalizeYamlValue(v)
		if err != nil {
			return nil, fmt.Errorf("default value %q of %q: %v", v, k, err)
		}
		gi.Defaults[k] = dflt.(map[string]interface{})
	}

	for i, gconn := range gi.Connections {
		if gconn.Plug.Empty() {
			return nil, fmt.Errorf("gadget connection plug cannot be empty")
		}
		if gconn.Slot.Empty() {
			gi.Connections[i].Slot.SnapID = "system"
			gi.Connections[i].Slot.Slot = gconn.Plug.Plug
		}
	}

	if classic && len(gi.Volumes) == 0 {
		// volumes can be left out on classic
		// can still specify defaults though
		return &gi, nil
	}

	// basic validation
	var bootloadersFound int
	for _, v := range gi.Volumes {
		switch v.Bootloader {
		case "":
			// pass
		case "grub", "u-boot", "android-boot":
			bootloadersFound += 1
		default:
			return nil, fmt.Errorf(errorFormat, "bootloader must be one of grub, u-boot or android-boot")
		}
	}
	switch {
	case bootloadersFound == 0:
		return nil, fmt.Errorf(errorFormat, "bootloader not declared in any volume")
	case bootloadersFound > 1:
		return nil, fmt.Errorf(errorFormat, fmt.Sprintf("too many (%d) bootloaders declared", bootloadersFound))
	}

	return &gi, nil
}
