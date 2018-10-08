// +build !go1.8

/*
 * Copyright (C) 2016-2017 Canonical Ltd
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

package httputil

import (
	"net/http"
	"strings"
)

func fixupHeadersForRedirect(req *http.Request, via []*http.Request) {
	// preserve some headers across redirects (needed for the CDN)
	// (this is done automatically, slightly more cleanly, from 1.8)
	for k, v := range via[0].Header {
		switch strings.ToLower(k) {
		case "authorization", "www-authenticate", "cookie", "cookie2":
			// Do not copy sensitive headers across
			// redirects. For rationale for which headers and
			// why, see https://golang.org/pkg/net/http/#Client
		default:
			req.Header[k] = v
		}
	}
}
