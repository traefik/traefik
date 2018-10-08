// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
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

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"

	"golang.org/x/net/context"
	"gopkg.in/macaroon.v1"

	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/asserts/sysdb"
	"github.com/snapcore/snapd/overlord/configstate/config"
	"github.com/snapcore/snapd/overlord/state"
)

// AuthState represents current authenticated users as tracked in state
type AuthState struct {
	LastID      int          `json:"last-id"`
	Users       []UserState  `json:"users"`
	Device      *DeviceState `json:"device,omitempty"`
	MacaroonKey []byte       `json:"macaroon-key,omitempty"`
}

// DeviceState represents the device's identity and store credentials
type DeviceState struct {
	Brand  string `json:"brand,omitempty"`
	Model  string `json:"model,omitempty"`
	Serial string `json:"serial,omitempty"`

	KeyID string `json:"key-id,omitempty"`

	SessionMacaroon string `json:"session-macaroon,omitempty"`
}

// UserState represents an authenticated user
type UserState struct {
	ID              int      `json:"id"`
	Username        string   `json:"username,omitempty"`
	Email           string   `json:"email,omitempty"`
	Macaroon        string   `json:"macaroon,omitempty"`
	Discharges      []string `json:"discharges,omitempty"`
	StoreMacaroon   string   `json:"store-macaroon,omitempty"`
	StoreDischarges []string `json:"store-discharges,omitempty"`
}

// HasStoreAuth returns true if the user has store authorization.
func (u *UserState) HasStoreAuth() bool {
	if u == nil {
		return false
	}
	return u.StoreMacaroon != ""
}

// MacaroonSerialize returns a store-compatible serialized representation of the given macaroon
func MacaroonSerialize(m *macaroon.Macaroon) (string, error) {
	marshalled, err := m.MarshalBinary()
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(marshalled)
	return encoded, nil
}

