package provider

import (
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
)

func instanceInfosCommand() *cmd.Command {
	return &cmd.Command{
		Use: "instanceinfos",
		Run: doInstanceInfo,
	}
}

func instanceInfosFile() newfs.File {
	return newfs.ChildXbee(newfs.CWD()).ChildFileJson("InstanceInfos")
}

func doInstanceInfo(_ []string) *cmd.XbeeError {
	infos, err := provider.InstanceInfos()
	if err == nil {
		f := instanceInfosFile()
		f.Save(infos)
	}
	return err
}
