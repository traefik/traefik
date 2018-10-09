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
	"sync"
	"time"

	"gopkg.in/tomb.v2"

	"github.com/snapcore/snapd/logger"
)

// HandlerFunc is the type of function for the handlers
type HandlerFunc func(task *Task, tomb *tomb.Tomb) error

// Retry is returned from a handler to signal that is ok to rerun the
// task at a later point. It's to be used also when a task goroutine
// is asked to stop through its tomb. After can be used to indicate
// how much to postpone the retry, 0 (the default) means at the next
// ensure pass and is what should be used if stopped through its tomb.
type Retry struct {
	After time.Duration
}

func (r *Retry) Error() string {
	return "task should be retried"
}

type blockedFunc func(t *Task, running []*Task) bool

// TaskRunner controls the running of goroutines to execute known task kinds.
type TaskRunner struct {
	state *State

	// locking
	mu       sync.Mutex
	handlers map[string]handlerPair
	optional []optionalHandler
	cleanups map[string]HandlerFunc
	stopped  bool

	blocked     []blockedFunc
	someBlocked bool

	// go-routines lifecycle
	tombs map[string]*tomb.Tomb
}

type handlerPair struct {
	do, undo HandlerFunc
}

type optionalHandler struct {
	match func(t *Task) bool
	handlerPair
}

// NewTaskRunner creates a new TaskRunner
func NewTaskRunner(s *State) *TaskRunner {
	return &TaskRunner{
		state:    s,
		handlers: make(map[string]handlerPair),
		cleanups: make(map[string]HandlerFunc),
		tombs:    make(map[string]*tomb.Tomb),
	}
}

// AddHandler registers the functions to concurrently call for doing and
// undoing tasks of the given kind. The undo handler may be nil.
func (r *TaskRunner) AddHandler(kind string, do, undo HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[kind] = handlerPair{do, undo}
}

// AddOptionalHandler register functions for doing and undoing tasks that match
// the given predicate if no explicit handler was registered for the task kind.
func (r *TaskRunner) AddOptionalHandler(match func(t *Task) bool, do, undo HandlerFunc) {
	r.optional = append(r.optional, optionalHandler{match, handlerPair{do, undo}})
}

func (r *TaskRunner) handlerPair(t *Task) handlerPair {
	if handler, ok := r.handlers[t.Kind()]; ok {
		return handler
	}
	for _, h := range r.optional {
		if h.match(t) {
			return h.handlerPair
		}
	}
	return handlerPair{}
}

// KnownTaskKinds returns all tasks kinds handled by this runner.
func (r *TaskRunner) KnownTaskKinds() []string {
	kinds := make([]string, 0, len(r.handlers))
	for h := range r.handlers {
		kinds = append(kinds, h)
	}
	return kinds
}

// AddCleanup registers a function to be called after the change completes,
// for cleaning up data left behind by tasks of the specified kind.
// The provided function will be called no matter what the final status of the
// task is. This mechanism enables keeping data around for a potential undo
// until there's no more chance of the task being undone.
//
// The cleanup function is run concurrently with other cleanup functions,
// despite any wait ordering between the tasks. If it returns an error,
// it will be retried later.
//
// The handler for tasks of the provided kind must have been previously
// registered before AddCleanup is called for it.
func (r *TaskRunner) AddCleanup(kind string, cleanup HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.handlers[kind]; !ok {
		panic("internal error: attempted to register cleanup for unknown task kind")
	}
	r.cleanups[kind] = cleanup
}

// SetBlocked sets a predicate function to decide whether to block a task from running based on the current running tasks. It can be used to control task serialisation.
func (r *TaskRunner) SetBlocked(pred func(t *Task, running []*Task) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.blocked = []blockedFunc{pred}
}

// AddBlocked adds a predicate function to decide whether to block a task from running based on the current running tasks. It can be used to control task serialisation. All added predicates are considered in turn until one returns true, or none.
func (r *TaskRunner) AddBlocked(pred func(t *Task, running []*Task) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.blocked = append(r.blocked, pred)
}

