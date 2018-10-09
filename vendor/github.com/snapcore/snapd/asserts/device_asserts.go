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
	"strings"
	"time"

	"github.com/snapcore/snapd/strutil"
)

// Model holds a model assertion, which is a statement by a brand
// about the properties of a device model.
type Model struct {
	assertionBase
	classic          bool
	requiredSnaps    []string
	sysUserAuthority []string
	timestamp        time.Time
}

// BrandID returns the brand identifier. Same as the authority id.
func (mod *Model) BrandID() string {
	return mod.HeaderString("brand-id")
}

// Model returns the model name identifier.
func (mod *Model) Model() string {
	return mod.HeaderString("model")
}

// DisplayName returns the human-friendly name of the model or
// falls back to Model if this was not set.
func (mod *Model) DisplayName() string {
	display := mod.HeaderString("display-name")
	if display == "" {
		return mod.Model()
	}
	return display
}

// Series returns the series of the core software the model uses.
func (mod *Model) Series() string {
	return mod.HeaderString("series")
}

// Classic returns whether the model is a classic system.
func (mod *Model) Classic() bool {
	return mod.classic
}

// Architecture returns the archicteture the model is based on.
func (mod *Model) Architecture() string {
	return mod.HeaderString("architecture")
}

// snapWithTrack represents a snap that includes optional track
// information like `snapName=trackName`
type snapWithTrack string

func (s snapWithTrack) Snap() string {
	return strings.SplitN(string(s), "=", 2)[0]
}

func (s snapWithTrack) Track() string {
	l := strings.SplitN(string(s), "=", 2)
	if len(l) > 1 {
		return l[1]
	}
	return ""
}

// Gadget returns the gadget snap the model uses.
func (mod *Model) Gadget() string {
	return snapWithTrack(mod.HeaderString("gadget")).Snap()
}

// GadgetTrack returns the gadget track the model uses.
func (mod *Model) GadgetTrack() string {
	return snapWithTrack(mod.HeaderString("gadget")).Track()
}

// Kernel returns the kernel snap the model uses.
func (mod *Model) Kernel() string {
	return snapWithTrack(mod.HeaderString("kernel")).Snap()
}

// KernelTrack returns the kernel track the model uses.
func (mod *Model) KernelTrack() string {
	return snapWithTrack(mod.HeaderString("kernel")).Track()
}

// Base returns the base snap the model uses.
func (mod *Model) Base() string {
	return mod.HeaderString("base")
}

// Store returns the snap store the model uses.
func (mod *Model) Store() string {
	return mod.HeaderString("store")
}

// RequiredSnaps returns the snaps that must be installed at all times and cannot be removed for this model.
func (mod *Model) RequiredSnaps() []string {
	return mod.requiredSnaps
}

// SystemUserAuthority returns the authority ids that are accepted as signers of system-user assertions for this model. Empty list means any.
func (mod *Model) SystemUserAuthority() []string {
	return mod.sysUserAuthority
}

// Timestamp returns the time when the model assertion was issued.
func (mod *Model) Timestamp() time.Time {
	return mod.timestamp
}

// Implement further consistency checks.
func (mod *Model) checkConsistency(db RODatabase, acck *AccountKey) error {
	// TODO: double check trust level of authority depending on class and possibly allowed-modes
	return nil
}

// sanity
var _ consistencyChecker = (*Model)(nil)

// limit model to only lowercase for now
var validModel = regexp.MustCompile("^[a-zA-Z0-9](?:-?[a-zA-Z0-9])*$")

func checkSnapWithTrackHeader(header string, headers map[string]interface{}) error {
	_, ok := headers[header]
	if !ok {
		return nil
	}
	value, ok := headers[header].(string)
	if !ok {
		return fmt.Errorf(`%q header must be a string`, header)
	}
	l := strings.SplitN(value, "=", 2)
	if len(l) == 1 {
		return nil
	}
	track := l[1]
	if strings.Count(track, "/") != 0 {
		return fmt.Errorf(`%q channel selector must be a track name only`, header)
	}
	channelRisks := []string{"stable", "candidate", "beta", "edge"}
	if strutil.ListContains(channelRisks, track) {
		return fmt.Errorf(`%q channel selector must be a track name`, header)
	}
	return nil
}

func checkModel(headers map[string]interface{}) (string, error) {
	s, err := checkStringMatches(headers, "model", validModel)
	if err != nil {
		return "", err
	}

	// TODO: support the concept of case insensitive/preserving string headers
	if strings.ToLower(s) != s {
		return "", fmt.Errorf(`"model" header cannot contain uppercase letters`)
	}
	return s, nil
}

