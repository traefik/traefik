package api

import (
	"fmt"
	"testing"

	"github.com/hashicorp/consul/testutil"
)

func TestHealth_Node(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	agent := c.Agent()
	health := c.Health()

	info, err := agent.Self()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	name := info["Config"]["NodeName"].(string)

	if err := testutil.WaitForResult(func() (bool, error) {
		checks, meta, err := health.Node(name, nil)
		if err != nil {
			return false, err
		}
		if meta.LastIndex == 0 {
			return false, fmt.Errorf("bad: %v", meta)
		}
		if len(checks) == 0 {
			return false, fmt.Errorf("bad: %v", checks)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestHealthChecks_AggregatedStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		checks HealthChecks
		exp    string
	}{
		{
			"empty",
			nil,
			HealthPassing,
		},
		{
			"passing",
			HealthChecks{
				&HealthCheck{
					Status: HealthPassing,
				},
			},
			HealthPassing,
		},
		{
			"warning",
			HealthChecks{
				&HealthCheck{
					Status: HealthWarning,
				},
			},
			HealthWarning,
		},
		{
			"critical",
			HealthChecks{
				&HealthCheck{
					Status: HealthCritical,
				},
			},
			HealthCritical,
		},
		{
			"node_maintenance",
			HealthChecks{
				&HealthCheck{
					CheckID: NodeMaint,
				},
			},
			HealthMaint,
		},
		{
			"service_maintenance",
			HealthChecks{
				&HealthCheck{
					CheckID: ServiceMaintPrefix + "service",
				},
			},
			HealthMaint,
		},
		{
			"unknown",
			HealthChecks{
				&HealthCheck{
					Status: "nope-nope-noper",
				},
			},
			"",
		},
		{
			"maintenance_over_critical",
			HealthChecks{
				&HealthCheck{
					CheckID: NodeMaint,
				},
				&HealthCheck{
					Status: HealthCritical,
				},
			},
			HealthMaint,
		},
		{
			"critical_over_warning",
			HealthChecks{
				&HealthCheck{
					Status: HealthCritical,
				},
				&HealthCheck{
					Status: HealthWarning,
				},
			},
			HealthCritical,
		},
		{
			"warning_over_passing",
			HealthChecks{
				&HealthCheck{
					Status: HealthWarning,
				},
				&HealthCheck{
					Status: HealthPassing,
				},
			},
			HealthWarning,
		},
		{
			"lots",
			HealthChecks{
				&HealthCheck{
					Status: HealthPassing,
				},
				&HealthCheck{
					Status: HealthPassing,
				},
				&HealthCheck{
					Status: HealthPassing,
				},
				&HealthCheck{
					Status: HealthWarning,
				},
			},
			HealthWarning,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			act := tc.checks.AggregatedStatus()
			if tc.exp != act {
				t.Errorf("\nexp: %#v\nact: %#v", tc.exp, act)
			}
		})
	}
}

func TestHealth_Checks(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	agent := c.Agent()
	health := c.Health()

	// Make a service with a check
	reg := &AgentServiceRegistration{
		Name: "foo",
		Check: &AgentServiceCheck{
			TTL: "15s",
		},
	}
	if err := agent.ServiceRegister(reg); err != nil {
		t.Fatalf("err: %v", err)
	}
	defer agent.ServiceDeregister("foo")

	if err := testutil.WaitForResult(func() (bool, error) {
		checks, meta, err := health.Checks("foo", nil)
		if err != nil {
			return false, err
		}
		if meta.LastIndex == 0 {
			return false, fmt.Errorf("bad: %v", meta)
		}
		if len(checks) == 0 {
			return false, fmt.Errorf("Bad: %v", checks)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestHealth_Checks_NodeMetaFilter(t *testing.T) {
	t.Parallel()
	meta := map[string]string{"somekey": "somevalue"}
	c, s := makeClientWithConfig(t, nil, func(conf *testutil.TestServerConfig) {
		conf.NodeMeta = meta
	})
	defer s.Stop()

	agent := c.Agent()
	health := c.Health()

	// Make a service with a check
	reg := &AgentServiceRegistration{
		Name: "foo",
		Check: &AgentServiceCheck{
			TTL: "15s",
		},
	}
	if err := agent.ServiceRegister(reg); err != nil {
		t.Fatalf("err: %v", err)
	}
	defer agent.ServiceDeregister("foo")

	if err := testutil.WaitForResult(func() (bool, error) {
		checks, meta, err := health.Checks("foo", &QueryOptions{NodeMeta: meta})
		if err != nil {
			return false, err
		}
		if meta.LastIndex == 0 {
			return false, fmt.Errorf("bad: %v", meta)
		}
		if len(checks) == 0 {
			return false, fmt.Errorf("Bad: %v", checks)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestHealth_Service(t *testing.T) {
	c, s := makeClient(t)
	defer s.Stop()

	health := c.Health()

	if err := testutil.WaitForResult(func() (bool, error) {
		// consul service should always exist...
		checks, meta, err := health.Service("consul", "", true, nil)
		if err != nil {
			return false, err
		}
		if meta.LastIndex == 0 {
			return false, fmt.Errorf("bad: %v", meta)
		}
		if len(checks) == 0 {
			return false, fmt.Errorf("Bad: %v", checks)
		}
		if _, ok := checks[0].Node.TaggedAddresses["wan"]; !ok {
			return false, fmt.Errorf("Bad: %v", checks[0].Node)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestHealth_Service_NodeMetaFilter(t *testing.T) {
	meta := map[string]string{"somekey": "somevalue"}
	c, s := makeClientWithConfig(t, nil, func(conf *testutil.TestServerConfig) {
		conf.NodeMeta = meta
	})
	defer s.Stop()

	health := c.Health()

	if err := testutil.WaitForResult(func() (bool, error) {
		// consul service should always exist...
		checks, meta, err := health.Service("consul", "", true, &QueryOptions{NodeMeta: meta})
		if err != nil {
			return false, err
		}
		if meta.LastIndex == 0 {
			return false, fmt.Errorf("bad: %v", meta)
		}
		if len(checks) == 0 {
			return false, fmt.Errorf("Bad: %v", checks)
		}
		if _, ok := checks[0].Node.TaggedAddresses["wan"]; !ok {
			return false, fmt.Errorf("Bad: %v", checks[0].Node)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestHealth_State(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	health := c.Health()

	if err := testutil.WaitForResult(func() (bool, error) {
		checks, meta, err := health.State("any", nil)
		if err != nil {
			return false, err
		}
		if meta.LastIndex == 0 {
			return false, fmt.Errorf("bad: %v", meta)
		}
		if len(checks) == 0 {
			return false, fmt.Errorf("Bad: %v", checks)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestHealth_State_NodeMetaFilter(t *testing.T) {
	t.Parallel()
	meta := map[string]string{"somekey": "somevalue"}
	c, s := makeClientWithConfig(t, nil, func(conf *testutil.TestServerConfig) {
		conf.NodeMeta = meta
	})
	defer s.Stop()

	health := c.Health()

	if err := testutil.WaitForResult(func() (bool, error) {
		checks, meta, err := health.State("any", &QueryOptions{NodeMeta: meta})
		if err != nil {
			return false, err
		}
		if meta.LastIndex == 0 {
			return false, fmt.Errorf("bad: %v", meta)
		}
		if len(checks) == 0 {
			return false, fmt.Errorf("Bad: %v", checks)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}
