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

const (
	defaultEventsURL = "/event"

	/* --- api related constants --- */
	marathonAPIVersion      = "v2"
	marathonAPIEventStream  = marathonAPIVersion + "/events"
	marathonAPISubscription = marathonAPIVersion + "/eventSubscriptions"
	marathonAPIApps         = marathonAPIVersion + "/apps"
	marathonAPITasks        = marathonAPIVersion + "/tasks"
	marathonAPIDeployments  = marathonAPIVersion + "/deployments"
	marathonAPIGroups       = marathonAPIVersion + "/groups"
	marathonAPIQueue        = marathonAPIVersion + "/queue"
	marathonAPIInfo         = marathonAPIVersion + "/info"
	marathonAPILeader       = marathonAPIVersion + "/leader"
	marathonAPIPing         = "ping"
)

const (
	// EventsTransportCallback activates callback events transport
	EventsTransportCallback EventsTransport = 1 << iota

	// EventsTransportSSE activates stream events transport
	EventsTransportSSE
)
