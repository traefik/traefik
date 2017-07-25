package cmd

import "github.com/jawher/mow.cli"

var apiKey *string

// CLI struct for main
type CLI struct {
	*cli.Cli
}

// NewCLI initializes new command line interface
func NewCLI() *CLI {
	c := &CLI{cli.App("vultr", "A Vultr CLI")}

	apiKey = c.String(cli.StringOpt{
		Name:      "k api-key",
		Desc:      "Vultr API-Key",
		EnvVar:    "VULTR_API_KEY",
		HideValue: true,
	})

	return c
}