// run must be called with the state lock in place
func (r *TaskRunner) run(t *Task) {
	var handler HandlerFunc
	switch t.Status() {
	case DoStatus:
		t.SetStatus(DoingStatus)
		fallthrough
	case DoingStatus:
		handler = r.handlerPair(t).do

	case UndoStatus:
		t.SetStatus(UndoingStatus)
		fallthrough
	case UndoingStatus:
		handler = r.handlerPair(t).undo

	default:
		panic("internal error: attempted to run task in status " + t.Status().String())
	}
	if handler == nil {
		panic("internal error: attempted to run task with nil handler for status " + t.Status().String())
	}

	t.At(time.Time{}) // clear schedule
	tomb := &tomb.Tomb{}
	r.tombs[t.ID()] = tomb
	tomb.Go(func() error {
		// Capture the error result with tomb.Kill so we can
		// use tomb.Err uniformily to consider both it or a
		// overriding previous Kill reason.
		tomb.Kill(handler(t, tomb))

		// Locks must be acquired in the same order everywhere.
		r.mu.Lock()
		defer r.mu.Unlock()
		r.state.Lock()
		defer r.state.Unlock()

		delete(r.tombs, t.ID())

		// some tasks were blocked, now there's chance the
		// blocked predicate will change its value
		if r.someBlocked {
			r.state.EnsureBefore(0)
		}

		err := tomb.Err()
		switch err.(type) {
		case nil:
			// we are ok
		case *Retry:
			// preserve
		default:
			if r.stopped {
				// we are shutting down, errors might be due
				// to cancellations, to be safe retry
				err = &Retry{}
			}
		}

		switch x := err.(type) {
		case *Retry:
			// Handler asked to be called again later.
			// TODO Allow postponing retries past the next Ensure.
			if t.Status() == AbortStatus {
				// Would work without it but might take two ensures.
				r.tryUndo(t)
			} else if x.After != 0 {
				t.At(timeNow().Add(x.After))
			}
		case nil:
			var next []*Task
			switch t.Status() {
			case DoingStatus:
				t.SetStatus(DoneStatus)
				fallthrough
			case DoneStatus:
				next = t.HaltTasks()
			case AbortStatus:
				// It was actually Done if it got here.
				t.SetStatus(UndoStatus)
				r.state.EnsureBefore(0)
			case UndoingStatus:
				t.SetStatus(UndoneStatus)
				fallthrough
			case UndoneStatus:
				next = t.WaitTasks()
			}
			if len(next) > 0 {
				r.state.EnsureBefore(0)
			}
		default:
			r.abortLanes(t.Change(), t.Lanes())
			t.SetStatus(ErrorStatus)
			t.Errorf("%s", err)
		}

		return nil
	})
}

func (r *TaskRunner) clean(t *Task) {
	if !t.Change().IsReady() {
		// Whole Change is not ready so don't run cleanups yet.
		return
	}

	cleanup, ok := r.cleanups[t.Kind()]
	if !ok {
		t.SetClean()
		return
	}

	tomb := &tomb.Tomb{}
	r.tombs[t.ID()] = tomb
	tomb.Go(func() error {
		tomb.Kill(cleanup(t, tomb))

		// Locks must be acquired in the same order everywhere.
		r.mu.Lock()
		defer r.mu.Unlock()
		r.state.Lock()
		defer r.state.Unlock()

		delete(r.tombs, t.ID())

		if tomb.Err() != nil {
			logger.Debugf("Cleaning task %s: %s", t.ID(), tomb.Err())
		} else {
			t.SetClean()
		}
		return nil
	})
}

func (r *TaskRunner) abortLanes(chg *Change, lanes []int) {
	chg.AbortLanes(lanes)
	ensureScheduled := false
	for _, t := range chg.Tasks() {
		status := t.Status()
		if status == AbortStatus {
			if tb, ok := r.tombs[t.ID()]; ok {
				tb.Kill(nil)
			}
		}
		if !ensureScheduled && !status.Ready() {
			ensureScheduled = true
			r.state.EnsureBefore(0)
		}
	}
}

// tryUndo replaces the status of a knowingly aborted task.
func (r *TaskRunner) tryUndo(t *Task) {
	if t.Status() == AbortStatus && r.handlerPair(t).undo == nil {
		// Cannot undo but it was stopped in flight.
		// Hold so it doesn't look like it finished.
		t.SetStatus(HoldStatus)
		if len(t.WaitTasks()) > 0 {
			r.state.EnsureBefore(0)
		}
	} else {
		t.SetStatus(UndoStatus)
		r.state.EnsureBefore(0)
	}
}

