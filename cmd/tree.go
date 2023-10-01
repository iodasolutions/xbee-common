package cmd

import (
	"fmt"
	"os"
)

// root MUST be initialized through Setup func
var root Command
var leaf *Command
var realArgs []string
var isHelp bool
var Args []string

var InContainer bool
var InVm bool
var ContainerOption = "--xbeeContainer"
var VmOption = "--xbeeVm"
var UserOption = "--xbeeUID"
var GroupOption = "--xbeeGID"
var AliasOption = "--xbeeAlias"
var Alias string
var EnvOption = "-e"

// var HostUser *user2.User
var UserId string
var GroupId string
var Envs []string

func init() {
	ok, args := filterBoolOption(ContainerOption, os.Args[1:])
	if ok {
		InContainer = true
	} else {
		ok, args = filterBoolOption(VmOption, args)
		if ok {
			InVm = true
		}
	}
	Alias, args = filterValueOption(AliasOption, args)
	UserId, args = filterValueOption(UserOption, args)
	GroupId, args = filterValueOption(GroupOption, args)
	Envs, args = filterValuesOption(EnvOption, args)
	Args = args
}

func filterBoolOption(option string, args []string) (bool, []string) {
	if len(args) > 0 {
		if args[0] == option {
			return true, args[1:]
		} else {
			return false, args
		}
	}
	return false, nil
}

func filterValueOption(option string, args []string) (string, []string) {
	if len(args) > 0 {
		if args[0] == option {
			return args[1], args[2:]
		} else {
			return "", args
		}
	}
	return "", nil
}
func filterValuesOption(option string, args []string) ([]string, []string) {
	var result []string
	var value string
	for {
		value, args = filterValueOption(option, args)
		if value != "" {
			result = append(result, value)
		} else {
			return result, args
		}
	}
}

func Setup(f func(*Command)) (bool, error) {
	f(&root)
	args := Args
	if len(args) > 0 && args[0] == "help" {
		isHelp = true
		args = args[1:]
	}
	var err error
	leaf, realArgs, err = findRunnable(&root, args)
	return leaf != nil, err
}

func RootCommand() *Command {
	return &root
}

func AddCommands(cmds ...*Command) {
	root.AddCommands(cmds...)
}
func IsHelp() bool {
	isH := GlobalOption("help").BooleanValue()
	return isH || isHelp
}

func Run() error {
	if leaf.ValidateArgs != nil {
		if err := leaf.ValidateArgs(realArgs); err != nil {
			fmt.Println(leaf.Usage())
			return err
		}
	}
	return leaf.Run(realArgs)
}

func Leaf() *Command     { return leaf }
func RealArgs() []string { return realArgs }

func findRunnable(c *Command, args []string) (*Command, []string, error) {
	if c.Run != nil {
		realArgs := NewArgsParser(c.Options).ParseArgs(args...)
		return c, realArgs, nil
	}
	if len(args) == 0 {
		return nil, nil, fmt.Errorf("Command %s needs a subcommand among %v\n", c.Use, c.SubCommandNames())
	}
	name := args[0]
	var childFound *Command
	for _, childC := range c.commands {
		if childC.Use == name {
			childFound = childC
			break
		}
		for _, alias := range childC.Aliases {
			if alias == name {
				childFound = childC
				break
			}
		}
	}
	if childFound == nil {
		return nil, []string{name}, nil // no command found
	} else {
		return findRunnable(childFound, args[1:])
	}
}
