// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2016 Canonical Ltd
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

var validAccountKeyName = regexp.MustCompile(`^(?:[a-z0-9]+-?)*[a-z](?:-?[a-z0-9])*$`)

// AccountKey holds an account-key assertion, asserting a public key
// belonging to the account.
type AccountKey struct {
	assertionBase
	since  time.Time
	until  time.Time
	pubKey PublicKey
}

// AccountID returns the account-id of this account-key.
func (ak *AccountKey) AccountID() string {
	return ak.HeaderString("account-id")
}

// Name returns the name of the account key.
func (ak *AccountKey) Name() string {
	return ak.HeaderString("name")
}

func IsValidAccountKeyName(name string) bool {
	return validAccountKeyName.MatchString(name)
}

// Since returns the time when the account key starts being valid.
func (ak *AccountKey) Since() time.Time {
	return ak.since
}

// Until returns the time when the account key stops being valid. A zero time means the key is valid forever.
func (ak *AccountKey) Until() time.Time {
	return ak.until
}

// PublicKeyID returns the key id used for lookup of the account key.
func (ak *AccountKey) PublicKeyID() string {
	return ak.pubKey.ID()
}

// isKeyValidAt returns whether the account key is valid at 'when' time.
func (ak *AccountKey) isKeyValidAt(when time.Time) bool {
	valid := when.After(ak.since) || when.Equal(ak.since)
	if valid && !ak.until.IsZero() {
		valid = when.Before(ak.until)
	}
	return valid
}

// publicKey returns the underlying public key of the account key.
func (ak *AccountKey) publicKey() PublicKey {
	return ak.pubKey
}

func checkPublicKey(ab *assertionBase, keyIDName string) (PublicKey, error) {
	pubKey, err := DecodePublicKey(ab.Body())
	if err != nil {
		return nil, err
	}
	keyID, err := checkNotEmptyString(ab.headers, keyIDName)
	if err != nil {
		return nil, err
	}
	if keyID != pubKey.ID() {
		return nil, fmt.Errorf("public key does not match provided key id")
	}
	return pubKey, nil
}

// Implement further consistency checks.
func (ak *AccountKey) checkConsistency(db RODatabase, acck *AccountKey) error {
	if !db.IsTrustedAccount(ak.AuthorityID()) {
		return fmt.Errorf("account-key assertion for %q is not signed by a directly trusted authority: %s", ak.AccountID(), ak.AuthorityID())
	}
	_, err := db.Find(AccountType, map[string]string{
		"account-id": ak.AccountID(),
	})
	if IsNotFound(err) {
		return fmt.Errorf("account-key assertion for %q does not have a matching account assertion", ak.AccountID())
	}
	if err != nil {
		return err
	}
	// XXX: Make this unconditional once account-key assertions are required to have a name.
	if ak.Name() != "" {
		// Check that we don't end up with multiple keys with
		// different IDs but the same account-id and name.
		// Note that this is a non-transactional check-then-add, so
		// is not a hard guarantee.  Backstores that can implement a
		// unique constraint should do so.
		assertions, err := db.FindMany(AccountKeyType, map[string]string{
			"account-id": ak.AccountID(),
			"name":       ak.Name(),
		})
		if err != nil && !IsNotFound(err) {
			return err
		}
		for _, assertion := range assertions {
			existingAccKey := assertion.(*AccountKey)
			if ak.PublicKeyID() != existingAccKey.PublicKeyID() {
				return fmt.Errorf("account-key assertion for %q with ID %q has the same name %q as existing ID %q", ak.AccountID(), ak.PublicKeyID(), ak.Name(), existingAccKey.PublicKeyID())
			}
		}
	}
	return nil
}

// sanity
var _ consistencyChecker = (*AccountKey)(nil)

// Prerequisites returns references to this account-key's prerequisite assertions.
func (ak *AccountKey) Prerequisites() []*Ref {
	return []*Ref{
		{Type: AccountType, PrimaryKey: []string{ak.AccountID()}},
	}
}

