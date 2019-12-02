package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

func TestCommand_PrintHelp(t *testing.T) {
	testCases := []struct {
		desc           string
		command        *Command
		expectedOutput string
		expectedError  error
	}{
		{
			desc:           "print default help",
			command:        &Command{},
			expectedOutput: "    \n\nUsage:  [command] [flags] [arguments]\n\nUse \" [command] --help\" for help on any command.\n\n",
		},
		{
			desc: "print custom help",
			command: &Command{
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
				CustomHelpFunc: func(w io.Writer, _ *Command) error {
					_, _ = fmt.Fprintln(w, "test")
					return nil
				},
			},
			expectedOutput: "test\n",
		},
		{
			desc: "error is returned from called help",
			command: &Command{
				CustomHelpFunc: func(_ io.Writer, _ *Command) error {
					return errors.New("test")
				},
			},
			expectedError: errors.New("test"),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			buffer := &bytes.Buffer{}
			err := test.command.PrintHelp(buffer)

			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedOutput, buffer.String())
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
			desc: "root command, with argument, command not found",
			args: []string{"", "echo"},
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
			expected: expected{error: true},
		},
		{
			desc: "root command, call help, with argument, command not found",
			args: []string{"", "echo", "--help"},
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
			expected: expected{error: true},
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
			desc: "one sub command, with argument, command not found",
			args: []string{"", "sub1", "echo"},
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
			expected: expected{error: true},
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
					AllowArg:      true,
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
					AllowArg:      true,
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
		{
			desc: "sub command help",
			args: []string{"", "test", "subtest", "--help"},
			command: func() *Command {
				rootCmd := &Command{
					Name:      "test",
					Resources: []ResourceLoader{&FlagLoader{}},
				}

				subCmd := &Command{
					Name:      "subtest",
					Resources: []ResourceLoader{&FlagLoader{}},
				}

				err := rootCmd.AddCommand(subCmd)
				require.NoError(t, err)

				subSubCmd := &Command{
					Name:      "subsubtest",
					Resources: []ResourceLoader{&FlagLoader{}},
				}

				err = subCmd.AddCommand(subSubCmd)
				require.NoError(t, err)

				subSubSubCmd := &Command{
					Name:      "subsubsubtest",
					Resources: []ResourceLoader{&FlagLoader{}},
					Run: func([]string) error {
						called = "subsubsubtest"
						return nil
					},
				}

				err = subSubCmd.AddCommand(subSubSubCmd)
				require.NoError(t, err)

				return rootCmd
			},
			expected: expected{},
		},
		{
			desc: "sub sub command help",
			args: []string{"", "test", "subtest", "subsubtest", "--help"},
			command: func() *Command {
				rootCmd := &Command{
					Name:      "test",
					Resources: []ResourceLoader{&FlagLoader{}},
				}

				subCmd := &Command{
					Name:      "subtest",
					Resources: []ResourceLoader{&FlagLoader{}},
				}

				err := rootCmd.AddCommand(subCmd)
				require.NoError(t, err)

				subSubCmd := &Command{
					Name:      "subsubtest",
					Resources: []ResourceLoader{&FlagLoader{}},
				}

				err = subCmd.AddCommand(subSubCmd)
				require.NoError(t, err)

				subSubSubCmd := &Command{
					Name:      "subsubsubtest",
					Resources: []ResourceLoader{&FlagLoader{}},
					Run: func([]string) error {
						called = "subsubsubtest"
						return nil
					},
				}

				err = subSubCmd.AddCommand(subSubSubCmd)
				require.NoError(t, err)

				return rootCmd
			},
			expected: expected{},
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
	testCases := []struct {
		desc string
		args []string
	}{
		{
			desc: "configFile arg in camel case",
			args: []string{"", "sub1", "--configFile=./fixtures/config.toml"},
		},
		{
			desc: "configfile arg in lower case",
			args: []string{"", "sub1", "--configfile=./fixtures/config.toml"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
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

			err = execute(rootCmd, test.args, true)
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
		})
	}
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

func TestName(t *testing.T) {
	rootCmd := &Command{
		Name:      "test",
		Resources: []ResourceLoader{&FlagLoader{}},
	}

	subCmd := &Command{
		Name:      "subtest",
		Resources: []ResourceLoader{&FlagLoader{}},
	}

	err := rootCmd.AddCommand(subCmd)
	require.NoError(t, err)

	subSubCmd := &Command{
		Name:      "subsubtest",
		Resources: []ResourceLoader{&FlagLoader{}},
		Run: func([]string) error {
			return nil
		},
	}

	err = subCmd.AddCommand(subSubCmd)
	require.NoError(t, err)

	subSubSubCmd := &Command{
		Name:      "subsubsubtest",
		Resources: []ResourceLoader{&FlagLoader{}},
		Run: func([]string) error {
			return nil
		},
	}

	err = subSubCmd.AddCommand(subSubSubCmd)
	require.NoError(t, err)

	err = execute(rootCmd, []string{"", "test", "subtest", "subsubtest", "subsubsubtest", "--help"}, true)
	require.NoError(t, err)
}
