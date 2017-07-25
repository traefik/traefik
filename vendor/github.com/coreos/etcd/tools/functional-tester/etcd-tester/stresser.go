// Copyright 2015 The etcd Authors
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

package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc/grpclog"
)

func init() { grpclog.SetLogger(plog) }

type Stresser interface {
	// Stress starts to stress the etcd cluster
	Stress() error
	// Pause stops the stresser from sending requests to etcd. Resume by calling Stress.
	Pause()
	// Close releases all of the Stresser's resources.
	Close()
	// ModifiedKeys reports the number of keys created and deleted by stresser
	ModifiedKeys() int64
	// Checker returns an invariant checker for after the stresser is canceled.
	Checker() Checker
}

// nopStresser implements Stresser that does nothing
type nopStresser struct {
	start time.Time
	qps   int
}

func (s *nopStresser) Stress() error { return nil }
func (s *nopStresser) Pause()        {}
func (s *nopStresser) Close()        {}
func (s *nopStresser) ModifiedKeys() int64 {
	return 0
}
func (s *nopStresser) Checker() Checker { return nil }

// compositeStresser implements a Stresser that runs a slice of
// stressers concurrently.
type compositeStresser struct {
	stressers []Stresser
}

func (cs *compositeStresser) Stress() error {
	for i, s := range cs.stressers {
		if err := s.Stress(); err != nil {
			for j := 0; j < i; j++ {
				cs.stressers[i].Close()
			}
			return err
		}
	}
	return nil
}

func (cs *compositeStresser) Pause() {
	var wg sync.WaitGroup
	wg.Add(len(cs.stressers))
	for i := range cs.stressers {
		go func(s Stresser) {
			defer wg.Done()
			s.Pause()
		}(cs.stressers[i])
	}
	wg.Wait()
}

func (cs *compositeStresser) Close() {
	var wg sync.WaitGroup
	wg.Add(len(cs.stressers))
	for i := range cs.stressers {
		go func(s Stresser) {
			defer wg.Done()
			s.Close()
		}(cs.stressers[i])
	}
	wg.Wait()
}

func (cs *compositeStresser) ModifiedKeys() (modifiedKey int64) {
	for _, stress := range cs.stressers {
		modifiedKey += stress.ModifiedKeys()
	}
	return modifiedKey
}

func (cs *compositeStresser) Checker() Checker {
	var chks []Checker
	for _, s := range cs.stressers {
		if chk := s.Checker(); chk != nil {
			chks = append(chks, chk)
		}
	}
	if len(chks) == 0 {
		return nil
	}
	return newCompositeChecker(chks)
}

type stressConfig struct {
	keyLargeSize   int
	keySize        int
	keySuffixRange int

	numLeases    int
	keysPerLease int

	rateLimiter *rate.Limiter

	etcdRunnerPath string
}

// NewStresser creates stresser from a comma separated list of stresser types.
func NewStresser(s string, sc *stressConfig, m *member) Stresser {
	types := strings.Split(s, ",")
	if len(types) > 1 {
		stressers := make([]Stresser, len(types))
		for i, stype := range types {
			stressers[i] = NewStresser(stype, sc, m)
		}
		return &compositeStresser{stressers}
	}
	switch s {
	case "nop":
		return &nopStresser{start: time.Now(), qps: int(sc.rateLimiter.Limit())}
	case "keys":
		// TODO: Too intensive stressers can panic etcd member with
		// 'out of memory' error. Put rate limits in server side.
		return &keyStresser{
			Endpoint:       m.grpcAddr(),
			keyLargeSize:   sc.keyLargeSize,
			keySize:        sc.keySize,
			keySuffixRange: sc.keySuffixRange,
			N:              100,
			rateLimiter:    sc.rateLimiter,
		}
	case "v2keys":
		return &v2Stresser{
			Endpoint:       m.ClientURL,
			keySize:        sc.keySize,
			keySuffixRange: sc.keySuffixRange,
			N:              100,
			rateLimiter:    sc.rateLimiter,
		}
	case "lease":
		return &leaseStresser{
			endpoint:     m.grpcAddr(),
			numLeases:    sc.numLeases,
			keysPerLease: sc.keysPerLease,
			rateLimiter:  sc.rateLimiter,
		}
	case "election-runner":
		reqRate := 100
		args := []string{
			"election",
			fmt.Sprintf("%v", time.Now().UnixNano()), // election name as current nano time
			"--dial-timeout=10s",
			"--endpoints", m.grpcAddr(),
			"--total-client-connections=10",
			"--rounds=0", // runs forever
			"--req-rate", fmt.Sprintf("%v", reqRate),
		}
		return newRunnerStresser(sc.etcdRunnerPath, args, sc.rateLimiter, reqRate)
	case "watch-runner":
		reqRate := 100
		args := []string{
			"watcher",
			"--prefix", fmt.Sprintf("%v", time.Now().UnixNano()), // prefix all keys with nano time
			"--total-keys=1",
			"--total-prefixes=1",
			"--watch-per-prefix=1",
			"--endpoints", m.grpcAddr(),
			"--rounds=0", // runs forever
			"--req-rate", fmt.Sprintf("%v", reqRate),
		}
		return newRunnerStresser(sc.etcdRunnerPath, args, sc.rateLimiter, reqRate)
	case "lock-racer-runner":
		reqRate := 100
		args := []string{
			"lock-racer",
			fmt.Sprintf("%v", time.Now().UnixNano()), // locker name as current nano time
			"--endpoints", m.grpcAddr(),
			"--total-client-connections=10",
			"--rounds=0", // runs forever
			"--req-rate", fmt.Sprintf("%v", reqRate),
		}
		return newRunnerStresser(sc.etcdRunnerPath, args, sc.rateLimiter, reqRate)
	case "lease-runner":
		args := []string{
			"lease-renewer",
			"--ttl=30",
			"--endpoints", m.grpcAddr(),
		}
		return newRunnerStresser(sc.etcdRunnerPath, args, sc.rateLimiter, 0)
	default:
		plog.Panicf("unknown stresser type: %s\n", s)
	}
	return nil // never reach here
}
