package ech

import (
	"bytes"
	"flag"
	"github.com/pkg/errors"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/traefik/v3/pkg/tls"
	"io"
	stdlog "log"
	"os"
)

// NewCmd builds a new Version command.
func NewCmd() *cli.Command {
	cmd := cli.Command{
		Name:        "ech",
		Description: `Encrypted Client Hello (ECH) utils.`,
		Run:         nil,
	}

	var err error
	err = cmd.AddCommand(generate())
	if err != nil {
		stdlog.Println(err)
		os.Exit(1)
	}

	err = cmd.AddCommand(request())
	if err != nil {
		stdlog.Println(err)
		os.Exit(1)
	}

	buf := bytes.NewBuffer(nil)
	err = cmd.PrintHelp(buf)
	if err != nil {
		stdlog.Println(err)
		os.Exit(1)
	}
	cmd.Run = func(_ []string) error {
		if _, err = os.Stdout.Write(buf.Bytes()); err != nil {
			return err
		}
		return nil
	}

	return &cmd
}

func generate() *cli.Command {
	help := []byte(`Usage: ech generate SNI [SNI ...]`)
	cmd := cli.Command{
		Name:        "generate",
		Description: "Generate a new ECH key with given outer sni.",
		AllowArg:    true,
		CustomHelpFunc: func(writer io.Writer, command *cli.Command) error {
			_, err := writer.Write(help)
			return err
		},
		Run: func(names []string) error {
			if len(names) == 0 {
				if _, err := os.Stdout.Write(help); err != nil {
					return err
				}
			}
			for _, name := range names {
				key, err := tls.NewECHKey(name)
				if err != nil {
					return errors.Wrapf(err, "failed to generate ECH key for %s", name)
				}
				data, err := tls.MarshalECHKey(key)
				if err != nil {
					return errors.Wrapf(err, "failed to marshal ECH key for %s", name)
				}
				if _, err = os.Stdout.Write(data); err != nil {
					return errors.Wrapf(err, "failed to write ECH key for %s", name)
				}
			}
			return nil
		},
	}
	return &cmd
}

func request() *cli.Command {
	return &cli.Command{
		Name:        "request",
		Description: "Make an ECH request.",
		AllowArg:    true,
		CustomHelpFunc: func(writer io.Writer, command *cli.Command) error {
			_, err := writer.Write([]byte(`Usage: ech request URL -e ECH [-h HOST] [-k])`))
			return err
		},
		Run: func(args []string) error {
			c := tls.ECHRequestConf[string]{}
			fs := flag.NewFlagSet("ech request", flag.ContinueOnError)
			c.URL = fs.Arg(0)
			c.ECH = *fs.String("ech", "", "A base64-encoded ECH public config list.")
			c.Host = *fs.String("h", "", "The host to use in the request. If not set, it will be derived from the URL.")
			c.Insecure = *fs.Bool("k", false, "Allow insecure server connections when using SSL.")
			fs.SetOutput(os.Stderr)
			if err := fs.Parse(args); err != nil {
				return err
			}

			_, err := tls.RequestWithECH(c)
			return err
		},
	}
}
