/*
Copyright 2017 Rohith All rights reserved.

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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnreachableStrategyAPI(t *testing.T) {
	app := Application{}
	require.Nil(t, app.UnreachableStrategy)
	app.SetUnreachableStrategy(UnreachableStrategy{}.SetExpungeAfterSeconds(30.0).SetInactiveAfterSeconds(5.0))
	us := app.UnreachableStrategy
	assert.Equal(t, 30.0, *us.ExpungeAfterSeconds)
	assert.Equal(t, 5.0, *us.InactiveAfterSeconds)

	app.EmptyUnreachableStrategy()
	us = app.UnreachableStrategy
	require.NotNil(t, us)
	assert.Nil(t, us.ExpungeAfterSeconds)
	assert.Nil(t, us.InactiveAfterSeconds)
}

func TestUnreachableStrategyUnmarshalEnabled(t *testing.T) {
	defaultConfig := NewDefaultConfig()
	configs := &configContainer{
		client: &defaultConfig,
		server: &serverConfig{
			scope: "unreachablestrategy-present",
		},
	}

	endpoint := newFakeMarathonEndpoint(t, configs)
	defer endpoint.Close()

	application, err := endpoint.Client.Application(fakeAppName)
	require.NoError(t, err)

	us := application.UnreachableStrategy
	require.NotNil(t, us)
	assert.Empty(t, us.AbsenceReason)
	if assert.NotNil(t, us.InactiveAfterSeconds) {
		assert.Equal(t, 3.0, *us.InactiveAfterSeconds)
	}
	if assert.NotNil(t, us.ExpungeAfterSeconds) {
		assert.Equal(t, 4.0, *us.ExpungeAfterSeconds)
	}
}

func TestUnreachableStrategyUnmarshalNonEnabled(t *testing.T) {
	defaultConfig := NewDefaultConfig()
	configs := &configContainer{
		client: &defaultConfig,
		server: &serverConfig{
			scope: "unreachablestrategy-absent",
		},
	}

	endpoint := newFakeMarathonEndpoint(t, configs)
	defer endpoint.Close()

	application, err := endpoint.Client.Application(fakeAppName)
	require.NoError(t, err)

	us := application.UnreachableStrategy
	require.NotNil(t, us)
	assert.Equal(t, UnreachableStrategyAbsenceReasonDisabled, us.AbsenceReason)
}

func TestUnreachableStrategyUnmarshalIllegal(t *testing.T) {
	j := []byte(`{false}`)
	us := UnreachableStrategy{}
	assert.Error(t, us.UnmarshalJSON(j))
}

func TestUnreachableStrategyMarshal(t *testing.T) {
	tests := []struct {
		name     string
		us       UnreachableStrategy
		wantJSON string
	}{
		{
			name: "present",
			us: UnreachableStrategy{
				EnabledUnreachableStrategy: EnabledUnreachableStrategy{
					InactiveAfterSeconds: float64p(3.5),
					ExpungeAfterSeconds:  float64p(4.5),
				},
				AbsenceReason: "",
			},
			wantJSON: `{"inactiveAfterSeconds":3.5,"expungeAfterSeconds":4.5}`,
		},
		{
			name: "absent",
			us: UnreachableStrategy{
				AbsenceReason: UnreachableStrategyAbsenceReasonDisabled,
			},
			wantJSON: fmt.Sprintf(`"%s"`, UnreachableStrategyAbsenceReasonDisabled),
		},
	}

	for _, test := range tests {
		label := fmt.Sprintf("test: %s", test.name)
		j, err := test.us.MarshalJSON()
		if assert.NoError(t, err, label) {
			assert.Equal(t, test.wantJSON, string(j), label)
		}
	}
}

func float64p(f float64) *float64 {
	return &f
}
