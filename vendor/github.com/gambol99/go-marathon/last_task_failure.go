/*

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

// LastTaskFailure provides details on the last error experienced by an application
type LastTaskFailure struct {
	AppID     string `json:"appId,omitempty"`
	Host      string `json:"host,omitempty"`
	Message   string `json:"message,omitempty"`
	State     string `json:"state,omitempty"`
	TaskID    string `json:"taskId,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Version   string `json:"version,omitempty"`
}
