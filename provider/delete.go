package provider

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
)

func deleteCommand() *cmd.Command {
	return &cmd.Command{
		Use: "delete",
		Run: doDelete,
	}
}

func doDelete(_ []string) *cmd.XbeeError {
	envName := EnvName()
	log2.Infof("Delete all instances from environment %s and wait...", envName)
	err := provider.Delete()
	if err == nil {
		log2.Infof(fmt.Sprintf("Environment %s Successfully destroyed", envName))
	}
	return err
}
