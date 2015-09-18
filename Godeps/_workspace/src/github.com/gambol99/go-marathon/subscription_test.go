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

func TestSubscriptions(t *testing.T) {
	client := NewFakeMarathonEndpoint(t)
	sub, err := client.Subscriptions()
	assert.Nil(t, err)
	assert.NotNil(t, sub)
	assert.NotNil(t, sub.CallbackURLs)
	assert.Equal(t, len(sub.CallbackURLs), 1)
}
