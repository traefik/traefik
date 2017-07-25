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

package command

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/report"

	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	checkPerfLoad   string
	checkPerfPrefix string
)

type checkPerfCfg struct {
	limit    int
	clients  int
	duration int
}

var checkPerfCfgMap = map[string]checkPerfCfg{
	// TODO: support read limit
	"s": {
		limit:    150,
		clients:  50,
		duration: 60,
	},
	"m": {
		limit:    1000,
		clients:  200,
		duration: 60,
	},
	"l": {
		limit:    8000,
		clients:  500,
		duration: 60,
	},
	"xl": {
		limit:    15000,
		clients:  1000,
		duration: 60,
	},
}

// NewCheckCommand returns the cobra command for "check".
func NewCheckCommand() *cobra.Command {
	cc := &cobra.Command{
		Use:   "check <subcommand>",
		Short: "commands for checking properties of the etcd cluster",
	}

	cc.AddCommand(NewCheckPerfCommand())

	return cc
}

// NewCheckPerfCommand returns the cobra command for "check perf".
func NewCheckPerfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perf [options]",
		Short: "Check the performance of the etcd cluster",
		Run:   newCheckPerfCommand,
	}

	// TODO: support customized configuration
	cmd.Flags().StringVar(&checkPerfLoad, "load", "s", "The performance check's workload model. Accepted workloads: s(small), m(medium), l(large), xl(xLarge)")
	cmd.Flags().StringVar(&checkPerfPrefix, "prefix", "/etcdctl-check-perf/", "The prefix for writing the performance check's keys.")

	return cmd
}

// newCheckPerfCommand executes the "check perf" command.
func newCheckPerfCommand(cmd *cobra.Command, args []string) {
	var checkPerfAlias = map[string]string{
		"s": "s", "small": "s",
		"m": "m", "medium": "m",
		"l": "l", "large": "l",
		"xl": "xl", "xLarge": "xl",
	}

	model, ok := checkPerfAlias[checkPerfLoad]
	if !ok {
		ExitWithError(ExitBadFeature, fmt.Errorf("unknown load option %v", checkPerfLoad))
	}
	cfg := checkPerfCfgMap[model]

	requests := make(chan v3.Op, cfg.clients)
	limit := rate.NewLimiter(rate.Limit(cfg.limit), 1)

	var clients []*v3.Client
	for i := 0; i < cfg.clients; i++ {
		clients = append(clients, mustClientFromCmd(cmd))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.duration)*time.Second)
	resp, err := clients[0].Get(ctx, checkPerfPrefix, v3.WithPrefix(), v3.WithLimit(1))
	cancel()
	if err != nil {
		ExitWithError(ExitError, err)
	}
	if len(resp.Kvs) > 0 {
		ExitWithError(ExitInvalidInput, fmt.Errorf("prefix %q has keys. Delete with etcdctl del --prefix %s first.", checkPerfPrefix, checkPerfPrefix))
	}

	ksize, vsize := 256, 1024
	k, v := make([]byte, ksize), string(make([]byte, vsize))

	bar := pb.New(cfg.duration)
	bar.Format("Bom !")
	bar.Start()

	r := report.NewReport("%4.4f")
	var wg sync.WaitGroup

	wg.Add(len(clients))
	for i := range clients {
		go func(c *v3.Client) {
			defer wg.Done()
			for op := range requests {
				st := time.Now()
				_, derr := c.Do(context.Background(), op)
				r.Results() <- report.Result{Err: derr, Start: st, End: time.Now()}
			}
		}(clients[i])
	}

	go func() {
		cctx, ccancel := context.WithTimeout(context.Background(), time.Duration(cfg.duration)*time.Second)
		defer ccancel()
		for limit.Wait(cctx) == nil {
			binary.PutVarint(k, int64(rand.Int63n(math.MaxInt64)))
			requests <- v3.OpPut(checkPerfPrefix+string(k), v)
		}
		close(requests)
	}()

	go func() {
		for i := 0; i < cfg.duration; i++ {
			time.Sleep(time.Second)
			bar.Add(1)
		}
		bar.Finish()
	}()

	sc := r.Stats()
	wg.Wait()
	close(r.Results())

	s := <-sc

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	_, err = clients[0].Delete(ctx, checkPerfPrefix, v3.WithPrefix())
	cancel()
	if err != nil {
		ExitWithError(ExitError, err)
	}

	ok = true
	if len(s.ErrorDist) != 0 {
		fmt.Println("FAIL: too many errors")
		for k, v := range s.ErrorDist {
			fmt.Printf("FAIL: ERROR(%v) -> %d\n", k, v)
		}
		ok = false
	}

	if s.RPS/float64(cfg.limit) <= 0.9 {
		fmt.Printf("FAIL: Throughput too low: %d writes/s\n", int(s.RPS)+1)
		ok = false
	} else {
		fmt.Printf("PASS: Throughput is %d writes/s\n", int(s.RPS)+1)
	}
	if s.Slowest > 0.5 { // slowest request > 500ms
		fmt.Printf("Slowest request took too long: %fs\n", s.Slowest)
		ok = false
	} else {
		fmt.Printf("PASS: Slowest request took %fs\n", s.Slowest)
	}
	if s.Stddev > 0.1 { // stddev > 100ms
		fmt.Printf("Stddev too high: %fs\n", s.Stddev)
		ok = false
	} else {
		fmt.Printf("PASS: Stddev is %fs\n", s.Stddev)
	}

	if ok {
		fmt.Println("PASS")
	} else {
		fmt.Println("FAIL")
		os.Exit(ExitError)
	}
}
