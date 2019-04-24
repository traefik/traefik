# Flæg

[![Travis branch](https://img.shields.io/travis/containous/flaeg/master.svg)](https://travis-ci.org/containous/flaeg)
[![Coverage Status](https://coveralls.io/repos/github/containous/flaeg/badge.svg?branch=master)](https://coveralls.io/github/containous/flaeg?branch=master)
[![license](https://img.shields.io/github/license/containous/flaeg.svg)](https://github.com/containous/flaeg/blob/master/LICENSE.md)

Flæg is a Go library for building dynamically a powerful modern Command Line Interface and loading a program configuration structure from arguments.
Go developers don't need to worry about keeping flags and commands updated anymore: it works by itself!

## Overview

You know how boring it is to keep your CLI up-to-date. You will be glad to use Flaeg ;-)
This package uses your own configuration structure to build your CLI.

You only need to describe every `StructField` with a `StructTag`,  flaeg will automatically build the CLI, parse data from args, and load Go values into Configuration structure via reflection!

We developed `flaeg` and [`staert`](https://github.com/containous/staert) in order to simplify configuration maintenance on [traefik](https://github.com/containous/traefik).

## Features

- Load your Configuration structure with program args
- Keep your Configuration structure values unchanged if no flags called (support defaults values)
- Many `Type` of `StructField` can be flagged :
	- type `bool`
	- type `int` (`int32`, `int64`, `uint`, `uint64`)
	- type `string`
	- type `float` (`float64`)
	- type `time.Duration`
    	- type `time.Time`
- Many `Kind` of `StructField` in the Configuration structure are supported :
	- Sub-Structure
	- Anonymous field (on Sub-Structure)
	- Pointers on anything
- You can add your "Parsers" on your own type like :
	- Arrays, Slices or Maps
	- Your structures
- Pointers flags are Boolean :
	- You can give  a structure of default values for those pointers
	- Pointer fields will get default values if their flag is called
- Flags names are fields names by default, but you can overwrite it in `StructTag`
- "Shorthand" flags (1 character) can be added in `StructTag` as well
- Flaeg is POSIX compliant using [pflag](https://github.com/ogier/pflag) package
- You only need to provide the root-Command which contains the function to run  
- You can add Sub-Commands to the root-Command

## Getting Started

### Installation

To install `Flaeg`, simply run:

```
$ go get github.com/containous/flaeg
```

### Configuration Structures

Flaeg works on any kind of structure, you only need to add a `StructTag` "description" on the fields to flag.
Like this:

```go
// Configuration is a struct which contains all differents type to field
// using parsers on string, time.Duration, pointer, bool, int, int64, time.Time, float64
type Configuration struct {
	Name     string        // no description struct tag, it will not be flaged
	LogLevel string        `short:"l" description:"Log level"`      // string type field, short flag "-l"
	Timeout  time.Duration `description:"Timeout duration"`         // time.Duration type field
	Db       *DatabaseInfo `description:"Enable database"`          // pointer type field (on DatabaseInfo)
	Owner    *OwnerInfo    `description:"Enable Owner description"` // another pointer type field (on OwnerInfo)
}
```

You can add sub-structures even if they are anonymous:

```go
type ServerInfo struct {
	Watch  bool   `description:"Watch device"`      // bool type
	IP     string `description:"Server ip address"` // string type field
	Load   int    `description:"Server load"`       // int type field
	Load64 int64  `description:"Server load"`       // int64 type field, same description just to be sure it works
}

type DatabaseInfo struct {
	ServerInfo             // anonymous sub-structures
	ConnectionMax   uint   `long:"comax" description:"Number max of connections on database"` // uint type field, long flag "--comax"
	ConnectionMax64 uint64 `description:"Number max of connections on database"`              // uint64 type field, same description just to be sure it works
}

type OwnerInfo struct {
	Name        *string      `description:"Owner name"`                     // pointer type field on string
	DateOfBirth time.Time    `long:"dob" description:"Owner date of birth"` // time.Time type field, long flag "--dob"
	Rate        float64      `description:"Owner rate"`                     // float64 type field
	Servers     []ServerInfo `description:"Owner Server"`                   // slice of ServerInfo type field, need a custom parser
}
```

### Flags

Flaeg is POSIX compliant using [pflag](https://github.com/ogier/pflag) package.
Flaeg concats the names of fields to generate the flags. They are not case sensitive.

For example, the field `ConnectionMax64` in `OwnerInfo` sub-Structure which is in `Configuration` Structure will be `--db.connectionmax64`.
But you can overwrite it with the `StructTag` `long` as like as the field `ConnectionMax` which is flagged `--db.comax`.

Finally, you can add a short flag (1 character) using the `StructTag` `short`, like in the field `LogLevel` with the short flags `-l` in addition to the flag`--loglevel`.

### Default values

Default values on fields come from the configuration structure. If it was not initialized, Golang default values are used.

For pointers, the `DefaultPointers` structure provides default values.

### Command

The `Command` structure contains program/command information (command name and description).
`Config` must be a pointer on the configuration struct to parse (it contains default values of field).
`DefaultPointersConfig` contains default pointers values: those values are set on pointers fields if their flags are called.

It must be the same type (struct) as `Config`.
`Run` is the func which launch the program using initialized configuration structure.

```go
type Command struct {
	Name                  string
	Description           string
	Config                interface{}
	DefaultPointersConfig interface{}
	Run                   func() error		
	Metadata              map[string]string
}
```

So, you can create Commands like this:

```go
rootCmd := &Command{
	Name: "flaegtest",
	Description: `flaegtest is a test program made to to test flaeg library.
	Complete documentation is available at https://github.com/containous/flaeg`,
	Config:                config,
	DefaultPointersConfig: defaultPointers,
	Run: func() error {
			fmt.Printf("Run flaegtest command with config : %+v\n", config)
			return nil
		},
	}
```

You have to create at least the root-Command, and you can add some sub-Command.

Metadata allows you to store some labels(Key-value) in the command and to use it elsewhere.
We needed that in [Stært](https://github.com/containous/staert).

### Help

The responsive help is auto-generated using the `description` `StructTag`, default value from configuration structure and/or `Command` structure.
Flag `--help` and short flag `-h` are bound to call the helper.
If the args parser fails, it will print the error and the helper will be call as well.

Here an example:

```
$./flaegtest --help
flaegtest is a test program made to test flaeg library.
Complete documentation is available at https://github.com/containous/flaeg

Usage: flaegtest [--flag=flag_argument] [-f[flag_argument]] ...     set flag_argument to flag(s)
   or: flaegtest [--flag[=true|false| ]] [-f[true|false| ]] ...     set true/false to boolean flag(s)

Available Commands:
	version                                            Print version
Use "flaegtest [command] --help" for more information about a command.

Flags:
    --db                 Enable database                       (default "false")
    --db.comax           Number max of connections on database (default "3200000000")
    --db.connectionmax64 Number max of connections on database (default "6400000000000000000")
    --db.ip              Server ip address                     (default "192.168.1.2")
    --db.load            Server load                           (default "32")
    --db.load64          Server load                           (default "64")
    --db.watch           Watch device                          (default "true")
-l, --loglevel           Log level                             (default "DEBUG")
    --owner              Enable Owner description              (default "true")
    --owner.dob          Owner date of birth                   (default "1993-09-12 07:32:00 +0000 UTC")
    --owner.name         Owner name                            (default "true")
    --owner.rate         Owner rate                            (default "0.999")
    --owner.servers      Owner Server                          (default "[]")
    --timeout            Timeout duration                      (default "1s")
-h, --help               Print Help (this message) and exit
```


### Run Flaeg

Let's run fleag now:

- `rootCmd` is the root-Command
- `versionCmd` is a sub-Command

```go
	// init flaeg
	flaeg := flaeg.New(rootCmd, os.Args[1:])
	// add sub-command Version
	flaeg.AddCommand(versionCmd)

	// run test
	if err := flaeg.Run(); err != nil {
		t.Errorf("Error %s", err.Error())
	}
}
```

### Custom Parsers

The function `flaeg.AddParser` adds a custom parser for a specified type.

```go
func (f *Flaeg) AddParser(typ reflect.Type, parser Parser)
```

It can be used like this:

```go
// add custom parser to fleag
flaeg.AddParser(reflect.TypeOf([]ServerInfo{}), &sliceServerValue{})
```

`sliceServerValue{}` need to implement `flaeg.Parser`:

```go
type Parser interface {
	flag.Getter
	SetValue(interface{})
}
```

like this:

```go
type sliceServerValue []ServerInfo

func (c *sliceServerValue) Set(s string) error {
	// could use RegExp
	srv := ServerInfo{IP: s}
	*c = append(*c, srv)
	return nil
}

func (c *sliceServerValue) Get() interface{} { return []ServerInfo(*c) }

func (c *sliceServerValue) String() string { return fmt.Sprintf("%v", *c) }

func (c *sliceServerValue) SetValue(val interface{}) {
	*c = sliceServerValue(val.([]ServerInfo))
}
```

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D
