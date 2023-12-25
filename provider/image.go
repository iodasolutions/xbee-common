package provider

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
)

func imageCommand() *cmd.Command {
	return &cmd.Command{
		Use: string(Image),
		Run: doImage,
	}
}

func doImage([]string) *cmd.XbeeError {
	envName := EnvName()
	log2.Infof("Create images from environment %s and wait...", envName)
	err := provider.Image()
	if err == nil {
		log2.Infof(fmt.Sprintf("Images from Environment %s Successfully created", envName))
	}
	return err
}
