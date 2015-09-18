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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	client := NewFakeMarathonEndpoint(t)
	info, err := client.Info()
	assert.Nil(t, err)
	assert.Equal(t, info.FrameworkId, "20140730-222531-1863654316-5050-10422-0000")
	assert.Equal(t, info.Leader, "127.0.0.1:8080")
	assert.Equal(t, info.Version, "0.7.0-SNAPSHOT")
}

func TestLeader(t *testing.T) {
	client := NewFakeMarathonEndpoint(t)
	leader, err := client.Leader()
	assert.Nil(t, err)
	assert.Equal(t, leader, "127.0.0.1:8080")
}

func TestAbdicateLeader(t *testing.T) {
	client := NewFakeMarathonEndpoint(t)
	message, err := client.AbdicateLeader()
	assert.Nil(t, err)
	assert.Equal(t, message, "Leadership abdicted")
}
