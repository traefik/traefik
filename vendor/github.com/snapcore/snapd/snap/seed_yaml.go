// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
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
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/snapcore/snapd/osutil"
)

// SeedSnap points to a snap in the seed to install, together with
// assertions (or alone if unasserted is true) it will be used to
// drive the installation and ultimately set SideInfo/SnapState for it.
type SeedSnap struct {
	Name string `yaml:"name"`

	// cross-reference/audit
	SnapID string `yaml:"snap-id,omitempty"`

	// bits that are orthongonal/not in assertions
	Channel string `yaml:"channel,omitempty"`
	DevMode bool   `yaml:"devmode,omitempty"`
	Classic bool   `yaml:"classic,omitempty"`

	Private bool `yaml:"private,omitempty"`

	Contact string `yaml:"contact,omitempty"`

	// no assertions are available in the seed for this snap
	Unasserted bool `yaml:"unasserted,omitempty"`

	File string `yaml:"file"`
}

type Seed struct {
	Snaps []*SeedSnap `yaml:"snaps"`
}

func ReadSeedYaml(fn string) (*Seed, error) {
	yamlData, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("cannot read seed yaml: %s", fn)
	}

	var seed Seed
	if err := yaml.Unmarshal(yamlData, &seed); err != nil {
		return nil, fmt.Errorf("cannot unmarshal %q: %s", yamlData, err)
	}

	// validate
	for _, sn := range seed.Snaps {
		if strings.Contains(sn.File, "/") {
			return nil, fmt.Errorf("%q must be a filename, not a path", sn.File)
		}
	}

	return &seed, nil
}

func (seed *Seed) Write(seedFn string) error {
	data, err := yaml.Marshal(&seed)
	if err != nil {
		return err
	}
	if err := osutil.AtomicWriteFile(seedFn, data, 0644, 0); err != nil {
		return err
	}
	return nil
}
