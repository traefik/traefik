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

package store

import (
	"fmt"
	"strconv"

	"github.com/snapcore/snapd/jsonutil/safejson"
	"github.com/snapcore/snapd/snap"
)

// storeSnap holds the information sent as JSON by the store for a snap.
type storeSnap struct {
	Architectures []string           `json:"architectures"`
	Base          string             `json:"base"`
	Confinement   string             `json:"confinement"`
	Contact       string             `json:"contact"`
	CreatedAt     string             `json:"created-at"` // revision timestamp
	Description   safejson.Paragraph `json:"description"`
	Download      storeSnapDownload  `json:"download"`
	Epoch         snap.Epoch         `json:"epoch"`
	License       string             `json:"license"`
	Name          string             `json:"name"`
	Prices        map[string]string  `json:"prices"` // currency->price,  free: {"USD": "0"}
	Private       bool               `json:"private"`
	Publisher     snap.StoreAccount  `json:"publisher"`
	Revision      int                `json:"revision"` // store revisions are ints starting at 1
	SnapID        string             `json:"snap-id"`
	SnapYAML      string             `json:"snap-yaml"` // optional
	Summary       safejson.String    `json:"summary"`
	Title         safejson.String    `json:"title"`
	Type          snap.Type          `json:"type"`
	Version       string             `json:"version"`

	// TODO: not yet defined: channel map

	// media
	Media []storeSnapMedia `json:"media"`

	CommonIDs []string `json:"common-ids"`
}

type storeSnapDownload struct {
	Sha3_384 string           `json:"sha3-384"`
	Size     int64            `json:"size"`
	URL      string           `json:"url"`
	Deltas   []storeSnapDelta `json:"deltas"`
}

type storeSnapDelta struct {
	Format   string `json:"format"`
	Sha3_384 string `json:"sha3-384"`
	Size     int64  `json:"size"`
	Source   int    `json:"source"`
	Target   int    `json:"target"`
	URL      string `json:"url"`
}

