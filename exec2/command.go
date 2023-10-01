package exec2

import (
	"context"
	"fmt"
	"github.com/iodasolutions/xbee-common/util"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

type Command struct {
	name string
	args []string

	bErr  *MachineReadableWriter
	bOut  *MachineReadableWriter
	quiet bool

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
	c.bOut = NewStdOutMachineReadableWriter()
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
	var cmd *exec.Cmd
	name := c.name
	args := c.args

	if c.user != "" {
		currentUser, err := user.Current()
		if err != nil { // this should not occur
			panic(util.Error("cannot get current user : %v", err))
		}
		if c.user != currentUser.Name {
			name = "su"
			args = append(args, "-l", "-c", c.String(), c.user)
		}
	}

	if ctx != nil {
		cmd = exec.CommandContext(ctx, name, args...)
	} else {
		cmd = exec.Command(c.name, args...)
	}

	if c.quiet {
		c.bErr = NewMachineOnlyReadableWriter()
	} else {
		c.bErr = NewStdErrMachineReadableWriter()
		if c.bOut == nil {
			cmd.Stdout = os.Stdout
		} else {
			cmd.Stdout = c.bOut
		}
		cmd.Stdin = os.Stdin
	}
	cmd.Stderr = c.bErr

	cmd.Dir = c.directory
	if c.env != nil {
		cmd.Env = c.env
	}

	return cmd
}

func (c *Command) Run(ctx context.Context) error {
	aCmd := c.createCmd(ctx)
	err := aCmd.Run()
	if err != nil {
		return fmt.Errorf("this command (%s) failed : %v", c.String(), err)
	}
	return nil
}

func (c *Command) RunReturnStdOut(ctx context.Context) (string, error) {
	c.bOut = NewStdOutMachineReadableWriter()
	aCmd := c.createCmd(ctx)
	err := aCmd.Run()
	return c.bOut.String(), err
}
