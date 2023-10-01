package cmd

import (
	"bytes"
	"fmt"
	"text/template"
)

type Command struct {
	Use          string
	Short        string
	Long         string
	Aliases      []string
	Hidden       bool
	Run          func([]string) error
	commands     map[string]*Command
	Options      []*Option
	ValidateArgs func([]string) error
	parent       *Command // used to display usage
}

func NewCommand(name string, aliases ...string) *Command {
	return &Command{
		Use:     name,
		Aliases: aliases,
	}
}

func (c *Command) WithRun(f func([]string) error) *Command {
	c.Run = f
	return c
}
func (c *Command) HasOptions() bool {
	return len(c.Options) > 0
}

func (c *Command) AddCommand(child *Command) (bool, error) {
	if c.commands == nil {
		c.commands = make(map[string]*Command)
	}
	if existingChild, ok := c.commands[child.Use]; !ok {
		c.commands[child.Use] = child
		child.parent = c
		return true, nil
	} else {
		if child.Run != nil {
			return false, fmt.Errorf("cannot add leaf command %s to %s, one already added", child.Use, c.Use)
		}
		existingChild.AddCommands(child.commandsList()...)
	}
	return false, nil
}
func (c *Command) AddCommands(cmds ...*Command) error {
	for _, aCmd := range cmds {
		if _, err := c.AddCommand(aCmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) AvailableSubCommandsToDisplay() (string, error) {
	t := template.New("subcommands")
	t, err := t.Parse(subCommandsTpl)
	if err != nil { //should not occur
		return "", fmt.Errorf("unexpected internal error when trying to parse template that list sub commands : %v", err)
	}
	sb := new(bytes.Buffer)
	err = t.Execute(sb, c.commands)
	if err != nil {
		return "", fmt.Errorf("unexpected internal error when trying to render the list of sub commands : %v", err)
	}
	return sb.String(), nil
}
func (c *Command) Usage() (string, error) {
	t := template.New("usage")
	t, err := t.Parse(usageTpl)
	if err != nil { //should not occur
		return "", fmt.Errorf("unexpected internal error when trying to parse template that show command usage : %v", err)
	}
	sb := new(bytes.Buffer)
	err = t.Execute(sb, c)
	if err != nil {
		return "", fmt.Errorf("unexpected internal error when trying to render command usage : %v", err)
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

func ExactArgs(n int) func([]string) error {
	f := func(args []string) error {
		if len(args) != n {
			return fmt.Errorf("expected %d args, actual is %d", n, len(args))
		}
		return nil
	}
	return f
}
func MinArgs(n int) func([]string) error {
	f := func(args []string) error {
		if len(args) < n {
			return fmt.Errorf("expected at least %d args, actual is %d", n, len(args))
		}
		return nil
	}
	return f
}
func MaxArgs(n int) func([]string) error {
	f := func(args []string) error {
		if len(args) > n {
			return fmt.Errorf("expected at most %d args, actual is %d", n, len(args))
		}
		return nil
	}
	return f
}
