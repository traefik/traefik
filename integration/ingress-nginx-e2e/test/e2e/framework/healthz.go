/*
Copyright 2020 The Kubernetes Authors.

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

package framework

import (
	"fmt"
	"net/http"
)

// VerifyHealthz verifies the status code of the healthz endpoint
func (f *Framework) VerifyHealthz(ip string, statusCode int) error {
	url := fmt.Sprintf("http://%v:10254/healthz", ip)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating GET request for URL %q failed: %v", url, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing GET request for URL %q failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != statusCode {
		return fmt.Errorf("GET request for URL %q returned HTTP status %s", url, resp.Status)
	}

	return nil
}
