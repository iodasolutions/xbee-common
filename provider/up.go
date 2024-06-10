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
	log2.Infof("Check SSH for each instance RUNNING")
	//up := r.AllUp()
	//
	//var wg2 sync.WaitGroup
	//wg2.Add(len(up))
	//ctx = context.Background()
	//if len(r.NotExisting) > 0 {
	//	InstallDockerAndXbee(ctx, r.NotExisting)
	//}
	//if len(r.NotExisting) > 0 || len(r.Down) > 0 {
	//	script := toEtcHosts(up)
	//	for _, info := range up {
	//		go func(info *InstanceInfo) {
	//			defer wg2.Done()
	//			client, err := info.Connect()
	//			if err != nil {
	//				log2.Errorf("instance %s is not reachable via SSH", info.Name)
	//			}
	//			if err := client.RunScriptQuiet(script); err != nil {
	//				log2.Errorf("Failed to update /etc/hosts on host %s", info.Name)
	//			}
	//		}(info)
	//	}
	//	wg2.Wait()
	//	log2.Infof("updated DNS for all instances created")
	//}

	infos := InstanceInfos(r)
	infos.Save()

	if err == nil {
		log2.Infof(fmt.Sprintf("Environment %s is now up", envName))
	}
	return err
}

func toEtcHosts(instanceInfos []*InstanceInfo) string {
	var s = `#!/usr/bin/env bash
cat <<EOF | sudo tee /etc/hosts
127.0.0.1 localhost

# The following lines are desirable for IPv6 capable hosts
::1 ip6-localhost ip6-loopback
fe00::0 ip6-localnet
ff00::0 ip6-mcastprefix
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
ff02::3 ip6-allhosts
`
	for _, v := range instanceInfos {
		s = s + fmt.Sprintf("%s %s\n", v.Ip, v.Name)
	}
	s = s + "EOF\n"
	return s
}
