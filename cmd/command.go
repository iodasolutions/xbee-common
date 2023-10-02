package cmd

import (
	"bytes"
	"text/template"
)

type Command struct {
	Use          string
	Short        string
	Long         string
	Aliases      []string
	Hidden       bool
	Run          func([]string) *XbeeError
	commands     map[string]*Command
	Options      []*Option
	ValidateArgs func([]string) *XbeeError
	parent       *Command // used to display usage
}

func NewCommand(name string, aliases ...string) *Command {
	return &Command{
		Use:     name,
		Aliases: aliases,
	}
}

func (c *Command) WithRun(f func([]string) *XbeeError) *Command {
	c.Run = f
	return c
}
func (c *Command) HasOptions() bool {
	return len(c.Options) > 0
}

func (c *Command) AddCommand(child *Command) (bool, *XbeeError) {
	if c.commands == nil {
		c.commands = make(map[string]*Command)
	}
	if existingChild, ok := c.commands[child.Use]; !ok {
		c.commands[child.Use] = child
		child.parent = c
		return true, nil
	} else {
		if child.Run != nil {
			return false, Error("cannot add leaf command %s to %s, one already added", child.Use, c.Use)
		}
		existingChild.AddCommands(child.commandsList()...)
	}
	return false, nil
}
func (c *Command) AddCommands(cmds ...*Command) *XbeeError {
	for _, aCmd := range cmds {
		if _, err := c.AddCommand(aCmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) AvailableSubCommandsToDisplay() (string, *XbeeError) {
	t := template.New("subcommands")
	t, err := t.Parse(subCommandsTpl)
	if err != nil { //should not occur
		return "", Error("unexpected internal error when trying to parse template that list sub commands : %v", err)
	}
	sb := new(bytes.Buffer)
	err = t.Execute(sb, c.commands)
	if err != nil {
		return "", Error("unexpected internal error when trying to render the list of sub commands : %v", err)
	}
	return sb.String(), nil
}
func (c *Command) Usage() (string, *XbeeError) {
	t := template.New("usage")
	t, err := t.Parse(usageTpl)
	if err != nil { //should not occur
		return "", Error("unexpected internal error when trying to parse template that show command usage : %v", err)
	}
	sb := new(bytes.Buffer)
	err = t.Execute(sb, c)
	if err != nil {
		return "", Error("unexpected internal error when trying to render command usage : %v", err)
	}
	return sb.String(), nil
}
func (c *Command) OptionsToDisplay() string {
	return displayOptions(c.Options)
}
func (c *Command) GlobalOptionsToDisplay() string {
	var options []*Option
	for _, option := range globalOptions {
		options = append(options, option)
	}
	return displayOptions(options)
}

func (c *Command) Path() (result string) {
	result = c.Use
	for aCmd := c.parent; aCmd != nil && aCmd.Use != ""; aCmd = aCmd.parent {
		result = aCmd.Use + " " + result
	}
	return
}

func (c *Command) SubCommandNames() (result []string) {
	for _, childC := range c.commands {
		result = append(result, childC.Use)
	}
	return
}
func (c *Command) commandsList() (result []*Command) {
	for _, elt := range c.commands {
		result = append(result, elt)
	}
	return
}

func ExactArgs(n int) func([]string) *XbeeError {
	f := func(args []string) *XbeeError {
		if len(args) != n {
			return Error("expected %d args, actual is %d", n, len(args))
		}
		return nil
	}
	return f
}
func MinArgs(n int) func([]string) *XbeeError {
	f := func(args []string) *XbeeError {
		if len(args) < n {
			return Error("expected at least %d args, actual is %d", n, len(args))
		}
		return nil
	}
	return f
}
func MaxArgs(n int) func([]string) *XbeeError {
	f := func(args []string) *XbeeError {
		if len(args) > n {
			return Error("expected at most %d args, actual is %d", n, len(args))
		}
		return nil
	}
	return f
}
