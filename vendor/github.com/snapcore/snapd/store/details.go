// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
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
	"github.com/snapcore/snapd/jsonutil/safejson"
	"github.com/snapcore/snapd/snap"
)

// snapDetails encapsulates the data sent to us from the store as JSON.
type snapDetails struct {
	AnonDownloadURL  string             `json:"anon_download_url,omitempty"`
	Architectures    []string           `json:"architecture"`
	Channel          string             `json:"channel,omitempty"`
	DownloadSha3_384 string             `json:"download_sha3_384,omitempty"`
	Summary          safejson.String    `json:"summary,omitempty"`
	Description      safejson.Paragraph `json:"description,omitempty"`
	DownloadSize     int64              `json:"binary_filesize,omitempty"`
	DownloadURL      string             `json:"download_url,omitempty"`
	Epoch            snap.Epoch         `json:"epoch"`
	LastUpdated      string             `json:"last_updated,omitempty"`
	Name             string             `json:"package_name"`
	Prices           map[string]float64 `json:"prices,omitempty"`
	// Note that the publisher is really the "display name" of the
	// publisher
	Publisher      string           `json:"publisher,omitempty"`
	RatingsAverage float64          `json:"ratings_average,omitempty"`
	Revision       int              `json:"revision"` // store revisions are ints starting at 1
	SnapID         string           `json:"snap_id"`
	License        string           `json:"license,omitempty"`
	Base           string           `json:"base,omitempty"`
	Media          []storeSnapMedia `json:"media,omitempty"`

	// FIXME: the store should send "contact" here, once it does we
	//        can remove support_url
	SupportURL string `json:"support_url"`
	Contact    string `json:"contact"`

	Title   safejson.String `json:"title"`
	Type    snap.Type       `json:"content,omitempty"`
	Version string          `json:"version"`

	Developer           string `json:"origin"`
	DeveloperID         string `json:"developer_id"`
	DeveloperName       string `json:"developer_name"`
	DeveloperValidation string `json:"developer_validation"`

	Private     bool   `json:"private"`
	Confinement string `json:"confinement"`

	CommonIDs []string `json:"common_ids,omitempty"`
}

func infoFromRemote(d *snapDetails) *snap.Info {
	info := &snap.Info{}
	info.Architectures = d.Architectures
	info.Type = d.Type
	info.Version = d.Version
	info.Epoch = d.Epoch
	info.RealName = d.Name
	info.SnapID = d.SnapID
	info.Revision = snap.R(d.Revision)
	info.EditedTitle = d.Title.Clean()
	info.EditedSummary = d.Summary.Clean()
	info.EditedDescription = d.Description.Clean()
	// Note that the store side is using confusing terminology here.
	// What the store calls "developer" is actually the publisher
	// username.
	//
	// It also sends "publisher" and "developer_name" which are
	// the "publisher display name" which we cannot use currently
	// because it is not validated (i.e. the publisher could put
	// anything in there and mislead the users this way).
	info.Publisher = snap.StoreAccount{
		ID:          d.DeveloperID,
		Username:    d.Developer,
		DisplayName: d.DeveloperName,
		Validation:  d.DeveloperValidation,
	}
	info.Channel = d.Channel
	info.Sha3_384 = d.DownloadSha3_384
	info.Size = d.DownloadSize
	info.AnonDownloadURL = d.AnonDownloadURL
	info.DownloadURL = d.DownloadURL
	info.Prices = d.Prices
	info.Private = d.Private
	info.Paid = len(info.Prices) > 0
	info.Confinement = snap.ConfinementType(d.Confinement)
	info.Contact = d.Contact
	info.License = d.License
	info.Base = d.Base
	info.CommonIDs = d.CommonIDs

	addMedia(info, d.Media)

	// FIXME: once the store sends "contact" for everything, remove
	//        the "SupportURL" part of the if
	if info.Contact == "" {
		info.Contact = d.SupportURL
	}

	return info
}
