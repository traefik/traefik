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

package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Status is used for status values for changes and tasks.
type Status int

// Admitted status values for changes and tasks.
const (
	// DefaultStatus is the standard computed status for a change or task.
	// For tasks it's always mapped to DoStatus, and for change its mapped
	// to an aggregation of its tasks' statuses. See Change.Status for details.
	DefaultStatus Status = 0

	// HoldStatus means the task should not run, perhaps as a consequence of an error on another task.
	HoldStatus Status = 1

	// DoStatus means the change or task is ready to start.
	DoStatus Status = 2

	// DoingStatus means the change or task is running or an attempt was made to run it.
	DoingStatus Status = 3

	// DoneStatus means the change or task was accomplished successfully.
	DoneStatus Status = 4

	// AbortStatus means the task should stop doing its activities and then undo.
	AbortStatus Status = 5

	// UndoStatus means the change or task should be undone, probably due to an error elsewhere.
	UndoStatus Status = 6

	// UndoingStatus means the change or task is being undone or an attempt was made to undo it.
	UndoingStatus Status = 7

	// UndoneStatus means a task was first done and then undone after an error elsewhere.
	// Changes go directly into the error status instead of being marked as undone.
	UndoneStatus Status = 8

	// ErrorStatus means the change or task has errored out while running or being undone.
	ErrorStatus Status = 9

	nStatuses = iota
)

// Ready returns whether a task or change with this status needs further
// work or has completed its attempt to perform the current goal.
func (s Status) Ready() bool {
	switch s {
	case DoneStatus, UndoneStatus, HoldStatus, ErrorStatus:
		return true
	}
	return false
}

func (s Status) String() string {
	switch s {
	case DefaultStatus:
		return "Default"
	case DoStatus:
		return "Do"
	case DoingStatus:
		return "Doing"
	case DoneStatus:
		return "Done"
	case AbortStatus:
		return "Abort"
	case UndoStatus:
		return "Undo"
	case UndoingStatus:
		return "Undoing"
	case UndoneStatus:
		return "Undone"
	case HoldStatus:
		return "Hold"
	case ErrorStatus:
		return "Error"
	}
	panic(fmt.Sprintf("internal error: unknown task status code: %d", s))
}

// Change represents a tracked modification to the system state.
//
// The Change provides both the justification for individual tasks
// to be performed and the grouping of them.
//
// As an example, if an administrator requests an interface connection,
// multiple hooks might be individually run to accomplish the task. The
// Change summary would reflect the request for an interface connection,
// while the individual Task values would track the running of
// the hooks themselves.
type Change struct {
	state   *State
	id      string
	kind    string
	summary string
	status  Status
	clean   bool
	data    customData
	taskIDs []string
	lanes   int
	ready   chan struct{}

	spawnTime time.Time
	readyTime time.Time
}

type byReadyTime []*Change

func (a byReadyTime) Len() int           { return len(a) }
func (a byReadyTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byReadyTime) Less(i, j int) bool { return a[i].readyTime.Before(a[j].readyTime) }

func newChange(state *State, id, kind, summary string) *Change {
	return &Change{
		state:   state,
		id:      id,
		kind:    kind,
		summary: summary,
		data:    make(customData),
		ready:   make(chan struct{}),

		spawnTime: timeNow(),
	}
}

type marshalledChange struct {
	ID      string                      `json:"id"`
	Kind    string                      `json:"kind"`
	Summary string                      `json:"summary"`
	Status  Status                      `json:"status"`
	Clean   bool                        `json:"clean,omitempty"`
	Data    map[string]*json.RawMessage `json:"data,omitempty"`
	TaskIDs []string                    `json:"task-ids,omitempty"`
	Lanes   int                         `json:"lanes,omitempty"`

	SpawnTime time.Time  `json:"spawn-time"`
	ReadyTime *time.Time `json:"ready-time,omitempty"`
}

