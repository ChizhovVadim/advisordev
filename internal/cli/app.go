package cli

import (
	"fmt"
	"os"
)

type Command struct {
	name    string
	handler func([]string) error
}

type App struct {
	commands []Command
}

func (app *App) AddCommand(name string, handler func([]string) error) {
	app.commands = append(app.commands, Command{
		name:    name,
		handler: handler,
	})
}

func (app *App) Run() error {
	var args = os.Args[1:]
	if len(args) == 0 {
		return fmt.Errorf("command not specified")
	}
	var cmdName = args[0]
	args = args[1:]
	for i := range app.commands {
		var cmd = &app.commands[i]
		if cmd.name == cmdName {
			return cmd.handler(args)
		}
	}
	return fmt.Errorf("bad command %v", cmdName)
}
