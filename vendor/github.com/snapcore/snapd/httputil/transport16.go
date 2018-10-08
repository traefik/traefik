// -*- Mode: Go; indent-tabs-mode: t -*-

// +build !go1.7

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

package httputil

import (
	"net/http"
	"time"
)

var origDefaultTransport *http.Transport = http.DefaultTransport.(*http.Transport)

// newDefaultTransport makes a fresh modifiable instance of Transport
// with the same parameters as http.DefaultTransport.
func newDefaultTransport() *http.Transport {
	// based on https://github.com/golang/go/blob/release-branch.go1.6/src/net/http/transport.go#L33
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  origDefaultTransport.Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
