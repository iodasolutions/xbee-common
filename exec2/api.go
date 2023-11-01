package exec2

import (
	"context"
	"github.com/iodasolutions/xbee-common/cmd"
)

func Run(ctx context.Context, name string, args ...string) *cmd.XbeeError {
	c := NewCommand(name, args...)
	return c.Run(ctx)
}
func RunReturnStdOut(ctx context.Context, name string, args ...string) (string, *cmd.XbeeError) {
	c := NewCommand(name, args...)
	err := c.Run(ctx)
	return c.Result(), err
}
