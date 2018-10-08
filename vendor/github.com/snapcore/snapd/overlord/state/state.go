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

// Package state implements the representation of system state.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/snapcore/snapd/logger"
)

// A Backend is used by State to checkpoint on every unlock operation
// and to mediate requests to ensure the state sooner or request restarts.
type Backend interface {
	Checkpoint(data []byte) error
	EnsureBefore(d time.Duration)
	// TODO: take flags to ask for reboot vs restart?
	RequestRestart(t RestartType)
}

type customData map[string]*json.RawMessage

func (data customData) get(key string, value interface{}) error {
	entryJSON := data[key]
	if entryJSON == nil {
		return ErrNoState
	}
	err := json.Unmarshal(*entryJSON, value)
	if err != nil {
		return fmt.Errorf("internal error: could not unmarshal state entry %q: %v", key, err)
	}
	return nil
}

func (data customData) has(key string) bool {
	return data[key] != nil
}

func (data customData) set(key string, value interface{}) {
	if value == nil {
		delete(data, key)
		return
	}
	serialized, err := json.Marshal(value)
	if err != nil {
		logger.Panicf("internal error: could not marshal value for state entry %q: %v", key, err)
	}
	entryJSON := json.RawMessage(serialized)
	data[key] = &entryJSON
}

type RestartType int

const (
	RestartUnset RestartType = iota
	RestartDaemon
	RestartSystem
	// RestartSocket will restart the daemon so that it goes into
	// socket activation mode.
	RestartSocket
)

// State represents an evolving system state that persists across restarts.
//
// The State is concurrency-safe, and all reads and writes to it must be
// performed with the state locked. It's a runtime error (panic) to perform
// operations without it.
//
// The state is persisted on every unlock operation via the StateBackend
// it was initialized with.
type State struct {
	mu  sync.Mutex
	muC int32

	lastTaskId   int
	lastChangeId int
	lastLaneId   int

	backend  Backend
	data     customData
	changes  map[string]*Change
	tasks    map[string]*Task
	warnings map[string]*Warning

	modified bool

	cache map[interface{}]interface{}

	restarting RestartType
	restartLck sync.Mutex
}

// New returns a new empty state.
func New(backend Backend) *State {
	return &State{
		backend:  backend,
		data:     make(customData),
		changes:  make(map[string]*Change),
		tasks:    make(map[string]*Task),
		warnings: make(map[string]*Warning),
		modified: true,
		cache:    make(map[interface{}]interface{}),
	}
}

// Modified returns whether the state was modified since the last checkpoint.
func (s *State) Modified() bool {
	return s.modified
}

// Lock acquires the state lock.
func (s *State) Lock() {
	s.mu.Lock()
	atomic.AddInt32(&s.muC, 1)
}

func (s *State) reading() {
	if atomic.LoadInt32(&s.muC) != 1 {
		panic("internal error: accessing state without lock")
	}
}

func (s *State) writing() {
	s.modified = true
	if atomic.LoadInt32(&s.muC) != 1 {
		panic("internal error: accessing state without lock")
	}
}

func (s *State) unlock() {
	atomic.AddInt32(&s.muC, -1)
	s.mu.Unlock()
}

type marshalledState struct {
	Data     map[string]*json.RawMessage `json:"data"`
	Changes  map[string]*Change          `json:"changes"`
	Tasks    map[string]*Task            `json:"tasks"`
	Warnings []*Warning                  `json:"warnings,omitempty"`

	LastChangeId int `json:"last-change-id"`
	LastTaskId   int `json:"last-task-id"`
	LastLaneId   int `json:"last-lane-id"`
}

// MarshalJSON makes State a json.Marshaller
func (s *State) MarshalJSON() ([]byte, error) {
	s.reading()
	return json.Marshal(marshalledState{
		Data:     s.data,
		Changes:  s.changes,
		Tasks:    s.tasks,
		Warnings: s.flattenWarnings(),

		LastTaskId:   s.lastTaskId,
		LastChangeId: s.lastChangeId,
		LastLaneId:   s.lastLaneId,
	})
}

