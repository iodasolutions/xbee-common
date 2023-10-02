package provider

import (
	"context"
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/exec2"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/util"
	"os"
	"os/exec"
)

type InstanceInfo struct {
	Name       string `json:"name,omitempty"`
	State      string `json:"state,omitempty"`
	ExternalIp string `json:"externalip,omitempty"`
	SSHPort    string `json:"sshport,omitempty"`
	Ip         string `json:"ip,omitempty"`
	User       string `json:"user,omitempty"`
}

func (info InstanceInfo) Connect() (*exec2.SSHClient, error) {
	return exec2.Connect(info.ExternalIp, info.SSHPort, info.User)
}

func (info InstanceInfo) CheckSSH(ctx context.Context, user string) bool {
	return exec2.CheckSSH(ctx, info.ExternalIp, info.SSHPort, user)
}
func (info InstanceInfo) Enter(ctx context.Context, user string) error {
	args := []string{"-i", newfs.NewRsaGen("").RootKeyPEM().String(),
		"-p", info.SSHPort,
		"-o", "StrictHostKeyChecking=no"}
	args = append(args, fmt.Sprintf("%s@%s", user, info.ExternalIp))
	aCmd := exec.CommandContext(ctx, "ssh", args...)
	aCmd.Stdout = os.Stdout
	aCmd.Stderr = os.Stderr
	aCmd.Stdin = os.Stdin
	return aCmd.Run()
}

type InstanceInfoForEnv map[string]*InstanceInfo

func InstanceInfosFromProvider() InstanceInfoForEnv {
	f := instanceInfosFile()
	if !f.Exists() {
		panic(cmd.Error("file %s MUST exist", f))
	}
	instanceInfos := map[string]*InstanceInfo{}
	f.Unmarshal(&instanceInfos)
	return instanceInfos
}

func (m InstanceInfoForEnv) FilterByState(states ...string) InstanceInfoForEnv {
	result := map[string]*InstanceInfo{}
	for k, v := range m {
		if util.Contains(states, v.State) {
			result[k] = v
		}
	}
	return result
}
func (m InstanceInfoForEnv) FilterByHost(hosts ...string) InstanceInfoForEnv {
	result := map[string]*InstanceInfo{}
	for k, v := range m {
		if util.Contains(hosts, v.Name) {
			result[k] = v
		}
	}
	return result
}
