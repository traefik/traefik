package connlimit

import (
	"testing"

	"github.com/codegangsta/cli"
	"github.com/vulcand/vulcand/plugin"
	. "gopkg.in/check.v1"
)

func TestCL(t *testing.T) { TestingT(t) }

type ConnLimitSuite struct {
}

var _ = Suite(&ConnLimitSuite{})

// One of the most important tests:
// Make sure the RateLimit spec is compatible and will be accepted by middleware registry
func (s *ConnLimitSuite) TestSpecIsOK(c *C) {
	c.Assert(plugin.NewRegistry().AddSpec(GetSpec()), IsNil)
}

func (s *ConnLimitSuite) TestNewConnLimitSuccess(c *C) {
	cl, err := NewConnLimit(10, "client.ip")
	c.Assert(cl, NotNil)
	c.Assert(err, IsNil)

	c.Assert(cl.String(), Not(Equals), "")

	out, err := cl.NewHandler(nil)
	c.Assert(out, NotNil)
	c.Assert(err, IsNil)
}

func (s *ConnLimitSuite) TestNewConnLimitBadParams(c *C) {
	// Unknown variable
	_, err := NewConnLimit(10, "client ip")
	c.Assert(err, NotNil)

	// Negative connections
	_, err = NewConnLimit(-10, "client.ip")
	c.Assert(err, NotNil)
}

func (s *ConnLimitSuite) TestNewConnLimitFromOther(c *C) {
	cl, err := NewConnLimit(10, "client.ip")
	c.Assert(cl, NotNil)
	c.Assert(err, IsNil)

	out, err := FromOther(*cl)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, cl)
}

func (s *ConnLimitSuite) TestNewConnLimitFromCli(c *C) {
	app := cli.NewApp()
	app.Name = "test"
	executed := false
	app.Action = func(ctx *cli.Context) {
		executed = true
		out, err := FromCli(ctx)
		c.Assert(out, NotNil)
		c.Assert(err, IsNil)

		cl := out.(*ConnLimit)
		c.Assert(cl.Variable, Equals, "client.ip")
		c.Assert(cl.Connections, Equals, int64(10))
	}
	app.Flags = CliFlags()
	app.Run([]string{"test", "--var=client.ip", "--connections=10"})
	c.Assert(executed, Equals, true)
}
