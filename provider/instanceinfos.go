package provider

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/util"
)

func instanceInfosCommand() *cmd.Command {
	return &cmd.Command{
		Use: string(Infos),
		Run: doInstanceInfo,
	}
}

func doInstanceInfo(_ []string) *cmd.XbeeError {
	value, err := provider.InstanceInfos()
	if err == nil {
		infos := InstanceInfos(value)
		infos.Save()
	}
	return err
}

type InstanceInfos []*InstanceInfo

func (i InstanceInfos) Save() {
	newfs.ChildXbee(newfs.CWD()).ChildFileJson("InstanceInfos").Save(i)
}
func (i InstanceInfos) FilterByHosts(names ...string) (result InstanceInfos) {
	for _, info := range i {
		if util.Contains(names, info.Name) {
			result = append(result, info)
		}
	}
	return
}

func (i InstanceInfos) ToMap() map[string]*InstanceInfo {
	result := make(map[string]*InstanceInfo)
	for _, info := range i {
		result[info.Name] = info
	}
	return result
}
func InstanceInfosFromProvider() (instanceInfos InstanceInfos, err *cmd.XbeeError) {
	return InstanceInfosFromProviderFor(newfs.CWD())
}

func InstanceInfosFromProviderFor(fd newfs.Folder) (instanceInfos InstanceInfos, err *cmd.XbeeError) {
	f := newfs.ChildXbee(fd).ChildFileJson("InstanceInfos")
	if !f.Exists() {
		err = cmd.Error("file %s MUST exist", f)
		return
	}
	defer func() {
		err2 := f.EnsureDelete()
		err = cmd.FollowedWith(err, err2)
	}()
	instanceInfos, err = newfs.Unmarshal[InstanceInfos](f)
	return
}

func (i InstanceInfos) ToEtcHosts() string {
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
	for _, v := range i {
		if v.Ip != "" {
			s = s + fmt.Sprintf("%s %s\n", v.Ip, v.Name)
		}
	}
	s = s + "EOF\n"
	return s
}
