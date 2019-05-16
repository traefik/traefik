package cli

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_AddCommand(t *testing.T) {
	testCases := []struct {
		desc          string
		subCommand    *Command
		expectedError bool
	}{
		{
			desc:       "sub command nil",
			subCommand: nil,
		},
		{
			desc: "add a simple command",
			subCommand: &Command{
				Name: "sub",
			},
		},
		{
			desc: "add a sub command with the same name as their parent",
			subCommand: &Command{
				Name: "root",
			},
			expectedError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rootCmd := &Command{
				Name: "root",
			}

			err := rootCmd.AddCommand(test.subCommand)

			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_execute(t *testing.T) {
	var called string

	type expected struct {
		result string
		error  bool
	}

	testCases := []struct {
		desc     string
		args     []string
		command  func() *Command
		expected expected
	}{
		{
			desc: "root command",
			args: []string{""},
			command: func() *Command {
				return &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called = "root"
						return nil
					},
				}

			},
			expected: expected{result: "root"},
		},
		{
			desc: "one sub command",
			args: []string{"", "sub1"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "test",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "root"
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub1"
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "sub1"},
		},
		{
			desc: "two sub commands",
			args: []string{"", "sub2"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "test",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "root"
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub1"
						return nil
					},
				})

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub2",
					Description:   "sub2",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub2"
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "sub2"},
		},
		{
			desc: "command with sub sub command, call sub command",
			args: []string{"", "sub1"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "test",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "root"
						return nil
					},
				}

				sub1 := &Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub1"
						return nil
					},
				}
				_ = rootCmd.AddCommand(sub1)

				_ = sub1.AddCommand(&Command{
					Name:          "sub2",
					Description:   "sub2",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub2"
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "sub1"},
		},
		{
			desc: "command with sub sub command, call sub sub command",
			args: []string{"", "sub1", "sub2"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "test",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "root"
						return nil
					},
				}

				sub1 := &Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub1"
						return nil
					},
				}
				_ = rootCmd.AddCommand(sub1)

				_ = sub1.AddCommand(&Command{
					Name:          "sub2",
					Description:   "sub2",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub2"
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "sub2"},
		},
		{
			desc: "command with sub command, call root command explicitly",
			args: []string{"", "root"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "root"
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub1"
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "root"},
		},
		{
			desc: "command with sub command, call root command implicitly",
			args: []string{""},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "root"
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub1"
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "root"},
		},
		{
			desc: "command with sub command, call sub command which has no run",
			args: []string{"", "sub1"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "root"
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
				})

				return rootCmd
			},
			expected: expected{error: true},
		},
		{
			desc: "command with sub command, call root command which has no run",
			args: []string{"", "root"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub1"
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{error: true},
		},
		{
			desc: "command with sub command, call implicitly root command which has no run",
			args: []string{""},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(_ []string) error {
						called += "sub1"
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{error: true},
		},
		{
			desc: "command with sub command, call sub command with arguments",
			args: []string{"", "sub1", "foobar.txt"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called = "root"
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(args []string) error {
						called += "sub1-" + strings.Join(args, "-")
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "sub1-foobar.txt"},
		},
		{
			desc: "command with sub command, call root command with arguments",
			args: []string{"", "foobar.txt"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(args []string) error {
						called += "root-" + strings.Join(args, "-")
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(args []string) error {
						called += "sub1-" + strings.Join(args, "-")
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "root-foobar.txt"},
		},
		{
			desc: "command with sub command, call sub command with flags",
			args: []string{"", "sub1", "--foo=bar", "--fii=bir"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(_ []string) error {
						called = "root"
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(args []string) error {
						called += "sub1-" + strings.Join(args, "")
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "sub1---foo=bar--fii=bir"},
		},
		{
			desc: "command with sub command, call explicitly root command with flags",
			args: []string{"", "root", "--foo=bar", "--fii=bir"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(args []string) error {
						called += "root-" + strings.Join(args, "")
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(args []string) error {
						called += "sub1-" + strings.Join(args, "")
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "root---foo=bar--fii=bir"},
		},
		{
			desc: "command with sub command, call implicitly root command with flags",
			args: []string{"", "--foo=bar", "--fii=bir"},
			command: func() *Command {
				rootCmd := &Command{
					Name:          "root",
					Description:   "This is a test",
					Configuration: nil,
					Run: func(args []string) error {
						called += "root-" + strings.Join(args, "")
						return nil
					},
				}

				_ = rootCmd.AddCommand(&Command{
					Name:          "sub1",
					Description:   "sub1",
					Configuration: nil,
					Run: func(args []string) error {
						called += "sub1-" + strings.Join(args, "")
						return nil
					},
				})

				return rootCmd
			},
			expected: expected{result: "root---foo=bar--fii=bir"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer func() {
				called = ""
			}()

			err := execute(test.command(), test.args, true)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.result, called)
			}
		})
	}
}

func Test_execute_configuration(t *testing.T) {
	rootCmd := &Command{
		Name:          "root",
		Description:   "This is a test",
		Configuration: nil,
		Run: func(_ []string) error {
			return nil
		},
	}

	element := &Yo{
		Fuu: "test",
	}

	sub1 := &Command{
		Name:          "sub1",
		Description:   "sub1",
		Configuration: element,
		Resources:     []ResourceLoader{&FlagLoader{}},
		Run: func(args []string) error {
			return nil
		},
	}
	err := rootCmd.AddCommand(sub1)
	require.NoError(t, err)

	args := []string{"", "sub1", "--foo=bar", "--fii=bir", "--yi"}

	err = execute(rootCmd, args, true)
	require.NoError(t, err)

	expected := &Yo{
		Foo: "bar",
		Fii: "bir",
		Fuu: "test",
		Yi: &Yi{
			Foo: "foo",
			Fii: "fii",
		},
	}
	assert.Equal(t, expected, element)
}

func Test_execute_configuration_file(t *testing.T) {
	rootCmd := &Command{
		Name:          "root",
		Description:   "This is a test",
		Configuration: nil,
		Run: func(_ []string) error {
			return nil
		},
	}

	element := &Yo{
		Fuu: "test",
	}

	sub1 := &Command{
		Name:          "sub1",
		Description:   "sub1",
		Configuration: element,
		Resources:     []ResourceLoader{&FileLoader{}, &FlagLoader{}},
		Run: func(args []string) error {
			return nil
		},
	}
	err := rootCmd.AddCommand(sub1)
	require.NoError(t, err)

	args := []string{"", "sub1", "--configFile=./fixtures/config.toml"}

	err = execute(rootCmd, args, true)
	require.NoError(t, err)

	expected := &Yo{
		Foo: "bar",
		Fii: "bir",
		Fuu: "test",
		Yi: &Yi{
			Foo: "foo",
			Fii: "fii",
		},
	}
	assert.Equal(t, expected, element)
}

func Test_execute_help(t *testing.T) {
	element := &Yo{
		Fuu: "test",
	}

	rooCmd := &Command{
		Name:          "root",
		Description:   "Description for root",
		Configuration: element,
		Run: func(args []string) error {
			return nil
		},
	}

	args := []string{"", "--help", "--foo"}

	backupStdout := os.Stdout
	defer func() {
		os.Stdout = backupStdout
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w

	err := execute(rooCmd, args, true)
	if err != nil {
		return
	}

	// read and restore stdout
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = backupStdout

	assert.Equal(t, `root    Description for root

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

`, string(out))
}

type Yo struct {
	Foo string `description:"Foo description"`
	Fii string `description:"Fii description"`
	Fuu string `description:"Fuu description"`
	Yi  *Yi    `label:"allowEmpty"`
	Yu  *Yi
}

func (y *Yo) SetDefaults() {
	y.Foo = "foo"
	y.Fii = "fii"
}

type Yi struct {
	Foo string
	Fii string
	Fuu string
}

func (y *Yi) SetDefaults() {
	y.Foo = "foo"
	y.Fii = "fii"
}
