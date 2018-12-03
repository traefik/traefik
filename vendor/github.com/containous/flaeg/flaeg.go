package flaeg

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/containous/flaeg/parse"
	flag "github.com/ogier/pflag"
)

// ErrParserNotFound is thrown when a field is flaged but not parser match its type
var ErrParserNotFound = errors.New("parser not found or custom parser missing")

// GetTypesRecursive links in flagMap a flag with its reflect.StructField
// You can whether provide objValue on a structure or a pointer to structure as first argument
// Flags are generated from field name or from StructTag
func getTypesRecursive(objValue reflect.Value, flagMap map[string]reflect.StructField, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:
		for i := 0; i < objValue.NumField(); i++ {
			if objValue.Type().Field(i).Anonymous {
				if err := getTypesRecursive(objValue.Field(i), flagMap, name); err != nil {
					return err
				}
			} else if len(objValue.Type().Field(i).Tag.Get("description")) > 0 {
				fieldName := objValue.Type().Field(i).Name
				if !isExported(fieldName) {
					return fmt.Errorf("field %s is an unexported field", fieldName)
				}

				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}

				if len(key) == 0 {
					name = strings.ToLower(fieldName)
				} else {
					name = key + "." + strings.ToLower(fieldName)
				}

				if _, ok := flagMap[name]; ok {
					return fmt.Errorf("tag already exists: %s", name)
				}
				flagMap[name] = objValue.Type().Field(i)

				if err := getTypesRecursive(objValue.Field(i), flagMap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Ptr:
		if len(key) > 0 {
			field := flagMap[name]
			field.Type = reflect.TypeOf(false)
			flagMap[name] = field
		}

		typ := objValue.Type().Elem()
		inst := reflect.New(typ).Elem()

		if err := getTypesRecursive(inst, flagMap, name); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

// GetBoolFlags returns flags on pointers
func GetBoolFlags(config interface{}) ([]string, error) {
	flagMap := make(map[string]reflect.StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagMap, ""); err != nil {
		return []string{}, err
	}

	flags := make([]string, 0, len(flagMap))
	for f, structField := range flagMap {
		if structField.Type.Kind() == reflect.Bool {
			flags = append(flags, f)
		}
	}
	return flags, nil
}

// GetFlags returns flags
func GetFlags(config interface{}) ([]string, error) {
	flagMap := make(map[string]reflect.StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagMap, ""); err != nil {
		return []string{}, err
	}

	flags := make([]string, 0, len(flagMap))
	for f := range flagMap {
		flags = append(flags, f)
	}
	return flags, nil
}

// ParseArgs : parses args return a map[flag]Getter, using parsers map[type]Getter
// args must be formatted as like as flag documentation. See https://golang.org/pkg/flag
func parseArgs(args []string, flagMap map[string]reflect.StructField, parsers map[reflect.Type]parse.Parser) (map[string]parse.Parser, error) {
	newParsers := map[string]parse.Parser{}
	flagSet := flag.NewFlagSet("flaeg.Load", flag.ContinueOnError)

	// Disable output
	flagSet.SetOutput(ioutil.Discard)

	var err error
	for flg, structField := range flagMap {
		if parser, ok := parsers[structField.Type]; ok {
			newParserValue := reflect.New(reflect.TypeOf(parser).Elem())
			newParserValue.Elem().Set(reflect.ValueOf(parser).Elem())
			newParser := newParserValue.Interface().(parse.Parser)

			if short := structField.Tag.Get("short"); len(short) == 1 {
				flagSet.VarP(newParser, flg, short, structField.Tag.Get("description"))
			} else {
				flagSet.Var(newParser, flg, structField.Tag.Get("description"))
			}
			newParsers[flg] = newParser
		} else {
			err = ErrParserNotFound
		}
	}

	// prevents case sensitivity issue
	args = argsToLower(args)
	if errParse := flagSet.Parse(args); errParse != nil {
		return nil, errParse
	}

	// Visitor in flag.Parse
	var flagList []*flag.Flag
	visitor := func(fl *flag.Flag) {
		flagList = append(flagList, fl)
	}

	// Fill flagList with parsed flags
	flagSet.Visit(visitor)

	// Return var
	valMap := make(map[string]parse.Parser)

	// Return parsers on parsed flag
	for _, flg := range flagList {
		valMap[flg.Name] = newParsers[flg.Name]
	}

	return valMap, err
}

func getDefaultValue(defaultValue reflect.Value, defaultPointersValue reflect.Value, defaultValmap map[string]reflect.Value, key string) error {
	if defaultValue.Type() != defaultPointersValue.Type() {
		return fmt.Errorf("parameters defaultValue and defaultPointersValue must be the same struct. defaultValue type: %s is not defaultPointersValue type: %s", defaultValue.Type().String(), defaultPointersValue.Type().String())
	}

	name := key
	switch defaultValue.Kind() {
	case reflect.Struct:
		for i := 0; i < defaultValue.NumField(); i++ {
			if defaultValue.Type().Field(i).Anonymous {
				if err := getDefaultValue(defaultValue.Field(i), defaultPointersValue.Field(i), defaultValmap, name); err != nil {
					return err
				}
			} else if len(defaultValue.Type().Field(i).Tag.Get("description")) > 0 {
				fieldName := defaultValue.Type().Field(i).Name
				if tag := defaultValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}

				if len(key) == 0 {
					name = strings.ToLower(fieldName)
				} else {
					name = key + "." + strings.ToLower(fieldName)
				}

				if defaultValue.Field(i).Kind() != reflect.Ptr {
					defaultValmap[name] = defaultValue.Field(i)
				}
				if err := getDefaultValue(defaultValue.Field(i), defaultPointersValue.Field(i), defaultValmap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Ptr:
		if !defaultPointersValue.IsNil() {
			if len(key) != 0 {
				// turn ptr fields to nil
				defaultPointersNilValue, err := setPointersNil(defaultPointersValue)
				if err != nil {
					return err
				}
				defaultValmap[name] = defaultPointersNilValue
			}

			if !defaultValue.IsNil() {
				if err := getDefaultValue(defaultValue.Elem(), defaultPointersValue.Elem(), defaultValmap, name); err != nil {
					return err
				}
			} else {
				if err := getDefaultValue(defaultPointersValue.Elem(), defaultPointersValue.Elem(), defaultValmap, name); err != nil {
					return err
				}
			}
		} else {
			instValue := reflect.New(defaultPointersValue.Type().Elem())
			if len(key) != 0 {
				defaultValmap[name] = instValue
			}

			if !defaultValue.IsNil() {
				if err := getDefaultValue(defaultValue.Elem(), instValue.Elem(), defaultValmap, name); err != nil {
					return err
				}
			} else {
				if err := getDefaultValue(instValue.Elem(), instValue.Elem(), defaultValmap, name); err != nil {
					return err
				}
			}
		}
	default:
		return nil
	}
	return nil
}

// objValue a reflect.Value of a not-nil pointer on a struct
func setPointersNil(objValue reflect.Value) (reflect.Value, error) {
	if objValue.Kind() != reflect.Ptr {
		return objValue, fmt.Errorf("parameters objValue must be a not-nil pointer on a struct, not a %s", objValue.Kind())
	} else if objValue.IsNil() {
		return objValue, errors.New("parameters objValue must be a not-nil pointer")
	} else if objValue.Elem().Kind() != reflect.Struct {
		// fmt.Printf("Parameters objValue must be a not-nil pointer on a struct, not a pointer on a %s\n", objValue.Elem().Kind().String())
		return objValue, nil
	}

	// Clone
	starObjValue := objValue.Elem()
	nilPointersObjVal := reflect.New(starObjValue.Type())
	starNilPointersObjVal := nilPointersObjVal.Elem()
	starNilPointersObjVal.Set(starObjValue)

	for i := 0; i < nilPointersObjVal.Elem().NumField(); i++ {
		if field := nilPointersObjVal.Elem().Field(i); field.Kind() == reflect.Ptr && field.CanSet() {
			field.Set(reflect.Zero(field.Type()))
		}
	}
	return nilPointersObjVal, nil
}

// FillStructRecursive initialize a value of any tagged Struct given by reference
func fillStructRecursive(objValue reflect.Value, defaultPointerValMap map[string]reflect.Value, valMap map[string]parse.Parser, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:

		for i := 0; i < objValue.Type().NumField(); i++ {
			if objValue.Type().Field(i).Anonymous {
				if err := fillStructRecursive(objValue.Field(i), defaultPointerValMap, valMap, name); err != nil {
					return err
				}
			} else if len(objValue.Type().Field(i).Tag.Get("description")) > 0 {
				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}

				if len(key) == 0 {
					name = strings.ToLower(fieldName)
				} else {
					name = key + "." + strings.ToLower(fieldName)
				}

				if objValue.Field(i).Kind() != reflect.Ptr {
					if val, ok := valMap[name]; ok {
						if err := setFields(objValue.Field(i), val); err != nil {
							return err
						}
					}
				}

				if err := fillStructRecursive(objValue.Field(i), defaultPointerValMap, valMap, name); err != nil {
					return err
				}
			}
		}

	case reflect.Ptr:
		if len(key) == 0 && !objValue.IsNil() {
			return fillStructRecursive(objValue.Elem(), defaultPointerValMap, valMap, name)
		}

		contains := false
		for flg := range valMap {
			// TODO replace by regexp
			if strings.HasPrefix(flg, name+".") {
				contains = true
				break
			}
		}

		needDefault := false
		if _, ok := valMap[name]; ok {
			needDefault = valMap[name].Get().(bool)
		}
		if contains && objValue.IsNil() {
			needDefault = true
		}

		if needDefault {
			if defVal, ok := defaultPointerValMap[name]; ok {
				// set default pointer value
				objValue.Set(defVal)
			} else {
				return fmt.Errorf("flag %s default value not provided", name)
			}
		}

		if !objValue.IsNil() && contains {
			if objValue.Type().Elem().Kind() == reflect.Struct {
				if err := fillStructRecursive(objValue.Elem(), defaultPointerValMap, valMap, name); err != nil {
					return err
				}
			}
		}
	default:
		return nil
	}
	return nil
}

// SetFields sets value to fieldValue using tag as key in valMap
func setFields(fieldValue reflect.Value, val parse.Parser) error {
	if fieldValue.CanSet() {
		fieldValue.Set(reflect.ValueOf(val).Elem().Convert(fieldValue.Type()))
	} else {
		return fmt.Errorf("%s is not settable", fieldValue.Type().String())
	}
	return nil
}

// PrintHelp generates and prints command line help
func PrintHelp(flagMap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]parse.Parser) error {
	return PrintHelpWithCommand(flagMap, defaultValmap, parsers, nil, nil)
}

// PrintError takes a not nil error and prints command line help
func PrintError(err error, flagMap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]parse.Parser) error {
	if err != flag.ErrHelp {
		fmt.Printf("Error: %s\n", err)
	}
	if !strings.Contains(err.Error(), ":No parser for type") {
		PrintHelp(flagMap, defaultValmap, parsers)
	}
	return err
}

// LoadWithParsers initializes config : struct fields given by reference, with args : arguments.
// Some custom parsers may be given.
func LoadWithParsers(config interface{}, defaultValue interface{}, args []string, customParsers map[reflect.Type]parse.Parser) error {
	cmd := &Command{
		Config:                config,
		DefaultPointersConfig: defaultValue,
	}
	_, cmd.Name = path.Split(os.Args[0])
	return LoadWithCommand(cmd, args, customParsers, nil)
}

// Load initializes config : struct fields given by reference, with args : arguments.
// Some custom parsers may be given.
func Load(config interface{}, defaultValue interface{}, args []string) error {
	return LoadWithParsers(config, defaultValue, args, nil)
}

// Command structure contains program/command information (command name and description)
// Config must be a pointer on the configuration struct to parse (it contains default values of field)
// DefaultPointersConfig contains default pointers values: those values are set on pointers fields if their flags are called
// It must be the same type(struct) as Config
// Run is the func which launch the program using initialized configuration structure
type Command struct {
	Name                  string
	Description           string
	Config                interface{}
	DefaultPointersConfig interface{} // TODO: case DefaultPointersConfig is nil
	Run                   func() error
	Metadata              map[string]string
	HideHelp              bool
}

// LoadWithCommand initializes config : struct fields given by reference, with args : arguments.
// Some custom parsers and some subCommand may be given.
func LoadWithCommand(cmd *Command, cmdArgs []string, customParsers map[reflect.Type]parse.Parser, subCommand []*Command) error {
	parsers, err := parse.LoadParsers(customParsers)
	if err != nil {
		return err
	}

	tagsMap := make(map[string]reflect.StructField)
	if err := getTypesRecursive(reflect.ValueOf(cmd.Config), tagsMap, ""); err != nil {
		return err
	}
	defaultValMap := make(map[string]reflect.Value)
	if err := getDefaultValue(reflect.ValueOf(cmd.Config), reflect.ValueOf(cmd.DefaultPointersConfig), defaultValMap, ""); err != nil {
		return err
	}

	valMap, errParseArgs := parseArgs(cmdArgs, tagsMap, parsers)
	if errParseArgs != nil && errParseArgs != ErrParserNotFound {
		return PrintErrorWithCommand(errParseArgs, tagsMap, defaultValMap, parsers, cmd, subCommand)
	}

	if err := fillStructRecursive(reflect.ValueOf(cmd.Config), defaultValMap, valMap, ""); err != nil {
		return err
	}

	if errParseArgs == ErrParserNotFound {
		return errParseArgs
	}

	return nil
}

// PrintHelpWithCommand generates and prints command line help for a Command
func PrintHelpWithCommand(flagMap map[string]reflect.StructField, defaultValMap map[string]reflect.Value, parsers map[reflect.Type]parse.Parser, cmd *Command, subCmd []*Command) error {
	// Hide command from help
	if cmd != nil && cmd.HideHelp {
		return fmt.Errorf("command %s not found", cmd.Name)
	}

	// Define a templates
	// Using POSXE STD : http://pubs.opengroup.org/onlinepubs/9699919799/
	const helper = `{{if .ProgDescription}}{{.ProgDescription}}

{{end}}Usage: {{.ProgName}} [flags] <command> [<arguments>]

Use "{{.ProgName}} <command> --help" for help on any command.
{{if .SubCommands}}
Commands:{{range $subCmdName, $subCmdDesc := .SubCommands}}
{{printf "\t%-50s %s" $subCmdName $subCmdDesc}}{{end}}
{{end}}
Flag's usage: {{.ProgName}} [--flag=flag_argument] [-f[flag_argument]] ...     set flag_argument to flag(s)
          or: {{.ProgName}} [--flag[=true|false| ]] [-f[true|false| ]] ...     set true/false to boolean flag(s)

Flags:
`
	// Use a struct to give data to template
	type TempStruct struct {
		ProgName        string
		ProgDescription string
		SubCommands     map[string]string
	}
	tempStruct := TempStruct{}
	if cmd != nil {
		tempStruct.ProgName = cmd.Name
		tempStruct.ProgDescription = cmd.Description
		tempStruct.SubCommands = map[string]string{}
		if len(subCmd) > 1 && cmd == subCmd[0] {
			for _, c := range subCmd[1:] {
				if !c.HideHelp {
					tempStruct.SubCommands[c.Name] = c.Description
				}
			}
		}
	} else {
		_, tempStruct.ProgName = path.Split(os.Args[0])
	}

	// Run Template
	tmplHelper, err := template.New("helper").Parse(helper)
	if err != nil {
		return err
	}
	err = tmplHelper.Execute(os.Stdout, tempStruct)
	if err != nil {
		return err
	}

	return printFlagsDescriptionsDefaultValues(flagMap, defaultValMap, parsers, os.Stdout)
}

func printFlagsDescriptionsDefaultValues(flagMap map[string]reflect.StructField, defaultValMap map[string]reflect.Value, parsers map[reflect.Type]parse.Parser, output io.Writer) error {
	// Sort alphabetically & Delete unparsable flags in a slice
	var flags []string
	for flg, field := range flagMap {
		if _, ok := parsers[field.Type]; ok {
			flags = append(flags, flg)
		}
	}
	sort.Strings(flags)

	// Process data
	var descriptions []string
	var defaultValues []string
	var flagsWithDash []string
	var shortFlagsWithDash []string
	for _, flg := range flags {
		field := flagMap[flg]
		if short := field.Tag.Get("short"); len(short) == 1 {
			shortFlagsWithDash = append(shortFlagsWithDash, "-"+short+",")
		} else {
			shortFlagsWithDash = append(shortFlagsWithDash, "")
		}
		flagsWithDash = append(flagsWithDash, "--"+flg)

		// flag on pointer ?
		if defVal, ok := defaultValMap[flg]; ok {
			if defVal.Kind() != reflect.Ptr {
				// Set defaultValue on parsers
				parsers[field.Type].SetValue(defaultValMap[flg].Interface())
			}

			if defVal := parsers[field.Type].String(); len(defVal) > 0 {
				defaultValues = append(defaultValues, fmt.Sprintf("(default \"%s\")", defVal))
			} else {
				defaultValues = append(defaultValues, "")
			}
		}

		splittedDescriptions := split(field.Tag.Get("description"), 80)
		for i, description := range splittedDescriptions {
			descriptions = append(descriptions, description)
			if i != 0 {
				defaultValues = append(defaultValues, "")
				flagsWithDash = append(flagsWithDash, "")
				shortFlagsWithDash = append(shortFlagsWithDash, "")
			}
		}
	}

	// add help flag
	shortFlagsWithDash = append(shortFlagsWithDash, "-h,")
	flagsWithDash = append(flagsWithDash, "--help")
	descriptions = append(descriptions, "Print Help (this message) and exit")
	defaultValues = append(defaultValues, "")

	return displayTab(output, shortFlagsWithDash, flagsWithDash, descriptions, defaultValues)
}

func split(str string, width int) []string {
	if len(str) > width {
		index := strings.LastIndex(str[:width], " ")
		if index == -1 {
			index = width
		}

		return append([]string{strings.TrimSpace(str[:index])}, split(strings.TrimSpace(str[index:]), width)...)
	}
	return []string{str}
}

func displayTab(output io.Writer, columns ...[]string) error {
	w := new(tabwriter.Writer)
	w.Init(output, 0, 4, 1, ' ', 0)

	nbRow := len(columns[0])
	nbCol := len(columns)

	for i := 0; i < nbRow; i++ {
		row := ""
		for j, col := range columns {
			row += col[i]
			if j != nbCol-1 {
				row += "\t"
			}
		}
		fmt.Fprintln(w, row)
	}

	return w.Flush()
}

// PrintErrorWithCommand takes a not nil error and prints command line help
func PrintErrorWithCommand(err error, flagMap map[string]reflect.StructField, defaultValMap map[string]reflect.Value, parsers map[reflect.Type]parse.Parser, cmd *Command, subCmd []*Command) error {
	if err != flag.ErrHelp {
		fmt.Printf("Error here : %s\n", err)
	}

	if errHelp := PrintHelpWithCommand(flagMap, defaultValMap, parsers, cmd, subCmd); errHelp != nil {
		return errHelp
	}

	return err
}

// Flaeg struct contains commands (at least the root one)
// and row arguments (command and/or flags)
// a map of custom parsers could be use
type Flaeg struct {
	calledCommand *Command
	commands      []*Command // rootCommand is th fist one in this slice
	args          []string
	commandArgs   []string
	customParsers map[reflect.Type]parse.Parser
}

// New creates and initialize a pointer on Flaeg
func New(rootCommand *Command, args []string) *Flaeg {
	var f Flaeg
	f.commands = []*Command{rootCommand}
	f.args = args
	f.customParsers = map[reflect.Type]parse.Parser{}
	return &f
}

// AddCommand adds sub-command to the root command
func (f *Flaeg) AddCommand(command *Command) {
	f.commands = append(f.commands, command)
}

// AddParser adds custom parser for a type to the map of custom parsers
func (f *Flaeg) AddParser(typ reflect.Type, parser parse.Parser) {
	f.customParsers[typ] = parser
}

// Run calls the command with flags given as arguments
func (f *Flaeg) Run() error {
	if f.calledCommand == nil {
		if _, _, err := f.findCommandWithCommandArgs(); err != nil {
			return err
		}
	}

	if _, err := f.Parse(f.calledCommand); err != nil {
		return err
	}
	return f.calledCommand.Run()
}

// Parse calls Flaeg Load Function end returns the parsed command structure (by reference)
// It returns nil and a not nil error if it fails
func (f *Flaeg) Parse(cmd *Command) (*Command, error) {
	if f.calledCommand == nil {
		f.commandArgs = f.args
	}

	if err := LoadWithCommand(cmd, f.commandArgs, f.customParsers, f.commands); err != nil {
		return cmd, err
	}
	return cmd, nil
}

// splitArgs takes args (type []string) and return command ("" if rootCommand) and command's args
func splitArgs(args []string) (string, []string) {
	if len(args) >= 1 && len(args[0]) >= 1 && string(args[0][0]) != "-" {
		if len(args) == 1 {
			return strings.ToLower(args[0]), []string{}
		}
		return strings.ToLower(args[0]), args[1:]
	}
	return "", args
}

// findCommandWithCommandArgs returns the called command (by reference) and command's args
// the error returned is not nil if it fails
func (f *Flaeg) findCommandWithCommandArgs() (*Command, []string, error) {
	var commandName string
	commandName, f.commandArgs = splitArgs(f.args)
	if len(commandName) > 0 {
		for _, command := range f.commands {
			if commandName == command.Name {
				f.calledCommand = command
				return f.calledCommand, f.commandArgs, nil
			}
		}
		return nil, []string{}, fmt.Errorf("command %s not found", commandName)
	}

	f.calledCommand = f.commands[0]
	return f.calledCommand, f.commandArgs, nil
}

// GetCommand splits args and returns the called command (by reference)
// It returns nil and a not nil error if it fails
func (f *Flaeg) GetCommand() (*Command, error) {
	if f.calledCommand == nil {
		_, _, err := f.findCommandWithCommandArgs()
		return f.calledCommand, err
	}
	return f.calledCommand, nil
}

// isExported return true is the field (from fieldName) is exported,
// else false
func isExported(fieldName string) bool {
	if len(fieldName) < 1 {
		return false
	}

	if string(fieldName[0]) == strings.ToUpper(string(fieldName[0])) {
		return true
	}

	return false
}

func argToLower(inArg string) string {
	if len(inArg) < 2 {
		return strings.ToLower(inArg)
	}

	var outArg string
	dashIndex := strings.Index(inArg, "--")
	if dashIndex == -1 {
		if dashIndex = strings.Index(inArg, "-"); dashIndex == -1 {
			return inArg
		}
		// -fValue
		outArg = strings.ToLower(inArg[dashIndex:dashIndex+2]) + inArg[dashIndex+2:]
		return outArg
	}

	// --flag
	if equalIndex := strings.Index(inArg, "="); equalIndex != -1 {
		// --flag=value
		outArg = strings.ToLower(inArg[dashIndex:equalIndex]) + inArg[equalIndex:]
	} else {
		// --boolflag
		outArg = strings.ToLower(inArg[dashIndex:])
	}

	return outArg
}

func argsToLower(inArgs []string) []string {
	outArgs := make([]string, len(inArgs))
	for i, inArg := range inArgs {
		outArgs[i] = argToLower(inArg)
	}
	return outArgs
}
