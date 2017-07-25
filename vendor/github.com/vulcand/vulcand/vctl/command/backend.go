package command

import (
	"github.com/codegangsta/cli"
	"github.com/vulcand/vulcand/engine"
)

func NewBackendCommand(cmd *Command) cli.Command {
	return cli.Command{
		Name:  "backend",
		Usage: "Operations with backends",
		Subcommands: []cli.Command{
			{
				Name:   "upsert",
				Usage:  "Update or insert a new backend to vulcan",
				Action: cmd.upsertBackendAction,
				Flags: append(append([]cli.Flag{
					cli.StringFlag{Name: "id", Usage: "backend id"}},
					backendOptions()...),
					getTLSFlags()...),
			},
			{
				Name:   "rm",
				Usage:  "Remove backend from vulcan",
				Action: cmd.deleteBackendAction,
				Flags: []cli.Flag{
					cli.StringFlag{Name: "id", Usage: "backend id"},
				},
			},
			{
				Name:   "ls",
				Usage:  "List backends",
				Action: cmd.listBackendsAction,
			},
			{
				Name:   "show",
				Usage:  "Show backend",
				Action: cmd.printBackendAction,
				Flags: []cli.Flag{
					cli.StringFlag{Name: "id", Usage: "backend id"},
				},
			},
		},
	}
}

func (cmd *Command) upsertBackendAction(c *cli.Context) {
	settings, err := getBackendSettings(c)
	if err != nil {
		cmd.printError(err)
		return
	}
	b, err := engine.NewHTTPBackend(c.String("id"), settings)
	if err != nil {
		cmd.printError(err)
		return
	}
	cmd.printResult("%s upserted", b, cmd.client.UpsertBackend(*b))
}

func (cmd *Command) deleteBackendAction(c *cli.Context) {
	if err := cmd.client.DeleteBackend(engine.BackendKey{Id: c.String("id")}); err != nil {
		cmd.printError(err)
	} else {
		cmd.printOk("backend deleted")
	}
}

func (cmd *Command) printBackendAction(c *cli.Context) {
	bk := engine.BackendKey{Id: c.String("id")}
	b, err := cmd.client.GetBackend(bk)
	if err != nil {
		cmd.printError(err)
		return
	}
	srvs, err := cmd.client.GetServers(bk)
	if err != nil {
		cmd.printError(err)
		return
	}
	cmd.printBackend(b, srvs)
}

func (cmd *Command) listBackendsAction(c *cli.Context) {
	out, err := cmd.client.GetBackends()
	if err != nil {
		cmd.printError(err)
	} else {
		cmd.printBackends(out)
	}
}

func getBackendSettings(c *cli.Context) (engine.HTTPBackendSettings, error) {
	s := engine.HTTPBackendSettings{}

	s.Timeouts.Read = c.Duration("readTimeout").String()
	s.Timeouts.Dial = c.Duration("dialTimeout").String()
	s.Timeouts.TLSHandshake = c.Duration("handshakeTimeout").String()

	s.KeepAlive.Period = c.Duration("keepAlivePeriod").String()
	s.KeepAlive.MaxIdleConnsPerHost = c.Int("maxIdleConns")

	tlsSettings, err := getTLSSettings(c)
	if err != nil {
		return s, err
	}
	s.TLS = tlsSettings
	return s, nil
}

func backendOptions() []cli.Flag {
	return []cli.Flag{
		// Timeouts
		cli.DurationFlag{Name: "readTimeout", Usage: "read timeout"},
		cli.DurationFlag{Name: "dialTimeout", Usage: "dial timeout"},
		cli.DurationFlag{Name: "handshakeTimeout", Usage: "TLS handshake timeout"},

		// Keep-alive parameters
		cli.StringFlag{Name: "keepAlivePeriod", Usage: "keep-alive period"},
		cli.IntFlag{Name: "maxIdleConns", Usage: "maximum idle connections per host"},
	}
}
