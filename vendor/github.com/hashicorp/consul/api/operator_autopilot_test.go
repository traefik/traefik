package api

import (
	"fmt"
	"testing"

	"github.com/hashicorp/consul/testutil"
)

func TestOperator_AutopilotGetSetConfiguration(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	operator := c.Operator()
	config, err := operator.AutopilotGetConfiguration(nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !config.CleanupDeadServers {
		t.Fatalf("bad: %v", config)
	}

	// Change a config setting
	newConf := &AutopilotConfiguration{CleanupDeadServers: false}
	if err := operator.AutopilotSetConfiguration(newConf, nil); err != nil {
		t.Fatalf("err: %v", err)
	}

	config, err = operator.AutopilotGetConfiguration(nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if config.CleanupDeadServers {
		t.Fatalf("bad: %v", config)
	}
}

func TestOperator_AutopilotCASConfiguration(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	operator := c.Operator()
	config, err := operator.AutopilotGetConfiguration(nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !config.CleanupDeadServers {
		t.Fatalf("bad: %v", config)
	}

	// Pass an invalid ModifyIndex
	{
		newConf := &AutopilotConfiguration{
			CleanupDeadServers: false,
			ModifyIndex:        config.ModifyIndex - 1,
		}
		resp, err := operator.AutopilotCASConfiguration(newConf, nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if resp {
			t.Fatalf("bad: %v", resp)
		}
	}

	// Pass a valid ModifyIndex
	{
		newConf := &AutopilotConfiguration{
			CleanupDeadServers: false,
			ModifyIndex:        config.ModifyIndex,
		}
		resp, err := operator.AutopilotCASConfiguration(newConf, nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !resp {
			t.Fatalf("bad: %v", resp)
		}
	}
}

func TestOperator_AutopilotServerHealth(t *testing.T) {
	t.Parallel()
	c, s := makeClientWithConfig(t, nil, func(c *testutil.TestServerConfig) {
		c.RaftProtocol = 3
	})
	defer s.Stop()

	operator := c.Operator()
	if err := testutil.WaitForResult(func() (bool, error) {
		out, err := operator.AutopilotServerHealth(nil)
		if err != nil {
			return false, fmt.Errorf("err: %v", err)
		}
		if len(out.Servers) != 1 ||
			!out.Servers[0].Healthy ||
			out.Servers[0].Name != s.Config.NodeName {
			return false, fmt.Errorf("bad: %v", out)
		}

		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
}
