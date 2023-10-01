package exec2

import "context"

func Run(ctx context.Context, name string, args ...string) error {
	c := NewCommand(name, args...)
	return c.Run(ctx)
}
func RunReturnStdOut(ctx context.Context, name string, args ...string) (string, error) {
	c := NewCommand(name, args...)
	return c.RunReturnStdOut(ctx)
}