func checkAuthorityMatchesBrand(a Assertion) error {
	typeName := a.Type().Name
	authorityID := a.AuthorityID()
	brand := a.HeaderString("brand-id")
	if brand != authorityID {
		return fmt.Errorf("authority-id and brand-id must match, %s assertions are expected to be signed by the brand: %q != %q", typeName, authorityID, brand)
	}
	return nil
}

func checkOptionalSystemUserAuthority(headers map[string]interface{}, brandID string) ([]string, error) {
	const name = "system-user-authority"
	v, ok := headers[name]
	if !ok {
		return []string{brandID}, nil
	}
	switch x := v.(type) {
	case string:
		if x == "*" {
			return nil, nil
		}
	case []interface{}:
		lst, err := checkStringListMatches(headers, name, validAccountID)
		if err == nil {
			return lst, nil
		}
	}
	return nil, fmt.Errorf("%q header must be '*' or a list of account ids", name)
}

var (
	modelMandatory       = []string{"architecture", "gadget", "kernel"}
	classicModelOptional = []string{"architecture", "gadget"}
)

func assembleModel(assert assertionBase) (Assertion, error) {
	err := checkAuthorityMatchesBrand(&assert)
	if err != nil {
		return nil, err
	}

	_, err = checkModel(assert.headers)
	if err != nil {
		return nil, err
	}

	classic, err := checkOptionalBool(assert.headers, "classic")
	if err != nil {
		return nil, err
	}

	if classic {
		if _, ok := assert.headers["kernel"]; ok {
			return nil, fmt.Errorf("cannot specify a kernel with a classic model")
		}
		if _, ok := assert.headers["base"]; ok {
			return nil, fmt.Errorf("cannot specify a base with a classic model")
		}
	}

	checker := checkNotEmptyString
	toCheck := modelMandatory
	if classic {
		checker = checkOptionalString
		toCheck = classicModelOptional
	}

	for _, h := range toCheck {
		if _, err := checker(assert.headers, h); err != nil {
			return nil, err
		}
	}

	// kernel/gadget can have (optional) tracks - validate those
	if err := checkSnapWithTrackHeader("kernel", assert.headers); err != nil {
		return nil, err
	}
	if err := checkSnapWithTrackHeader("gadget", assert.headers); err != nil {
		return nil, err
	}

	// store is optional but must be a string, defaults to the ubuntu store
	_, err = checkOptionalString(assert.headers, "store")
	if err != nil {
		return nil, err
	}

	// display-name is optional but must be a string
	_, err = checkOptionalString(assert.headers, "display-name")
	if err != nil {
		return nil, err
	}

	// TODO parallel-install: verify if snap names are valid store names
	reqSnaps, err := checkStringList(assert.headers, "required-snaps")
	if err != nil {
		return nil, err
	}

	sysUserAuthority, err := checkOptionalSystemUserAuthority(assert.headers, assert.HeaderString("brand-id"))
	if err != nil {
		return nil, err
	}

	timestamp, err := checkRFC3339Date(assert.headers, "timestamp")
	if err != nil {
		return nil, err
	}

	// NB:
	// * core is not supported at this time, it defaults to ubuntu-core
	// in prepare-image until rename and/or introduction of the header.
	// * some form of allowed-modes, class are postponed,
	//
	// prepare-image takes care of not allowing them for now

	// ignore extra headers and non-empty body for future compatibility
	return &Model{
		assertionBase:    assert,
		classic:          classic,
		requiredSnaps:    reqSnaps,
		sysUserAuthority: sysUserAuthority,
		timestamp:        timestamp,
	}, nil
}

// Serial holds a serial assertion, which is a statement binding a
// device identity with the device public key.
type Serial struct {
	assertionBase
	timestamp time.Time
	pubKey    PublicKey
}

// BrandID returns the brand identifier of the device.
func (ser *Serial) BrandID() string {
	return ser.HeaderString("brand-id")
}

// Model returns the model name identifier of the device.
func (ser *Serial) Model() string {
	return ser.HeaderString("model")
}

// Serial returns the serial identifier of the device, together with
// brand id and model they form the unique identifier of the device.
func (ser *Serial) Serial() string {
	return ser.HeaderString("serial")
}

// DeviceKey returns the public key of the device.
func (ser *Serial) DeviceKey() PublicKey {
	return ser.pubKey
}

// Timestamp returns the time when the serial assertion was issued.
func (ser *Serial) Timestamp() time.Time {
	return ser.timestamp
}

// TODO: implement further consistency checks for Serial but first review approach

