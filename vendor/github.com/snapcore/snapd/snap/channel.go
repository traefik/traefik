// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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
	"strings"

	"github.com/snapcore/snapd/arch"
	"github.com/snapcore/snapd/strutil"
)

var channelRisks = []string{"stable", "candidate", "beta", "edge"}

// Channel identifies and describes completely a store channel.
type Channel struct {
	Architecture string `json:"architecture"`
	Name         string `json:"name"`
	Track        string `json:"track"`
	Risk         string `json:"risk"`
	Branch       string `json:"branch,omitempty"`
}

// ParseChannel parses a string representing a store channel and includes the given architecture, if architecture is "" the system architecture is included.
func ParseChannel(s string, architecture string) (Channel, error) {
	if s == "" {
		return Channel{}, fmt.Errorf("channel name cannot be empty")
	}
	p := strings.Split(s, "/")
	var risk, track, branch string
	switch len(p) {
	default:
		return Channel{}, fmt.Errorf("channel name has too many components: %s", s)
	case 3:
		track, risk, branch = p[0], p[1], p[2]
	case 2:
		if strutil.ListContains(channelRisks, p[0]) {
			risk, branch = p[0], p[1]
		} else {
			track, risk = p[0], p[1]
		}
	case 1:
		if strutil.ListContains(channelRisks, p[0]) {
			risk = p[0]
		} else {
			track = p[0]
			risk = "stable"
		}
	}

	if !strutil.ListContains(channelRisks, risk) {
		return Channel{}, fmt.Errorf("invalid risk in channel name: %s", s)
	}

	if architecture == "" {
		architecture = arch.UbuntuArchitecture()
	}

	return Channel{
		Architecture: architecture,
		Track:        track,
		Risk:         risk,
		Branch:       branch,
	}.Clean(), nil
}

// Clean returns a Channel with a normalized track and name.
func (c Channel) Clean() Channel {
	track := c.Track

	if track == "latest" {
		track = ""
	}

	// normalized name
	name := c.Risk
	if track != "" {
		name = track + "/" + name
	}
	if c.Branch != "" {
		name = name + "/" + c.Branch
	}

	return Channel{
		Architecture: c.Architecture,
		Name:         name,
		Track:        track,
		Risk:         c.Risk,
		Branch:       c.Branch,
	}
}

func (c Channel) String() string {
	return c.Name
}

// Full returns the full name of the channel, inclusive the default track "latest".
func (c *Channel) Full() string {
	if c.Track == "" {
		return "latest/" + c.Name
	}
	return c.String()
}

func riskLevel(risk string) int {
	for i, r := range channelRisks {
		if r == risk {
			return i
		}
	}
	return -1
}

// ChannelMatch represents on which fields two channels are matching.
type ChannelMatch struct {
	Architecture bool
	Track        bool
	Risk         bool
}

// String returns the string represantion of the match, results can be:
//  "architecture:track:risk"
//  "architecture:track"
//  "architecture:risk"
//  "track:risk"
//  "architecture"
//  "track"
//  "risk"
//  ""
func (cm ChannelMatch) String() string {
	matching := []string{}
	if cm.Architecture {
		matching = append(matching, "architecture")
	}
	if cm.Track {
		matching = append(matching, "track")
	}
	if cm.Risk {
		matching = append(matching, "risk")
	}
	return strings.Join(matching, ":")

}

// Match returns a ChannelMatch of which fields among architecture,track,risk match between c and c1 store channels, risk is matched taking channel inheritance into account and considering c the requested channel.
func (c *Channel) Match(c1 *Channel) ChannelMatch {
	requestedRiskLevel := riskLevel(c.Risk)
	rl1 := riskLevel(c1.Risk)
	return ChannelMatch{
		Architecture: c.Architecture == c1.Architecture,
		Track:        c.Track == c1.Track,
		Risk:         requestedRiskLevel >= rl1,
	}
}
