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

package httputil

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

type ClientOptions struct {
	Timeout    time.Duration
	TLSConfig  *tls.Config
	MayLogBody bool
	Proxy      func(*http.Request) (*url.URL, error)
}

// NewHTTPCLient returns a new http.Client with a LoggedTransport, a
// Timeout and preservation of range requests across redirects
func NewHTTPClient(opts *ClientOptions) *http.Client {
	if opts == nil {
		opts = &ClientOptions{}
	}

	transport := newDefaultTransport()
	transport.TLSClientConfig = opts.TLSConfig
	if opts.Proxy != nil {
		transport.Proxy = opts.Proxy
	}

	return &http.Client{
		Transport: &LoggedTransport{
			Transport: transport,
			Key:       "SNAPD_DEBUG_HTTP",
			body:      opts.MayLogBody,
		},
		Timeout:       opts.Timeout,
		CheckRedirect: checkRedirect,
	}
}
