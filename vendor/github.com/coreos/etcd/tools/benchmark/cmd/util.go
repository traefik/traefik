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
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/report"
	"golang.org/x/net/context"
)

var (
	// dialTotal counts the number of mustCreateConn calls so that endpoint
	// connections can be handed out in round-robin order
	dialTotal int

	// leaderEps is a cache for holding endpoints of a leader node
	leaderEps []string
)

func mustFindLeaderEndpoints(c *clientv3.Client) {
	resp, lerr := c.MemberList(context.TODO())
	if lerr != nil {
		fmt.Fprintf(os.Stderr, "failed to get a member list: %s\n", lerr)
		os.Exit(1)
	}

	leaderId := uint64(0)
	for _, ep := range c.Endpoints() {
		if sresp, serr := c.Status(context.TODO(), ep); serr == nil {
			leaderId = sresp.Leader
			break
		}
	}

	for _, m := range resp.Members {
		if m.ID == leaderId {
			leaderEps = m.ClientURLs
			return
		}
	}

	fmt.Fprintf(os.Stderr, "failed to find a leader endpoint\n")
	os.Exit(1)
}

func mustCreateConn() *clientv3.Client {
	connEndpoints := leaderEps
	if len(connEndpoints) == 0 {
		connEndpoints = []string{endpoints[dialTotal%len(endpoints)]}
		dialTotal++
	}
	cfg := clientv3.Config{
		Endpoints:   connEndpoints,
		DialTimeout: dialTimeout,
	}
	if !tls.Empty() {
		cfgtls, err := tls.ClientConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "bad tls config: %v\n", err)
			os.Exit(1)
		}
		cfg.TLS = cfgtls
	}

	if len(user) != 0 {
		splitted := strings.SplitN(user, ":", 2)
		if len(splitted) != 2 {
			fmt.Fprintf(os.Stderr, "bad user information: %s\n", user)
			os.Exit(1)
		}

		cfg.Username = splitted[0]
		cfg.Password = splitted[1]
	}

	client, err := clientv3.New(cfg)
	if targetLeader && len(leaderEps) == 0 {
		mustFindLeaderEndpoints(client)
		client.Close()
		return mustCreateConn()
	}

	clientv3.SetLogger(log.New(os.Stderr, "grpc", 0))

	if err != nil {
		fmt.Fprintf(os.Stderr, "dial error: %v\n", err)
		os.Exit(1)
	}

	return client
}

func mustCreateClients(totalClients, totalConns uint) []*clientv3.Client {
	conns := make([]*clientv3.Client, totalConns)
	for i := range conns {
		conns[i] = mustCreateConn()
	}

	clients := make([]*clientv3.Client, totalClients)
	for i := range clients {
		clients[i] = conns[i%int(totalConns)]
	}
	return clients
}

func mustRandBytes(n int) []byte {
	rb := make([]byte, n)
	_, err := rand.Read(rb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate value: %v\n", err)
		os.Exit(1)
	}
	return rb
}

func newReport() report.Report {
	p := "%4.4f"
	if precise {
		p = "%g"
	}
	if sample {
		return report.NewReportSample(p)
	}
	return report.NewReport(p)
}
