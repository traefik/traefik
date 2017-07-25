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

package cmd

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/report"

	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
	"gopkg.in/cheggaaa/pb.v1"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Benchmark watch",
	Long: `Benchmark watch tests the performance of processing watch requests and 
sending events to watchers. It tests the sending performance by 
changing the value of the watched keys with concurrent put 
requests.

During the test, each watcher watches (--total/--watchers) keys 

(a watcher might watch on the same key multiple times if 
--watched-key-total is small).

Each key is watched by (--total/--watched-key-total) watchers.
`,
	Run: watchFunc,
}

var (
	watchStreams          int
	watchWatchesPerStream int
	watchedKeyTotal       int

	watchPutRate  int
	watchPutTotal int

	watchKeySize      int
	watchKeySpaceSize int
	watchSeqKeys      bool
)

type watchedKeys struct {
	watched     []string
	numWatchers map[string]int

	watches []clientv3.WatchChan

	// ctx to control all watches
	ctx    context.Context
	cancel context.CancelFunc
}

func init() {
	RootCmd.AddCommand(watchCmd)
	watchCmd.Flags().IntVar(&watchStreams, "streams", 10, "Total watch streams")
	watchCmd.Flags().IntVar(&watchWatchesPerStream, "watch-per-stream", 100, "Total watchers per stream")
	watchCmd.Flags().IntVar(&watchedKeyTotal, "watched-key-total", 1, "Total number of keys to be watched")

	watchCmd.Flags().IntVar(&watchPutRate, "put-rate", 0, "Number of keys to put per second")
	watchCmd.Flags().IntVar(&watchPutTotal, "put-total", 1000, "Number of put requests")

	watchCmd.Flags().IntVar(&watchKeySize, "key-size", 32, "Key size of watch request")
	watchCmd.Flags().IntVar(&watchKeySpaceSize, "key-space-size", 1, "Maximum possible keys")
	watchCmd.Flags().BoolVar(&watchSeqKeys, "sequential-keys", false, "Use sequential keys")
}

func watchFunc(cmd *cobra.Command, args []string) {
	if watchKeySpaceSize <= 0 {
		fmt.Fprintf(os.Stderr, "expected positive --key-space-size, got (%v)", watchKeySpaceSize)
		os.Exit(1)
	}
	grpcConns := int(totalClients)
	if totalClients > totalConns {
		grpcConns = int(totalConns)
	}
	wantedConns := 1 + (watchStreams / 100)
	if grpcConns < wantedConns {
		fmt.Fprintf(os.Stderr, "warning: grpc limits 100 streams per client connection, have %d but need %d\n", grpcConns, wantedConns)
	}
	clients := mustCreateClients(totalClients, totalConns)
	wk := newWatchedKeys()
	benchMakeWatches(clients, wk)
	benchPutWatches(clients, wk)
}

func benchMakeWatches(clients []*clientv3.Client, wk *watchedKeys) {
	streams := make([]clientv3.Watcher, watchStreams)
	for i := range streams {
		streams[i] = clientv3.NewWatcher(clients[i%len(clients)])
	}

	keyc := make(chan string, watchStreams)
	bar = pb.New(watchStreams * watchWatchesPerStream)
	bar.Format("Bom !")
	bar.Start()

	r := newReport()
	rch := r.Results()

	wg.Add(len(streams) + 1)
	wc := make(chan []clientv3.WatchChan, len(streams))
	for _, s := range streams {
		go func(s clientv3.Watcher) {
			defer wg.Done()
			var ws []clientv3.WatchChan
			for i := 0; i < watchWatchesPerStream; i++ {
				k := <-keyc
				st := time.Now()
				wch := s.Watch(wk.ctx, k)
				rch <- report.Result{Start: st, End: time.Now()}
				ws = append(ws, wch)
				bar.Increment()
			}
			wc <- ws
		}(s)
	}
	go func() {
		defer func() {
			close(keyc)
			wg.Done()
		}()
		for i := 0; i < watchStreams*watchWatchesPerStream; i++ {
			key := wk.watched[i%len(wk.watched)]
			keyc <- key
			wk.numWatchers[key]++
		}
	}()

	rc := r.Run()
	wg.Wait()
	bar.Finish()
	close(r.Results())
	fmt.Printf("Watch creation summary:\n%s", <-rc)

	for i := 0; i < len(streams); i++ {
		wk.watches = append(wk.watches, (<-wc)...)
	}
}

func newWatchedKeys() *watchedKeys {
	watched := make([]string, watchedKeyTotal)
	for i := range watched {
		k := make([]byte, watchKeySize)
		if watchSeqKeys {
			binary.PutVarint(k, int64(i%watchKeySpaceSize))
		} else {
			binary.PutVarint(k, int64(rand.Intn(watchKeySpaceSize)))
		}
		watched[i] = string(k)
	}
	ctx, cancel := context.WithCancel(context.TODO())
	return &watchedKeys{
		watched:     watched,
		numWatchers: make(map[string]int),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func benchPutWatches(clients []*clientv3.Client, wk *watchedKeys) {
	eventsTotal := 0
	for i := 0; i < watchPutTotal; i++ {
		eventsTotal += wk.numWatchers[wk.watched[i%len(wk.watched)]]
	}

	bar = pb.New(eventsTotal)
	bar.Format("Bom !")
	bar.Start()

	r := newReport()

	wg.Add(len(wk.watches))
	nrRxed := int32(eventsTotal)
	for _, w := range wk.watches {
		go func(wc clientv3.WatchChan) {
			defer wg.Done()
			recvWatchChan(wc, r.Results(), &nrRxed)
			wk.cancel()
		}(w)
	}

	putreqc := make(chan clientv3.Op, len(clients))
	go func() {
		defer close(putreqc)
		for i := 0; i < watchPutTotal; i++ {
			putreqc <- clientv3.OpPut(wk.watched[i%(len(wk.watched))], "data")
		}
	}()

	limit := rate.NewLimiter(rate.Limit(watchPutRate), 1)
	for _, cc := range clients {
		go func(c *clientv3.Client) {
			for op := range putreqc {
				if err := limit.Wait(context.TODO()); err != nil {
					panic(err)
				}
				if _, err := c.Do(context.TODO(), op); err != nil {
					panic(err)
				}
			}
		}(cc)
	}

	rc := r.Run()
	wg.Wait()
	bar.Finish()
	close(r.Results())
	fmt.Printf("Watch events received summary:\n%s", <-rc)

}

func recvWatchChan(wch clientv3.WatchChan, results chan<- report.Result, nrRxed *int32) {
	for r := range wch {
		st := time.Now()
		for range r.Events {
			results <- report.Result{Start: st, End: time.Now()}
			bar.Increment()
			if atomic.AddInt32(nrRxed, -1) <= 0 {
				return
			}
		}
	}
}
