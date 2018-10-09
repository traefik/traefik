// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/snapcore/snapd/snap"
)

var (
	ErrSnapshotSetNotFound   = errors.New("no snapshot set with the given ID")
	ErrSnapshotSnapsNotFound = errors.New("no snapshot for the requested snaps found in the set with the given ID")
)

// A snapshotAction is used to request an operation on a snapshot.
type snapshotAction struct {
	SetID  uint64   `json:"set"`
	Action string   `json:"action"`
	Snaps  []string `json:"snaps,omitempty"`
	Users  []string `json:"users,omitempty"`
}

// A Snapshot is a collection of archives with a simple metadata json file
// (and hashsums of everything).
type Snapshot struct {
	// SetID is the ID of the snapshot set (a snapshot set is the result of a "snap save" invocation)
	SetID uint64 `json:"set"`
	// the time this snapshot's data collection was started
	Time time.Time `json:"time"`

	// information about the snap this data is for
	Snap     string        `json:"snap"`
	Revision snap.Revision `json:"revision"`
	SnapID   string        `json:"snap-id,omitempty"`
	Epoch    snap.Epoch    `json:"epoch,omitempty"`
	Summary  string        `json:"summary"`
	Version  string        `json:"version"`

	// the snap's configuration at snapshot time
	Conf map[string]interface{} `json:"conf,omitempty"`

	// the hash of the archives' data, keyed by archive path
	// (either 'archive.tgz' for the system archive, or
	// user/<username>.tgz for each user)
	SHA3_384 map[string]string `json:"sha3-384"`
	// the sum of the archive sizes
	Size int64 `json:"size,omitempty"`
	// if the snapshot failed to open this will be the reason why
	Broken string `json:"broken,omitempty"`
}

// IsValid checks whether the snapshot is missing information that
// should be there for a snapshot that's just been opened.
func (sh *Snapshot) IsValid() bool {
	return !(sh == nil || sh.SetID == 0 || sh.Snap == "" || sh.Revision.Unset() || len(sh.SHA3_384) == 0 || sh.Time.IsZero())
}

// A SnapshotSet is a set of snapshots created by a single "snap save".
type SnapshotSet struct {
	ID        uint64      `json:"id"`
	Snapshots []*Snapshot `json:"snapshots"`
}

// Time returns the earliest time in the set.
func (ss SnapshotSet) Time() time.Time {
	if len(ss.Snapshots) == 0 {
		return time.Time{}
	}
	mint := ss.Snapshots[0].Time
	for _, sh := range ss.Snapshots {
		if sh.Time.Before(mint) {
			mint = sh.Time
		}
	}
	return mint
}

// Size returns the sum of the set's sizes.
func (ss SnapshotSet) Size() int64 {
	var sum int64
	for _, sh := range ss.Snapshots {
		sum += sh.Size
	}
	return sum
}

// SnapshotSets lists the snapshot sets in the system that belong to the
// given set (if non-zero) and are for the given snaps (if non-empty).
func (client *Client) SnapshotSets(setID uint64, snapNames []string) ([]SnapshotSet, error) {
	q := make(url.Values)
	if setID > 0 {
		q.Add("set", strconv.FormatUint(setID, 10))
	}
	if len(snapNames) > 0 {
		q.Add("snaps", strings.Join(snapNames, ","))
	}

	var snapshotSets []SnapshotSet
	_, err := client.doSync("GET", "/v2/snapshots", q, nil, nil, &snapshotSets)
	return snapshotSets, err
}

// ForgetSnapshots permanently removes the snapshot set, limited to the
// given snaps (if non-empty).
func (client *Client) ForgetSnapshots(setID uint64, snaps []string) (changeID string, err error) {
	return client.snapshotAction(&snapshotAction{
		SetID:  setID,
		Action: "forget",
		Snaps:  snaps,
	})
}

// CheckSnapshots verifies the archive checksums in the given snapshot set.
//
// If snaps or users are non-empty, limit to checking only those
// archives of the snapshot.
func (client *Client) CheckSnapshots(setID uint64, snaps []string, users []string) (changeID string, err error) {
	return client.snapshotAction(&snapshotAction{
		SetID:  setID,
		Action: "check",
		Snaps:  snaps,
		Users:  users,
	})
}

// RestoreSnapshots extracts the given snapshot set.
//
// If snaps or users are non-empty, limit to checking only those
// archives of the snapshot.
func (client *Client) RestoreSnapshots(setID uint64, snaps []string, users []string) (changeID string, err error) {
	return client.snapshotAction(&snapshotAction{
		SetID:  setID,
		Action: "restore",
		Snaps:  snaps,
		Users:  users,
	})
}

func (client *Client) snapshotAction(action *snapshotAction) (changeID string, err error) {
	data, err := json.Marshal(action)
	if err != nil {
		return "", fmt.Errorf("cannot marshal snapshot action: %v", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return client.doAsync("POST", "/v2/snapshots", nil, headers, bytes.NewBuffer(data))
}
