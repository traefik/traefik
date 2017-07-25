// Copyright 2017 The etcd Authors
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

// +build linux,!cov

package osutil

import (
	"syscall"
	"unsafe"
)

// dflSignal sets the given signal to SIG_DFL
func dflSignal(sig syscall.Signal) {
	// clearing out the sigact sets the signal to SIG_DFL
	var sigactBuf [32]uint64
	ptr := unsafe.Pointer(&sigactBuf)
	syscall.Syscall6(uintptr(syscall.SYS_RT_SIGACTION), uintptr(sig), uintptr(ptr), 0, 8, 0, 0)
}
