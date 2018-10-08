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

package asserts

import (
	"fmt"
)

type fetchProgress int

const (
	fetchNotSeen fetchProgress = iota
	fetchRetrieved
	fetchSaved
)

// A Fetcher helps fetching assertions and their prerequisites.
type Fetcher interface {
	// Fetch retrieves the assertion indicated by ref then its prerequisites
	// recursively, along the way saving prerequisites before dependent assertions.
	Fetch(*Ref) error
	// Save retrieves the prerequisites of the assertion recursively,
	// along the way saving them, and finally saves the assertion.
	Save(Assertion) error
}

type fetcher struct {
	db       RODatabase
	retrieve func(*Ref) (Assertion, error)
	save     func(Assertion) error

	fetched map[string]fetchProgress
}

// NewFetcher creates a Fetcher which will use trustedDB to determine trusted assertions, will fetch assertions following prerequisites using retrieve, and then will pass them to save, saving prerequisites before dependent assertions.
func NewFetcher(trustedDB RODatabase, retrieve func(*Ref) (Assertion, error), save func(Assertion) error) Fetcher {
	return &fetcher{
		db:       trustedDB,
		retrieve: retrieve,
		save:     save,
		fetched:  make(map[string]fetchProgress),
	}
}

func (f *fetcher) chase(ref *Ref, a Assertion) error {
	// check if ref points to predefined assertion, in which case
	// there is nothing to do
	_, err := ref.Resolve(f.db.FindPredefined)
	if err == nil {
		return nil
	}
	if !IsNotFound(err) {
		return err
	}
	u := ref.Unique()
	switch f.fetched[u] {
	case fetchSaved:
		return nil // nothing to do
	case fetchRetrieved:
		return fmt.Errorf("circular assertions are not expected: %s", ref)
	}
	if a == nil {
		retrieved, err := f.retrieve(ref)
		if err != nil {
			return err
		}
		a = retrieved
	}
	f.fetched[u] = fetchRetrieved
	for _, preref := range a.Prerequisites() {
		if err := f.Fetch(preref); err != nil {
			return err
		}
	}
	if err := f.fetchAccountKey(a.SignKeyID()); err != nil {
		return err
	}
	if err := f.save(a); err != nil {
		return err
	}
	f.fetched[u] = fetchSaved
	return nil
}

// Fetch retrieves the assertion indicated by ref then its prerequisites
// recursively, along the way saving prerequisites before dependent assertions.
func (f *fetcher) Fetch(ref *Ref) error {
	return f.chase(ref, nil)
}

// fetchAccountKey behaves like Fetch for the account-key with the given key id.
func (f *fetcher) fetchAccountKey(keyID string) error {
	keyRef := &Ref{
		Type:       AccountKeyType,
		PrimaryKey: []string{keyID},
	}
	return f.Fetch(keyRef)
}

// Save retrieves the prerequisites of the assertion recursively,
// along the way saving them, and finally saves the assertion.
func (f *fetcher) Save(a Assertion) error {
	return f.chase(a.Ref(), a)
}
