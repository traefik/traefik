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
	MEMBER_AVAILABLE   = 0
	MEMBER_UNAVAILABLE = 1
)

type Cluster interface {
	Url() string
	/* retrieve a member from the cluster */
	GetMember() (string, error)
	/* make the last member as down */
	MarkDown()
	/* the size of the cluster */
	Size() int
	/* the members which are available */
	Active() []string
	/* the members which are NOT available */
	NonActive() []string
}

type MarathonCluster struct {
	sync.RWMutex
	/* the cluster url */
	url string
	/* a link list of members */
	members *Member
	/* the number of members */
	size int
	/* the protocol */
	protocol string
	/* the current host */
	active *Member
}

func (cluster MarathonCluster) String() string {
	return fmt.Sprintf("url: %s|%s, members: %s, size: %d, active: %s",
		cluster.protocol, cluster.url, cluster.members, cluster.size, cluster.active)
}

func (cluster *MarathonCluster) ClusterState() []string {
	list := make([]string, 0)
	member := cluster.members
	for i := 0; i < cluster.size; i++ {
		list = append(list, fmt.Sprintf("%s", member))
		member = member.next
	}
	return list
}

type Member struct {
	/* the name / ip address of the host */
	hostname string
	/* the status of the host */
	status int
	/* the next member in the list */
	next *Member
}

func (member Member) String() string {
	status := "UP"
	if member.status == MEMBER_UNAVAILABLE {
		status = "DOWN"
	}
	return fmt.Sprintf("member: %s:%s", member.hostname, status)
}

func NewMarathonCluster(marathon_url string) (Cluster, error) {
	cluster := new(MarathonCluster)
	/* step: parse the marathon url */
	if marathon, err := url.Parse(marathon_url); err != nil {
		return nil, ErrInvalidEndpoint
	} else {
		/* step: check the protocol */
		if marathon.Scheme != "http" && marathon.Scheme != "https" {
			return nil, ErrInvalidEndpoint
		}
		cluster.protocol = marathon.Scheme
		cluster.url = marathon_url

		/* step: create a link list of the hosts */
		var previous *Member = nil
		for index, host := range strings.SplitN(marathon.Host, ",", -1) {
			/* step: create a new cluster member */
			member := new(Member)
			member.hostname = host
			cluster.size += 1
			/* step: if the first member */
			if index == 0 {
				cluster.members = member
				cluster.active = member
				previous = member
			} else {
				previous.next = member
				previous = member
			}
		}
		/* step: close the link list */
		previous.next = cluster.active
	}
	return cluster, nil
}

func (cluster *MarathonCluster) Url() string {
	return cluster.url
}

// Retrieve a list of active members
func (cluster *MarathonCluster) Active() []string {
	cluster.RLock()
	defer cluster.RUnlock()
	member := cluster.active
	list := make([]string, 0)
	for i := 0; i < cluster.size; i++ {
		if member.status == MEMBER_AVAILABLE {
			list = append(list, member.hostname)
		}
	}
	return list
}

// Retrieve a list of endpoints which are non-active
func (cluster *MarathonCluster) NonActive() []string {
	cluster.RLock()
	defer cluster.RUnlock()
	member := cluster.active
	list := make([]string, 0)
	for i := 0; i < cluster.size; i++ {
		if member.status == MEMBER_UNAVAILABLE {
			list = append(list, member.hostname)
		}
	}
	return list
}

// Retrieve the current member, i.e. the current endpoint in use
func (cluster *MarathonCluster) GetMember() (string, error) {
	cluster.Lock()
	defer cluster.Unlock()
	for i := 0; i < cluster.size; i++ {
		if cluster.active.status == MEMBER_AVAILABLE {
			return cluster.GetMarathonURL(cluster.active), nil
		}
		/* move to the next member */
		if cluster.active.next != nil {
			cluster.active = cluster.active.next
		} else {
			return "", errors.New("No cluster memebers available at the moment")
		}
	}
	/* we reached the end and there were no members available */
	return "", errors.New("No cluster memebers available at the moment")
}

// Retrieves the current marathon url
func (cluster *MarathonCluster) GetMarathonURL(member *Member) string {
	return fmt.Sprintf("%s://%s", cluster.protocol, member.hostname)
}

// Marks the current endpoint as down and waits for it to come back only
func (cluster *MarathonCluster) MarkDown() {
	cluster.Lock()
	defer cluster.Unlock()

	/* step: mark the current host as down */
	member := cluster.active
	member.status = MEMBER_UNAVAILABLE

	/* step: create a go-routine to place the member back in */
	go func() {
		http_client := &http.Client{}

		/* step: we wait a ping from the host to work */
		for {
			if response, err := http_client.Get(cluster.GetMarathonURL(member) + "/ping"); err == nil && response.StatusCode == 200 {
				member.status = MEMBER_AVAILABLE
				return
			} else {
				time.Sleep(10 * time.Second)
			}
		}
	}()

	/* step: move to the next member */
	if cluster.active.next != nil {
		cluster.active = cluster.active.next
	}
}

// Retrieve the size of the cluster
func (cluster *MarathonCluster) Size() int {
	return cluster.size
}
