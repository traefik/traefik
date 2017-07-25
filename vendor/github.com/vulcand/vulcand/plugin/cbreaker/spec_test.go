package cbreaker

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/codegangsta/cli"
	"github.com/vulcand/vulcand/plugin"
	. "gopkg.in/check.v1"
)

func TestCL(t *testing.T) { TestingT(t) }

type SpecSuite struct {
}

var _ = Suite(&SpecSuite{})

// One of the most important tests:
// Make sure the spec is compatible and will be accepted by middleware registry
func (s *SpecSuite) TestSpecIsOK(c *C) {
	c.Assert(plugin.NewRegistry().AddSpec(GetSpec()), IsNil)
}

func (s *SpecSuite) TestNewCircuitBreakerFromJSON(c *C) {
	r := plugin.NewRegistry()
	c.Assert(r.AddSpec(GetSpec()), IsNil)

	bytes := []byte(`{
                 "Condition":"LatencyAtQuantileMS(50.0) < 20",
                 "Fallback":{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}},
                 "OnTripped": {"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "GET"}},
                 "OnStandby": {"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "POST"}},
                 "FallbackDuration": 10000000000,
                 "RecoveryDuration": 10000000000,
                 "CheckPeriod": 100000000}`)

	spec := r.GetSpec(GetSpec().Type)
	c.Assert(spec, NotNil)

	out, err := spec.FromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)

	spec2 := out.(*Spec)
	c.Assert(spec2.Fallback, NotNil)
	c.Assert(spec2.OnTripped, NotNil)
	c.Assert(spec2.OnStandby, NotNil)
}

func (s *SpecSuite) TestNewCircuitBreakerFromJSONEmptyStrings(c *C) {
	r := plugin.NewRegistry()
	c.Assert(r.AddSpec(GetSpec()), IsNil)

	bytes := []byte(`{
                 "Condition":"LatencyAtQuantileMS(50.0) < 20",
                 "Fallback":{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}},
                 "OnTripped": "",
                 "OnStandby": "",
                 "FallbackDuration": 10000000000,
                 "RecoveryDuration": 10000000000,
                 "CheckPeriod": 100000000}`)

	spec := r.GetSpec(GetSpec().Type)
	c.Assert(spec, NotNil)

	out, err := spec.FromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)

	spec2 := out.(*Spec)
	c.Assert(spec2.Fallback, NotNil)
	c.Assert(spec2.OnTripped, Equals, "")
	c.Assert(spec2.OnStandby, Equals, "")
}

func (s *SpecSuite) TestNewCircuitBreakerFromJSONDefaults(c *C) {
	r := plugin.NewRegistry()
	c.Assert(r.AddSpec(GetSpec()), IsNil)

	bytes := []byte(`{
                 "Condition":"LatencyAtQuantileMS(50.0) < 20",
                 "Fallback":{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}}`)

	spec := r.GetSpec(GetSpec().Type)
	c.Assert(spec, NotNil)

	out, err := spec.FromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)

	spec2 := out.(*Spec)
	c.Assert(spec2.Fallback, NotNil)
	c.Assert(spec2.OnTripped, IsNil)
	c.Assert(spec2.OnStandby, IsNil)
}

func (s *SpecSuite) TestNewCircuitBreakerSerializationCycle(c *C) {
	cl, err := NewSpec(
		`LatencyAtQuantileMS(50.0) < 20`,
		`{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}`,
		`{"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "POST", "Form": {"Key": ["Val"]}}}`,
		`{"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "POST", "Form": {"Key": ["Val"]}}}`,
		defaultFallbackDuration,
		defaultRecoveryDuration,
		defaultCheckPeriod,
	)

	bytes, err := json.Marshal(cl)
	c.Assert(err, IsNil)

	r := plugin.NewRegistry()
	c.Assert(r.AddSpec(GetSpec()), IsNil)

	spec := r.GetSpec(GetSpec().Type)
	c.Assert(spec, NotNil)

	out, err := spec.FromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)

	spec2 := out.(*Spec)
	c.Assert(spec2.Fallback, NotNil)
	c.Assert(spec2.OnTripped, NotNil)
	c.Assert(spec2.OnStandby, NotNil)
}

func (s *SpecSuite) TestNewCircuitBreakerSuccess(c *C) {
	cl, err := NewSpec(
		`LatencyAtQuantileMS(50.0) < 20`,
		`{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}`,
		`{"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "POST", "Form": {"Key": ["Val"]}}}`,
		`{"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "POST", "Form": {"Key": ["Val"]}}}`,
		defaultFallbackDuration,
		defaultRecoveryDuration,
		defaultCheckPeriod,
	)
	c.Assert(err, IsNil)
	c.Assert(cl, NotNil)

	c.Assert(cl.String(), Not(Equals), "")

	out, err := cl.NewHandler(nil)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)
}