// MarshalJSON makes Change a json.Marshaller
func (c *Change) MarshalJSON() ([]byte, error) {
	c.state.reading()
	var readyTime *time.Time
	if !c.readyTime.IsZero() {
		readyTime = &c.readyTime
	}
	return json.Marshal(marshalledChange{
		ID:      c.id,
		Kind:    c.kind,
		Summary: c.summary,
		Status:  c.status,
		Clean:   c.clean,
		Data:    c.data,
		TaskIDs: c.taskIDs,
		Lanes:   c.lanes,

		SpawnTime: c.spawnTime,
		ReadyTime: readyTime,
	})
}

// UnmarshalJSON makes Change a json.Unmarshaller
func (c *Change) UnmarshalJSON(data []byte) error {
	if c.state != nil {
		c.state.writing()
	}
	var unmarshalled marshalledChange
	err := json.Unmarshal(data, &unmarshalled)
	if err != nil {
		return err
	}
	c.id = unmarshalled.ID
	c.kind = unmarshalled.Kind
	c.summary = unmarshalled.Summary
	c.status = unmarshalled.Status
	c.clean = unmarshalled.Clean
	custData := unmarshalled.Data
	if custData == nil {
		custData = make(customData)
	}
	c.data = custData
	c.taskIDs = unmarshalled.TaskIDs
	c.lanes = unmarshalled.Lanes
	c.ready = make(chan struct{})
	c.spawnTime = unmarshalled.SpawnTime
	if unmarshalled.ReadyTime != nil {
		c.readyTime = *unmarshalled.ReadyTime
	}
	return nil
}

// finishUnmarshal is called after the state and tasks are accessible.
func (c *Change) finishUnmarshal() {
	if c.Status().Ready() {
		close(c.ready)
	}
}

// ID returns the individual random key for the change.
func (c *Change) ID() string {
	return c.id
}

// Kind returns the nature of the change for managers to know how to handle it.
func (c *Change) Kind() string {
	return c.kind
}

// Summary returns a summary describing what the change is about.
func (c *Change) Summary() string {
	return c.summary
}

// Set associates value with key for future consulting by managers.
// The provided value must properly marshal and unmarshal with encoding/json.
func (c *Change) Set(key string, value interface{}) {
	c.state.writing()
	c.data.set(key, value)
}

// Get unmarshals the stored value associated with the provided key
// into the value parameter.
func (c *Change) Get(key string, value interface{}) error {
	c.state.reading()
	return c.data.get(key, value)
}

var statusOrder = []Status{
	AbortStatus,
	UndoingStatus,
	UndoStatus,
	DoingStatus,
	DoStatus,
	ErrorStatus,
	UndoneStatus,
	DoneStatus,
	HoldStatus,
}

func init() {
	if len(statusOrder) != nStatuses-1 {
		panic("statusOrder has wrong number of elements")
	}
}

// Status returns the current status of the change.
// If the status was not explicitly set the result is derived from the status
// of the individual tasks related to the change, according to the following
// decision sequence:
//
//     - With at least one task in DoStatus, return DoStatus
//     - With at least one task in ErrorStatus, return ErrorStatus
//     - Otherwise, return DoneStatus
//
func (c *Change) Status() Status {
	c.state.reading()
	if c.status == DefaultStatus {
		if len(c.taskIDs) == 0 {
			return HoldStatus
		}
		statusStats := make([]int, nStatuses)
		for _, tid := range c.taskIDs {
			statusStats[c.state.tasks[tid].Status()]++
		}
		for _, s := range statusOrder {
			if statusStats[s] > 0 {
				return s
			}
		}
		panic(fmt.Sprintf("internal error: cannot process change status: %v", statusStats))
	}
	return c.status
}

// SetStatus sets the change status, overriding the default behavior (see Status method).
func (c *Change) SetStatus(s Status) {
	c.state.writing()
	c.status = s
	if s.Ready() {
		c.markReady()
	}
}

func (c *Change) markReady() {
	select {
	case <-c.ready:
	default:
		close(c.ready)
	}
	if c.readyTime.IsZero() {
		c.readyTime = timeNow()
	}
}

// Ready returns a channel that is closed the first time the change becomes ready.
func (c *Change) Ready() <-chan struct{} {
	return c.ready
}

