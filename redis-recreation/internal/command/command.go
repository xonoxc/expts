package command

const DEFAULT_ARG_CAPACITY = 3

type Command struct {
	Name string
	Args []string
}

func NewCommand() *Command {
	return &Command{
		Name: "",
		Args: make([]string, DEFAULT_ARG_CAPACITY),
	}
}
