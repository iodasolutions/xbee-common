package provider

import (
	"bytes"
	"context"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/template"
	"sync"
)

var script = `#!/bin/bash
set -e
if [ ! -f {{ .XbeePath }} ]; then
	apt-get update
#docker
	apt-get install -y \
		ca-certificates \
		curl \
		gnupg \
		lsb-release
	mkdir -m 0755 -p /etc/apt/keyrings
	curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
	echo \
	  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
	  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
	apt-get update
	apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
	usermod -aG docker {{ .user }}
#xbee
	{{ if .remote }}
	archi=$(uname -m)
	if [ "${archi}" == "x86_64" ];then
		archi=amd64
	elif [ "${archi}" == "aarch64" ]; then
	   archi=arm64
	fi
	curl -O "https://s3.eu-west-3.amazonaws.com/xbee.repository.public/linux_${archi}/xbee.gz"
	gunzip xbee.gz && mv ./xbee /usr/bin && chmod +x /usr/bin/xbee
	mkdir -p /xbee/packs
	{{ end }}
	
fi
cat > /etc/hostname <<EOF
{{ .name }}
EOF
hostname {{ .name }}
`

func dockerScript(info *InstanceInfo) string {
	model := map[string]interface{}{
		"user":     info.User,
		"remote":   !cmd.OptionFrom("local").BooleanValue(),
		"name":     info.Name,
		"XbeePath": XbeePath(),
	}
	w := &bytes.Buffer{}
	if err := template.OutputWithTemplate(script, w, model, nil); err != nil {
		panic(cmd.Error("failed to parse network template : %v", err))
	}
	return w.String()
}

func InstallDockerAndXbee(ctx context.Context, up map[string]*InstanceInfo) {
	var wg sync.WaitGroup
	wg.Add(len(up))
	for _, info := range up {
		go func(info *InstanceInfo) {
			defer wg.Done()
			Hosts()
			client, err := info.Connect()
			if err != nil {
				log2.Errorf("instance %s is not reachable via SSH", info.Name)
				return
			}
			if err := client.RunScript(dockerScript(info)); err != nil {
				log2.Errorf("Failed to install docker on host %s", info.Name)
			}
		}(info)
	}
	wg.Wait()
	log2.Infof("docker installed")
}