// taskStatusChanged is called by tasks when their status is changed,
// to give the opportunity for the change to close its ready channel.
func (c *Change) taskStatusChanged(t *Task, old, new Status) {
	if old.Ready() == new.Ready() {
		return
	}
	for _, tid := range c.taskIDs {
		task := c.state.tasks[tid]
		if task != t && !task.status.Ready() {
			return
		}
	}
	// Here is the exact moment when a change goes from unready to ready,
	// and from ready to unready. For now handle only the first of those.
	// For the latter the channel might be replaced in the future.
	if c.IsReady() && !c.Status().Ready() {
		panic(fmt.Errorf("change %s unexpectedly became unready (%s)", c.ID(), c.Status()))
	}
	c.markReady()
}

// IsClean returns whether all tasks in the change have been cleaned. See SetClean.
func (c *Change) IsClean() bool {
	c.state.reading()
	return c.clean
}

// IsReady returns whether the change is considered ready.
//
// The result is similar to calling Ready on the status returned by the Status
// method, but this function is more efficient as it doesn't need to recompute
// the aggregated state of tasks on every call.
//
// As an exception, IsReady returns false for a Change without any tasks that
// never had its status explicitly set and was never unmarshalled out of the
// persistent state, despite its initial status being Hold. This is how the
// system represents changes right after they are created.
func (c *Change) IsReady() bool {
	select {
	case <-c.ready:
		return true
	default:
	}
	return false
}

func (c *Change) taskCleanChanged() {
	if !c.IsReady() {
		panic("internal error: attempted to set a task clean while change not ready")
	}
	for _, tid := range c.taskIDs {
		task := c.state.tasks[tid]
		if !task.clean {
			return
		}
	}
	c.clean = true
}

// SpawnTime returns the time when the change was created.
func (c *Change) SpawnTime() time.Time {
	c.state.reading()
	return c.spawnTime
}

// ReadyTime returns the time when the change became ready.
func (c *Change) ReadyTime() time.Time {
	c.state.reading()
	return c.readyTime
}

// changeError holds a set of task errors.
type changeError struct {
	errors []taskError
}

type taskError struct {
	task  string
	error string
}

