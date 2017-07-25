// Copyright 2016 The etcd Authors
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
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	v3sync "github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/etcdserver/api/v3lock/v3lockpb"
	"github.com/coreos/etcd/pkg/report"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"gopkg.in/cheggaaa/pb.v1"
)

// stmCmd represents the STM benchmark command
var stmCmd = &cobra.Command{
	Use:   "stm",
	Short: "Benchmark STM",

	Run: stmFunc,
}

type stmApply func(v3sync.STM) error

var (
	stmIsolation string
	stmIso       v3sync.Isolation

	stmTotal        int
	stmKeysPerTxn   int
	stmKeyCount     int
	stmValSize      int
	stmWritePercent int
	stmLocker       string
	stmRate         int
)

func init() {
	RootCmd.AddCommand(stmCmd)

	stmCmd.Flags().StringVar(&stmIsolation, "isolation", "r", "Read Committed (c), Repeatable Reads (r), Serializable (s), or Snapshot (ss)")
	stmCmd.Flags().IntVar(&stmKeyCount, "keys", 1, "Total unique keys accessible by the benchmark")
	stmCmd.Flags().IntVar(&stmTotal, "total", 10000, "Total number of completed STM transactions")
	stmCmd.Flags().IntVar(&stmKeysPerTxn, "keys-per-txn", 1, "Number of keys to access per transaction")
	stmCmd.Flags().IntVar(&stmWritePercent, "txn-wr-percent", 50, "Percentage of keys to overwrite per transaction")
	stmCmd.Flags().StringVar(&stmLocker, "stm-locker", "stm", "Wrap STM transaction with a custom locking mechanism (stm, lock-client, lock-rpc)")
	stmCmd.Flags().IntVar(&stmValSize, "val-size", 8, "Value size of each STM put request")
	stmCmd.Flags().IntVar(&stmRate, "rate", 0, "Maximum STM transactions per second (0 is no limit)")
}

func stmFunc(cmd *cobra.Command, args []string) {
	if stmKeyCount <= 0 {
		fmt.Fprintf(os.Stderr, "expected positive --keys, got (%v)", stmKeyCount)
		os.Exit(1)
	}

	if stmWritePercent < 0 || stmWritePercent > 100 {
		fmt.Fprintf(os.Stderr, "expected [0, 100] --txn-wr-percent, got (%v)", stmWritePercent)
		os.Exit(1)
	}

	if stmKeysPerTxn < 0 || stmKeysPerTxn > stmKeyCount {
		fmt.Fprintf(os.Stderr, "expected --keys-per-txn between 0 and %v, got (%v)", stmKeyCount, stmKeysPerTxn)
		os.Exit(1)
	}

	switch stmIsolation {
	case "c":
		stmIso = v3sync.ReadCommitted
	case "r":
		stmIso = v3sync.RepeatableReads
	case "s":
		stmIso = v3sync.Serializable
	case "ss":
		stmIso = v3sync.SerializableSnapshot
	default:
		fmt.Fprintln(os.Stderr, cmd.Usage())
		os.Exit(1)
	}

	if stmRate == 0 {
		stmRate = math.MaxInt32
	}
	limit := rate.NewLimiter(rate.Limit(stmRate), 1)

	requests := make(chan stmApply, totalClients)
	clients := mustCreateClients(totalClients, totalConns)

	bar = pb.New(stmTotal)
	bar.Format("Bom !")
	bar.Start()

	r := newReport()
	for i := range clients {
		wg.Add(1)
		go doSTM(clients[i], requests, r.Results())
	}

	go func() {
		for i := 0; i < stmTotal; i++ {
			kset := make(map[string]struct{})
			for len(kset) != stmKeysPerTxn {
				k := make([]byte, 16)
				binary.PutVarint(k, int64(rand.Intn(stmKeyCount)))
				s := string(k)
				kset[s] = struct{}{}
			}

			applyf := func(s v3sync.STM) error {
				limit.Wait(context.Background())
				wrs := int(float32(len(kset)*stmWritePercent) / 100.0)
				for k := range kset {
					s.Get(k)
					if wrs > 0 {
						s.Put(k, string(mustRandBytes(stmValSize)))
						wrs--
					}
				}
				return nil
			}

			requests <- applyf
		}
		close(requests)
	}()

	rc := r.Run()
	wg.Wait()
	close(r.Results())
	bar.Finish()
	fmt.Printf("%s", <-rc)
}

func doSTM(client *v3.Client, requests <-chan stmApply, results chan<- report.Result) {
	defer wg.Done()

	lock, unlock := func() error { return nil }, func() error { return nil }
	switch stmLocker {
	case "lock-client":
		s, err := v3sync.NewSession(client)
		if err != nil {
			panic(err)
		}
		defer s.Close()
		m := v3sync.NewMutex(s, "stmlock")
		lock = func() error { return m.Lock(context.TODO()) }
		unlock = func() error { return m.Unlock(context.TODO()) }
	case "lock-rpc":
		var lockKey []byte
		s, err := v3sync.NewSession(client)
		if err != nil {
			panic(err)
		}
		defer s.Close()
		lc := v3lockpb.NewLockClient(client.ActiveConnection())
		lock = func() error {
			req := &v3lockpb.LockRequest{Name: []byte("stmlock"), Lease: int64(s.Lease())}
			resp, err := lc.Lock(context.TODO(), req)
			if resp != nil {
				lockKey = resp.Key
			}
			return err
		}
		unlock = func() error {
			req := &v3lockpb.UnlockRequest{Key: lockKey}
			_, err := lc.Unlock(context.TODO(), req)
			return err
		}
	case "stm":
	default:
		fmt.Fprintf(os.Stderr, "unexpected stm locker %q\n", stmLocker)
		os.Exit(1)
	}

	for applyf := range requests {
		st := time.Now()
		if lerr := lock(); lerr != nil {
			panic(lerr)
		}
		_, err := v3sync.NewSTM(client, applyf, v3sync.WithIsolation(stmIso))
		if lerr := unlock(); lerr != nil {
			panic(lerr)
		}
		results <- report.Result{Err: err, Start: st, End: time.Now()}
		bar.Increment()
	}
}
