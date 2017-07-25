package connlimit

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/vulcand/oxy/connlimit"
	"github.com/vulcand/oxy/utils"
	"github.com/vulcand/vulcand/plugin"
)

const Type = "connlimit"

func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      Type,
		FromOther: FromOther,
		FromCli:   FromCli,
		CliFlags:  CliFlags(),
	}
}

// Control simultaneous connections for a location per some variable.
type ConnLimit struct {
	Connections int64
	Variable    string // Variable defines how the limiting should be done. e.g. 'client.ip' or 'request.header.X-My-Header'
}

// Returns vulcan library compatible middleware
func (c *ConnLimit) NewHandler(next http.Handler) (http.Handler, error) {
	extract, err := utils.NewExtractor(c.Variable)
	if err != nil {
		return nil, err
	}
	return connlimit.New(next, extract, c.Connections)
}

func NewConnLimit(connections int64, variable string) (*ConnLimit, error) {
	if _, err := utils.NewExtractor(variable); err != nil {
		return nil, err
	}
	if connections < 0 {
		return nil, fmt.Errorf("connections should be > 0, got %d", connections)
	}
	return &ConnLimit{
		Connections: connections,
		Variable:    variable,
	}, nil
}

func (c *ConnLimit) String() string {
	return fmt.Sprintf("connections=%d, variable=%s", c.Connections, c.Variable)
}

func FromOther(c ConnLimit) (plugin.Middleware, error) {
	return NewConnLimit(c.Connections, c.Variable)
}

// Constructs the middleware from the command line
func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return NewConnLimit(int64(c.Int("connections")), c.String("var"))
}

func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "variable, var", Value: "client.ip", Usage: "variable to rate against, e.g. client.ip, request.host or request.header.X-Header"},
		cli.IntFlag{Name: "connections, conns", Value: 1, Usage: "amount of simultaneous connections allowed per variable value"},
	}
}