func (s *SpecSuite) TestNewCircuitBreakerFromOther(c *C) {
	cl, err := NewSpec(
		"LatencyAtQuantileMS(50.0) < 20",
		`{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}`,
		`{"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "GET"}}`,
		`{"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "POST"}}`,
		defaultFallbackDuration,
		defaultRecoveryDuration,
		defaultCheckPeriod,
	)
	c.Assert(cl, NotNil)
	c.Assert(err, IsNil)

	out, err := FromOther(*cl)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, cl)
}

func (s *SpecSuite) TestNewCircuitBreakerFromCli(c *C) {
	app := cli.NewApp()
	app.Name = "test"
	executed := false
	app.Action = func(ctx *cli.Context) {
		executed = true
		out, err := FromCli(ctx)
		c.Assert(out, NotNil)
		c.Assert(err, IsNil)

		cl := out.(*Spec)
		c.Assert(cl.Condition, Equals, "LatencyAtQuantileMS(50.0) < 20")
		c.Assert(cl.Fallback, Equals, `{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}`)
		c.Assert(cl.OnTripped, Equals, `{"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "GET"}}`)
		c.Assert(cl.OnStandby, Equals, `{"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "POST"}}`)
		c.Assert(cl.FallbackDuration, Equals, 11*time.Second)
		c.Assert(cl.RecoveryDuration, Equals, 12*time.Second)
		c.Assert(cl.CheckPeriod, Equals, 14*time.Millisecond)
	}
	app.Flags = CliFlags()
	app.Run([]string{"test",
		"--condition=LatencyAtQuantileMS(50.0) < 20",
		`--fallback={"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}`,
		`--onTripped={"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "GET"}}`,
		`--onStandby={"Type": "webhook", "Action": {"URL": "http://localhost", "Method": "POST"}}`,
		`--fallbackDuration=11s`,
		`--recoveryDuration=12s`,
		`--checkPeriod=14ms`,
	})
	c.Assert(executed, Equals, true)
}

func (s *SpecSuite) TestNewCircuitBreakerBadParams(c *C) {
	params := []struct {
		Condition        string
		Fallback         string
		OnTripped        string
		OnStandby        string
		FallbackDuration time.Duration
		RecoveryDuration time.Duration
		CheckPeriod      time.Duration
	}{
		{
			Condition:        "whut?",
			Fallback:         `{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}`,
			OnTripped:        "",
			OnStandby:        "",
			FallbackDuration: defaultFallbackDuration,
			RecoveryDuration: defaultRecoveryDuration,
			CheckPeriod:      defaultCheckPeriod,
		},
		{
			Condition:        "LatencyAtQuantileMS(50.0) < 20",
			Fallback:         "", // No fallback is bad
			OnTripped:        "",
			OnStandby:        "",
			FallbackDuration: defaultFallbackDuration,
			RecoveryDuration: defaultRecoveryDuration,
			CheckPeriod:      defaultCheckPeriod,
		},
		{
			Condition:        "LatencyAtQuantileMS(50.0) < 20",
			Fallback:         `{"Type": "panic", "Action": {"Body": "AAAAA!!!!"}}`, // Unknown fallback type
			OnTripped:        "",
			OnStandby:        "",
			FallbackDuration: defaultFallbackDuration,
			RecoveryDuration: defaultRecoveryDuration,
			CheckPeriod:      defaultCheckPeriod,
		},
		{
			Condition:        "LatencyAtQuantileMS(50.0) < 20",
			Fallback:         `{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}`,
			OnTripped:        `{"Type": "what?", "Action": {"Body": "Come back later"}}`, // unknown side effect
			OnStandby:        "",
			FallbackDuration: defaultFallbackDuration,
			RecoveryDuration: defaultRecoveryDuration,
			CheckPeriod:      defaultCheckPeriod,
		},
		{
			Condition:        "LatencyAtQuantileMS(50.0) < 20",
			Fallback:         `{"Type": "response", "Action": {"StatusCode": 400, "Body": "Come back later"}}`,
			OnTripped:        "",
			OnStandby:        `{"Type": "what?", "Action": {"Body": "Come back later"}}`, // unknown side effect
			FallbackDuration: defaultFallbackDuration,
			RecoveryDuration: defaultRecoveryDuration,
			CheckPeriod:      defaultCheckPeriod,
		},
	}
	for _, p := range params {
		_, err := NewSpec(p.Condition, p.Fallback, p.OnTripped, p.OnStandby, p.FallbackDuration, p.RecoveryDuration, p.CheckPeriod)
		c.Assert(err, NotNil)
	}
}
