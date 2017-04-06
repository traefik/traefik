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
	"time"

	flag "github.com/ogier/pflag"
)

// ErrParserNotFound is thrown when a field is flaged but not parser match its type
var ErrParserNotFound = errors.New("Parser not found or custom parser missing")

// GetTypesRecursive links in flagmap a flag with its reflect.StructField
// You can whether provide objValue on a structure or a pointer to structure as first argument
// Flags are genereted from field name or from StructTag
func getTypesRecursive(objValue reflect.Value, flagmap map[string]reflect.StructField, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:

		for i := 0; i < objValue.NumField(); i++ {
			if objValue.Type().Field(i).Anonymous {
				if err := getTypesRecursive(objValue.Field(i), flagmap, name); err != nil {
					return err
				}
			} else if len(objValue.Type().Field(i).Tag.Get("description")) > 0 {
				fieldName := objValue.Type().Field(i).Name
				if !isExported(fieldName) {
					return fmt.Errorf("Field %s is an unexported field", fieldName)
				}

				name += objValue.Type().Name()
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if len(key) == 0 {
					//Lower Camel Case
					//name = strings.ToLower(string(fieldName[0])) + fieldName[1:]
					name = strings.ToLower(fieldName)
				} else {
					name = key + "." + strings.ToLower(fieldName)
				}
				if _, ok := flagmap[name]; ok {
					return errors.New("Tag already exists: " + name)
				}
				flagmap[name] = objValue.Type().Field(i)

				if err := getTypesRecursive(objValue.Field(i), flagmap, name); err != nil {
					return err
				}
			}

		}
	case reflect.Ptr:
		if len(key) > 0 {
			field := flagmap[name]
			field.Type = reflect.TypeOf(false)
			flagmap[name] = field
		}
		typ := objValue.Type().Elem()
		inst := reflect.New(typ).Elem()
		if err := getTypesRecursive(inst, flagmap, name); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

//GetPointerFlags returns flags on pointers
func GetBoolFlags(config interface{}) ([]string, error) {
	flagmap := make(map[string]reflect.StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		return []string{}, err
	}
	flags := make([]string, 0, len(flagmap))
	for f, structField := range flagmap {
		if structField.Type.Kind() == reflect.Bool {
			flags = append(flags, f)
		}
	}
	return flags, nil
}

//GetFlags returns flags
func GetFlags(config interface{}) ([]string, error) {
	flagmap := make(map[string]reflect.StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		return []string{}, err
	}
	flags := make([]string, 0, len(flagmap))
	for f := range flagmap {
		flags = append(flags, f)
	}
	return flags, nil
}

//loadParsers loads default parsers and custom parsers given as parameter. Return a map [reflect.Type]parsers
// bool, int, int64, uint, uint64, float64,
func loadParsers(customParsers map[reflect.Type]Parser) (map[reflect.Type]Parser, error) {
	parsers := map[reflect.Type]Parser{}

	var boolParser boolValue
	parsers[reflect.TypeOf(true)] = &boolParser

	var intParser intValue
	parsers[reflect.TypeOf(1)] = &intParser

	var int64Parser int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser

	var uintParser uintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser

	var uint64Parser uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser

	var stringParser stringValue
	parsers[reflect.TypeOf("")] = &stringParser

	var float64Parser float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser

	var durationParser Duration
	parsers[reflect.TypeOf(Duration(time.Second))] = &durationParser

	var timeParser timeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser

	for rType, parser := range customParsers {
		parsers[rType] = parser
	}
	return parsers, nil
}

//ParseArgs : parses args return valmap map[flag]Getter, using parsers map[type]Getter
//args must be formated as like as flag documentation. See https://golang.org/pkg/flag
func parseArgs(args []string, flagmap map[string]reflect.StructField, parsers map[reflect.Type]Parser) (map[string]Parser, error) {
	//Return var
	valmap := make(map[string]Parser)
	//Visitor in flag.Parse
	flagList := []*flag.Flag{}
	visitor := func(fl *flag.Flag) {
		flagList = append(flagList, fl)
	}
	newParsers := map[string]Parser{}
	flagSet := flag.NewFlagSet("flaeg.Load", flag.ContinueOnError)
	//Disable output
	flagSet.SetOutput(ioutil.Discard)
	var err error
	for flag, structField := range flagmap {
		//for _, flag := range flags {
		//structField := flagmap[flag]
		if parser, ok := parsers[structField.Type]; ok {
			newparserValue := reflect.New(reflect.TypeOf(parser).Elem())
			newparserValue.Elem().Set(reflect.ValueOf(parser).Elem())
			newparser := newparserValue.Interface().(Parser)
			if short := structField.Tag.Get("short"); len(short) == 1 {
				// fmt.Printf("short : %s long : %s\n", short, flag)
				flagSet.VarP(newparser, flag, short, structField.Tag.Get("description"))
			} else {
				flagSet.Var(newparser, flag, structField.Tag.Get("description"))
			}
			newParsers[flag] = newparser
		} else {
			err = ErrParserNotFound
		}
	}

	// prevents case sensitivity issue
	args = argsToLower(args)
	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	//Fill flagList with parsed flags
	flagSet.Visit(visitor)
	//Return parsers on parsed flag
	for _, flag := range flagList {
		valmap[flag.Name] = newParsers[flag.Name]
	}

	return valmap, err
}

func getDefaultValue(defaultValue reflect.Value, defaultPointersValue reflect.Value, defaultValmap map[string]reflect.Value, key string) error {
	if defaultValue.Type() != defaultPointersValue.Type() {
		return fmt.Errorf("Parameters defaultValue and defaultPointersValue must be the same struct. defaultValue type : %s is not defaultPointersValue type : %s", defaultValue.Type().String(), defaultPointersValue.Type().String())
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
				name += defaultValue.Type().Name()
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
					// if _, ok := defaultValmap[name]; ok {
					// 	return errors.New("Tag already exists: " + name)
					// }
					defaultValmap[name] = defaultValue.Field(i)
					// fmt.Printf("%s: got default value %+v\n", name, defaultValue.Field(i))
				}
				if err := getDefaultValue(defaultValue.Field(i), defaultPointersValue.Field(i), defaultValmap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Ptr:
		if !defaultPointersValue.IsNil() {
			if len(key) != 0 {
				//turn ptr fields to nil
				defaultPointersNilValue, err := setPointersNil(defaultPointersValue)
				if err != nil {
					return err
				}
				defaultValmap[name] = defaultPointersNilValue
				// fmt.Printf("%s: got default value %+v\n", name, defaultPointersNilValue)
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
				// fmt.Printf("%s: got default value %+v\n", name, instValue)
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

//objValue a reflect.Value of a not-nil pointer on a struct
func setPointersNil(objValue reflect.Value) (reflect.Value, error) {
	if objValue.Kind() != reflect.Ptr {
		return objValue, fmt.Errorf("Parameters objValue must be a not-nil pointer on a struct, not a %s", objValue.Kind().String())
	} else if objValue.IsNil() {
		return objValue, fmt.Errorf("Parameters objValue must be a not-nil pointer")
	} else if objValue.Elem().Kind() != reflect.Struct {
		// fmt.Printf("Parameters objValue must be a not-nil pointer on a struct, not a pointer on a %s\n", objValue.Elem().Kind().String())
		return objValue, nil
	}
	//Clone
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

//FillStructRecursive initialize a value of any taged Struct given by reference
func fillStructRecursive(objValue reflect.Value, defaultPointerValmap map[string]reflect.Value, valmap map[string]Parser, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:

		for i := 0; i < objValue.Type().NumField(); i++ {
			if objValue.Type().Field(i).Anonymous {
				if err := fillStructRecursive(objValue.Field(i), defaultPointerValmap, valmap, name); err != nil {
					return err
				}
			} else if len(objValue.Type().Field(i).Tag.Get("description")) > 0 {
				name += objValue.Type().Name()
				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if len(key) == 0 {
					name = strings.ToLower(fieldName)
				} else {
					name = key + "." + strings.ToLower(fieldName)
				}
				// fmt.Println(name)
				if objValue.Field(i).Kind() != reflect.Ptr {

					if val, ok := valmap[name]; ok {
						// fmt.Printf("%s : set def val\n", name)
						if err := setFields(objValue.Field(i), val); err != nil {
							return err
						}
					}
				}
				if err := fillStructRecursive(objValue.Field(i), defaultPointerValmap, valmap, name); err != nil {
					return err
				}
			}
		}

	case reflect.Ptr:
		if len(key) == 0 && !objValue.IsNil() {
			if err := fillStructRecursive(objValue.Elem(), defaultPointerValmap, valmap, name); err != nil {
				return err
			}
			return nil
		}
		contains := false
		for flag := range valmap {
			// TODO replace by regexp
			if strings.Contains(flag, name+".") {
				contains = true
				break
			}
		}
		needDefault := false
		if _, ok := valmap[name]; ok {
			needDefault = valmap[name].Get().(bool)
		}
		if contains && objValue.IsNil() {
			needDefault = true
		}

		if needDefault {
			if defVal, ok := defaultPointerValmap[name]; ok {
				//set default pointer value
				// fmt.Printf("%s  : set default value %+v\n", name, defVal)
				objValue.Set(defVal)
			} else {
				return fmt.Errorf("flag %s default value not provided", name)
			}
		}
		if !objValue.IsNil() && contains {
			if objValue.Type().Elem().Kind() == reflect.Struct {
				if err := fillStructRecursive(objValue.Elem(), defaultPointerValmap, valmap, name); err != nil {
					return err
				}
			}
		}
	default:
		return nil
	}
	return nil
}

// SetFields sets value to fieldValue using tag as key in valmap
func setFields(fieldValue reflect.Value, val Parser) error {
	if fieldValue.CanSet() {
		fieldValue.Set(reflect.ValueOf(val).Elem().Convert(fieldValue.Type()))
	} else {
		return errors.New(fieldValue.Type().String() + " is not settable.")
	}
	return nil
}

//PrintHelp generates and prints command line help
func PrintHelp(flagmap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser) error {
	return PrintHelpWithCommand(flagmap, defaultValmap, parsers, nil, nil)
}

//PrintError takes a not nil error and prints command line help
func PrintError(err error, flagmap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser) error {
	if err != flag.ErrHelp {
		fmt.Printf("Error : %s\n", err)
	}
	if !strings.Contains(err.Error(), ":No parser for type") {
		PrintHelp(flagmap, defaultValmap, parsers)
	}
	return err
}

//LoadWithParsers initializes config : struct fields given by reference, with args : arguments.
//Some custom parsers may be given.
func LoadWithParsers(config interface{}, defaultValue interface{}, args []string, customParsers map[reflect.Type]Parser) error {
	cmd := &Command{
		Config:                config,
		DefaultPointersConfig: defaultValue,
	}
	_, cmd.Name = path.Split(os.Args[0])
	return LoadWithCommand(cmd, args, customParsers, nil)
}

//Load initializes config : struct fields given by reference, with args : arguments.
//Some custom parsers may be given.
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
	DefaultPointersConfig interface{} //TODO:case DefaultPointersConfig is nil
	Run                   func() error
	Metadata              map[string]string
}

//LoadWithCommand initializes config : struct fields given by reference, with args : arguments.
//Some custom parsers and some subCommand may be given.
func LoadWithCommand(cmd *Command, cmdArgs []string, customParsers map[reflect.Type]Parser, subCommand []*Command) error {

	parsers, err := loadParsers(customParsers)
	if err != nil {
		return err
	}

	tagsmap := make(map[string]reflect.StructField)
	if err := getTypesRecursive(reflect.ValueOf(cmd.Config), tagsmap, ""); err != nil {
		return err
	}
	defaultValmap := make(map[string]reflect.Value)
	if err := getDefaultValue(reflect.ValueOf(cmd.Config), reflect.ValueOf(cmd.DefaultPointersConfig), defaultValmap, ""); err != nil {
		return err
	}

	valmap, errParseArgs := parseArgs(cmdArgs, tagsmap, parsers)
	if errParseArgs != nil && errParseArgs != ErrParserNotFound {
		return PrintErrorWithCommand(errParseArgs, tagsmap, defaultValmap, parsers, cmd, subCommand)
	}

	if err := fillStructRecursive(reflect.ValueOf(cmd.Config), defaultValmap, valmap, ""); err != nil {
		return err
	}

	if errParseArgs == ErrParserNotFound {
		return errParseArgs
	}

	return nil
}

//PrintHelpWithCommand generates and prints command line help for a Command
func PrintHelpWithCommand(flagmap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser, cmd *Command, subCmd []*Command) error {
	// Define a templates
	// Using POSXE STD : http://pubs.opengroup.org/onlinepubs/9699919799/
	const helper = `{{if .ProgDescription}}{{.ProgDescription}}

{{end}}Usage: {{.ProgName}} [--flag=flag_argument] [-f[flag_argument]] ...     set flag_argument to flag(s)
   or: {{.ProgName}} [--flag[=true|false| ]] [-f[true|false| ]] ...     set true/false to boolean flag(s)
{{if .SubCommands}}
Available Commands:{{range $subCmdName, $subCmdDesc := .SubCommands}}
{{printf "\t%-50s %s" $subCmdName $subCmdDesc}}{{end}}
Use "{{.ProgName}} [command] --help" for more information about a command.
{{end}}
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
				tempStruct.SubCommands[c.Name] = c.Description
			}
		}
	} else {
		_, tempStruct.ProgName = path.Split(os.Args[0])
	}

	//Run Template
	tmplHelper, err := template.New("helper").Parse(helper)
	if err != nil {
		return err
	}
	err = tmplHelper.Execute(os.Stdout, tempStruct)
	if err != nil {
		return err
	}

	return printFlagsDescriptionsDefaultValues(flagmap, defaultValmap, parsers, os.Stdout)
}

func printFlagsDescriptionsDefaultValues(flagmap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser, output io.Writer) error {
	// Sort alphabetically & Delete unparsable flags in a slice
	flags := []string{}
	for flag, field := range flagmap {
		if _, ok := parsers[field.Type]; ok {
			flags = append(flags, flag)
		}
	}
	sort.Strings(flags)

	// Process data
	descriptions := []string{}
	defaultValues := []string{}
	flagsWithDashs := []string{}
	shortFlagsWithDash := []string{}
	for _, flag := range flags {
		field := flagmap[flag]
		if short := field.Tag.Get("short"); len(short) == 1 {
			shortFlagsWithDash = append(shortFlagsWithDash, "-"+short+",")
		} else {
			shortFlagsWithDash = append(shortFlagsWithDash, "")
		}
		flagsWithDashs = append(flagsWithDashs, "--"+flag)

		//flag on pointer ?
		if defVal, ok := defaultValmap[flag]; ok {
			if defVal.Kind() != reflect.Ptr {
				// Set defaultValue on parsers
				parsers[field.Type].SetValue(defaultValmap[flag].Interface())
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
				flagsWithDashs = append(flagsWithDashs, "")
				shortFlagsWithDash = append(shortFlagsWithDash, "")
			}
		}
	}
	//add help flag
	shortFlagsWithDash = append(shortFlagsWithDash, "-h,")
	flagsWithDashs = append(flagsWithDashs, "--help")
	descriptions = append(descriptions, "Print Help (this message) and exit")
	defaultValues = append(defaultValues, "")
	return displayTab(output, shortFlagsWithDash, flagsWithDashs, descriptions, defaultValues)
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
	nbRow := len(columns[0])
	nbCol := len(columns)
	w := new(tabwriter.Writer)
	w.Init(output, 0, 4, 1, ' ', 0)
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
	w.Flush()
	return nil
}

//PrintErrorWithCommand takes a not nil error and prints command line help
func PrintErrorWithCommand(err error, flagmap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser, cmd *Command, subCmd []*Command) error {
	if err != flag.ErrHelp {
		fmt.Printf("Error here : %s\n", err)
	}
	PrintHelpWithCommand(flagmap, defaultValmap, parsers, cmd, subCmd)
	return err
}

//Flaeg struct contains commands (at least the root one)
//and row arguments (command and/or flags)
//a map of custom parsers could be use
type Flaeg struct {
	calledCommand *Command
	commands      []*Command ///rootCommand is th fist one in this slice
	args          []string
	commmandArgs  []string
	customParsers map[reflect.Type]Parser
}

//New creats and initialize a pointer on Flaeg
func New(rootCommand *Command, args []string) *Flaeg {
	var f Flaeg
	f.commands = []*Command{rootCommand}
	f.args = args
	f.customParsers = map[reflect.Type]Parser{}
	return &f
}

//AddCommand adds sub-command to the root command
func (f *Flaeg) AddCommand(command *Command) {
	f.commands = append(f.commands, command)
}

//AddParser adds custom parser for a type to the map of custom parsers
func (f *Flaeg) AddParser(typ reflect.Type, parser Parser) {
	f.customParsers[typ] = parser
}

// Run calls the command with flags given as agruments
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
		f.commmandArgs = f.args
	}
	if err := LoadWithCommand(cmd, f.commmandArgs, f.customParsers, f.commands); err != nil {
		return cmd, err
	}
	return cmd, nil
}

//splitArgs takes args (type []string) and return command ("" if rootCommand) and command's args
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
	commandName := ""
	commandName, f.commmandArgs = splitArgs(f.args)
	if len(commandName) > 0 {
		for _, command := range f.commands {
			if commandName == command.Name {
				f.calledCommand = command
				return f.calledCommand, f.commmandArgs, nil
			}
		}
		return nil, []string{}, fmt.Errorf("Command %s not found", commandName)
	}

	f.calledCommand = f.commands[0]
	return f.calledCommand, f.commmandArgs, nil
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

//isExported return true is the field (from fieldName) is exported,
//else false
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
		//-fValue
		outArg = strings.ToLower(inArg[dashIndex:dashIndex+2]) + inArg[dashIndex+2:]
		return outArg
	}
	//--flag
	if equalIndex := strings.Index(inArg, "="); equalIndex != -1 {
		//--flag=value
		outArg = strings.ToLower(inArg[dashIndex:equalIndex]) + inArg[equalIndex:]
	} else {
		//--boolflag
		outArg = strings.ToLower(inArg[dashIndex:])
	}

	return outArg
}

func argsToLower(inArgs []string) []string {
	outArgs := make([]string, len(inArgs), len(inArgs))
	for i, inArg := range inArgs {
		outArgs[i] = argToLower(inArg)
	}
	return outArgs
}
