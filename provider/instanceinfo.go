package provider

import (
	"bytes"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/template"
)

type InstanceInfo struct {
	Name          string `json:"name,omitempty"`
	State         string `json:"state,omitempty"`
	InitialState  string `json:"initialstate,omitempty"`
	ExternalIp    string `json:"externalip,omitempty"`
	SSHPort       string `json:"sshport,omitempty"`
	Ip            string `json:"ip,omitempty"`
	User          string `json:"user,omitempty"`
	PackIdExist   bool   `json:"packidexist,omitempty"`
	SystemIdExist bool   `json:"systemidexist,omitempty"`
}

func (info *InstanceInfo) hostnameScript() string {
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
