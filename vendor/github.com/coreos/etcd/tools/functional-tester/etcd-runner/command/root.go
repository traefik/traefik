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

// Package command implements individual etcd-runner commands for the etcd-runner utility.
package command

import (
	"log"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
)

const (
	cliName        = "etcd-runner"
	cliDescription = "Stress tests using clientv3 functionality.."

	defaultDialTimeout = 2 * time.Second
)

var (
	rootCmd = &cobra.Command{
		Use:        cliName,
		Short:      cliDescription,
		SuggestFor: []string{"etcd-runner"},
	}
)

func init() {
	cobra.EnablePrefixMatching = true

	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.Lmicroseconds)

	rootCmd.PersistentFlags().StringSliceVar(&endpoints, "endpoints", []string{"127.0.0.1:2379"}, "gRPC endpoints")
	rootCmd.PersistentFlags().DurationVar(&dialTimeout, "dial-timeout", defaultDialTimeout, "dial timeout for client connections")
	rootCmd.PersistentFlags().IntVar(&reqRate, "req-rate", 30, "maximum number of requests per second")
	rootCmd.PersistentFlags().IntVar(&rounds, "rounds", 100, "number of rounds to run; 0 to run forever")

	rootCmd.AddCommand(
		NewElectionCommand(),
		NewLeaseRenewerCommand(),
		NewLockRacerCommand(),
		NewWatchCommand(),
	)
}

func Start() {
	rootCmd.SetUsageFunc(usageFunc)

	// Make help just show the usage
	rootCmd.SetHelpTemplate(`{{.UsageString}}`)

	if err := rootCmd.Execute(); err != nil {
		ExitWithError(ExitError, err)
	}
}
