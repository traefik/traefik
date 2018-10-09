// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2018 Canonical Ltd
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
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/strutil"
)

var (
	// ErrBadQuery is returned from Find when the query has special characters in strange places.
	ErrBadQuery = errors.New("bad query")

	// ErrSnapNotFound is returned when a snap can not be found
	ErrSnapNotFound = errors.New("snap not found")

	// ErrUnauthenticated is returned when authentication is needed to complete the query
	ErrUnauthenticated = errors.New("you need to log in first")

	// ErrAuthenticationNeeds2fa is returned if the authentication needs 2factor
	ErrAuthenticationNeeds2fa = errors.New("two factor authentication required")

	// Err2faFailed is returned when 2fa failed (e.g., a bad token was given)
	Err2faFailed = errors.New("two factor authentication failed")

	// ErrInvalidCredentials is returned on login error
	// It can also be returned when refreshing the discharge
	// macaroon if the user has changed their password.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrTOSNotAccepted is returned when the user has not accepted the store's terms of service.
	ErrTOSNotAccepted = errors.New("terms of service not accepted")

	// ErrNoPaymentMethods is returned when the user has no valid payment methods associated with their account.
	ErrNoPaymentMethods = errors.New("no payment methods")

	// ErrPaymentDeclined is returned when the user's payment method was declined by the upstream payment provider.
	ErrPaymentDeclined = errors.New("payment declined")

	// ErrLocalSnap is returned when an operation that only applies to snaps that come from a store was attempted on a local snap.
	ErrLocalSnap = errors.New("cannot perform operation on local snap")

	// ErrNoUpdateAvailable is returned when an update is attempetd for a snap that has no update available.
	ErrNoUpdateAvailable = errors.New("snap has no updates available")
)

// RevisionNotAvailableError is returned when an install is attempted for a snap but the/a revision is not available (given install constraints).
type RevisionNotAvailableError struct {
	Action   string
	Channel  string
	Releases []snap.Channel
}

func (e *RevisionNotAvailableError) Error() string {
	return "no snap revision available as specified"
}

// DownloadError represents a download error
type DownloadError struct {
	Code int
	URL  *url.URL
}

func (e *DownloadError) Error() string {
	return fmt.Sprintf("received an unexpected http response code (%v) when trying to download %s", e.Code, e.URL)
}

// PasswordPolicyError is returned in a few corner cases, most notably
// when the password has been force-reset.
type PasswordPolicyError map[string]stringList

func (e PasswordPolicyError) Error() string {
	var msg string

	if reason, ok := e["reason"]; ok && len(reason) == 1 {
		msg = reason[0]
		if location, ok := e["location"]; ok && len(location) == 1 {
			msg += "\nTo address this, go to: " + location[0] + "\n"
		}
	} else {
		for k, vs := range e {
			msg += fmt.Sprintf("%s: %s\n", k, strings.Join(vs, "  "))
		}
	}

	return msg
}

// InvalidAuthDataError signals that the authentication data didn't pass validation.
type InvalidAuthDataError map[string]stringList

func (e InvalidAuthDataError) Error() string {
	var es []string
	for _, v := range e {
		es = append(es, v...)
	}
	// XXX: confirm with server people that extra args are all
	//      full sentences (with periods and capitalization)
	//      (empirically this checks out)
	return strings.Join(es, "  ")
}

// SnapActionError conveys errors that were reported on otherwise overall successful snap action (install/refresh) request.
type SnapActionError struct {
	// NoResults is set if the there were no results in the response
	NoResults bool
	// Refresh errors by snap name.
	Refresh map[string]error
	// Install errors by snap name.
	Install map[string]error
	// Download errors by snap name.
	Download map[string]error
	// Other errors.
	Other []error
}

func (e SnapActionError) Error() string {
	nRefresh := len(e.Refresh)
	nInstall := len(e.Install)
	nDownload := len(e.Download)
	nOther := len(e.Other)

	// single error
	switch nRefresh + nInstall + nDownload + nOther {
	case 0:
		if e.NoResults {
			// this is an atypical result
			return "no install/refresh information results from the store"
		}
	case 1:
		if nOther == 0 {
			var op string
			var errs map[string]error
			switch {
			case nRefresh > 0:
				op = "refresh"
				errs = e.Refresh
			case nInstall > 0:
				op = "install"
				errs = e.Install
			case nDownload > 0:
				op = "download"
				errs = e.Download
			}
			for name, e := range errs {
				return fmt.Sprintf("cannot %s snap %q: %v", op, name, e)
			}
		} else {
			return fmt.Sprintf("cannot refresh, install, or download: %v", e.Other[0])
		}
	}

	header := "cannot refresh, install, or download:"
	if nOther == 0 {
		// at least one of nDownload, nInstall, or nRefresh is > 0
		switch {
		case nDownload == 0 && nRefresh == 0:
			header = "cannot install:"
		case nDownload == 0 && nInstall == 0:
			header = "cannot refresh:"
		case nRefresh == 0 && nInstall == 0:
			header = "cannot download:"
		case nDownload == 0:
			header = "cannot refresh or install:"
		case nInstall == 0:
			header = "cannot refresh or download:"
		case nRefresh == 0:
			header = "cannot install or download:"
		}
	}

	// reverse the "snap->error" map to "error->snap", as the
	// common case is that all snaps fail with the same error
	// (e.g. "no refresh available")
	errToSnaps := map[string][]string{}
	errKeys := []string{} // poorman's ordered map

	for _, m := range []map[string]error{e.Refresh, e.Install, e.Download} {
		for snapName, err := range m {
			k := err.Error()
			v, ok := errToSnaps[k]
			if !ok {
				errKeys = append(errKeys, k)
			}
			errToSnaps[k] = append(v, snapName)
		}
	}

	es := make([]string, 1, 1+len(errToSnaps)+nOther)
	es[0] = header
	for _, k := range errKeys {
		sort.Strings(errToSnaps[k])
		es = append(es, fmt.Sprintf("%s: %s", k, strutil.Quoted(errToSnaps[k])))
	}

	for _, e := range e.Other {
		es = append(es, e.Error())
	}

	if len(es) == 2 {
		// header + 1 reason
		return strings.Join(es, " ")
	}

	return strings.Join(es, "\n")
}

// Authorization soft-expiry errors that get handled automatically.
var (
	errUserAuthorizationNeedsRefresh   = errors.New("soft-expired user authorization needs refresh")
	errDeviceAuthorizationNeedsRefresh = errors.New("soft-expired device authorization needs refresh")
)

func translateSnapActionError(action, channel, code, message string, releases []snapRelease) error {
	switch code {
	case "revision-not-found":
		e := &RevisionNotAvailableError{
			Action:  action,
			Channel: channel,
		}
		if len(releases) != 0 {
			parsedReleases := make([]snap.Channel, len(releases))
			for i := 0; i < len(releases); i++ {
				var err error
				parsedReleases[i], err = snap.ParseChannel(releases[i].Channel, releases[i].Architecture)
				if err != nil {
					// shouldn't happen, return error without Releases
					return e
				}
			}
			e.Releases = parsedReleases
		}
		return e
	case "id-not-found", "name-not-found":
		return ErrSnapNotFound
	case "user-authorization-needs-refresh":
		return errUserAuthorizationNeedsRefresh
	case "device-authorization-needs-refresh":
		return errDeviceAuthorizationNeedsRefresh
	default:
		return fmt.Errorf("%v", message)
	}
}