func assembleSerial(assert assertionBase) (Assertion, error) {
	err := checkAuthorityMatchesBrand(&assert)
	if err != nil {
		return nil, err
	}

	_, err = checkModel(assert.headers)
	if err != nil {
		return nil, err
	}

	encodedKey, err := checkNotEmptyString(assert.headers, "device-key")
	if err != nil {
		return nil, err
	}
	pubKey, err := DecodePublicKey([]byte(encodedKey))
	if err != nil {
		return nil, err
	}
	keyID, err := checkNotEmptyString(assert.headers, "device-key-sha3-384")
	if err != nil {
		return nil, err
	}
	if keyID != pubKey.ID() {
		return nil, fmt.Errorf("device key does not match provided key id")
	}

	timestamp, err := checkRFC3339Date(assert.headers, "timestamp")
	if err != nil {
		return nil, err
	}

	// ignore extra headers and non-empty body for future compatibility
	return &Serial{
		assertionBase: assert,
		timestamp:     timestamp,
		pubKey:        pubKey,
	}, nil
}

// SerialRequest holds a serial-request assertion, which is a self-signed request to obtain a full device identity bound to the device public key.
type SerialRequest struct {
	assertionBase
	pubKey PublicKey
}

// BrandID returns the brand identifier of the device making the request.
func (sreq *SerialRequest) BrandID() string {
	return sreq.HeaderString("brand-id")
}

// Model returns the model name identifier of the device making the request.
func (sreq *SerialRequest) Model() string {
	return sreq.HeaderString("model")
}

// Serial returns the optional proposed serial identifier for the device, the service taking the request might use it or ignore it.
func (sreq *SerialRequest) Serial() string {
	return sreq.HeaderString("serial")
}

// RequestID returns the id for the request, obtained from and to be presented to the serial signing service.
func (sreq *SerialRequest) RequestID() string {
	return sreq.HeaderString("request-id")
}

// DeviceKey returns the public key of the device making the request.
func (sreq *SerialRequest) DeviceKey() PublicKey {
	return sreq.pubKey
}

func assembleSerialRequest(assert assertionBase) (Assertion, error) {
	_, err := checkNotEmptyString(assert.headers, "brand-id")
	if err != nil {
		return nil, err
	}

	_, err = checkModel(assert.headers)
	if err != nil {
		return nil, err
	}

	_, err = checkNotEmptyString(assert.headers, "request-id")
	if err != nil {
		return nil, err
	}

	_, err = checkOptionalString(assert.headers, "serial")
	if err != nil {
		return nil, err
	}

	encodedKey, err := checkNotEmptyString(assert.headers, "device-key")
	if err != nil {
		return nil, err
	}
	pubKey, err := DecodePublicKey([]byte(encodedKey))
	if err != nil {
		return nil, err
	}

	if pubKey.ID() != assert.SignKeyID() {
		return nil, fmt.Errorf("device key does not match included signing key id")
	}

	// ignore extra headers and non-empty body for future compatibility
	return &SerialRequest{
		assertionBase: assert,
		pubKey:        pubKey,
	}, nil
}

// DeviceSessionRequest holds a device-session-request assertion, which is a request wrapping a store-provided nonce to start a session by a device signed with its key.
type DeviceSessionRequest struct {
	assertionBase
	timestamp time.Time
}

// BrandID returns the brand identifier of the device making the request.
func (req *DeviceSessionRequest) BrandID() string {
	return req.HeaderString("brand-id")
}

// Model returns the model name identifier of the device making the request.
func (req *DeviceSessionRequest) Model() string {
	return req.HeaderString("model")
}

// Serial returns the serial identifier of the device making the request,
// together with brand id and model it forms the unique identifier of
// the device.
func (req *DeviceSessionRequest) Serial() string {
	return req.HeaderString("serial")
}

// Nonce returns the nonce obtained from store and to be presented when requesting a device session.
func (req *DeviceSessionRequest) Nonce() string {
	return req.HeaderString("nonce")
}

// Timestamp returns the time when the device-session-request was created.
func (req *DeviceSessionRequest) Timestamp() time.Time {
	return req.timestamp
}

func assembleDeviceSessionRequest(assert assertionBase) (Assertion, error) {
	_, err := checkModel(assert.headers)
	if err != nil {
		return nil, err
	}

	_, err = checkNotEmptyString(assert.headers, "nonce")
	if err != nil {
		return nil, err
	}

	timestamp, err := checkRFC3339Date(assert.headers, "timestamp")
	if err != nil {
		return nil, err
	}

	// ignore extra headers and non-empty body for future compatibility
	return &DeviceSessionRequest{
		assertionBase: assert,
		timestamp:     timestamp,
	}, nil
}
