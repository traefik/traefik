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

package osutil

import (
	"bytes"
	"fmt"
)

// OutputErr formats an error based on output if its length is not zero,
// or returns err otherwise.
func OutputErr(output []byte, err error) error {
	output = bytes.TrimSpace(output)
	if len(output) > 0 {
		if bytes.Contains(output, []byte{'\n'}) {
			err = fmt.Errorf("\n-----\n%s\n-----", output)
		} else {
			err = fmt.Errorf("%s", output)
		}
	}
	return err
}
