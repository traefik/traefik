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

package cmd

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/report"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"gopkg.in/cheggaaa/pb.v1"
)

// txnPutCmd represents the txnPut command
var txnPutCmd = &cobra.Command{
	Use:   "txn-put",
	Short: "Benchmark txn-put",

	Run: txnPutFunc,
}

var (
	txnPutTotal     int
	txnPutRate      int
	txnPutOpsPerTxn int
)

func init() {
	RootCmd.AddCommand(txnPutCmd)
	txnPutCmd.Flags().IntVar(&keySize, "key-size", 8, "Key size of txn put")
	txnPutCmd.Flags().IntVar(&valSize, "val-size", 8, "Value size of txn put")
	txnPutCmd.Flags().IntVar(&txnPutOpsPerTxn, "txn-ops", 1, "Number of puts per txn")
	txnPutCmd.Flags().IntVar(&txnPutRate, "rate", 0, "Maximum txns per second (0 is no limit)")

	txnPutCmd.Flags().IntVar(&txnPutTotal, "total", 10000, "Total number of txn requests")
	txnPutCmd.Flags().IntVar(&keySpaceSize, "key-space-size", 1, "Maximum possible keys")
}

func txnPutFunc(cmd *cobra.Command, args []string) {
	if keySpaceSize <= 0 {
		fmt.Fprintf(os.Stderr, "expected positive --key-space-size, got (%v)", keySpaceSize)
		os.Exit(1)
	}

	requests := make(chan []v3.Op, totalClients)
	if txnPutRate == 0 {
		txnPutRate = math.MaxInt32
	}
	limit := rate.NewLimiter(rate.Limit(txnPutRate), 1)
	clients := mustCreateClients(totalClients, totalConns)
	k, v := make([]byte, keySize), string(mustRandBytes(valSize))

	bar = pb.New(txnPutTotal)
	bar.Format("Bom !")
	bar.Start()

	r := newReport()
	for i := range clients {
		wg.Add(1)
		go func(c *v3.Client) {
			defer wg.Done()
			for ops := range requests {
				limit.Wait(context.Background())
				st := time.Now()
				_, err := c.Txn(context.TODO()).Then(ops...).Commit()
				r.Results() <- report.Result{Err: err, Start: st, End: time.Now()}
				bar.Increment()
			}
		}(clients[i])
	}

	go func() {
		for i := 0; i < txnPutTotal; i++ {
			ops := make([]v3.Op, txnPutOpsPerTxn)
			for j := 0; j < txnPutOpsPerTxn; j++ {
				binary.PutVarint(k, int64(((i*txnPutOpsPerTxn)+j)%keySpaceSize))
				ops[j] = v3.OpPut(string(k), v)
			}
			requests <- ops
		}
		close(requests)
	}()

	rc := r.Run()
	wg.Wait()
	close(r.Results())
	bar.Finish()
	fmt.Println(<-rc)
}
