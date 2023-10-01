package provider

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/newfs"
	"os"
)

var provider Provider
var admin Admin

type Provider interface {
	Up() (*InitialStatus, error)
	Delete() error
	InstanceInfos() (map[string]*InstanceInfo, error)
}

type Admin interface {
	DestroyVolumes([]string) error
}

func Execute(p Provider, a Admin) {
	defer func() {
		log2.Close()
		newfs.DeleteTmp()
	}()
	provider = p
	admin = a
	ok, err := cmd.Setup(buildCmdTree)
	if !ok {
		err = fmt.Errorf("unknown action : %s", os.Args[1])
	}
	if err == nil {
		err = cmd.Run()
	}
	if err != nil {
		newfs.DoExitOnError(err)
	}
}

func buildCmdTree(root *cmd.Command) {
	root.AddCommands(upCommand(),
		deleteCommand(),
		destroyVolumesCommand(),
		instanceInfosCommand())
}