// MacaroonDeserialize returns a deserialized macaroon from a given store-compatible serialization
func MacaroonDeserialize(serializedMacaroon string) (*macaroon.Macaroon, error) {
	var m macaroon.Macaroon
	decoded, err := base64.RawURLEncoding.DecodeString(serializedMacaroon)
	if err != nil {
		return nil, err
	}
	err = m.UnmarshalBinary(decoded)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// generateMacaroonKey generates a random key to sign snapd macaroons
func generateMacaroonKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

const snapdMacaroonLocation = "snapd"

// newUserMacaroon returns a snapd macaroon for the given username
func newUserMacaroon(macaroonKey []byte, userID int) (string, error) {
	userMacaroon, err := macaroon.New(macaroonKey, strconv.Itoa(userID), snapdMacaroonLocation)
	if err != nil {
		return "", fmt.Errorf("cannot create macaroon for snapd user: %s", err)
	}

	serializedMacaroon, err := MacaroonSerialize(userMacaroon)
	if err != nil {
		return "", fmt.Errorf("cannot serialize macaroon for snapd user: %s", err)
	}

	return serializedMacaroon, nil
}

// NewUser tracks a new authenticated user and saves its details in the state
func NewUser(st *state.State, username, email, macaroon string, discharges []string) (*UserState, error) {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err == state.ErrNoState {
		authStateData = AuthState{}
	} else if err != nil {
		return nil, err
	}

	if authStateData.MacaroonKey == nil {
		authStateData.MacaroonKey, err = generateMacaroonKey()
		if err != nil {
			return nil, err
		}
	}

	authStateData.LastID++

	localMacaroon, err := newUserMacaroon(authStateData.MacaroonKey, authStateData.LastID)
	if err != nil {
		return nil, err
	}

	sort.Strings(discharges)
	authenticatedUser := UserState{
		ID:              authStateData.LastID,
		Username:        username,
		Email:           email,
		Macaroon:        localMacaroon,
		Discharges:      nil,
		StoreMacaroon:   macaroon,
		StoreDischarges: discharges,
	}
	authStateData.Users = append(authStateData.Users, authenticatedUser)

	st.Set("auth", authStateData)

	return &authenticatedUser, nil
}

var ErrInvalidUser = errors.New("invalid user")

// RemoveUser removes a user from the state given its ID
func RemoveUser(st *state.State, userID int) error {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err == state.ErrNoState {
		return ErrInvalidUser
	}
	if err != nil {
		return err
	}

	for i := range authStateData.Users {
		if authStateData.Users[i].ID == userID {
			// delete without preserving order
			n := len(authStateData.Users) - 1
			authStateData.Users[i] = authStateData.Users[n]
			authStateData.Users[n] = UserState{}
			authStateData.Users = authStateData.Users[:n]
			st.Set("auth", authStateData)
			return nil
		}
	}

	return ErrInvalidUser
}

func Users(st *state.State) ([]*UserState, error) {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err == state.ErrNoState {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	users := make([]*UserState, len(authStateData.Users))
	for i := range authStateData.Users {
		users[i] = &authStateData.Users[i]
	}
	return users, nil
}

// User returns a user from the state given its ID
func User(st *state.State, id int) (*UserState, error) {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err == state.ErrNoState {
		return nil, ErrInvalidUser
	}
	if err != nil {
		return nil, err
	}

	for _, user := range authStateData.Users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, ErrInvalidUser
}

// UpdateUser updates user in state
func UpdateUser(st *state.State, user *UserState) error {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err == state.ErrNoState {
		return ErrInvalidUser
	}
	if err != nil {
		return err
	}

	for i := range authStateData.Users {
		if authStateData.Users[i].ID == user.ID {
			authStateData.Users[i] = *user
			st.Set("auth", authStateData)
			return nil
		}
	}

	return ErrInvalidUser
}

// Device returns the device details from the state.
func Device(st *state.State) (*DeviceState, error) {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err == state.ErrNoState {
		return &DeviceState{}, nil
	} else if err != nil {
		return nil, err
	}

	if authStateData.Device == nil {
		return &DeviceState{}, nil
	}

	return authStateData.Device, nil
}

// SetDevice updates the device details in the state.
func SetDevice(st *state.State, device *DeviceState) error {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err == state.ErrNoState {
		authStateData = AuthState{}
	} else if err != nil {
		return err
	}

	authStateData.Device = device
	st.Set("auth", authStateData)

	return nil
}

var ErrInvalidAuth = fmt.Errorf("invalid authentication")

// CheckMacaroon returns the UserState for the given macaroon/discharges credentials
func CheckMacaroon(st *state.State, macaroon string, discharges []string) (*UserState, error) {
	var authStateData AuthState
	err := st.Get("auth", &authStateData)
	if err != nil {
		return nil, ErrInvalidAuth
	}

	snapdMacaroon, err := MacaroonDeserialize(macaroon)
	if err != nil {
		return nil, ErrInvalidAuth
	}
	// attempt snapd macaroon verification
	if snapdMacaroon.Location() == snapdMacaroonLocation {
		// no caveats to check so far
		check := func(caveat string) error { return nil }
		// ignoring discharges, unused for snapd macaroons atm
		err = snapdMacaroon.Verify(authStateData.MacaroonKey, check, nil)
		if err != nil {
			return nil, ErrInvalidAuth
		}
		macaroonID := snapdMacaroon.Id()
		userID, err := strconv.Atoi(macaroonID)
		if err != nil {
			return nil, ErrInvalidAuth
		}
		user, err := User(st, userID)
		if err != nil {
			return nil, ErrInvalidAuth
		}
		if macaroon != user.Macaroon {
			return nil, ErrInvalidAuth
		}
		return user, nil
	}

	// if macaroon is not a snapd macaroon, fallback to previous token-style check
NextUser:
	for _, user := range authStateData.Users {
		if user.Macaroon != macaroon {
			continue
		}
		if len(user.Discharges) != len(discharges) {
			continue
		}
		// sort discharges (stored users' discharges are already sorted)
		sort.Strings(discharges)
		for i, d := range user.Discharges {
			if d != discharges[i] {
				continue NextUser
			}
		}
		return &user, nil
	}
	return nil, ErrInvalidAuth
}

// DeviceSessionRequestParams gathers the assertions and information to be sent to request a device session.
type DeviceSessionRequestParams struct {
	Request *asserts.DeviceSessionRequest
	Serial  *asserts.Serial
	Model   *asserts.Model
}

func (p *DeviceSessionRequestParams) EncodedRequest() string {
	return string(asserts.Encode(p.Request))
}

func (p *DeviceSessionRequestParams) EncodedSerial() string {
	return string(asserts.Encode(p.Serial))
}

func (p *DeviceSessionRequestParams) EncodedModel() string {
	return string(asserts.Encode(p.Model))
}

// DeviceAssertions helps exposing the assertions about device identity.
// All methods should return state.ErrNoState if the underlying needed
// information is not (yet) available.
type DeviceAssertions interface {
	// Model returns the device model assertion.
	Model() (*asserts.Model, error)
	// Serial returns the device serial assertion.
	Serial() (*asserts.Serial, error)

	// DeviceSessionRequestParams produces a device-session-request with the given nonce, together with other required parameters, the device serial and model assertions.
	DeviceSessionRequestParams(nonce string) (*DeviceSessionRequestParams, error)
	// ProxyStore returns the store assertion for the proxy store if one is set.
	ProxyStore() (*asserts.Store, error)
}

var (
	// ErrNoSerial indicates that a device serial is not set yet.
	ErrNoSerial = errors.New("no device serial yet")
)

// CloudInfo reflects cloud information for the system (as captured in the core configuration).
type CloudInfo struct {
	Name             string `json:"name"`
	Region           string `json:"region,omitempty"`
	AvailabilityZone string `json:"availability-zone,omitempty"`
}

// TODO: move AuthContext to something like a storecontext package, it
// is about more than just authorization now.

// An AuthContext exposes authorization data and handles its updates.
type AuthContext interface {
	Device() (*DeviceState, error)

	UpdateDeviceAuth(device *DeviceState, sessionMacaroon string) (actual *DeviceState, err error)

	UpdateUserAuth(user *UserState, discharges []string) (actual *UserState, err error)

	StoreID(fallback string) (string, error)

	DeviceSessionRequestParams(nonce string) (*DeviceSessionRequestParams, error)
	ProxyStoreParams(defaultURL *url.URL) (proxyStoreID string, proxySroreURL *url.URL, err error)

	CloudInfo() (*CloudInfo, error)
}

// authContext helps keeping track of auth data in the state and exposing it.
type authContext struct {
	state         *state.State
	deviceAsserts DeviceAssertions
}

// NewAuthContext returns an AuthContext for state.
func NewAuthContext(st *state.State, deviceAsserts DeviceAssertions) AuthContext {
	return &authContext{state: st, deviceAsserts: deviceAsserts}
}

// Device returns current device state.
func (ac *authContext) Device() (*DeviceState, error) {
	ac.state.Lock()
	defer ac.state.Unlock()

	return Device(ac.state)
}

// UpdateDeviceAuth updates the device auth details in state.
// The last update wins but other device details are left unchanged.
// It returns the updated device state value.
func (ac *authContext) UpdateDeviceAuth(device *DeviceState, newSessionMacaroon string) (actual *DeviceState, err error) {
	ac.state.Lock()
	defer ac.state.Unlock()

	cur, err := Device(ac.state)
	if err != nil {
		return nil, err
	}

	// just do it, last update wins
	cur.SessionMacaroon = newSessionMacaroon
	if err := SetDevice(ac.state, cur); err != nil {
		return nil, fmt.Errorf("internal error: cannot update just read device state: %v", err)
	}

	return cur, nil
}

// UpdateUserAuth updates the user auth details in state.
// The last update wins but other user details are left unchanged.
// It returns the updated user state value.
func (ac *authContext) UpdateUserAuth(user *UserState, newDischarges []string) (actual *UserState, err error) {
	ac.state.Lock()
	defer ac.state.Unlock()

	cur, err := User(ac.state, user.ID)
	if err != nil {
		return nil, err
	}

	// just do it, last update wins
	cur.StoreDischarges = newDischarges
	if err := UpdateUser(ac.state, cur); err != nil {
		return nil, fmt.Errorf("internal error: cannot update just read user state: %v", err)
	}

	return cur, nil
}

// StoreID returns the store set in the model assertion, if mod != nil
// and it's not the generic classic model, or the override from the
// UBUNTU_STORE_ID envvar.
func StoreID(mod *asserts.Model) string {
	if mod != nil && mod.Ref().Unique() != sysdb.GenericClassicModel().Ref().Unique() {
		return mod.Store()
	}
	return os.Getenv("UBUNTU_STORE_ID")
}

// StoreID returns the store id according to system state or
// the fallback one if the state has none set (yet).
func (ac *authContext) StoreID(fallback string) (string, error) {
	var mod *asserts.Model
	if ac.deviceAsserts != nil {
		var err error
		mod, err = ac.deviceAsserts.Model()
		if err != nil && err != state.ErrNoState {
			return "", err
		}
	}
	storeID := StoreID(mod)
	if storeID != "" {
		return storeID, nil
	}
	return fallback, nil
}

// DeviceSessionRequestParams produces a device-session-request with the given nonce, together with other required parameters, the device serial and model assertions. It returns ErrNoSerial if the device serial is not yet initialized.
func (ac *authContext) DeviceSessionRequestParams(nonce string) (*DeviceSessionRequestParams, error) {
	if ac.deviceAsserts == nil {
		return nil, ErrNoSerial
	}
	params, err := ac.deviceAsserts.DeviceSessionRequestParams(nonce)
	if err == state.ErrNoState {
		return nil, ErrNoSerial
	}
	if err != nil {
		return nil, err
	}
	return params, nil
}

// ProxyStoreParams returns the id and URL of the proxy store if one is set. Returns the defaultURL otherwise and id = "".
func (ac *authContext) ProxyStoreParams(defaultURL *url.URL) (proxyStoreID string, proxySroreURL *url.URL, err error) {
	var sto *asserts.Store
	if ac.deviceAsserts != nil {
		var err error
		sto, err = ac.deviceAsserts.ProxyStore()
		if err != nil && err != state.ErrNoState {
			return "", nil, err
		}
	}
	if sto != nil {
		return sto.Store(), sto.URL(), nil
	}
	return "", defaultURL, nil
}

// CloudInfo returns the cloud instance information (if available).
func (ac *authContext) CloudInfo() (*CloudInfo, error) {
	ac.state.Lock()
	defer ac.state.Unlock()
	tr := config.NewTransaction(ac.state)
	var cloudInfo CloudInfo
	err := tr.Get("core", "cloud", &cloudInfo)
	if err != nil && !config.IsNoOption(err) {
		return nil, err
	}
	if cloudInfo.Name != "" {
		return &cloudInfo, nil
	}
	return nil, nil
}

type ensureContextKey struct{}

// EnsureContextTODO returns a provisional context marked as
// pertaining to an Ensure loop.
// TODO: see Overlord.Loop to replace it with a proper context passed to all Ensures.
func EnsureContextTODO() context.Context {
	ctx := context.TODO()
	return context.WithValue(ctx, ensureContextKey{}, struct{}{})
}

// IsEnsureContext returns whether context was marked as pertaining to an Ensure loop.
func IsEnsureContext(ctx context.Context) bool {
	return ctx.Value(ensureContextKey{}) != nil
}
