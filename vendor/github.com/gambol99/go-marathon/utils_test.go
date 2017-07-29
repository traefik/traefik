/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stubAddr struct {
	addr string
}

func (sa stubAddr) Network() string {
	return "network"
}

func (sa stubAddr) String() string {
	return sa.addr + "/8"
}

func TestUtilsAtomicIsSwitched(t *testing.T) {
	var sw atomicSwitch
	assert.False(t, sw.IsSwitched())
	sw.SwitchOn()
	assert.True(t, sw.IsSwitched())
}

func TestUtilsAtomicIsSwitchedOff(t *testing.T) {
	var sw atomicSwitch
	assert.False(t, sw.IsSwitched())
	sw.SwitchOn()
	assert.True(t, sw.IsSwitched())
	sw.SwitchedOff()
	assert.False(t, sw.IsSwitched())
}

func TestUtilsDeadline(t *testing.T) {
	err := deadline(time.Duration(5)*time.Millisecond, func(chan bool) error {
		<-time.After(time.Duration(1) * time.Second)
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrTimeoutError, err)

	err = deadline(time.Duration(5)*time.Second, func(chan bool) error {
		<-time.After(time.Duration(5) * time.Millisecond)
		return nil
	})

	assert.NoError(t, err)
}

func TestUtilsContains(t *testing.T) {
	list := []string{"1", "2", "3"}
	assert.True(t, contains(list, "2"))
	assert.False(t, contains(list, "12"))
}

func TestUtilsValidateID(t *testing.T) {
	path := "test/path"
	assert.Equal(t, validateID(path), "/test/path")
	path = "/test/path"
	assert.Equal(t, validateID(path), "/test/path")
}

func TestUtilsGetInterfaceAddress(t *testing.T) {
	// Find actual IP address we can test against.
	interfaces, err := net.Interfaces()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(interfaces))
	iface := interfaces[0]
	expectedName := iface.Name
	addresses, err := iface.Addrs()
	assert.NoError(t, err)
	expectedIPAddress := parseIPAddr(addresses[0])

	// Execute test.
	address, err := getInterfaceAddress(expectedName)
	assert.NoError(t, err)
	assert.Equal(t, expectedIPAddress, address)
	address, err = getInterfaceAddress("nothing")
	assert.Error(t, err)
	assert.Equal(t, "", address)
}

func TestUtilsTrimRootPath(t *testing.T) {
	path := "/test/path"
	assert.Equal(t, trimRootPath(path), "test/path")
	path = "test/path"
	assert.Equal(t, trimRootPath(path), "test/path")
}

func TestParseIPAddr(t *testing.T) {
	ipAddr := "127.0.0.1"
	addr := stubAddr{ipAddr}
	assert.Equal(t, ipAddr, parseIPAddr(addr))
}
