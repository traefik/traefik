/*
Copyright 2015 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package trace implements utility functions for capturing logs
package trace

import (
	"bytes"
	"fmt"
	"regexp"
	rundebug "runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"runtime"
)

const (
	// FileField is a field with code file added to structured traces
	FileField = "file"
	// FunctionField is a field with function name
	FunctionField = "func"
	// LevelField returns logging level as set by logrus
	LevelField = "level"
	// Component is a field that represents component - e.g. service or
	// function
	Component = "trace.component"
	// ComponentFields is a fields component
	ComponentFields = "trace.fields"
	// DefaultComponentPadding is a default padding for component field
	DefaultComponentPadding = 11
	// DefaultLevelPadding is a default padding for level field
	DefaultLevelPadding = 4
)

// TextFormatter is logrus-compatible formatter and adds
// file and line details to every logged entry.
type TextFormatter struct {
	// DisableTimestamp disables timestamp output (useful when outputting to
	// systemd logs)
	DisableTimestamp bool
	// ComponentPadding is a padding to pick when displaying
	// and formatting component field, defaults to DefaultComponentPadding
	ComponentPadding int
}

// Format implements logrus.Formatter interface and adds file and line
func (tf *TextFormatter) Format(e *log.Entry) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			data = append([]byte("panic in log formatter\n"), rundebug.Stack()...)
			return
		}
	}()
	var file string
	if frameNo := findFrame(); frameNo != -1 {
		t := newTrace(frameNo, nil)
		file = t.Loc()
	}

	w := &writer{}

	// time
	if !tf.DisableTimestamp {
		w.writeField(e.Time.Format(time.RFC3339))
	}

	// level
	w.writeField(strings.ToUpper(padMax(e.Level.String(), DefaultLevelPadding)))

	// component, always output
	componentI, ok := e.Data[Component]
	if !ok {
		componentI = ""
	}
	component, ok := componentI.(string)
	if !ok {
		component = fmt.Sprintf("%v", componentI)
	}
	padding := DefaultComponentPadding
	if tf.ComponentPadding != 0 {
		padding = tf.ComponentPadding
	}
	if w.Len() > 0 {
		w.WriteByte(' ')
	}
	if component != "" {
		component = fmt.Sprintf("[%v]", component)
	}
	component = strings.ToUpper(padMax(component, padding))
	if component[len(component)-1] != ' ' {
		component = component[:len(component)-1] + "]"
	}
	w.WriteString(component)

	// message
	if e.Message != "" {
		w.writeField(e.Message)
	}

	// rest of the fields
	if len(e.Data) > 0 {
		w.writeMap(e.Data)
	}

	// file, if present, always last
	if file != "" {
		w.writeField(file)
	}

	w.WriteByte('\n')
	data = w.Bytes()
	return
}

// JSONFormatter implements logrus.Formatter interface and adds file and line
// properties to JSON entries
type JSONFormatter struct {
	log.JSONFormatter
}

// Format implements logrus.Formatter interface
func (j *JSONFormatter) Format(e *log.Entry) ([]byte, error) {
	if frameNo := findFrame(); frameNo != -1 {
		t := newTrace(frameNo, nil)
		new := e.WithFields(log.Fields{
			FileField:     t.Loc(),
			FunctionField: t.FuncName(),
		})
		new.Time = e.Time
		new.Level = e.Level
		new.Message = e.Message
		e = new
	}
	return (&j.JSONFormatter).Format(e)
}

var r = regexp.MustCompile(`github\.com/(S|s)irupsen/logrus`)

func findFrame() int {
	for i := 3; i < 10; i++ {
		_, file, _, ok := runtime.Caller(i)
		if !ok {
			return -1
		}
		if !r.MatchString(file) {
			return i
		}
	}
	return -1
}

type writer struct {
	bytes.Buffer
}

func (w *writer) writeField(value interface{}) {
	if w.Len() > 0 {
		w.WriteByte(' ')
	}
	w.writeValue(value)
}

func (w *writer) writeValue(value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	if !needsQuoting(stringVal) {
		w.WriteString(stringVal)
	} else {
		w.WriteString(fmt.Sprintf("%q", stringVal))
	}
}

func (w *writer) writeKeyValue(key string, value interface{}) {
	if w.Len() > 0 {
		w.WriteByte(' ')
	}
	w.WriteString(key)
	w.WriteByte(':')
	w.writeValue(value)
}

func (w *writer) writeMap(m map[string]interface{}) {
	if len(m) == 0 {
		return
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if key == Component {
			continue
		}
		switch val := m[key].(type) {
		case log.Fields:
			w.writeMap(val)
		default:
			w.writeKeyValue(key, val)
		}
	}
}

func needsQuoting(text string) bool {
	for _, r := range text {
		if !strconv.IsPrint(r) {
			return true
		}
	}
	return false
}

func padMax(in string, chars int) string {
	switch {
	case len(in) < chars:
		return in + strings.Repeat(" ", chars-len(in))
	default:
		return in[:chars]
	}
}
