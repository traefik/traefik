package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/plugin"
)

func NewMiddlewareCommands(cmd *Command) []cli.Command {
	out := []cli.Command{}
	for _, spec := range cmd.registry.GetSpecs() {
		if spec.CliFlags != nil && spec.FromCli != nil {
			out = append(out, makeMiddlewareCommands(cmd, spec))
		}
	}
	return out
}

func makeMiddlewareCommands(cmd *Command, spec *plugin.MiddlewareSpec) cli.Command {
	flags := append([]cli.Flag{}, spec.CliFlags...)
	flags = append(flags,
		cli.StringFlag{Name: "frontend, f", Usage: "location id"},
		cli.DurationFlag{Name: "ttl", Usage: "ttl"},
		cli.IntFlag{Name: "priority", Value: 1, Usage: "middleware priority, smaller values are lower"},
		cli.StringFlag{Name: "id", Usage: fmt.Sprintf("%s id", spec.Type)})

	return cli.Command{
		Name:  spec.Type,
		Usage: fmt.Sprintf("Operations on %s middlewares", spec.Type),
		Subcommands: []cli.Command{
			{
				Name:   "upsert",
				Usage:  fmt.Sprintf("Add new or update new %s to frontend", spec.Type),
				Flags:  flags,
				Action: makeUpsertMiddlewareAction(cmd, spec),
			},
			{
				Name:   "rm",
				Usage:  fmt.Sprintf("Remove %s from frontend", spec.Type),
				Action: makeDeleteMiddlewareAction(cmd, spec),
				Flags: []cli.Flag{
					cli.StringFlag{Name: "frontend, f", Usage: "Frontend id"},
					cli.StringFlag{Name: "id", Usage: fmt.Sprintf("%s id", spec.Type)},
				},
			},
		},
	}
}

func makeUpsertMiddlewareAction(cmd *Command, spec *plugin.MiddlewareSpec) func(c *cli.Context) {
	return func(c *cli.Context) {
		m, err := spec.FromCli(c)
		if err != nil {
			cmd.printError(err)
		} else {
			mi := engine.Middleware{Id: c.String("id"), Middleware: m, Type: spec.Type, Priority: c.Int("priority")}
			err := cmd.client.UpsertMiddleware(engine.FrontendKey{Id: c.String("frontend")}, mi, c.Duration("ttl"))
			if err != nil {
				cmd.printError(err)
				return
			}
			cmd.printOk("%v upserted", spec.Type)
		}
	}
}

func makeDeleteMiddlewareAction(cmd *Command, spec *plugin.MiddlewareSpec) func(c *cli.Context) {
	return func(c *cli.Context) {
		mk := engine.MiddlewareKey{FrontendKey: engine.FrontendKey{Id: c.String("frontend")}, Id: c.String("id")}
		if err := cmd.client.DeleteMiddleware(mk); err != nil {
			cmd.printError(err)
			return
		}
		cmd.printOk("%v deleted", spec.Type)
	}
}
