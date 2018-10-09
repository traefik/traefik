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

package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/snapcore/snapd/logger"
)

var (
	DefaultRepeatAfter = time.Hour * 24
	DefaultExpireAfter = time.Hour * 24 * 28

	errNoWarningMessage     = errors.New("warning has no message")
	errBadWarningMessage    = errors.New("malformed warning message")
	errNoWarningFirstAdded  = errors.New("warning has no first-added timestamp")
	errNoWarningExpireAfter = errors.New("warning has no expire-after duration")
	errNoWarningRepeatAfter = errors.New("warning has no repeat-after duration")
)

type jsonWarning struct {
	Message     string     `json:"message"`
	FirstAdded  time.Time  `json:"first-added"`
	LastAdded   time.Time  `json:"last-added"`
	LastShown   *time.Time `json:"last-shown,omitempty"`
	ExpireAfter string     `json:"expire-after,omitempty"`
	RepeatAfter string     `json:"repeat-after,omitempty"`
}

type Warning struct {
	// the warning text itself. Only one of these in the system at a time.
	message string
	// the first time one of these messages was created
	firstAdded time.Time
	// the last time one of these was created
	lastAdded time.Time
	// the last time one of these was shown to the user
	lastShown time.Time
	// how much time since one of these was last added should we drop the message
	expireAfter time.Duration
	// how much time since one of these was last shown should we repeat it
	repeatAfter time.Duration
}

func (w *Warning) String() string {
	return w.message
}

func (w *Warning) MarshalJSON() ([]byte, error) {
	jw := jsonWarning{
		Message:     w.message,
		FirstAdded:  w.firstAdded,
		LastAdded:   w.lastAdded,
		ExpireAfter: w.expireAfter.String(),
		RepeatAfter: w.repeatAfter.String(),
	}
	if !w.lastShown.IsZero() {
		jw.LastShown = &w.lastShown
	}

	return json.Marshal(jw)
}

func (w *Warning) UnmarshalJSON(data []byte) error {
	var jw jsonWarning
	err := json.Unmarshal(data, &jw)
	if err != nil {
		return err
	}
	w.message = jw.Message
	w.firstAdded = jw.FirstAdded
	w.lastAdded = jw.LastAdded
	if jw.LastShown != nil {
		w.lastShown = *jw.LastShown
	}
	if jw.ExpireAfter != "" {
		w.expireAfter, err = time.ParseDuration(jw.ExpireAfter)
		if err != nil {
			return err
		}
	}
	if jw.RepeatAfter != "" {
		w.repeatAfter, err = time.ParseDuration(jw.RepeatAfter)
		if err != nil {
			return err
		}
	}

	return w.validate()
}

func (w *Warning) validate() (e error) {
	if w.message == "" {
		return errNoWarningMessage
	}
	if strings.TrimSpace(w.message) != w.message {
		return errBadWarningMessage
	}
	if w.firstAdded.IsZero() {
		return errNoWarningFirstAdded
	}
	if w.expireAfter == 0 {
		return errNoWarningExpireAfter
	}
	if w.repeatAfter == 0 {
		return errNoWarningRepeatAfter
	}
	return nil
}

func (w *Warning) ExpiredBefore(now time.Time) bool {
	return w.lastAdded.Add(w.expireAfter).Before(now)
}

func (w *Warning) ShowAfter(t time.Time) bool {
	if w.lastShown.IsZero() {
		// warning was never shown before; was it added after the cutoff?
		return !w.firstAdded.After(t)
	}

	return w.lastShown.Add(w.repeatAfter).Before(t)
}

// flattenWarning loops over the warnings map, and returns all
// non-expired warnings therein as a flat list, for serialising.
// Call with the lock held.
func (s *State) flattenWarnings() []*Warning {
	now := time.Now()
	flat := make([]*Warning, 0, len(s.warnings))
	for _, w := range s.warnings {
		if w.ExpiredBefore(now) {
			continue
		}
		flat = append(flat, w)
	}
	return flat
}

// unflattenWarnings takes a flat list of warnings and replaces the
// warning map with them, ignoring expired warnings in the process.
// Call with the lock held.
func (s *State) unflattenWarnings(flat []*Warning) {
	now := time.Now()
	s.warnings = make(map[string]*Warning, len(flat))
	for _, w := range flat {
		if w.ExpiredBefore(now) {
			continue
		}
		s.warnings[w.message] = w
	}
}

// Warnf records a warning: if it's the first Warning with this
// message it'll be added (with its firstAdded and lastAdded set to the
// current time), otherwise the existing one will have its lastAdded
// updated.
func (s *State) Warnf(template string, args ...interface{}) {
	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(template, args...)
	} else {
		message = template
	}
	s.addWarning(Warning{
		message:     message,
		expireAfter: DefaultExpireAfter,
		repeatAfter: DefaultRepeatAfter,
	}, time.Now().UTC())
}

func (s *State) addWarning(w Warning, t time.Time) {
	s.writing()

	if s.warnings[w.message] == nil {
		w.firstAdded = t
		if err := w.validate(); err != nil {
			// programming error!
			logger.Panicf("internal error, please report: attempted to add invalid warning: %v", err)
			return
		}
		s.warnings[w.message] = &w
	}
	s.warnings[w.message].lastAdded = t
}

type byLastAdded []*Warning

func (a byLastAdded) Len() int           { return len(a) }
func (a byLastAdded) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLastAdded) Less(i, j int) bool { return a[i].lastAdded.Before(a[j].lastAdded) }

// AllWarnings returns all the warnings in the system, whether they're
// due to be shown or not. They'll be sorted by lastAdded.
func (s *State) AllWarnings() []*Warning {
	s.reading()

	all := s.flattenWarnings()
	sort.Sort(byLastAdded(all))

	return all
}

// OkayWarnings marks warnings that were showable at the given time as shown.
func (s *State) OkayWarnings(t time.Time) int {
	t = t.UTC()
	s.writing()

	n := 0
	for _, w := range s.warnings {
		if w.ShowAfter(t) {
			w.lastShown = t
			n++
		}
	}

	return n
}

// PendingWarnings returns the list of warnings to show the user, sorted by
// lastAdded, and a timestamp than can be used to refer to these warnings.
//
// Warnings to show to the user are those that have not been shown before,
// or that have been shown earlier than repeatAfter ago.
func (s *State) PendingWarnings() ([]*Warning, time.Time) {
	s.reading()
	now := time.Now().UTC()

	var toShow []*Warning
	for _, w := range s.warnings {
		if !w.ShowAfter(now) {
			continue
		}
		toShow = append(toShow, w)
	}

	sort.Sort(byLastAdded(toShow))
	return toShow, now
}

// WarningsSummary returns the number of warnings that are ready to be
// shown to the user, and the timestamp of the most recently added
// warning (useful for silencing the warning alerts, and OKing the
// returned warnings).
func (s *State) WarningsSummary() (int, time.Time) {
	s.reading()
	now := time.Now().UTC()
	var last time.Time

	var n int
	for _, w := range s.warnings {
		if w.ShowAfter(now) {
			n++
			if w.lastAdded.After(last) {
				last = w.lastAdded
			}
		}
	}

	return n, last
}

// UnshowAllWarnings clears the lastShown timestamp from all the
// warnings. For use in debugging.
func (s *State) UnshowAllWarnings() {
	s.writing()
	for _, w := range s.warnings {
		w.lastShown = time.Time{}
	}
}
