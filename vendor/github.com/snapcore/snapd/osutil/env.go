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

package osutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetenvBool returns whether the given key may be considered "set" in the
// environment (i.e. it is set to one of "1", "true", etc).
//
// An optional second argument can be provided, which determines how to
// treat missing or unparsable values; default is to treat them as false.
func GetenvBool(key string, dflt ...bool) bool {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}

	if len(dflt) > 0 {
		return dflt[0]
	}

	return false
}

// GetenvInt64 interprets the value of the given environment variable
// as an int64 and returns the corresponding value. The base can be
// implied via the prefix (0x for 16, 0 for 8; otherwise 10).
//
// An optional second argument can be provided, which determines how to
// treat missing or unparsable values; default is to treat them as 0.
func GetenvInt64(key string, dflt ...int64) int64 {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		if b, err := strconv.ParseInt(val, 0, 64); err == nil {
			return b
		}
	}

	if len(dflt) > 0 {
		return dflt[0]
	}

	return 0
}

// EnvMap takes a list of "key=value" strings and transforms them into
// a map.
func EnvMap(env []string) map[string]string {
	out := make(map[string]string, len(env))
	for _, kv := range env {
		l := strings.SplitN(kv, "=", 2)
		if len(l) == 2 {
			out[l[0]] = l[1]
		}
	}
	return out
}

// SubstituteEnv takes a list of environment strings like:
// - K1=BAR
// - K2=$K1
// - K3=${K2}
// and substitutes them top-down from the given environment
// and from the os environment.
//
// Input strings that do not have the form "k=v" will be dropped
// from the output.
//
// The result will be a list of environment strings in the same
// order as the input.
func SubstituteEnv(env []string) []string {
	envMap := map[string]string{}
	out := make([]string, 0, len(env))

	for _, s := range env {
		l := strings.SplitN(s, "=", 2)
		if len(l) < 2 {
			continue
		}
		k := l[0]
		v := l[1]
		v = os.Expand(v, func(k string) string {
			if s, ok := envMap[k]; ok {
				return s
			}
			return os.Getenv(k)
		})
		out = append(out, fmt.Sprintf("%s=%s", k, v))
		envMap[k] = v
	}

	return out
}
