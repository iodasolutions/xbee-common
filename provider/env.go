package provider

import (
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/types"
	"github.com/iodasolutions/xbee-common/util"
	"sync"
)

func EnvJson() newfs.File {
	return newfs.ChildXbee(newfs.CWD()).ChildFileJson("env")
}

type Env struct {
	Origin  string    `json:"origin,omitempty"`
	Commit  string    `json:"commit,omitempty"`
	Hosts   []*Host   `json:"hosts,omitempty"`
	Volumes []*Volume `json:"volumes,omitempty"`
	Nets    []*Net    `json:"nets,omitempty"`
}

type Host struct {
	Name       string                 `json:"name,omitempty"`
	Provider   map[string]interface{} `json:"provider,omitempty"`
	Ports      []string               `json:"ports,omitempty"`
	Volumes    []string               `json:"volumes,omitempty"`
	User       string                 `json:"user,omitempty"`
	ExternalIp string                 `json:"externalip,omitempty"`
}

type Volume struct {
	Name     string                 `json:"name,omitempty"`
	Provider map[string]interface{} `json:"provider,omitempty"`
	Size     int                    `json:"size,omitempty"`
}

type Net struct {
	Name     string                 `json:"name,omitempty"`
	Provider map[string]interface{} `json:"provider,omitempty"`
	Cidr     string                 `json:"cidr,omitempty"`
}

var env struct {
	Env  *Env
	once sync.Once
}

func initEnv() {
	var err *cmd.XbeeError
	if env.Env, err = newfs.Unmarshal[*Env](EnvJson()); err != nil {
		newfs.DoExitOnError(err)
	}
}

func Hosts() (result []*Host) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Hosts
}

func VolumesForEnv() (result []*Volume) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Volumes
}
func NetsForEnv() (result []*Net) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Nets
}

func EnvId() *types.IdJson {
	env.once.Do(func() {
		initEnv()
	})
	return &types.IdJson{
		Origin: env.Env.Origin,
		Commit: env.Env.Commit,
	}
}

func VolumesFromEnvironment(names []string) (result []*Volume) {
	for _, v := range VolumesForEnv() {
		if util.Contains(names, v.Name) {
			result = append(result, v)
		}
	}
	return
}
