package provider

import (
	"bytes"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/template"
	"strings"
)

var authorizedKey = `
mkdir -p /home/{{ .user }}/.ssh
touch /home/{{ .user }}/.ssh/authorized_keys
cat >> /home/{{ .user }}/.ssh/authorized_keys <<EOF
{{ .xbeepublickey }}
EOF
cat >> /etc/ssh/sshd_config <<EOF
PubkeyAcceptedKeyTypes=+ssh-rsa
EOF
systemctl restart sshd
`

func authorizedKeyModel(user string) map[string]interface{} {
	aMap := map[string]interface{}{
		"xbeepublickey": strings.TrimSpace(newfs.NewRsaGen("").RootAuthorizedKey().Content()),
		"user":          user,
	}
	return aMap
}

func AuthorizedKeyScript(user string) string {
	w := &bytes.Buffer{}
	model := authorizedKeyModel(user)
	if err := template.OutputWithTemplate(authorizedKey, w, model, nil); err != nil {
		panic(cmd.Error("failed to parse userData template : %v", err))
	}
	return w.String()
}
