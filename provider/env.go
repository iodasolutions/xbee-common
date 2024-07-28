package provider

import (
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/types"
	"github.com/iodasolutions/xbee-common/util"
	"sync"
)

func envJson() newfs.File {
	return newfs.ChildXbee(newfs.CWD()).ChildFileJson("env")
}

func Save(e *Env) {
	envJson().Save(e)
}

type Env struct {
	Id      string        `json:"id"`
	Name    string        `json:"name"`
	Hosts   []*XbeeHost   `json:"hosts,omitempty"`
	Volumes []*XbeeVolume `json:"volumes,omitempty"`
	Nets    []*XbeeNet    `json:"nets,omitempty"`
}

type XbeeHost struct {
	Name       string                 `json:"name,omitempty"`
	Provider   map[string]interface{} `json:"provider,omitempty"`
	Ports      []string               `json:"ports,omitempty"`
	Volumes    []string               `json:"volumes,omitempty"`
	User       string                 `json:"user,omitempty"`
	ExternalIp string                 `json:"externalip,omitempty"`
	SystemId   *types.IdJson          `json:"systemid,omitempty"`
	SystemHash string                 `json:"systemhash,omitempty"`
	PackId     *types.IdJson          `json:"packid,omitempty"`
	PackHash   string                 `json:"packhash,omitempty"`
}

type XbeeVolume struct {
	Name     string                 `json:"name,omitempty"`
	Provider map[string]interface{} `json:"provider,omitempty"`
	Size     int                    `json:"size,omitempty"`
}

type XbeeNet struct {
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
	if env.Env, err = newfs.Unmarshal[*Env](envJson()); err != nil {
		newfs.DoExitOnError(err)
	}
}

func Hosts() (result []*XbeeHost) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Hosts
}

func VolumesForEnv() (result []*XbeeVolume) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Volumes
}
func NetsForEnv() (result []*XbeeNet) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Nets
}

// EnvName should be used for logging purpose.
func EnvName() string {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Name
}

func EnvId() string {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Id
}

func VolumesFromEnvironment(names []string) (result []*XbeeVolume) {
	for _, v := range VolumesForEnv() {
		if util.Contains(names, v.Name) {
			result = append(result, v)
		}
	}
	return
}
