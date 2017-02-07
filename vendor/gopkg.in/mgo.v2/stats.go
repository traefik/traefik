// mgo - MongoDB driver for Go
//
// Copyright (c) 2010-2012 - Gustavo Niemeyer <gustavo@niemeyer.net>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mgo

import (
	"sync"
)

var stats *Stats
var statsMutex sync.Mutex

func SetStats(enabled bool) {
	statsMutex.Lock()
	if enabled {
		if stats == nil {
			stats = &Stats{}
		}
	} else {
		stats = nil
	}
	statsMutex.Unlock()
}

func GetStats() (snapshot Stats) {
	statsMutex.Lock()
	snapshot = *stats
	statsMutex.Unlock()
	return
}

func ResetStats() {
	statsMutex.Lock()
	debug("Resetting stats")
	old := stats
	stats = &Stats{}
	// These are absolute values:
	stats.Clusters = old.Clusters
	stats.SocketsInUse = old.SocketsInUse
	stats.SocketsAlive = old.SocketsAlive
	stats.SocketRefs = old.SocketRefs
	statsMutex.Unlock()
	return
}

type Stats struct {
	Clusters     int
	MasterConns  int
	SlaveConns   int
	SentOps      int
	ReceivedOps  int
	ReceivedDocs int
	SocketsAlive int
	SocketsInUse int
	SocketRefs   int
}

func (stats *Stats) cluster(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.Clusters += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) conn(delta int, master bool) {
	if stats != nil {
		statsMutex.Lock()
		if master {
			stats.MasterConns += delta
		} else {
			stats.SlaveConns += delta
		}
		statsMutex.Unlock()
	}
}

func (stats *Stats) sentOps(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.SentOps += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) receivedOps(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.ReceivedOps += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) receivedDocs(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.ReceivedDocs += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) socketsInUse(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.SocketsInUse += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) socketsAlive(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.SocketsAlive += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) socketRefs(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.SocketRefs += delta
		statsMutex.Unlock()
	}
}
