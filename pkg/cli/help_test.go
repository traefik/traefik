package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintHelp(t *testing.T) {
	testCases := []struct {
		desc     string
		command  *Command
		expected string
	}{
		{
			desc: "no sub-command, with flags",
			command: func() *Command {
				element := &Yo{
					Fuu: "test",
				}

				return &Command{
					Name:          "root",
					Description:   "Description for root",
					Configuration: element,
					Run: func(args []string) error {
						return nil
					},
				}
			}(),
			expected: `root    Description for root

Usage: root [command] [flags] [arguments]

Use "root [command] --help" for help on any command.

Flag's usage: root [--flag=flag_argument] [-f [flag_argument]]    # set flag_argument to flag(s)
          or: root [--flag[=true|false| ]] [-f [true|false| ]]    # set true/false to boolean flag(s)

Flags:
    --fii  (Default: "fii")
        Fii description

    --foo  (Default: "foo")
        Foo description

    --fuu  (Default: "test")
        Fuu description

    --yi  (Default: "false")

    --yi.fii  (Default: "fii")

    --yi.foo  (Default: "foo")

    --yi.fuu  (Default: "")

    --yu.fii  (Default: "fii")

    --yu.foo  (Default: "foo")

    --yu.fuu  (Default: "")

`,
		},
		{
			desc: "with sub-commands, with flags, call root help",
			command: func() *Command {
				element := &Yo{
					Fuu: "test",
				}

				rootCmd := &Command{
					Name:          "root",
					Description:   "Description for root",
					Configuration: element,
					Run: func(_ []string) error {
						return nil
					},
				}

				err := rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "Description for sub1",
					Configuration: element,
					Run: func(args []string) error {
						return nil
					},
				})
				require.NoError(t, err)

				err = rootCmd.AddCommand(&Command{
					Name:          "sub2",
					Description:   "Description for sub2",
					Configuration: element,
					Run: func(args []string) error {
						return nil
					},
				})
				require.NoError(t, err)

				return rootCmd
			}(),
			expected: `root    Description for root

Usage: root [command] [flags] [arguments]

Use "root [command] --help" for help on any command.

Commands:
    sub1    Description for sub1
    sub2    Description for sub2

Flag's usage: root [--flag=flag_argument] [-f [flag_argument]]    # set flag_argument to flag(s)
          or: root [--flag[=true|false| ]] [-f [true|false| ]]    # set true/false to boolean flag(s)

Flags:
    --fii  (Default: "fii")
        Fii description

    --foo  (Default: "foo")
        Foo description

    --fuu  (Default: "test")
        Fuu description

    --yi  (Default: "false")

    --yi.fii  (Default: "fii")

    --yi.foo  (Default: "foo")

    --yi.fuu  (Default: "")

    --yu.fii  (Default: "fii")

    --yu.foo  (Default: "foo")

    --yu.fuu  (Default: "")

`,
		},
		{
			desc: "no sub-command, no flags",
			command: func() *Command {
				return &Command{
					Name:          "root",
					Description:   "Description for root",
					Configuration: nil,
					Run: func(args []string) error {
						return nil
					},
				}
			}(),
			expected: `root    Description for root

Usage: root [command] [flags] [arguments]

Use "root [command] --help" for help on any command.

`,
		},
		{
			desc: "no sub-command, slice flags",
			command: func() *Command {
				return &Command{
					Name:        "root",
					Description: "Description for root",
					Configuration: &struct {
						Foo []struct {
							Field string
						}
					}{},
					Run: func(args []string) error {
						return nil
					},
				}
			}(),
			expected: `root    Description for root

Usage: root [command] [flags] [arguments]

Use "root [command] --help" for help on any command.

Flag's usage: root [--flag=flag_argument] [-f [flag_argument]]    # set flag_argument to flag(s)
          or: root [--flag[=true|false| ]] [-f [true|false| ]]    # set true/false to boolean flag(s)

Flags:
    --foo  (Default: "")

    --foo[n].field  (Default: "")

`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			buffer := &bytes.Buffer{}
			err := PrintHelp(buffer, test.command)
			require.NoError(t, err)

			assert.Equal(t, test.expected, buffer.String())
		})
	}
}