// UnmarshalJSON makes State a json.Unmarshaller
func (s *State) UnmarshalJSON(data []byte) error {
	s.writing()
	var unmarshalled marshalledState
	err := json.Unmarshal(data, &unmarshalled)
	if err != nil {
		return err
	}
	s.data = unmarshalled.Data
	s.changes = unmarshalled.Changes
	s.tasks = unmarshalled.Tasks
	s.unflattenWarnings(unmarshalled.Warnings)
	s.lastChangeId = unmarshalled.LastChangeId
	s.lastTaskId = unmarshalled.LastTaskId
	s.lastLaneId = unmarshalled.LastLaneId
	// backlink state again
	for _, t := range s.tasks {
		t.state = s
	}
	for _, chg := range s.changes {
		chg.state = s
		chg.finishUnmarshal()
	}
	return nil
}

func (s *State) checkpointData() []byte {
	data, err := json.Marshal(s)
	if err != nil {
		// this shouldn't happen, because the actual delicate serializing happens at various Set()s
		logger.Panicf("internal error: could not marshal state for checkpointing: %v", err)
	}
	return data
}

// unlock checkpoint retry parameters (5 mins of retries by default)
var (
	unlockCheckpointRetryMaxTime  = 5 * time.Minute
	unlockCheckpointRetryInterval = 3 * time.Second
)

// Unlock releases the state lock and checkpoints the state.
// It does not return until the state is correctly checkpointed.
// After too many unsuccessful checkpoint attempts, it panics.
func (s *State) Unlock() {
	defer s.unlock()

	if !s.modified || s.backend == nil {
		return
	}

	data := s.checkpointData()
	var err error
	start := time.Now()
	for time.Since(start) <= unlockCheckpointRetryMaxTime {
		if err = s.backend.Checkpoint(data); err == nil {
			s.modified = false
			return
		}
		time.Sleep(unlockCheckpointRetryInterval)
	}
	logger.Panicf("cannot checkpoint even after %v of retries every %v: %v", unlockCheckpointRetryMaxTime, unlockCheckpointRetryInterval, err)
}

// EnsureBefore asks for an ensure pass to happen sooner within duration from now.
func (s *State) EnsureBefore(d time.Duration) {
	if s.backend != nil {
		s.backend.EnsureBefore(d)
	}
}

// RequestRestart asks for a restart of the managing process.
func (s *State) RequestRestart(t RestartType) {
	if s.backend != nil {
		s.restartLck.Lock()
		s.restarting = t
		s.restartLck.Unlock()
		s.backend.RequestRestart(t)
	}
}

// Restarting returns whether a restart was requested with RequestRestart and of which type.
func (s *State) Restarting() (bool, RestartType) {
	s.restartLck.Lock()
	defer s.restartLck.Unlock()
	return s.restarting != RestartUnset, s.restarting
}

func MockRestarting(s *State, restarting RestartType) RestartType {
	s.restartLck.Lock()
	defer s.restartLck.Unlock()
	old := s.restarting
	s.restarting = restarting
	return old
}

// ErrNoState represents the case of no state entry for a given key.
var ErrNoState = errors.New("no state entry for key")

// Get unmarshals the stored value associated with the provided key
// into the value parameter.
// It returns ErrNoState if there is no entry for key.
func (s *State) Get(key string, value interface{}) error {
	s.reading()
	return s.data.get(key, value)
}

// Set associates value with key for future consulting by managers.
// The provided value must properly marshal and unmarshal with encoding/json.
func (s *State) Set(key string, value interface{}) {
	s.writing()
	s.data.set(key, value)
}

// Cached returns the cached value associated with the provided key.
// It returns nil if there is no entry for key.
func (s *State) Cached(key interface{}) interface{} {
	s.reading()
	return s.cache[key]
}

// Cache associates value with key for future consulting by managers.
// The cached value is not persisted.
func (s *State) Cache(key, value interface{}) {
	s.reading() // Doesn't touch persisted data.
	if value == nil {
		delete(s.cache, key)
	} else {
		s.cache[key] = value
	}
}

// NewChange adds a new change to the state.
func (s *State) NewChange(kind, summary string) *Change {
	s.writing()
	s.lastChangeId++
	id := strconv.Itoa(s.lastChangeId)
	chg := newChange(s, id, kind, summary)
	s.changes[id] = chg
	return chg
}

