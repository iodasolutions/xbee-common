package provider

import (
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

func instanceInfosFile() newfs.File {
	return newfs.ChildXbee(newfs.CWD()).ChildFileJson("InstanceInfos")
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
	instanceInfosFile().Save(i)
}
func (i InstanceInfos) filterByHosts(names ...string) (result InstanceInfos) {
	for _, info := range i {
		if util.Contains(names, info.Name) {
			result = append(result, info)
		}
	}
	return
}

func (i InstanceInfos) filterByStates(states ...string) (result InstanceInfos) {
	for _, info := range i {
		if util.Contains(states, info.State) {
			result = append(result, info)
		}
	}
	return
}

func InstanceInfosFromProvider() (instanceInfos InstanceInfos, err *cmd.XbeeError) {
	f := instanceInfosFile()
	if !f.Exists() {
		panic(cmd.Error("file %s MUST exist", f))
	}
	defer func() {
		err2 := f.EnsureDelete()
		err = cmd.FollowedWith(err, err2)
	}()
	instanceInfos, err = newfs.Unmarshal[InstanceInfos](f)
	return
}
