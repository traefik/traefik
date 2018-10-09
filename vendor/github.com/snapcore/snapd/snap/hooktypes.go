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

package snap

import (
	"regexp"
)

var supportedHooks = []*HookType{
	NewHookType(regexp.MustCompile("^prepare-device$")),
	NewHookType(regexp.MustCompile("^configure$")),
	NewHookType(regexp.MustCompile("^install$")),
	NewHookType(regexp.MustCompile("^pre-refresh$")),
	NewHookType(regexp.MustCompile("^post-refresh$")),
	NewHookType(regexp.MustCompile("^remove$")),
	NewHookType(regexp.MustCompile("^prepare-(?:plug|slot)-[-a-z0-9]+$")),
	NewHookType(regexp.MustCompile("^unprepare-(?:plug|slot)-[-a-z0-9]+$")),
	NewHookType(regexp.MustCompile("^connect-(?:plug|slot)-[-a-z0-9]+$")),
	NewHookType(regexp.MustCompile("^disconnect-(?:plug|slot)-[-a-z0-9]+$")),
}

// HookType represents a pattern of supported hook names.
type HookType struct {
	pattern *regexp.Regexp
}

// NewHookType returns a new HookType with the given pattern.
func NewHookType(pattern *regexp.Regexp) *HookType {
	return &HookType{
		pattern: pattern,
	}
}

// Match returns true if the given hook name matches this hook type.
func (hookType HookType) Match(hookName string) bool {
	return hookType.pattern.MatchString(hookName)
}

// IsHookSupported returns true if the given hook name matches one of the
// supported hooks.
func IsHookSupported(hookName string) bool {
	for _, hookType := range supportedHooks {
		if hookType.Match(hookName) {
			return true
		}
	}

	return false
}

func MockSupportedHookTypes(hookTypes []*HookType) (restore func()) {
	old := supportedHooks
	supportedHooks = hookTypes
	return func() { supportedHooks = old }
}

func MockAppendSupportedHookTypes(hookTypes []*HookType) (restore func()) {
	old := supportedHooks
	supportedHooks = append(supportedHooks, hookTypes...)
	return func() { supportedHooks = old }
}
