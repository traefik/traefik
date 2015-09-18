package kingpin

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrCommandNotSpecified = fmt.Errorf("command not specified")
)

// Action callback executed at various stages after all values are populated.
// The application, commands, arguments and flags all have corresponding
// actions.
type Action func(*ParseContext) error

type ApplicationValidator func(*Application) error

// An Application contains the definitions of flags, arguments and commands
// for an application.
type Application struct {
	*flagGroup
	*argGroup
	*cmdGroup
	initialized    bool
	Name           string
	Help           string
	author         string
	version        string
	writer         io.Writer // Destination for usage and errors.
	usageTemplate  string
	action         Action
	preAction      Action
	validator      ApplicationValidator
	terminate      func(status int) // See Terminate()
	noInterspersed bool             // can flags be interspersed with args (or must they come first)
}

// New creates a new Kingpin application instance.
func New(name, help string) *Application {
	a := &Application{
		flagGroup:     newFlagGroup(),
		argGroup:      newArgGroup(),
		Name:          name,
		Help:          help,
		writer:        os.Stderr,
		usageTemplate: DefaultUsageTemplate,
		terminate:     os.Exit,
	}
	a.cmdGroup = newCmdGroup(a)
	a.Flag("help", "Show help (also see --help-long and --help-man).").Bool()
	a.Flag("help-long", "Generate long help.").Hidden().PreAction(a.generateLongHelp).Bool()
	a.Flag("help-man", "Generate a man page.").Hidden().PreAction(a.generateManPage).Bool()
	return a
}

func (a *Application) generateLongHelp(c *ParseContext) error {
	a.Writer(os.Stdout)
	if err := a.UsageForContextWithTemplate(c, 2, LongHelpTemplate); err != nil {
		return err
	}
	a.terminate(0)
	return nil
}

func (a *Application) generateManPage(c *ParseContext) error {
	a.Writer(os.Stdout)
	if err := a.UsageForContextWithTemplate(c, 2, ManPageTemplate); err != nil {
		return err
	}
	a.terminate(0)
	return nil
}

// Terminate specifies the termination handler. Defaults to os.Exit(status).
// If nil is passed, a no-op function will be used.
func (a *Application) Terminate(terminate func(int)) *Application {
	if terminate == nil {
		terminate = func(int) {}
	}
	a.terminate = terminate
	return a
}

// Specify the writer to use for usage and errors. Defaults to os.Stderr.
func (a *Application) Writer(w io.Writer) *Application {
	a.writer = w
	return a
}

// UsageTemplate specifies the text template to use when displaying usage
// information. The default is UsageTemplate.
func (a *Application) UsageTemplate(template string) *Application {
	a.usageTemplate = template
	return a
}

// Validate sets a validation function to run when parsing.
func (a *Application) Validate(validator ApplicationValidator) *Application {
	a.validator = validator
	return a
}

// ParseContext parses the given command line and returns the fully populated
// ParseContext.
func (a *Application) ParseContext(args []string) (*ParseContext, error) {
	if err := a.init(); err != nil {
		return nil, err
	}
	context := tokenize(args)
	_, err := parse(context, a)
	return context, err
}

// Parse parses command-line arguments. It returns the selected command and an
// error. The selected command will be a space separated subcommand, if
// subcommands have been configured.
//
// This will populate all flag and argument values, call all callbacks, and so
// on.
func (a *Application) Parse(args []string) (command string, err error) {
	context, err := a.ParseContext(args)
	if err != nil {
		if a.hasHelp(args) {
			a.writeUsage(context, err)
		}
		return "", err
	}
	a.maybeHelp(context)
	if !context.EOL() {
		return "", fmt.Errorf("unexpected argument '%s'", context.Peek())
	}
	command, err = a.execute(context)
	if err == ErrCommandNotSpecified {
		a.writeUsage(context, nil)
	}
	return command, err
}

func (a *Application) writeUsage(context *ParseContext, err error) {
	if err != nil {
		a.Errorf("%s", err)
	}
	if err := a.UsageForContext(context); err != nil {
		panic(err)
	}
	a.terminate(1)
}

func (a *Application) hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" {
			return true
		}
	}
	return false
}

func (a *Application) maybeHelp(context *ParseContext) {
	for _, element := range context.Elements {
		if flag, ok := element.Clause.(*FlagClause); ok && flag.name == "help" {
			a.writeUsage(context, nil)
		}
	}
}

// findCommandFromArgs finds a command (if any) from the given command line arguments.
func (a *Application) findCommandFromArgs(args []string) (command string, err error) {
	if err := a.init(); err != nil {
		return "", err
	}
	context := tokenize(args)
	if err := a.parse(context); err != nil {
		return "", err
	}
	return a.findCommandFromContext(context), nil
}