func (e *changeError) Error() string {
	var buf bytes.Buffer
	buf.WriteString("cannot perform the following tasks:\n")
	for _, te := range e.errors {
		fmt.Fprintf(&buf, "- %s (%s)\n", te.task, te.error)
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

func stripErrorMsg(msg string) (string, bool) {
	i := strings.Index(msg, " ")
	if i >= 0 && strings.HasPrefix(msg[i:], " ERROR ") {
		return msg[i+len(" ERROR "):], true
	}
	return "", false
}

// Err returns an error value based on errors that were logged for tasks registered
// in this change, or nil if the change is not in ErrorStatus.
func (c *Change) Err() error {
	c.state.reading()
	if c.Status() != ErrorStatus {
		return nil
	}
	var errors []taskError
	for _, tid := range c.taskIDs {
		task := c.state.tasks[tid]
		if task.Status() != ErrorStatus {
			continue
		}
		for _, msg := range task.Log() {
			if s, ok := stripErrorMsg(msg); ok {
				errors = append(errors, taskError{task.Summary(), s})
			}
		}
	}
	if len(errors) == 0 {
		return fmt.Errorf("internal inconsistency: change %q in ErrorStatus with no task errors logged", c.Kind())
	}
	return &changeError{errors}
}

// State returns the system State
func (c *Change) State() *State {
	return c.state
}

// AddTask registers a task as required for the state change to
// be accomplished.
func (c *Change) AddTask(t *Task) {
	c.state.writing()
	if t.change != "" {
		panic(fmt.Sprintf("internal error: cannot add one %q task to multiple changes", t.Kind()))
	}
	t.change = c.id
	c.taskIDs = addOnce(c.taskIDs, t.ID())
}

// AddAll registers all tasks in the set as required for the state
// change to be accomplished.
func (c *Change) AddAll(ts *TaskSet) {
	c.state.writing()
	for _, t := range ts.tasks {
		c.AddTask(t)
	}
}

// Tasks returns all the tasks this state change depends on.
func (c *Change) Tasks() []*Task {
	c.state.reading()
	return c.state.tasksIn(c.taskIDs)
}

// LaneTasks returns all tasks from given lanes the state change depends on.
func (c *Change) LaneTasks(lanes ...int) []*Task {
	laneLookup := make(map[int]bool)
	for _, l := range lanes {
		laneLookup[l] = true
	}

	c.state.reading()
	var tasks []*Task
	for _, tid := range c.taskIDs {
		t := c.state.tasks[tid]
		if len(t.lanes) == 0 && laneLookup[0] {
			tasks = append(tasks, t)
		}
		for _, l := range t.lanes {
			if laneLookup[l] {
				tasks = append(tasks, t)
				break
			}
		}
	}
	return tasks
}

// Abort flags the change for cancellation, whether in progress or not.
// Cancellation will proceed at the next ensure pass.
func (c *Change) Abort() {
	c.state.writing()
	tasks := make([]*Task, len(c.taskIDs))
	for i, tid := range c.taskIDs {
		tasks[i] = c.state.tasks[tid]
	}
	c.abortTasks(tasks, make(map[int]bool), make(map[string]bool))
}

// AbortLanes aborts all tasks in the provided lanes and any tasks waiting on them,
// except for tasks that are also in a healthy lane (not aborted, and not waiting
// on aborted).
func (c *Change) AbortLanes(lanes []int) {
	c.state.writing()
	c.abortLanes(lanes, make(map[int]bool), make(map[string]bool))
}

func (c *Change) abortLanes(lanes []int, abortedLanes map[int]bool, seenTasks map[string]bool) {
	var hasLive = make(map[int]bool)
	var hasDead = make(map[int]bool)
	var laneTasks []*Task
NextChangeTask:
	for _, tid := range c.taskIDs {
		t := c.state.tasks[tid]

		var live bool
		switch t.Status() {
		case DoStatus, DoingStatus, DoneStatus:
			live = true
		}

		for _, tlane := range t.Lanes() {
			for _, lane := range lanes {
				if tlane == lane {
					laneTasks = append(laneTasks, t)
					continue NextChangeTask
				}
			}

			// Track opinion about lanes not in the kill list.
			// If the lane ends up being entirely live, we'll
			// preserve this task alive too.
			if live {
				hasLive[tlane] = true
			} else {
				hasDead[tlane] = true
			}
		}
	}

	abortTasks := make([]*Task, 0, len(laneTasks))
NextLaneTask:
	for _, t := range laneTasks {
		for _, tlane := range t.Lanes() {
			if hasLive[tlane] && !hasDead[tlane] {
				continue NextLaneTask
			}
		}
		abortTasks = append(abortTasks, t)
	}

	for _, lane := range lanes {
		abortedLanes[lane] = true
	}
	if len(abortTasks) > 0 {
		c.abortTasks(abortTasks, abortedLanes, seenTasks)
	}
}

func (c *Change) abortTasks(tasks []*Task, abortedLanes map[int]bool, seenTasks map[string]bool) {
	var lanes []int
	for i := 0; i < len(tasks); i++ {
		t := tasks[i]
		if seenTasks[t.id] {
			continue
		}
		seenTasks[t.id] = true
		switch t.Status() {
		case DoStatus:
			// Still pending so don't even start.
			t.SetStatus(HoldStatus)
		case DoingStatus:
			// In progress so stop and undo it.
			t.SetStatus(AbortStatus)
		case DoneStatus:
			// Already done so undo it.
			t.SetStatus(UndoStatus)
		}

		for _, lane := range t.Lanes() {
			if !abortedLanes[lane] {
				lanes = append(lanes, t.Lanes()...)
			}
		}

		for _, halted := range t.HaltTasks() {
			if !seenTasks[halted.id] {
				tasks = append(tasks, halted)
			}
		}
	}
	if len(lanes) > 0 {
		c.abortLanes(lanes, abortedLanes, seenTasks)
	}
}
