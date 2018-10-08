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
	"encoding/json"

	"github.com/snapcore/snapd/store"
)

func (client *Client) Buy(opts *store.BuyOptions) (*store.BuyResult, error) {
	if opts == nil {
		opts = &store.BuyOptions{}
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(opts); err != nil {
		return nil, err
	}

	var result store.BuyResult
	_, err := client.doSync("POST", "/v2/buy", nil, nil, &body, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (client *Client) ReadyToBuy() error {
	var result bool
	_, err := client.doSync("GET", "/v2/buy/ready", nil, nil, nil, &result)
	return err
}
