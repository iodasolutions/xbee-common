package provider

import (
	"sync"

	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/types"
	"github.com/iodasolutions/xbee-common/util"
	"github.com/iodasolutions/xbee-common/yaml2"
	"gopkg.in/yaml.v3"
)

type Env struct {
	Provider           *yaml.Node             `json:"provider,omitempty"`
	Id                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Hosts              map[string]*XbeeHost   `json:"hosts,omitempty"`
	Volumes            map[string]*XbeeVolume `json:"volumes,omitempty"`
	Nets               []*XbeeNet             `json:"nets,omitempty"`
	SystemProviderData map[string]interface{} `json:"system_provider_data,omitempty"`
}

func (e *Env) VolumesLinkedToHosts() (result []*XbeeVolume) {
	for _, h := range e.Hosts {
		for _, name := range h.Volumes {
			result = append(result, e.Volumes[name])
		}
	}
	return
}

func envYaml() newfs.File {
	return newfs.ChildXbee(newfs.CWD()).ChildFileYml("env")
}
func (e *Env) Save() {
	envYaml().Save(e)
}

type XbeeHost struct {
	Provider     *yaml.Node    `json:"provider,omitempty"`
	Name         string        `json:"name,omitempty"`
	Ports        []string      `json:"ports,omitempty"`
	Volumes      []string      `json:"volumes,omitempty"`
	User         string        `json:"user,omitempty"`
	ExternalIp   string        `json:"externalip,omitempty"`
	SystemName   string        `json:"system_name,omitempty"`
	SystemOrigin *types.Origin `json:"system_origin,omitempty"`
	SystemHash   string        `json:"systemhash,omitempty"`
	PackName     string        `json:"pack_name,omitempty"`
	PackOrigin   *types.Origin `json:"pack_origin,omitempty"`
	PackHash     string        `json:"packhash,omitempty"`
	OsArch       string        `json:"osarch,omitempty"`
}

func (ph *XbeeHost) EffectivePackOrigin() *types.Origin {
	if ph.PackOrigin == nil {
		return ph.SystemOrigin
	}
	return ph.PackOrigin
}
func (ph *XbeeHost) EffectivePackName() string {
	if ph.PackName != "" {
		return ph.PackName
	}
	return ph.SystemName
}
func (ph *XbeeHost) EffectiveHash() string {
	if ph.PackOrigin == nil {
		return ph.SystemHash
	}
	return ph.PackHash
}
func (ph *XbeeHost) DisplayName() string {
	name := ph.EffectivePackName()
	if ph.PackOrigin != nil {
		name += "-" + ph.SystemName
	}
	return name
}

type XbeeVolume struct {
	Provider *yaml.Node `json:"provider,omitempty"`
	Name     string     `json:"name,omitempty"`
	Size     int        `json:"size,omitempty"`
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
	if env.Env, err = newfs.Unmarshal[*Env](envYaml()); err != nil {
		newfs.DoExitOnError(err)
	}

	hostProvider := yaml2.FindNodeNoError(env.Env.Provider, "host")
	for index := range env.Env.Hosts {
		yaml2.MergeNodes(env.Env.Hosts[index].Provider, hostProvider)
	}
	volumeProvider := yaml2.FindNodeNoError(env.Env.Provider, "volume")
	for index := range env.Env.Volumes {
		yaml2.MergeNodes(env.Env.Volumes[index].Provider, volumeProvider)
	}
}

func Hosts() (result map[string]*XbeeHost) {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.Hosts
}

func VolumesForEnv() (result map[string]*XbeeVolume) {
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

func SystemProviderDataFor(systemHash string) map[string]interface{} {
	env.once.Do(func() {
		initEnv()
	})
	return env.Env.SystemProviderData[systemHash].(map[string]interface{})
}