func assembleAccountKey(assert assertionBase) (Assertion, error) {
	_, err := checkNotEmptyString(assert.headers, "account-id")
	if err != nil {
		return nil, err
	}

	// XXX: We should require name to be present after backfilling existing assertions.
	_, ok := assert.headers["name"]
	if ok {
		_, err = checkStringMatches(assert.headers, "name", validAccountKeyName)
		if err != nil {
			return nil, err
		}
	}

	since, err := checkRFC3339Date(assert.headers, "since")
	if err != nil {
		return nil, err
	}

	until, err := checkRFC3339DateWithDefault(assert.headers, "until", time.Time{})
	if err != nil {
		return nil, err
	}
	if !until.IsZero() && until.Before(since) {
		return nil, fmt.Errorf("'until' time cannot be before 'since' time")
	}

	pubk, err := checkPublicKey(&assert, "public-key-sha3-384")
	if err != nil {
		return nil, err
	}

	// ignore extra headers for future compatibility
	return &AccountKey{
		assertionBase: assert,
		since:         since,
		until:         until,
		pubKey:        pubk,
	}, nil
}

// AccountKeyRequest holds an account-key-request assertion, which is a self-signed request to prove that the requester holds the private key and wishes to create an account-key assertion for it.
type AccountKeyRequest struct {
	assertionBase
	since  time.Time
	until  time.Time
	pubKey PublicKey
}

// AccountID returns the account-id of this account-key-request.
func (akr *AccountKeyRequest) AccountID() string {
	return akr.HeaderString("account-id")
}

// Name returns the name of the account key.
func (akr *AccountKeyRequest) Name() string {
	return akr.HeaderString("name")
}

// Since returns the time when the requested account key starts being valid.
func (akr *AccountKeyRequest) Since() time.Time {
	return akr.since
}

// Until returns the time when the requested account key stops being valid. A zero time means the key is valid forever.
func (akr *AccountKeyRequest) Until() time.Time {
	return akr.until
}

// PublicKeyID returns the underlying public key ID of the requested account key.
func (akr *AccountKeyRequest) PublicKeyID() string {
	return akr.pubKey.ID()
}

// signKey returns the underlying public key of the requested account key.
func (akr *AccountKeyRequest) signKey() PublicKey {
	return akr.pubKey
}

// Implement further consistency checks.
func (akr *AccountKeyRequest) checkConsistency(db RODatabase, acck *AccountKey) error {
	_, err := db.Find(AccountType, map[string]string{
		"account-id": akr.AccountID(),
	})
	if IsNotFound(err) {
		return fmt.Errorf("account-key-request assertion for %q does not have a matching account assertion", akr.AccountID())
	}
	if err != nil {
		return err
	}
	return nil
}

// sanity
var (
	_ consistencyChecker = (*AccountKeyRequest)(nil)
	_ customSigner       = (*AccountKeyRequest)(nil)
)

// Prerequisites returns references to this account-key-request's prerequisite assertions.
func (akr *AccountKeyRequest) Prerequisites() []*Ref {
	return []*Ref{
		{Type: AccountType, PrimaryKey: []string{akr.AccountID()}},
	}
}

func assembleAccountKeyRequest(assert assertionBase) (Assertion, error) {
	_, err := checkNotEmptyString(assert.headers, "account-id")
	if err != nil {
		return nil, err
	}

	_, err = checkStringMatches(assert.headers, "name", validAccountKeyName)
	if err != nil {
		return nil, err
	}

	since, err := checkRFC3339Date(assert.headers, "since")
	if err != nil {
		return nil, err
	}

	until, err := checkRFC3339DateWithDefault(assert.headers, "until", time.Time{})
	if err != nil {
		return nil, err
	}
	if !until.IsZero() && until.Before(since) {
		return nil, fmt.Errorf("'until' time cannot be before 'since' time")
	}

	pubk, err := checkPublicKey(&assert, "public-key-sha3-384")
	if err != nil {
		return nil, err
	}

	// ignore extra headers for future compatibility
	return &AccountKeyRequest{
		assertionBase: assert,
		since:         since,
		until:         until,
		pubKey:        pubk,
	}, nil
}
