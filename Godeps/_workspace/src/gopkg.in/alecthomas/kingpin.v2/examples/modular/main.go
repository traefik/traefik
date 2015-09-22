package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
)

// Context for "ls" command
type LsCommand struct {
	All bool
}

func (l *LsCommand) run(c *kingpin.ParseContext) error {
	fmt.Printf("all=%v\n", l.All)
	return nil
}

func configureLsCommand(app *kingpin.Application) {
	c := &LsCommand{}
	ls := app.Command("ls", "List files.").Action(c.run)
	ls.Flag("all", "List all files.").Short('a').BoolVar(&c.All)
}

func main() {
	app := kingpin.New("modular", "My modular application.")
	configureLsCommand(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
