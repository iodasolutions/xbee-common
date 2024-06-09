package provider

import (
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/newfs"
	"os"
)

var provider Provider
var admin Admin

type Provider interface {
	Up() (*InitialStatus, *cmd.XbeeError)
	Delete() *cmd.XbeeError
	InstanceInfos() ([]*InstanceInfo, *cmd.XbeeError)
	Image() *cmd.XbeeError
}

type Admin interface {
	DestroyVolumes([]string) *cmd.XbeeError
}

func Execute(p Provider, a Admin) {
	defer func() {
		log2.Close()
	}()
	provider = p
	admin = a
	ok, err := cmd.Setup(buildCmdTree)
	if !ok {
		err = cmd.Error("unknown action : %s", os.Args[1])
	}
	if err == nil {
		err = cmd.Run()
	}
	if err != nil {
		newfs.DoExitOnError(err)
	}
}

func buildCmdTree(root *cmd.Command) *cmd.XbeeError {
	return root.AddCommands(upCommand(),
		deleteCommand(),
		destroyVolumesCommand(),
		instanceInfosCommand(),
		imageCommand())
}