type storeSnapMedia struct {
	Type   string `json:"type"` // icon/screenshot
	URL    string `json:"url"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}

// storeInfoChannel is the channel description included in info results
type storeInfoChannel struct {
	Architecture string `json:"architecture"`
	Name         string `json:"name"`
	Risk         string `json:"risk"`
	Track        string `json:"track"`
}

// storeInfoChannelSnap is the snap-in-a-channel of which the channel map is made
type storeInfoChannelSnap struct {
	storeSnap
	Channel storeInfoChannel `json:"channel"`
}

// storeInfo is the result of v2/info calls
type storeInfo struct {
	ChannelMap []*storeInfoChannelSnap `json:"channel-map"`
	Snap       storeSnap               `json:"snap"`
	Name       string                  `json:"name"`
	SnapID     string                  `json:"snap-id"`
}

func infoFromStoreInfo(si *storeInfo) (*snap.Info, error) {
	if len(si.ChannelMap) == 0 {
		// if a snap has no released revisions, it _could_ be returned
		// (currently no, but spec is purposely ambiguous)
		// we treat it as a 'not found' for now at least
		return nil, ErrSnapNotFound
	}

	thisOne := si.ChannelMap[0]
	thisSnap := thisOne.storeSnap // copy it as we're about to modify it
	// here we assume that the ChannelSnapInfo can be populated with data
	// that's in the channel map and not the outer snap. This is a
	// reasonable assumption today, but copyNonZeroFrom can easily be
	// changed to copy to a list if needed.
	copyNonZeroFrom(&si.Snap, &thisSnap)

	info, err := infoFromStoreSnap(&thisSnap)
	if err != nil {
		return nil, err
	}
	info.Channel = thisOne.Channel.Name
	info.Channels = make(map[string]*snap.ChannelSnapInfo, len(si.ChannelMap))
	seen := make(map[string]bool, len(si.ChannelMap))
	for _, s := range si.ChannelMap {
		ch := s.Channel
		info.Channels[ch.Track+"/"+ch.Risk] = &snap.ChannelSnapInfo{
			Revision:    snap.R(s.Revision),
			Confinement: snap.ConfinementType(s.Confinement),
			Version:     s.Version,
			Channel:     ch.Name,
			Epoch:       s.Epoch,
			Size:        s.Download.Size,
		}
		if !seen[ch.Track] {
			seen[ch.Track] = true
			info.Tracks = append(info.Tracks, ch.Track)
		}
	}

	return info, nil
}

// copy non-zero fields from src to dst
func copyNonZeroFrom(src, dst *storeSnap) {
	if len(src.Architectures) > 0 {
		dst.Architectures = src.Architectures
	}
	if src.Base != "" {
		dst.Base = src.Base
	}
	if src.Confinement != "" {
		dst.Confinement = src.Confinement
	}
	if src.Contact != "" {
		dst.Contact = src.Contact
	}
	if src.CreatedAt != "" {
		dst.CreatedAt = src.CreatedAt
	}
	if src.Description.Clean() != "" {
		dst.Description = src.Description
	}
	if src.Download.URL != "" {
		dst.Download = src.Download
	}
	if src.Epoch.String() != "0" {
		dst.Epoch = src.Epoch
	}
	if src.License != "" {
		dst.License = src.License
	}
	if src.Name != "" {
		dst.Name = src.Name
	}
	if len(src.Prices) > 0 {
		dst.Prices = src.Prices
	}
	if src.Private {
		dst.Private = src.Private
	}
	if src.Publisher.ID != "" {
		dst.Publisher = src.Publisher
	}
	if src.Revision > 0 {
		dst.Revision = src.Revision
	}
	if src.SnapID != "" {
		dst.SnapID = src.SnapID
	}
	if src.SnapYAML != "" {
		dst.SnapYAML = src.SnapYAML
	}
	if src.Summary.Clean() != "" {
		dst.Summary = src.Summary
	}
	if src.Title.Clean() != "" {
		dst.Title = src.Title
	}
	if src.Type != "" {
		dst.Type = src.Type
	}
	if src.Version != "" {
		dst.Version = src.Version
	}
	if len(src.Media) > 0 {
		dst.Media = src.Media
	}
	if len(src.CommonIDs) > 0 {
		dst.CommonIDs = src.CommonIDs
	}
}

func infoFromStoreSnap(d *storeSnap) (*snap.Info, error) {
	info := &snap.Info{}
	info.RealName = d.Name
	info.Revision = snap.R(d.Revision)
	info.SnapID = d.SnapID
	info.EditedTitle = d.Title.Clean()
	info.EditedSummary = d.Summary.Clean()
	info.EditedDescription = d.Description.Clean()
	info.Private = d.Private
	info.Contact = d.Contact
	info.Architectures = d.Architectures
	info.Type = d.Type
	info.Version = d.Version
	info.Epoch = d.Epoch
	info.Confinement = snap.ConfinementType(d.Confinement)
	info.Base = d.Base
	info.License = d.License
	info.Publisher = d.Publisher
	info.DownloadURL = d.Download.URL
	info.Size = d.Download.Size
	info.Sha3_384 = d.Download.Sha3_384
	if len(d.Download.Deltas) > 0 {
		deltas := make([]snap.DeltaInfo, len(d.Download.Deltas))
		for i, d := range d.Download.Deltas {
			deltas[i] = snap.DeltaInfo{
				FromRevision: d.Source,
				ToRevision:   d.Target,
				Format:       d.Format,
				DownloadURL:  d.URL,
				Size:         d.Size,
				Sha3_384:     d.Sha3_384,
			}
		}
		info.Deltas = deltas
	}
	info.CommonIDs = d.CommonIDs

	// fill in the plug/slot data
	if rawYamlInfo, err := snap.InfoFromSnapYaml([]byte(d.SnapYAML)); err == nil {
		if info.Plugs == nil {
			info.Plugs = make(map[string]*snap.PlugInfo)
		}
		for k, v := range rawYamlInfo.Plugs {
			info.Plugs[k] = v
			info.Plugs[k].Snap = info
		}
		if info.Slots == nil {
			info.Slots = make(map[string]*snap.SlotInfo)
		}
		for k, v := range rawYamlInfo.Slots {
			info.Slots[k] = v
			info.Slots[k].Snap = info
		}
	}

	// convert prices
	if len(d.Prices) > 0 {
		prices := make(map[string]float64, len(d.Prices))
		for currency, priceStr := range d.Prices {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse snap price: %v", err)
			}
			prices[currency] = price
		}
		info.Paid = true
		info.Prices = prices
	}

	// media
	addMedia(info, d.Media)

	return info, nil
}

func addMedia(info *snap.Info, media []storeSnapMedia) {
	if len(media) == 0 {
		return
	}
	info.Media = make(snap.MediaInfos, len(media))
	for i, mediaObj := range media {
		info.Media[i].Type = mediaObj.Type
		info.Media[i].URL = mediaObj.URL
		info.Media[i].Width = mediaObj.Width
		info.Media[i].Height = mediaObj.Height
	}
}