// findCommandFromContext finds a command (if any) from a parsed context.
func (a *Application) findCommandFromContext(context *ParseContext) string {
	commands := []string{}
	for _, element := range context.Elements {
		if c, ok := element.Clause.(*CmdClause); ok {
			commands = append(commands, c.name)
		}
	}
	return strings.Join(commands, " ")
}

// Version adds a --version flag for displaying the application version.
func (a *Application) Version(version string) *Application {
	a.version = version
	a.Flag("version", "Show application version.").PreAction(func(*ParseContext) error {
		fmt.Fprintln(a.writer, version)
		a.terminate(0)
		return nil
	}).Bool()
	return a
}

func (a *Application) Author(author string) *Application {
	a.author = author
	return a
}

// Action callback to call when all values are populated and parsing is
// complete, but before any command, flag or argument actions.
//
// All Action() callbacks are called in the order they are encountered on the
// command line.
func (a *Application) Action(action Action) *Application {
	a.action = action
	return a
}

// Action called after parsing completes but before validation and execution.
func (a *Application) PreAction(action Action) *Application {
	a.preAction = action
	return a
}

// Command adds a new top-level command.
func (a *Application) Command(name, help string) *CmdClause {
	return a.addCommand(name, help)
}

// Interspersed control if flags can be interspersed with positional arguments
//
// true (the default) means that they can, false means that all the flags must appear before the first positional arguments.
func (a *Application) Interspersed(interspersed bool) *Application {
	a.noInterspersed = !interspersed
	return a
}

func (a *Application) init() error {
	if a.initialized {
		return nil
	}
	if a.cmdGroup.have() && a.argGroup.have() {
		return fmt.Errorf("can't mix top-level Arg()s with Command()s")
	}

	// If we have subcommands, add a help command at the top-level.
	if a.cmdGroup.have() {
		var command []string
		help := a.Command("help", "Show help.").Action(func(c *ParseContext) error {
			a.Usage(command)
			a.terminate(0)
			return nil
		})
		help.Arg("command", "Show help on command.").StringsVar(&command)
		// Make help first command.
		l := len(a.commandOrder)
		a.commandOrder = append(a.commandOrder[l-1:l], a.commandOrder[:l-1]...)
	}

	if err := a.flagGroup.init(); err != nil {
		return err
	}
	if err := a.cmdGroup.init(); err != nil {
		return err
	}
	if err := a.argGroup.init(); err != nil {
		return err
	}
	for _, cmd := range a.commands {
		if err := cmd.init(); err != nil {
			return err
		}
	}
	flagGroups := []*flagGroup{a.flagGroup}
	for _, cmd := range a.commandOrder {
		if err := checkDuplicateFlags(cmd, flagGroups); err != nil {
			return err
		}
	}
	a.initialized = true
	return nil
}

