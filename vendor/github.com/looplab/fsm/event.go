// Copyright (c) 2013 - Max Persson <max@looplab.se>
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

package fsm

// Event is the info that get passed as a reference in the callbacks.
type Event struct {
	// FSM is a reference to the current FSM.
	FSM *FSM

	// Event is the event name.
	Event string

	// Src is the state before the transition.
	Src string

	// Dst is the state after the transition.
	Dst string

	// Err is an optional error that can be returned from a callback.
	Err error

	// Args is a optinal list of arguments passed to the callback.
	Args []interface{}

	// canceled is an internal flag set if the transition is canceled.
	canceled bool

	// async is an internal flag set if the transition should be asynchronous
	async bool
}

// Cancel can be called in before_<EVENT> or leave_<STATE> to cancel the
// current transition before it happens. It takes an opitonal error, which will
// overwrite e.Err if set before.
func (e *Event) Cancel(err ...error) {
	e.canceled = true

	if len(err) > 0 {
		e.Err = err[0]
	}
}

// Async can be called in leave_<STATE> to do an asynchronous state transition.
//
// The current state transition will be on hold in the old state until a final
// call to Transition is made. This will comlete the transition and possibly
// call the other callbacks.
func (e *Event) Async() {
	e.async = true
}
