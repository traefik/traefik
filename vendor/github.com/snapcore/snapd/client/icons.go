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
	"fmt"
	"io/ioutil"
	"regexp"
)

// Icon represents the icon of an installed snap
type Icon struct {
	Filename string
	Content  []byte
}

// Icon returns the Icon belonging to an installed snap
func (c *Client) Icon(pkgID string) (*Icon, error) {
	const errPrefix = "cannot retrieve icon"

	response, err := c.raw("GET", fmt.Sprintf("/v2/icons/%s/icon", pkgID), nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to communicate with server: %s", errPrefix, err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("%s: Not Found", errPrefix)
	}

	re := regexp.MustCompile(`attachment; filename=(.+)`)
	matches := re.FindStringSubmatch(response.Header.Get("Content-Disposition"))

	if matches == nil || matches[1] == "" {
		return nil, fmt.Errorf("%s: cannot determine filename", errPrefix)
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", errPrefix, err)
	}

	icon := &Icon{
		Filename: matches[1],
		Content:  content,
	}

	return icon, nil
}
