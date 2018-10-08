// -*- Mode: Go; indent-tabs-mode: t -*-

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

package asserts

import (
	"fmt"
	"net/url"
	"time"
)

// Store holds a store assertion, defining the configuration needed to connect
// a device to the store.
type Store struct {
	assertionBase
	url       *url.URL
	timestamp time.Time
}

// Store returns the identifying name of the operator's store.
func (store *Store) Store() string {
	return store.HeaderString("store")
}

// OperatorID returns the account id of the store's operator.
func (store *Store) OperatorID() string {
	return store.HeaderString("operator-id")
}

// URL returns the URL of the store's API.
func (store *Store) URL() *url.URL {
	return store.url
}

// Location returns a summary of the store's location/purpose.
func (store *Store) Location() string {
	return store.HeaderString("location")
}

// Timestamp returns the time when the store assertion was issued.
func (store *Store) Timestamp() time.Time {
	return store.timestamp
}

func (store *Store) checkConsistency(db RODatabase, acck *AccountKey) error {
	// Will be applied to a system's snapd so must be signed by a trusted authority.
	if !db.IsTrustedAccount(store.AuthorityID()) {
		return fmt.Errorf("store assertion %q is not signed by a directly trusted authority: %s",
			store.Store(), store.AuthorityID())
	}

	_, err := db.Find(AccountType, map[string]string{"account-id": store.OperatorID()})
	if err != nil {
		if IsNotFound(err) {
			return fmt.Errorf(
				"store assertion %q does not have a matching account assertion for the operator %q",
				store.Store(), store.OperatorID())
		}
		return err
	}

	return nil
}

// Prerequisites returns references to this store's prerequisite assertions.
func (store *Store) Prerequisites() []*Ref {
	return []*Ref{
		{AccountType, []string{store.OperatorID()}},
	}
}

// checkStoreURL validates the "url" header and returns a full URL or nil.
func checkStoreURL(headers map[string]interface{}) (*url.URL, error) {
	s, err := checkOptionalString(headers, "url")
	if err != nil {
		return nil, err
	}

	if s == "" {
		return nil, nil
	}

	errWhat := `"url" header`

	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("%s must be a valid URL: %s", errWhat, s)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf(`%s scheme must be "https" or "http": %s`, errWhat, s)
	}
	if u.Host == "" {
		return nil, fmt.Errorf(`%s must have a host: %s`, errWhat, s)
	}
	if u.RawQuery != "" {
		return nil, fmt.Errorf(`%s must not have a query: %s`, errWhat, s)
	}
	if u.Fragment != "" {
		return nil, fmt.Errorf(`%s must not have a fragment: %s`, errWhat, s)
	}

	return u, nil
}

func assembleStore(assert assertionBase) (Assertion, error) {
	_, err := checkNotEmptyString(assert.headers, "operator-id")
	if err != nil {
		return nil, err
	}

	url, err := checkStoreURL(assert.headers)
	if err != nil {
		return nil, err
	}

	_, err = checkOptionalString(assert.headers, "location")
	if err != nil {
		return nil, err
	}

	timestamp, err := checkRFC3339Date(assert.headers, "timestamp")
	if err != nil {
		return nil, err
	}

	return &Store{
		assertionBase: assert,
		url:           url,
		timestamp:     timestamp,
	}, nil
}
