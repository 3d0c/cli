package cli

var commands = map[string]Command{}

func Register(name string, c Command) {
	commands[name] = c
}

func List() map[string]Command {
	return commands
}
