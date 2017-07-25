package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/vulcand/vulcand/api"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/plugin"
	"github.com/vulcand/vulcand/secret"
)

type Command struct {
	vulcanUrl string
	client    *api.Client
	out       io.Writer
	registry  *plugin.Registry
}

func NewCommand(registry *plugin.Registry) *Command {
	return &Command{
		out:      os.Stdout,
		registry: registry,
	}
}

func (cmd *Command) Run(args []string) error {
	url, args, err := findVulcanUrl(args)
	if err != nil {
		return err
	}
	cmd.vulcanUrl = url
	cmd.client = api.NewClient(cmd.vulcanUrl, cmd.registry)

	app := cli.NewApp()
	app.Name = "vctl"
	app.Usage = "Command line interface to a running vulcan instance"
	app.Flags = flags()

	app.Commands = []cli.Command{
		NewLogCommand(cmd),
		NewKeyCommand(cmd),
		NewTopCommand(cmd),
		NewHostCommand(cmd),
		NewBackendCommand(cmd),
		NewFrontendCommand(cmd),
		NewServerCommand(cmd),
		NewListenerCommand(cmd),
	}
	app.Commands = append(app.Commands, NewMiddlewareCommands(cmd)...)
	return app.Run(args)
}

// This function extracts vulcan url from the command line regardless of it's position
// this is a workaround, as cli libary does not support "superglobal" urls yet.
func findVulcanUrl(args []string) (string, []string, error) {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--vulcan=") || strings.HasPrefix(arg, "-vulcan=") {
			out := strings.Split(arg, "=")
			return out[1], cut(i, i+1, args), nil
		} else if strings.HasPrefix(arg, "-vulcan") || strings.HasPrefix(arg, "--vulcan") {
			// This argument should not be the last one
			if i > len(args)-2 {
				return "", nil, fmt.Errorf("provide a valid vulcan URL")
			}
			return args[i+1], cut(i, i+2, args), nil
		}
	}
	return "http://localhost:8182", args, nil
}

func cut(i, j int, args []string) []string {
	s := []string{}
	s = append(s, args[:i]...)
	return append(s, args[j:]...)
}

func flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "vulcan", Value: "http://localhost:8182", Usage: "Url for vulcan server"},
	}
}

func readKeyPair(certPath, keyPath string) (*engine.KeyPair, error) {
	fKey, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	defer fKey.Close()
	key, err := ioutil.ReadAll(fKey)
	if err != nil {
		return nil, err
	}

	fCert, err := os.Open(certPath)
	if err != nil {
		return nil, err
	}
	defer fCert.Close()
	cert, err := ioutil.ReadAll(fCert)
	if err != nil {
		return nil, err
	}
	return engine.NewKeyPair(cert, key)
}

func readBox(key string) (*secret.Box, error) {
	keyB, err := secret.KeyFromString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption key: %s", err)
	}
	return secret.NewBox(keyB)
}