// Recursively check commands for duplicate flags.
func checkDuplicateFlags(current *CmdClause, flagGroups []*flagGroup) error {
	// Check for duplicates.
	for _, flags := range flagGroups {
		for _, flag := range current.flagOrder {
			if flag.shorthand != 0 {
				if _, ok := flags.short[string(flag.shorthand)]; ok {
					return fmt.Errorf("duplicate short flag -%c", flag.shorthand)
				}
			}
			if _, ok := flags.long[flag.name]; ok {
				return fmt.Errorf("duplicate long flag --%s", flag.name)
			}
		}
	}
	flagGroups = append(flagGroups, current.flagGroup)
	// Check subcommands.
	for _, subcmd := range current.commandOrder {
		if err := checkDuplicateFlags(subcmd, flagGroups); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) execute(context *ParseContext) (string, error) {
	var err error
	selected := []string{}

	if err = a.setDefaults(context); err != nil {
		return "", err
	}

	selected, err = a.setValues(context)
	if err != nil {
		return "", err
	}

	if err = a.applyPreActions(context); err != nil {
		return "", err
	}

	if err = a.validateRequired(context); err != nil {
		return "", err
	}

	if err = a.applyValidators(context); err != nil {
		return "", err
	}

	if err = a.applyActions(context); err != nil {
		return "", err
	}

	command := strings.Join(selected, " ")
	if command == "" && a.cmdGroup.have() {
		return "", ErrCommandNotSpecified
	}
	return command, err
}

func (a *Application) setDefaults(context *ParseContext) error {
	flagElements := map[string]*ParseElement{}
	for _, element := range context.Elements {
		if flag, ok := element.Clause.(*FlagClause); ok {
			flagElements[flag.name] = element
		}
	}

	argElements := map[string]*ParseElement{}
	for _, element := range context.Elements {
		if arg, ok := element.Clause.(*ArgClause); ok {
			argElements[arg.name] = element
		}
	}

	// Check required flags and set defaults.
	for _, flag := range context.flags.long {
		if flagElements[flag.name] == nil {
			// Set defaults, if any.
			if flag.defaultValue != "" {
				if err := flag.value.Set(flag.defaultValue); err != nil {
					return err
				}
			}
		}
	}

	for _, arg := range context.arguments.args {
		if argElements[arg.name] == nil {
			// Set defaults, if any.
			if arg.defaultValue != "" {
				if err := arg.value.Set(arg.defaultValue); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (a *Application) validateRequired(context *ParseContext) error {
	flagElements := map[string]*ParseElement{}
	for _, element := range context.Elements {
		if flag, ok := element.Clause.(*FlagClause); ok {
			flagElements[flag.name] = element
		}
	}

	argElements := map[string]*ParseElement{}
	for _, element := range context.Elements {
		if arg, ok := element.Clause.(*ArgClause); ok {
			argElements[arg.name] = element
		}
	}

	// Check required flags and set defaults.
	for _, flag := range context.flags.long {
		if flagElements[flag.name] == nil {
			// Check required flags were provided.
			if flag.needsValue() {
				return fmt.Errorf("required flag --%s not provided", flag.name)
			}
		}
	}

	for _, arg := range context.arguments.args {
		if argElements[arg.name] == nil {
			if arg.required {
				return fmt.Errorf("required argument '%s' not provided", arg.name)
			}
		}
	}
	return nil
}

func (a *Application) setValues(context *ParseContext) (selected []string, err error) {
	// Set all arg and flag values.
	var lastCmd *CmdClause
	for _, element := range context.Elements {
		switch clause := element.Clause.(type) {
		case *FlagClause:
			if err = clause.value.Set(*element.Value); err != nil {
				return
			}

		case *ArgClause:
			if err = clause.value.Set(*element.Value); err != nil {
				return
			}

		case *CmdClause:
			if clause.validator != nil {
				if err = clause.validator(clause); err != nil {
					return
				}
			}
			selected = append(selected, clause.name)
			lastCmd = clause
		}
	}

	if lastCmd != nil && len(lastCmd.commands) > 0 {
		return nil, fmt.Errorf("must select a subcommand of '%s'", lastCmd.FullCommand())
	}

	return
}

func (a *Application) applyValidators(context *ParseContext) (err error) {
	// Call command validation functions.
	for _, element := range context.Elements {
		if cmd, ok := element.Clause.(*CmdClause); ok && cmd.validator != nil {
			if err = cmd.validator(cmd); err != nil {
				return err
			}
		}
	}

	if a.validator != nil {
		err = a.validator(a)
	}
	return err
}

func (a *Application) applyPreActions(context *ParseContext) error {
	if a.preAction != nil {
		if err := a.preAction(context); err != nil {
			return err
		}
	}
	// Dispatch to actions.
	for _, element := range context.Elements {
		switch clause := element.Clause.(type) {
		case *ArgClause:
			if clause.preAction != nil {
				if err := clause.preAction(context); err != nil {
					return err
				}
			}
		case *CmdClause:
			if clause.preAction != nil {
				if err := clause.preAction(context); err != nil {
					return err
				}
			}
		case *FlagClause:
			if clause.preAction != nil {
				if err := clause.preAction(context); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (a *Application) applyActions(context *ParseContext) error {
	if a.action != nil {
		if err := a.action(context); err != nil {
			return err
		}
	}
	// Dispatch to actions.
	for _, element := range context.Elements {
		switch clause := element.Clause.(type) {
		case *ArgClause:
			if clause.action != nil {
				if err := clause.action(context); err != nil {
					return err
				}
			}
		case *CmdClause:
			if clause.action != nil {
				if err := clause.action(context); err != nil {
					return err
				}
			}
		case *FlagClause:
			if clause.action != nil {
				if err := clause.action(context); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Errorf prints an error message to w in the format "<appname>: error: <message>".
func (a *Application) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(a.writer, a.Name+": error: "+format+"\n", args...)
}

// Fatalf writes a formatted error to w then terminates with exit status 1.
func (a *Application) Fatalf(format string, args ...interface{}) {
	a.Errorf(format, args...)
	a.terminate(1)
}

// FatalUsage prints an error message followed by usage information, then
// exits with a non-zero status.
func (a *Application) FatalUsage(format string, args ...interface{}) {
	a.Errorf(format, args...)
	a.Usage([]string{})
	a.terminate(1)
}

// FatalUsageContext writes a printf formatted error message to w, then usage
// information for the given ParseContext, before exiting.
func (a *Application) FatalUsageContext(context *ParseContext, format string, args ...interface{}) {
	a.Errorf(format, args...)
	if err := a.UsageForContext(context); err != nil {
		panic(err)
	}
	a.terminate(1)
}

// FatalIfError prints an error and exits if err is not nil. The error is printed
// with the given formatted string, if any.
func (a *Application) FatalIfError(err error, format string, args ...interface{}) {
	if err != nil {
		prefix := ""
		if format != "" {
			prefix = fmt.Sprintf(format, args...) + ": "
		}
		a.Errorf(prefix+"%s", err)
		a.terminate(1)
	}
}
