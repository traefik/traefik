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

func TestGroups(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	groups, err := endpoint.Client.Groups()
	assert.NoError(t, err)
	assert.NotNil(t, groups)
	assert.Equal(t, len(groups.Groups), 1)
	group := groups.Groups[0]
	assert.Equal(t, group.ID, fakeGroupName)
}

func TestGroup(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	group, err := endpoint.Client.Group(fakeGroupName)
	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, len(group.Apps), 1)
	assert.Equal(t, group.ID, fakeGroupName)

	group, err = endpoint.Client.Group(fakeGroupName1)

	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, group.ID, fakeGroupName1)
	assert.NotNil(t, group.Groups)
	assert.Equal(t, len(group.Groups), 1)

	frontend := group.Groups[0]
	assert.Equal(t, frontend.ID, "frontend")
	assert.Equal(t, len(frontend.Apps), 3)
	for _, app := range frontend.Apps {
		assert.NotNil(t, app.Container)
		assert.NotNil(t, app.Container.Docker)
		assert.Equal(t, app.Container.Docker.Network, "BRIDGE")
		if len(*app.Container.Docker.PortMappings) == 0 {
			t.Fail()
		}
	}
}
