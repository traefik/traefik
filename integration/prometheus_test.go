package integration

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/testhelpers"
	"github.com/go-check/check"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	checker "github.com/vdemeester/shakers"
)

type PrometheusSuite struct{ BaseSuite }

func (s *PrometheusSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "prometheus")
	s.composeProject.Start(c)
}

func (s *PrometheusSuite) TestSimpleConfiguration(c *check.C) {
	whoamiHost := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/prometheus/simple.toml", struct {
		Server string
	}{whoamiHost})
	defer os.Remove(file)
	cmd, _ := s.cmdTraefik(withConfigFile(file))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, try.BodyContains("Host:test.localhost"))
	c.Assert(err, checker.IsNil)

	okRequest := testhelpers.MustNewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	okRequest.Host = "test.localhost"

	for i := 0; i < 2; i++ {
		err = try.Request(okRequest, 3*time.Second, try.StatusCodeIs(http.StatusOK))
		c.Assert(err, checker.IsNil)
	}

	// fetch metrics output
	resp, err := http.Get("http://127.0.0.1:8080/metrics")
	c.Assert(err, checker.IsNil)

	s.verifyMetricsOutput(c, resp.Body)
}

func (s *PrometheusSuite) verifyMetricsOutput(c *check.C, responseBody io.ReadCloser) {
	defer responseBody.Close()

	var metricsParser expfmt.TextParser
	metricsFamilies, err := metricsParser.TextToMetricFamilies(responseBody)

	c.Assert(err, checker.IsNil)

	tests := []struct {
		name   string
		labels map[string]string
		assert func(*dto.MetricFamily)
	}{{
		name:   "traefik_config_reloads_total",
		assert: buildCounterAssert(c, "traefik_config_reloads_total", 1),
	},
		{
			name:   "traefik_config_reload_failures_total",
			assert: buildCounterAssert(c, "traefik_config_reload_failures_total", 0),
		},
		{
			name:   "traefik_config_last_reload_success",
			assert: buildTimestampAssert(c, "traefik_config_last_reload_success"),
		},
		{
			name: "traefik_entrypoint_requests_total",
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"entrypoint": "http",
			},
			assert: buildCounterAssert(c, "traefik_entrypoint_requests_total", 2),
		},
		{
			name: "traefik_entrypoint_request_duration_seconds",
			labels: map[string]string{
				"code":       "200",
				"method":     http.MethodGet,
				"entrypoint": "http",
			},
			assert: buildHistogramAssert(c, "traefik_entrypoint_request_duration_seconds", 2),
		},
		{
			name: "traefik_backend_requests_total",
			labels: map[string]string{
				"code":    "200",
				"method":  http.MethodGet,
				"backend": "backend1",
			},
			assert: buildCounterAssert(c, "traefik_backend_requests_total", 2),
		},
		{
			name: "traefik_backend_request_duration_seconds",
			labels: map[string]string{
				"code":    "200",
				"method":  http.MethodGet,
				"backend": "backend1",
			},
			assert: buildHistogramAssert(c, "traefik_backend_request_duration_seconds", 2),
		},
		{
			name: "traefik_backend_retries_total",
			labels: map[string]string{
				"backend": "backend1",
			},
			assert: buildGreaterThanCounterAssert(c, "traefik_backend_retries_total", 1),
		},
	}

	for _, test := range tests {
		family := findMetricFamily(test.name, metricsFamilies)
		if family == nil {
			c.Errorf("gathered metrics do not contain '%s'", test.name)
			continue
		}
		for _, label := range family.Metric[0].Label {
			val, ok := test.labels[*label.Name]
			if !ok {
				c.Errorf("gathered metrics %s contains unexpected label '%s'", test.name, *label.Name)
			} else if val != *label.Value {
				c.Errorf("label '%s' in metric '%s' has wrong value '%s'", label, test.name, *label.Value)
			}
		}
		if test.assert != nil {
			test.assert(family)
		}
	}

}

func findMetricFamily(name string, families map[string]*dto.MetricFamily) *dto.MetricFamily {
	for familyName, family := range families {
		if familyName == name {
			return family
		}
	}
	return nil
}

func buildCounterAssert(c *check.C, metricName string, expectedValue int) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if cv := int(family.Metric[0].Counter.GetValue()); cv != expectedValue {
			c.Errorf("metric %s has value %d, want %d", metricName, cv, expectedValue)
		}
	}
}

func buildGreaterThanCounterAssert(c *check.C, metricName string, expectedMinValue int) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if cv := int(family.Metric[0].Counter.GetValue()); cv < expectedMinValue {
			c.Errorf("metric %s has value %d, want at least %d", metricName, cv, expectedMinValue)
		}
	}
}

func buildHistogramAssert(c *check.C, metricName string, expectedSampleCount int) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if sc := int(family.Metric[0].Histogram.GetSampleCount()); sc != expectedSampleCount {
			c.Errorf("metric %s has sample count value %d, want %d", metricName, sc, expectedSampleCount)
		}
	}
}

func buildTimestampAssert(c *check.C, metricName string) func(family *dto.MetricFamily) {
	return func(family *dto.MetricFamily) {
		if ts := time.Unix(int64(family.Metric[0].Gauge.GetValue()), 0); time.Now().Sub(ts) < time.Second*1 {
			c.Errorf("metric %s has wrong timestamp %v", metricName, ts)
		}
	}
}
