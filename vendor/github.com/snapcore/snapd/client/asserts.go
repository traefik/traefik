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
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/snapcore/snapd/asserts" // for parsing
)

// Ack tries to add an assertion to the system assertion
// database. To succeed the assertion must be valid, its signature
// verified with a known public key and the assertion consistent with
// and its prerequisite in the database.
func (client *Client) Ack(b []byte) error {
	var rsp interface{}
	if _, err := client.doSync("POST", "/v2/assertions", nil, nil, bytes.NewReader(b), &rsp); err != nil {
		return err
	}

	return nil
}

// AssertionTypes returns a list of assertion type names.
func (client *Client) AssertionTypes() ([]string, error) {
	var types struct {
		Types []string `json:"types"`
	}
	_, err := client.doSync("GET", "/v2/assertions", nil, nil, nil, &types)
	if err != nil {
		return nil, fmt.Errorf("cannot get assertion type names: %v", err)
	}

	return types.Types, nil
}

// Known queries assertions with type assertTypeName and matching assertion headers.
func (client *Client) Known(assertTypeName string, headers map[string]string) ([]asserts.Assertion, error) {
	path := fmt.Sprintf("/v2/assertions/%s", assertTypeName)
	q := url.Values{}

	if len(headers) > 0 {
		for k, v := range headers {
			q.Set(k, v)
		}
	}

	response, err := client.raw("GET", path, q, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query assertions: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, parseError(response)
	}

	sanityCount, err := strconv.Atoi(response.Header.Get("X-Ubuntu-Assertions-Count"))
	if err != nil {
		return nil, fmt.Errorf("invalid assertions count")
	}

	dec := asserts.NewDecoder(response.Body)

	asserts := []asserts.Assertion{}

	// TODO: make sure asserts can decode and deal with unknown types
	for {
		a, err := dec.Decode()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode assertions: %v", err)
		}
		asserts = append(asserts, a)
	}

	if len(asserts) != sanityCount {
		return nil, fmt.Errorf("response did not have the expected number of assertions")
	}

	return asserts, nil
}
