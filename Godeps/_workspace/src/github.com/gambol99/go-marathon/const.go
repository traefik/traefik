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
	DEFAULT_EVENTS_URL = "/event"

	/* --- api related constants --- */
	MARATHON_API_VERSION      = "v2"
	MARATHON_API_SUBSCRIPTION = MARATHON_API_VERSION + "/eventSubscriptions"
	MARATHON_API_APPS         = MARATHON_API_VERSION + "/apps"
	MARATHON_API_TASKS        = MARATHON_API_VERSION + "/tasks"
	MARATHON_API_DEPLOYMENTS  = MARATHON_API_VERSION + "/deployments"
	MARATHON_API_GROUPS       = MARATHON_API_VERSION + "/groups"
	MARATHON_API_QUEUE        = MARATHON_API_VERSION + "/queue"
	MARATHON_API_INFO         = MARATHON_API_VERSION + "/info"
	MARATHON_API_LEADER       = MARATHON_API_VERSION + "/leader"
	MARATHON_API_PING         = "/ping"
	MARATHON_API_LOGGING      = "/logging"
	MARATHON_API_HELP         = "/help"
	MARATHON_API_METRICS      = "/metrics"
)
