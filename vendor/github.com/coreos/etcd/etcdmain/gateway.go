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

package etcdmain

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/coreos/etcd/proxy/tcpproxy"

	"github.com/spf13/cobra"
)

var (
	gatewayListenAddr        string
	gatewayEndpoints         []string
	gatewayDNSCluster        string
	gatewayInsecureDiscovery bool
	getewayRetryDelay        time.Duration
	gatewayCA                string
)

var (
	rootCmd = &cobra.Command{
		Use:        "etcd",
		Short:      "etcd server",
		SuggestFor: []string{"etcd"},
	}
)

func init() {
	rootCmd.AddCommand(newGatewayCommand())
}

// newGatewayCommand returns the cobra command for "gateway".
func newGatewayCommand() *cobra.Command {
	lpc := &cobra.Command{
		Use:   "gateway <subcommand>",
		Short: "gateway related command",
	}
	lpc.AddCommand(newGatewayStartCommand())

	return lpc
}

func newGatewayStartCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "start",
		Short: "start the gateway",
		Run:   startGateway,
	}

	cmd.Flags().StringVar(&gatewayListenAddr, "listen-addr", "127.0.0.1:23790", "listen address")
	cmd.Flags().StringVar(&gatewayDNSCluster, "discovery-srv", "", "DNS domain used to bootstrap initial cluster")
	cmd.Flags().BoolVar(&gatewayInsecureDiscovery, "insecure-discovery", false, "accept insecure SRV records")
	cmd.Flags().StringVar(&gatewayCA, "trusted-ca-file", "", "path to the client server TLS CA file.")

	cmd.Flags().StringSliceVar(&gatewayEndpoints, "endpoints", []string{"127.0.0.1:2379"}, "comma separated etcd cluster endpoints")

	cmd.Flags().DurationVar(&getewayRetryDelay, "retry-delay", time.Minute, "duration of delay before retrying failed endpoints")

	return &cmd
}

func stripSchema(eps []string) []string {
	var endpoints []string

	for _, ep := range eps {

		if u, err := url.Parse(ep); err == nil && u.Host != "" {
			ep = u.Host
		}

		endpoints = append(endpoints, ep)
	}

	return endpoints
}

func startGateway(cmd *cobra.Command, args []string) {
	srvs := discoverEndpoints(gatewayDNSCluster, gatewayCA, gatewayInsecureDiscovery)
	if len(srvs.Endpoints) == 0 {
		// no endpoints discovered, fall back to provided endpoints
		srvs.Endpoints = gatewayEndpoints
	}
	// Strip the schema from the endpoints because we start just a TCP proxy
	srvs.Endpoints = stripSchema(srvs.Endpoints)
	if len(srvs.SRVs) == 0 {
		for _, ep := range srvs.Endpoints {
			h, p, err := net.SplitHostPort(ep)
			if err != nil {
				plog.Fatalf("error parsing endpoint %q", ep)
			}
			var port uint16
			fmt.Sscanf(p, "%d", &port)
			srvs.SRVs = append(srvs.SRVs, &net.SRV{Target: h, Port: port})
		}
	}

	if len(srvs.Endpoints) == 0 {
		plog.Fatalf("no endpoints found")
	}

	l, err := net.Listen("tcp", gatewayListenAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tp := tcpproxy.TCPProxy{
		Listener:        l,
		Endpoints:       srvs.SRVs,
		MonitorInterval: getewayRetryDelay,
	}

	// At this point, etcd gateway listener is initialized
	notifySystemd()

	tp.Run()
}
