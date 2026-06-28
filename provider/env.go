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
	Provider           *yaml.Node             `yaml:"provider,omitempty"`
	Id                 string                 `yaml:"id"`
	Name               string                 `yaml:"name"`
	Hosts              map[string]*XbeeHost   `yaml:"hosts,omitempty"`
	Volumes            map[string]*XbeeVolume `yaml:"volumes,omitempty"`
	Nets               []*XbeeNet             `yaml:"nets,omitempty"`
	SystemProviderData map[string]interface{} `yaml:"system_provider_data,omitempty"`
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
	Provider     *yaml.Node    `yaml:"provider,omitempty"`
	Name         string        `yaml:"name,omitempty"`
	Ports        []string      `yaml:"ports,omitempty"`
	Volumes      []string      `yaml:"volumes,omitempty"`
	User         string        `yaml:"user,omitempty"`
	ExternalIp   string        `yaml:"externalip,omitempty"`
	SystemName   string        `yaml:"system_name,omitempty"`
	SystemOrigin *types.Origin `yaml:"system_origin,omitempty"`
	SystemHash   string        `yaml:"systemhash,omitempty"`
	PackName     string        `yaml:"pack_name,omitempty"`
	PackOrigin   *types.Origin `yaml:"pack_origin,omitempty"`
	PackHash     string        `yaml:"packhash,omitempty"`
	OsArch       string        `yaml:"osarch,omitempty"`
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
	Provider *yaml.Node `yaml:"provider,omitempty"`
	Name     string     `yaml:"name,omitempty"`
	Size     int        `yaml:"size,omitempty"`
}

type XbeeNet struct {
	Name string `yaml:"name,omitempty"`
	Cidr string `yaml:"cidr,omitempty"`
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
