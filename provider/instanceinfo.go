package provider

import (
	"bytes"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/template"
)

type InstanceInfo struct {
	Name          string `yaml:"name,omitempty"`
	State         string `yaml:"state,omitempty"`
	ExternalIp    string `yaml:"externalip,omitempty"`
	SSHPort       string `yaml:"sshport,omitempty"`
	Ip            string `yaml:"ip,omitempty"`
	User          string `yaml:"user,omitempty"`
	PackIdExist   bool   `yaml:"packidexist,omitempty"`
	SystemIdExist bool   `yaml:"systemidexist,omitempty"`
}

func (info *InstanceInfo) HostnameScript() string {
	model := map[string]interface{}{
		"name": info.Name,
	}
	script := `cat > /etc/hostname <<EOF
{{ .name }}
EOF
hostname {{ .name }}
`
	w := &bytes.Buffer{}
	if err := template.OutputWithTemplate(script, w, model, nil); err != nil {
		panic(cmd.Error("failed to parse network template : %v", err))
	}
	return w.String()
}
