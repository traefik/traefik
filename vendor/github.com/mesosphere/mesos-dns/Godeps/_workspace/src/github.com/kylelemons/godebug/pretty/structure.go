// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pretty

import (
	"bytes"
	"strconv"
	"strings"
)

type node interface {
	WriteTo(w *bytes.Buffer, indent string, cfg *Config)
}

func compactString(n node) string {
	switch k := n.(type) {
	case stringVal:
		return string(k)
	case rawVal:
		return string(k)
	}

	buf := new(bytes.Buffer)
	n.WriteTo(buf, "", &Config{Compact: true})
	return buf.String()
}

type stringVal string

func (str stringVal) WriteTo(w *bytes.Buffer, indent string, cfg *Config) {
	w.WriteString(strconv.Quote(string(str)))
}

type rawVal string

func (r rawVal) WriteTo(w *bytes.Buffer, indent string, cfg *Config) {
	w.WriteString(string(r))
}

type keyval struct {
	key string
	val node
}

type keyvals []keyval

func (l keyvals) Len() int           { return len(l) }
func (l keyvals) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l keyvals) Less(i, j int) bool { return l[i].key < l[j].key }

func (l keyvals) WriteTo(w *bytes.Buffer, indent string, cfg *Config) {
	keyWidth := 0

	for _, kv := range l {
		if kw := len(kv.key); kw > keyWidth {
			keyWidth = kw
		}
	}
	padding := strings.Repeat(" ", keyWidth+1)

	inner := indent + "  " + padding
	w.WriteByte('{')
	for i, kv := range l {
		if cfg.Compact {
			w.WriteString(kv.key)
			w.WriteByte(':')
		} else {
			if i > 0 || cfg.Diffable {
				w.WriteString("\n ")
				w.WriteString(indent)
			}
			w.WriteString(kv.key)
			w.WriteByte(':')
			w.WriteString(padding[len(kv.key):])
		}
		kv.val.WriteTo(w, inner, cfg)
		if i+1 < len(l) || cfg.Diffable {
			w.WriteByte(',')
		}
	}
	if !cfg.Compact && cfg.Diffable && len(l) > 0 {
		w.WriteString("\n")
		w.WriteString(indent)
	}
	w.WriteByte('}')
}

type list []node

func (l list) WriteTo(w *bytes.Buffer, indent string, cfg *Config) {
	if max := cfg.ShortList; max > 0 {
		short := compactString(l)
		if len(short) <= max {
			w.WriteString(short)
			return
		}
	}

	inner := indent + " "
	w.WriteByte('[')
	for i, v := range l {
		if !cfg.Compact && (i > 0 || cfg.Diffable) {
			w.WriteByte('\n')
			w.WriteString(inner)
		}
		v.WriteTo(w, inner, cfg)
		if i+1 < len(l) || cfg.Diffable {
			w.WriteByte(',')
		}
	}
	if !cfg.Compact && cfg.Diffable && len(l) > 0 {
		w.WriteByte('\n')
		w.WriteString(indent)
	}
	w.WriteByte(']')
}