// NewLane creates a new lane in the state.
func (s *State) NewLane() int {
	s.writing()
	s.lastLaneId++
	return s.lastLaneId
}

// Changes returns all changes currently known to the state.
func (s *State) Changes() []*Change {
	s.reading()
	res := make([]*Change, 0, len(s.changes))
	for _, chg := range s.changes {
		res = append(res, chg)
	}
	return res
}

// Change returns the change for the given ID.
func (s *State) Change(id string) *Change {
	s.reading()
	return s.changes[id]
}

// NewTask creates a new task.
// It usually will be registered with a Change using AddTask or
// through a TaskSet.
func (s *State) NewTask(kind, summary string) *Task {
	s.writing()
	s.lastTaskId++
	id := strconv.Itoa(s.lastTaskId)
	t := newTask(s, id, kind, summary)
	s.tasks[id] = t
	return t
}

// Tasks returns all tasks currently known to the state and linked to changes.
func (s *State) Tasks() []*Task {
	s.reading()
	res := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if t.Change() == nil { // skip unlinked tasks
			continue
		}
		res = append(res, t)
	}
	return res
}

// Task returns the task for the given ID if the task has been linked to a change.
func (s *State) Task(id string) *Task {
	s.reading()
	t := s.tasks[id]
	if t == nil || t.Change() == nil {
		return nil
	}
	return t
}

// TaskCount returns the number of tasks that currently exist in the state,
// whether linked to a change or not.
func (s *State) TaskCount() int {
	s.reading()
	return len(s.tasks)
}

func (s *State) tasksIn(tids []string) []*Task {
	res := make([]*Task, len(tids))
	for i, tid := range tids {
		res[i] = s.tasks[tid]
	}
	return res
}

// Prune does several cleanup tasks to the in-memory state:
//
//  * it removes changes that became ready for more than pruneWait and aborts
//    tasks spawned for more than abortWait.
//
//  * it removes tasks unlinked to changes after pruneWait. When there are more
//    changes than the limit set via "maxReadyChanges" those changes in ready
//    state will also removed even if they are below the pruneWait duration.
//
//  * it removes expired warnings.
func (s *State) Prune(pruneWait, abortWait time.Duration, maxReadyChanges int) {
	now := time.Now()
	pruneLimit := now.Add(-pruneWait)
	abortLimit := now.Add(-abortWait)

	// sort from oldest to newest
	changes := s.Changes()
	sort.Sort(byReadyTime(changes))

	readyChangesCount := 0
	for i := range changes {
		// changes are sorted (not-ready sorts first)
		// so we know we can iterate in reverse and break once we
		// find a ready time of "zero"
		chg := changes[len(changes)-i-1]
		if chg.ReadyTime().IsZero() {
			break
		}
		readyChangesCount++
	}

	for k, w := range s.warnings {
		if w.ExpiredBefore(now) {
			delete(s.warnings, k)
		}
	}

	for _, chg := range changes {
		spawnTime := chg.SpawnTime()
		readyTime := chg.ReadyTime()
		if readyTime.IsZero() {
			if spawnTime.Before(pruneLimit) && len(chg.Tasks()) == 0 {
				chg.Abort()
				delete(s.changes, chg.ID())
			} else if spawnTime.Before(abortLimit) {
				chg.Abort()
			}
			continue
		}
		// change old or we have too many changes
		if readyTime.Before(pruneLimit) || readyChangesCount > maxReadyChanges {
			s.writing()
			for _, t := range chg.Tasks() {
				delete(s.tasks, t.ID())
			}
			delete(s.changes, chg.ID())
			readyChangesCount--
		}
	}

	for tid, t := range s.tasks {
		// TODO: this could be done more aggressively
		if t.Change() == nil && t.SpawnTime().Before(pruneLimit) {
			s.writing()
			delete(s.tasks, tid)
		}
	}
}

// ReadState returns the state deserialized from r.
func ReadState(backend Backend, r io.Reader) (*State, error) {
	s := new(State)
	s.Lock()
	defer s.unlock()
	d := json.NewDecoder(r)
	err := d.Decode(&s)
	if err != nil {
		return nil, err
	}
	s.backend = backend
	s.modified = false
	s.cache = make(map[interface{}]interface{})
	return s, err
}
