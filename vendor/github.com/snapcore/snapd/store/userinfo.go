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

package store

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/snapcore/snapd/httputil"
)

var (
	httpClient = httputil.NewHTTPClient(&httputil.ClientOptions{
		Timeout:    10 * time.Second,
		MayLogBody: true,
	})
)

type keysReply struct {
	Username         string   `json:"username"`
	SSHKeys          []string `json:"ssh_keys"`
	OpenIDIdentifier string   `json:"openid_identifier"`
}

type User struct {
	Username         string
	SSHKeys          []string
	OpenIDIdentifier string
}

func UserInfo(email string) (userinfo *User, err error) {
	var v keysReply
	ssourl := fmt.Sprintf("%s/keys/%s", authURL(), url.QueryEscape(email))

	resp, err := httputil.RetryRequest(ssourl, func() (*http.Response, error) {
		return httpClient.Get(ssourl)
	}, func(resp *http.Response) error {
		if resp.StatusCode != 200 {
			// we recheck the status
			return nil
		}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&v); err != nil {
			return fmt.Errorf("cannot unmarshal: %v", err)
		}
		return nil
	}, defaultRetryStrategy)

	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200: // good
	case 404:
		return nil, fmt.Errorf("cannot find user %q", email)
	default:
		return nil, respToError(resp, fmt.Sprintf("look up user %q", email))
	}

	return &User{
		Username:         v.Username,
		SSHKeys:          v.SSHKeys,
		OpenIDIdentifier: v.OpenIDIdentifier,
	}, nil
}
