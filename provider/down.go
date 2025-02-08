package provider

import (
	"github.com/iodasolutions/xbee-common/cmd"
)

func downCommand() *cmd.Command {
	return &cmd.Command{
		Use: string(Down),
		Run: doDown,
	}
}

func doDown(_ []string) *cmd.XbeeError {
	err := provider.Down()
	return err
}
