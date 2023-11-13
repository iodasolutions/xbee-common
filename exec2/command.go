package exec2

import (
	"context"
	"github.com/iodasolutions/xbee-common/cmd"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

type Command struct {
	name string
	args []string

	bErr   *MachineReadableWriter
	bOut   *MachineReadableWriter
	quiet  bool
	result bool

	user      string
	directory string
	env       []string
}

func NewCommand(name string, args ...string) *Command {
	c := &Command{
		name: name,
		args: args,
	}
	return c
}

func (c *Command) String() string {
	var args []string
	args = append(args, c.name)
	args = append(args, c.args...)
	return strings.Join(args, " ")
}

func (c *Command) WithResult() *Command {
	c.result = true
	return c
}

func (c *Command) WithDirectory(dir string) *Command {
	c.directory = dir
	return c
}
func (c *Command) WithUser(user string) *Command {
	c.user = user
	return c
}
func (c *Command) WithEnv(env []string) *Command {
	c.env = os.Environ()
	c.env = append(c.env, env...)
	return c
}

func (c *Command) Quiet() *Command {
	c.quiet = true
	return c
}

func (c *Command) createCmd(ctx context.Context) *exec.Cmd {
	var aCmd *exec.Cmd
	name := c.name
	args := c.args

	if c.user != "" {
		currentUser, err := user.Current()
		if err != nil { // this should not occur
			panic(cmd.Error("cannot get current user : %v", err))
		}
		if c.user != currentUser.Name {
			name = "su"
			args = append(args, "-l", "-c", c.String(), c.user)
		}
	}

	if ctx != nil {
		aCmd = exec.CommandContext(ctx, name, args...)
	} else {
		aCmd = exec.Command(c.name, args...)
	}

	if c.quiet {
		c.bErr = NewMachineOnlyReadableWriter()
		if c.result {
			c.bOut = NewMachineOnlyReadableWriter()
		}
	} else {
		c.bErr = NewStdErrMachineReadableWriter()
		if c.result {
			c.bOut = NewStdOutMachineReadableWriter()
			aCmd.Stdout = c.bOut
		} else {
			aCmd.Stdout = os.Stdout
		}
		aCmd.Stdin = os.Stdin
	}
	aCmd.Stderr = c.bErr

	aCmd.Dir = c.directory
	if c.env != nil {
		aCmd.Env = c.env
	}

	return aCmd
}

func (c *Command) Run(ctx context.Context) *cmd.XbeeError {
	aCmd := c.createCmd(ctx)
	err := aCmd.Run()
	if err != nil {
		return cmd.Error("this command (%s) failed : %v", c.String(), err)
	}
	return nil
}

func (c *Command) Result() string {
	if c.bOut == nil {
		return ""
	} else {
		result := c.bOut.String()
		return result
	}
}
