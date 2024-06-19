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

func (i InstanceInfos) FilterByStates(states ...string) (result InstanceInfos) {
	for _, info := range i {
		if util.Contains(states, info.State) {
			result = append(result, info)
		}
	}
	return
}

func (i InstanceInfos) FilterByInitialStates(states ...string) (result InstanceInfos) {
	for _, info := range i {
		if util.Contains(states, info.InitialState) {
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
