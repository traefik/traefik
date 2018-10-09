// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribuLicenseidte it and/or modify
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

package spdx

import "bytes"

// ValidateLicense implements license validation for SPDX 2.1 License
// Expressions as described in Appendix IV of
// https://spdx.org/spdx-specification-21-web-version
//
// An error is returned if the license string is not conforming this
// spec.
//
// Note that the "license-ref" part of the spec is not supported
func ValidateLicense(license string) error {
	return newParser(bytes.NewBufferString(license)).Validate()
}
