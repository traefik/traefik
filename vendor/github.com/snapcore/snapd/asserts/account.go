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
	"regexp"
	"time"
)

var (
	// account ids look like snap-ids or a nice identifier
	validAccountID = regexp.MustCompile("^(?:[a-z0-9A-Z]{32}|[-a-z0-9]{2,28})$")
)

// Account holds an account assertion, which ties a name for an account
// to its identifier and provides the authority's confidence in the name's validity.
type Account struct {
	assertionBase
	validation string
	timestamp  time.Time
}

// AccountID returns the account-id of the account.
func (acc *Account) AccountID() string {
	return acc.HeaderString("account-id")
}

// Username returns the user name for the account.
func (acc *Account) Username() string {
	return acc.HeaderString("username")
}

// DisplayName returns the human-friendly name for the account.
func (acc *Account) DisplayName() string {
	return acc.HeaderString("display-name")
}

// Validation returns the level of confidence of the authority in the
// account's identity, expected to be "unproven" or "verified", and
// for forward compatibility any value != "unproven" can be considered
// at least "verified".
func (acc *Account) Validation() string {
	return acc.validation
}

// Timestamp returns the time when the account was issued.
func (acc *Account) Timestamp() time.Time {
	return acc.timestamp
}

// Implement further consistency checks.
func (acc *Account) checkConsistency(db RODatabase, acck *AccountKey) error {
	if !db.IsTrustedAccount(acc.AuthorityID()) {
		return fmt.Errorf("account assertion for %q is not signed by a directly trusted authority: %s", acc.AccountID(), acc.AuthorityID())
	}
	return nil
}

// sanity
var _ consistencyChecker = (*Account)(nil)

func assembleAccount(assert assertionBase) (Assertion, error) {
	_, err := checkNotEmptyString(assert.headers, "display-name")
	if err != nil {
		return nil, err
	}

	validation, err := checkNotEmptyString(assert.headers, "validation")
	if err != nil {
		return nil, err
	}
	// backward compatibility with the hard-coded trusted account
	// assertions
	// TODO: generate revision 1 of them with validation
	// s/certified/verified/
	if validation == "certified" {
		validation = "verified"
	}

	timestamp, err := checkRFC3339Date(assert.headers, "timestamp")
	if err != nil {
		return nil, err
	}

	_, err = checkOptionalString(assert.headers, "username")
	if err != nil {
		return nil, err
	}

	return &Account{
		assertionBase: assert,
		validation:    validation,
		timestamp:     timestamp,
	}, nil
}
