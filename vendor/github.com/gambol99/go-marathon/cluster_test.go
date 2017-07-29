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
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSize(t *testing.T) {
	cluster, err := newStandardCluster(fakeMarathonURL)
	assert.NoError(t, err)
	assert.Equal(t, cluster.size(), 3)
}

func TestActive(t *testing.T) {
	cluster, err := newStandardCluster(fakeMarathonURL)
	assert.NoError(t, err)
	assert.Equal(t, len(cluster.activeMembers()), 3)
}

func TestNonActive(t *testing.T) {
	cluster, err := newStandardCluster(fakeMarathonURL)
	assert.NoError(t, err)
	assert.Equal(t, len(cluster.nonActiveMembers()), 0)
}

func TestGetMember(t *testing.T) {
	cases := []struct {
		isDCOS      bool
		MarathonURL string
		member      string
	}{
		{
			isDCOS:      false,
			MarathonURL: fakeMarathonURL,
			member:      "http://127.0.0.1:3000",
		},
		{
			isDCOS:      false,
			MarathonURL: fakeMarathonURLWithPath,
			member:      "http://127.0.0.1:3000/path",
		},
		{
			isDCOS:      true,
			MarathonURL: fakeMarathonURL,
			member:      "http://127.0.0.1:3000/marathon",
		},
		{
			isDCOS:      true,
			MarathonURL: fakeMarathonURLWithPath,
			member:      "http://127.0.0.1:3000/path",
		},
	}
	for _, x := range cases {
		cluster, err := newCluster(&httpClient{config: Config{HTTPClient: http.DefaultClient}}, x.MarathonURL, x.isDCOS)
		assert.NoError(t, err)
		member, err := cluster.getMember()
		assert.NoError(t, err)
		assert.Equal(t, member, x.member)
	}
}

func TestMarkDown(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()
	cluster, err := newStandardCluster(endpoint.URL)
	assert.NoError(t, err)
	assert.Equal(t, len(cluster.activeMembers()), 3)

	members := cluster.activeMembers()
	cluster.markDown(members[0])
	cluster.markDown(members[1])
	assert.Equal(t, 1, len(cluster.activeMembers()))

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, len(cluster.activeMembers()), 3)
}

func TestValidClusterHosts(t *testing.T) {
	cs := []struct {
		URL    string
		Expect []string
	}{
		{
			URL:    "http://127.0.0.1",
			Expect: []string{"http://127.0.0.1"},
		},
		{
			URL:    "http://127.0.0.1:8080",
			Expect: []string{"http://127.0.0.1:8080"},
		},
		{
			URL:    "http://127.0.0.1:8080,http://127.0.0.2:8081",
			Expect: []string{"http://127.0.0.1:8080", "http://127.0.0.2:8081"},
		},
		{
			URL:    "https://127.0.0.1:8080,http://127.0.0.2:8081",
			Expect: []string{"https://127.0.0.1:8080", "http://127.0.0.2:8081"},
		},
		{
			URL:    "http://127.0.0.1:8080,127.0.0.2",
			Expect: []string{"http://127.0.0.1:8080", "http://127.0.0.2"},
		},
		{
			URL:    "https://127.0.0.1:8080,127.0.0.2",
			Expect: []string{"https://127.0.0.1:8080", "https://127.0.0.2"},
		},
		{
			URL:    "http://127.0.0.1:8080,127.0.0.2:8080",
			Expect: []string{"http://127.0.0.1:8080", "http://127.0.0.2:8080"},
		},
		{
			URL:    "http://127.0.0.1:8080,https://127.0.0.2",
			Expect: []string{"http://127.0.0.1:8080", "https://127.0.0.2"},
		},
		{
			URL:    "http://127.0.0.1:8080,https://127.0.0.2:8080",
			Expect: []string{"http://127.0.0.1:8080", "https://127.0.0.2:8080"},
		},
		{
			URL:    "http://127.0.0.1:8080/path1,127.0.0.2/path2",
			Expect: []string{"http://127.0.0.1:8080/path1", "http://127.0.0.2/path2"},
		},
	}
	for _, x := range cs {
		c, err := newStandardCluster(x.URL)
		if !assert.NoError(t, err, "URL '%s' should not have thrown an error: %s", x.URL, err) {
			continue
		}
		assert.Equal(t, x.Expect, c.activeMembers(), "URL '%s', expected: %v, got: %s", x.URL, x.Expect, c.activeMembers())
	}
}

func TestInvalidClusterHosts(t *testing.T) {
	for _, invalidHost := range []string{
		"",
		"://",
		"http://",
		"http://,,",
		"http://%42",
		"http://,127.0.0.1:3000,127.0.0.1:3000",
		"http://127.0.0.1:3000,,127.0.0.1:3000",
		"http://127.0.0.1:3000,127.0.0.1:3000,",
		"foo://127.0.0.1:3000",
	} {
		_, err := newStandardCluster(invalidHost)
		if !assert.Error(t, err) {
			t.Errorf("undetected invalid host: %s", invalidHost)
		}
	}
}

func newStandardCluster(url string) (*cluster, error) {
	return newCluster(&httpClient{config: Config{HTTPClient: http.DefaultClient}}, url, false)
}
