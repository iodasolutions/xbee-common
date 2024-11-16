package provider

import (
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/types"
	"github.com/iodasolutions/xbee-common/util"
	"sync"
)

type XbeeElement[T any] struct {
	Provider map[string]interface{} `json:"provider,omitempty"`
	Element  T                      `json:"element,omitempty"`
}

func envJson() newfs.File {
	return newfs.ChildXbee(newfs.CWD()).ChildFileJson("env")
}

func Save(e *Env) {
	envJson().Save(e)
}

type Env struct {
	Id                 string                    `json:"id"`
	Name               string                    `json:"name"`
	Hosts              []XbeeElement[XbeeHost]   `json:"hosts,omitempty"`
	Volumes            []XbeeElement[XbeeVolume] `json:"volumes,omitempty"`
	Nets               []XbeeElement[XbeeNet]    `json:"nets,omitempty"`
	SystemProviderData map[string]interface{}    `json:"system_provider_data,omitempty"`
	Provider           map[string]interface{}    `json:"provider,omitempty"`
}

type XbeeHost struct {
	Name       string        `json:"name,omitempty"`
	Ports      []string      `json:"ports,omitempty"`
	Volumes    []string      `json:"volumes,omitempty"`
	User       string        `json:"user,omitempty"`
	ExternalIp string        `json:"externalip,omitempty"`
	SystemId   *types.IdJson `json:"systemid,omitempty"`
	SystemHash string        `json:"systemhash,omitempty"`
	PackId     *types.IdJson `json:"packid,omitempty"`
	PackHash   string        `json:"packhash,omitempty"`
	OsArch     string        `json:"osarch,omitempty"`
}

type XbeeVolume struct {
	Name string `json:"name,omitempty"`
	Size int    `json:"size,omitempty"`
}

type XbeeNet struct {
	Name string `json:"name,omitempty"`
	Cidr string `json:"cidr,omitempty"`
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

func mergeProviders() {
	hostProvider := env.Env.Provider["host"].(map[string]interface{})
	for _, h := range env.Env.Hosts {
		h.Provider = util.MergeMaps(hostProvider, h.Provider)
	}
	volumeProvider := env.Env.Provider["volume"].(map[string]interface{})
	for _, v := range env.Env.Volumes {
		v.Provider = util.MergeMaps(volumeProvider, v.Provider)
	}
}

func Hosts() (result []XbeeElement[XbeeHost]) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Hosts
}

func VolumesForEnv() (result []XbeeElement[XbeeVolume]) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Volumes
}
func NetsForEnv() (result []XbeeElement[XbeeNet]) {
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

func VolumesFromEnvironment(names []string) (result []XbeeElement[XbeeVolume]) {
	for _, v := range VolumesForEnv() {
		if util.Contains(names, v.Element.Name) {
			result = append(result, v)
		}
	}
	return
}

func SystemProviderDataFor(systemHash string) map[string]interface{} {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.SystemProviderData[systemHash].(map[string]interface{})
}