// Ensure starts new goroutines for all known tasks with no pending
// dependencies.
// Note that Ensure will lock the state.
func (r *TaskRunner) Ensure() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		// we are stopping, don't run another ensure
		return nil
	}

	// Locks must be acquired in the same order everywhere.
	r.state.Lock()
	defer r.state.Unlock()

	r.someBlocked = false
	running := make([]*Task, 0, len(r.tombs))
	for tid := range r.tombs {
		t := r.state.Task(tid)
		if t != nil {
			running = append(running, t)
		}
	}

	ensureTime := timeNow()
	nextTaskTime := time.Time{}
ConsiderTasks:
	for _, t := range r.state.Tasks() {
		handlers := r.handlerPair(t)
		if handlers.do == nil {
			// Handled by a different runner instance.
			continue
		}

		tb := r.tombs[t.ID()]

		if t.Status() == AbortStatus {
			if tb != nil {
				tb.Kill(nil)
				continue
			}
			r.tryUndo(t)
		}

		if tb != nil {
			// Already being handled.
			continue
		}

		status := t.Status()
		if status.Ready() {
			if !t.IsClean() {
				r.clean(t)
			}
			continue
		}

		if mustWait(t) {
			// Dependencies still unhandled.
			continue
		}

		if status == UndoStatus && handlers.undo == nil {
			// Although this has no dependencies itself, it must have waited
			// above too since follow up tasks may have handlers again.
			// Cannot undo. Revert to done status.
			t.SetStatus(DoneStatus)
			if len(t.WaitTasks()) > 0 {
				r.state.EnsureBefore(0)
			}
			continue
		}

		// skip tasks scheduled for later and also track the earliest one
		tWhen := t.AtTime()
		if !tWhen.IsZero() && ensureTime.Before(tWhen) {
			if nextTaskTime.IsZero() || nextTaskTime.After(tWhen) {
				nextTaskTime = tWhen
			}
			continue
		}

		// check if any of the blocked predicates returns true
		// and skip the task if so
		for _, blocked := range r.blocked {
			if blocked(t, running) {
				r.someBlocked = true
				continue ConsiderTasks
			}
		}

		logger.Debugf("Running task %s on %s: %s", t.ID(), t.Status(), t.Summary())
		r.run(t)

		running = append(running, t)
	}

	// schedule next Ensure no later than the next task time
	if !nextTaskTime.IsZero() {
		r.state.EnsureBefore(nextTaskTime.Sub(ensureTime))
	}

	return nil
}

// mustWait returns whether task t must wait for other tasks to be done.
func mustWait(t *Task) bool {
	switch t.Status() {
	case DoStatus:
		for _, wt := range t.WaitTasks() {
			if wt.Status() != DoneStatus {
				return true
			}
		}
	case UndoStatus:
		for _, ht := range t.HaltTasks() {
			if !ht.Status().Ready() {
				return true
			}
		}
	}
	return false
}

// wait expects to be called with th r.mu lock held
func (r *TaskRunner) wait() {
	for len(r.tombs) > 0 {
		for _, t := range r.tombs {
			r.mu.Unlock()
			t.Wait()
			r.mu.Lock()
			break
		}
	}
}

// Stop kills all concurrent activities and returns after that's done.
func (r *TaskRunner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stopped = true

	for _, tb := range r.tombs {
		tb.Kill(nil)
	}

	r.wait()
}

// Wait waits for all concurrent activities and returns after that's done.
func (r *TaskRunner) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.wait()
}

// StopKinds kills all concurrent tasks of the given kinds and returns
// after that's done.
func (r *TaskRunner) StopKinds(kind ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	kinds := make(map[string]bool, len(kind))
	for _, k := range kind {
		kinds[k] = true
	}

	var tombs []*tomb.Tomb
	// Locks must be acquired in the same order everywhere:
	// r.mu, r.state
	r.state.Lock()
	for tid, tb := range r.tombs {
		task := r.state.Task(tid)
		if task == nil || !kinds[task.Kind()] {
			continue
		}
		tombs = append(tombs, tb)
		tb.Kill(nil)
	}
	r.state.Unlock()

	for _, tb := range tombs {
		r.mu.Unlock()
		tb.Wait()
		r.mu.Lock()
	}
}
