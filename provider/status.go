package provider

import (
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/util"
)

func upStatusFile() newfs.File { return newfs.ChildXbee(newfs.CWD()).ChildFileJson("InitialStatus") }

func UpStatusFromProvider() (*InitialStatus, *cmd.XbeeError) {
	f := upStatusFile()
	if !f.Exists() {
		panic(cmd.Error("file %s MUST exist", f))
	}
	if status, err := newfs.Unmarshal[*InitialStatus](f); err != nil {
		return nil, err
	} else {
		return status, nil
	}
}

type InitialStatus struct {
	NotExisting map[string]*InstanceInfo `json:"notexisting,omitempty"`
	Down        map[string]*InstanceInfo `json:"down,omitempty"`
	Up          map[string]*InstanceInfo `json:"up,omitempty"`
	Other       map[string]*InstanceInfo `json:"other,omitempty"`
}

func (ups *InitialStatus) String() string {
	s, _ := util.NewJsonIO(ups).SaveAsString()
	return s
}

func (ups *InitialStatus) AllUp() (result []*InstanceInfo) {
	for _, v := range ups.NotExisting {
		result = append(result, v)
	}
	for _, v := range ups.Down {
		result = append(result, v)
	}
	for _, v := range ups.Up {
		result = append(result, v)
	}
	return
}
