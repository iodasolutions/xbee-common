package provider

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
)

func upCommand() *cmd.Command {
	return &cmd.Command{
		Options: []*cmd.Option{
			cmd.NewBooleanOption("local", "", false),
		},
		Use: string(Up),
		Run: doUp,
	}
}

func doUp(_ []string) *cmd.XbeeError {
	envName := EnvName()
	log2.Infof("Create/Start all instances from environment %s and wait...", envName)
	r, err := provider.Up()
	if err != nil {
		return err
	}

	infos := InstanceInfos(r)
	infos.Save()

	if err == nil {
		log2.Infof(fmt.Sprintf("Environment %s is now up", envName))
	}
	return err
}
