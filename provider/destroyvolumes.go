package provider

import "github.com/iodasolutions/xbee-common/cmd"

func destroyVolumesCommand() *cmd.Command {
	return &cmd.Command{
		Use: "destroyvolumes",
		Run: doDestroyVolumes,
	}
}

func doDestroyVolumes(args []string) error {
	return admin.DestroyVolumes(args)
}