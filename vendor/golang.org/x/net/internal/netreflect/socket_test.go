// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netreflect_test

import (
	"net"
	"os"
	"testing"

	"golang.org/x/net/internal/netreflect"
	"golang.org/x/net/internal/nettest"
)

func TestSocketOf(t *testing.T) {
	for _, network := range []string{"tcp", "unix", "unixpacket"} {
		if !nettest.TestableNetwork(network) {
			continue
		}
		ln, err := nettest.NewLocalListener(network)
		if err != nil {
			t.Error(err)
			continue
		}
		defer func() {
			path := ln.Addr().String()
			ln.Close()
			if network == "unix" || network == "unixpacket" {
				os.Remove(path)
			}
		}()
		c, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			t.Error(err)
			continue
		}
		defer c.Close()
		if _, err := netreflect.SocketOf(c); err != nil {
			t.Error(err)
			continue
		}
	}
}

func TestPacketSocketOf(t *testing.T) {
	for _, network := range []string{"udp", "unixgram"} {
		if !nettest.TestableNetwork(network) {
			continue
		}
		c, err := nettest.NewLocalPacketListener(network)
		if err != nil {
			t.Error(err)
			continue
		}
		defer c.Close()
		if _, err := netreflect.PacketSocketOf(c); err != nil {
			t.Error(err)
			continue
		}
	}
}
