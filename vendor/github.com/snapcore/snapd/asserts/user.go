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
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var validSystemUserUsernames = regexp.MustCompile(`^[a-z0-9][-a-z0-9+.-_]*$`)

// SystemUser holds a system-user assertion which allows creating local
// system users.
type SystemUser struct {
	assertionBase
	series  []string
	models  []string
	sshKeys []string
	since   time.Time
	until   time.Time
}

// BrandID returns the brand identifier that signed this assertion.
func (su *SystemUser) BrandID() string {
	return su.HeaderString("brand-id")
}

// Email returns the email address that this assertion is valid for.
func (su *SystemUser) Email() string {
	return su.HeaderString("email")
}

// Series returns the series that this assertion is valid for.
func (su *SystemUser) Series() []string {
	return su.series
}

// Models returns the models that this assertion is valid for.
func (su *SystemUser) Models() []string {
	return su.models
}

// Name returns the full name of the user (e.g. Random Guy).
func (su *SystemUser) Name() string {
	return su.HeaderString("name")
}

// Username returns the system user name that should be created (e.g. "foo").
func (su *SystemUser) Username() string {
	return su.HeaderString("username")
}

// Password returns the crypt(3) compatible password for the user.
// Note that only ID: $6$ or stronger is supported (sha512crypt).
func (su *SystemUser) Password() string {
	return su.HeaderString("password")
}

// SSHKeys returns the ssh keys for the user.
func (su *SystemUser) SSHKeys() []string {
	return su.sshKeys
}

// Since returns the time since the assertion is valid.
func (su *SystemUser) Since() time.Time {
	return su.since
}

// Until returns the time until the assertion is valid.
func (su *SystemUser) Until() time.Time {
	return su.until
}

// ValidAt returns whether the system-user is valid at 'when' time.
func (su *SystemUser) ValidAt(when time.Time) bool {
	valid := when.After(su.since) || when.Equal(su.since)
	if valid {
		valid = when.Before(su.until)
	}
	return valid
}

// Implement further consistency checks.
func (su *SystemUser) checkConsistency(db RODatabase, acck *AccountKey) error {
	// Do the cross-checks when this assertion is actually used,
	// i.e. in the create-user code. See also Model.checkConsitency

	return nil
}

// sanity
var _ consistencyChecker = (*SystemUser)(nil)

type shadow struct {
	ID     string
	Rounds string
	Salt   string
	Hash   string
}

// crypt(3) compatible hashes have the forms:
// - $id$salt$hash
// - $id$rounds=N$salt$hash
func parseShadowLine(line string) (*shadow, error) {
	l := strings.SplitN(line, "$", 5)
	if len(l) != 4 && len(l) != 5 {
		return nil, fmt.Errorf(`hashed password must be of the form "$integer-id$salt$hash", see crypt(3)`)
	}

	// if rounds is the second field, the line must consist of 4
	if strings.HasPrefix(l[2], "rounds=") && len(l) == 4 {
		return nil, fmt.Errorf(`missing hash field`)
	}

	// shadow line without $rounds=N$
	if len(l) == 4 {
		return &shadow{
			ID:   l[1],
			Salt: l[2],
			Hash: l[3],
		}, nil
	}
	// shadow line with rounds
	return &shadow{
		ID:     l[1],
		Rounds: l[2],
		Salt:   l[3],
		Hash:   l[4],
	}, nil
}

func checkHashedPassword(headers map[string]interface{}, name string) (string, error) {
	pw, err := checkOptionalString(headers, name)
	if err != nil {
		return "", err
	}
	// the pw string is optional, so just return if its empty
	if pw == "" {
		return "", nil
	}

	// parse the shadow line
	shd, err := parseShadowLine(pw)
	if err != nil {
		return "", fmt.Errorf(`%q header invalid: %s`, name, err)
	}

	// and verify it

	// see crypt(3), ID 6 means SHA-512 (since glibc 2.7)
	ID, err := strconv.Atoi(shd.ID)
	if err != nil {
		return "", fmt.Errorf(`%q header must start with "$integer-id$", got %q`, name, shd.ID)
	}
	// double check that we only allow modern hashes
	if ID < 6 {
		return "", fmt.Errorf("%q header only supports $id$ values of 6 (sha512crypt) or higher", name)
	}

	// the $rounds=N$ part is optional
	if strings.HasPrefix(shd.Rounds, "rounds=") {
		rounds, err := strconv.Atoi(strings.SplitN(shd.Rounds, "=", 2)[1])
		if err != nil {
			return "", fmt.Errorf("%q header has invalid number of rounds: %s", name, err)
		}
		if rounds < 5000 || rounds > 999999999 {
			return "", fmt.Errorf("%q header rounds parameter out of bounds: %d", name, rounds)
		}
	}

	// see crypt(3) for the legal chars
	validSaltAndHash := regexp.MustCompile(`^[a-zA-Z0-9./]+$`)
	if !validSaltAndHash.MatchString(shd.Salt) {
		return "", fmt.Errorf("%q header has invalid chars in salt %q", name, shd.Salt)
	}
	if !validSaltAndHash.MatchString(shd.Hash) {
		return "", fmt.Errorf("%q header has invalid chars in hash %q", name, shd.Hash)
	}

	return pw, nil
}

func assembleSystemUser(assert assertionBase) (Assertion, error) {
	// brand-id here can be different from authority-id,
	// the code using the assertion must use the policy set
	// by the model assertion system-user-authority header
	email, err := checkNotEmptyString(assert.headers, "email")
	if err != nil {
		return nil, err
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf(`"email" header must be a RFC 5322 compliant email address: %s`, err)
	}

	series, err := checkStringList(assert.headers, "series")
	if err != nil {
		return nil, err
	}
	models, err := checkStringList(assert.headers, "models")
	if err != nil {
		return nil, err
	}
	if _, err := checkOptionalString(assert.headers, "name"); err != nil {
		return nil, err
	}
	if _, err := checkStringMatches(assert.headers, "username", validSystemUserUsernames); err != nil {
		return nil, err
	}
	if _, err := checkHashedPassword(assert.headers, "password"); err != nil {
		return nil, err
	}

	sshKeys, err := checkStringList(assert.headers, "ssh-keys")
	if err != nil {
		return nil, err
	}
	since, err := checkRFC3339Date(assert.headers, "since")
	if err != nil {
		return nil, err
	}
	until, err := checkRFC3339Date(assert.headers, "until")
	if err != nil {
		return nil, err
	}
	if until.Before(since) {
		return nil, fmt.Errorf("'until' time cannot be before 'since' time")
	}

	// "global" system-user assertion can only be valid for 1y
	if len(models) == 0 && until.After(since.AddDate(1, 0, 0)) {
		return nil, fmt.Errorf("'until' time cannot be more than 365 days in the future when no models are specified")
	}

	return &SystemUser{
		assertionBase: assert,
		series:        series,
		models:        models,
		sshKeys:       sshKeys,
		since:         since,
		until:         until,
	}, nil
}
