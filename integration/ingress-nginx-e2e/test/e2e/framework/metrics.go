/*
Copyright 2019 The Kubernetes Authors.

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

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

// GetMetric returns the current prometheus metric exposed by NGINX
func (f *Framework) GetMetric(metricName, ip string) (*dto.MetricFamily, error) {
	url := fmt.Sprintf("http://%v:10254/metrics", ip)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating GET request for URL %q failed: %v", url, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing GET request for URL %q failed: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request for URL %q returned HTTP status %s", url, resp.Status)
	}

	parser := expfmt.NewTextParser(model.UTF8Validation)
	metrics, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading text format failed: %v", err)
	}

	for _, m := range metrics {
		if m.GetName() == metricName {
			return m, nil
		}
	}

	return nil, fmt.Errorf("there is no metric with name %v", metricName)
}
