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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	marathonNodeUp   = 0
	marathonNodeDown = 1
)

// Cluster is the interface for the marathon cluster impl
type Cluster interface {
	URL() string
	// retrieve a member from the cluster
	GetMember() (string, error)
	// make the last member as down
	MarkDown()
	// the size of the cluster
	Size() int
	// the members which are available
	Active() []string
	// the members which are NOT available
	NonActive() []string
}

type marathonCluster struct {
	sync.RWMutex
	// the cluster url
	url string
	// a link list of members
	members *marathonNode
	//  the number of members
	size int
	// the protocol
	protocol string
	// the current host
	active *marathonNode
	// the http client
	client *http.Client
}

// String returns a string representation of the cluster
func (r *marathonCluster) String() string {
	return fmt.Sprintf("url: %s|%s, members: %s, size: %d, active: %s",
		r.protocol, r.url, r.members, r.size, r.active)
}

type marathonNode struct {
	// the name / ip address of the host
	hostname string
	// the status of the host
	status int
	// the next member in the list
	next *marathonNode
}

func (member marathonNode) String() string {
	status := "UP"
	if member.status == marathonNodeDown {
		status = "DOWN"
	}

	return fmt.Sprintf("member: %s:%s", member.hostname, status)
}

func newCluster(client *http.Client, marathonURL string) (Cluster, error) {
	// step: parse the marathon url
	marathon, err := url.Parse(marathonURL)
	if err != nil {
		return nil, ErrInvalidEndpoint
	}

	// step: check the protocol
	if marathon.Scheme != "http" && marathon.Scheme != "https" {
		return nil, ErrInvalidEndpoint
	}

	cluster := &marathonCluster{
		client:   client,
		protocol: marathon.Scheme,
		url:      marathonURL,
	}

	/* step: create a link list of the hosts */
	var previous *marathonNode
	for index, host := range strings.SplitN(marathon.Host, ",", -1) {
		if len(host) == 0 {
			return nil, ErrInvalidEndpoint
		}

		// step: create a new cluster member
		node := new(marathonNode)
		node.hostname = host
		cluster.size++
		// step: if the first member
		if index == 0 {
			cluster.members = node
			cluster.active = node
			previous = node
		} else {
			previous.next = node
			previous = node
		}
	}
	// step: close the link list
	previous.next = cluster.active

	return cluster, nil
}

func (r *marathonCluster) URL() string {
	return r.url
}

func (r *marathonCluster) Active() []string {
	return r.memberStatus(marathonNodeUp)
}

func (r *marathonCluster) NonActive() []string {
	return r.memberStatus(marathonNodeDown)
}

func (r *marathonCluster) memberStatus(status int) []string {
	var list []string

	r.RLock()
	defer r.RUnlock()
	member := r.members

	for i := 0; i < r.size; i++ {
		if member.status == status {
			list = append(list, member.hostname)
		}
		member = member.next
	}

	return list
}

// Retrieve the current member, i.e. the current endpoint in use
func (r *marathonCluster) GetMember() (string, error) {
	r.Lock()
	defer r.Unlock()
	for i := 0; i < r.size; i++ {
		if r.active.status == marathonNodeUp {
			return r.GetMarathonURL(r.active), nil
		}
		// move to the next member
		if r.active.next != nil {
			r.active = r.active.next
		} else {
			return "", errors.New("no cluster members available at the moment")
		}
	}

	// we reached the end and there were no members available
	defer r.MarkDown()
	return "", ErrMarathonDown
}

// Retrieves the current marathon url
func (r *marathonCluster) GetMarathonURL(node *marathonNode) string {
	return fmt.Sprintf("%s://%s", r.protocol, node.hostname)
}

// MarkDown downs node the current endpoint as down and waits for it to come back only
func (r *marathonCluster) MarkDown() {
	r.Lock()
	defer r.Unlock()

	node := r.active
	node.status = marathonNodeDown

	// step: create a go-routine to place the member back in
	go func() {
		for {
			response, err := r.client.Get(r.GetMarathonURL(node) + "/ping")
			if err == nil && response.StatusCode == 200 {
				node.status = marathonNodeUp
				return
			}
			<-time.After(time.Duration(5 * time.Second))
		}
	}()

	// step: move to the next member
	if r.active.next != nil {
		r.active = r.active.next
	}
}

// Six retrieve the size of the cluster
func (r *marathonCluster) Size() int {
	return r.size
}
