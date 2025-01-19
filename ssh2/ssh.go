package ssh

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/newfs"
	"golang.org/x/crypto/ssh"
	"io"
)

func Connect(host string, port string, user string) (conn *ssh.Client, err *cmd.XbeeError) {
	var aConf *ssh.ClientConfig
	rg := newfs.NewRsaGen(newfs.NewFolder(""))
	xbeeKey := rg.RootKeyPEM().Content()
	pemBytes := []byte(xbeeKey)
	signer, err2 := ssh.ParsePrivateKey(pemBytes)
	if err2 != nil {
		err = cmd.Error("ssh : parse key %s failed for host %s : %v", xbeeKey, host, err2)
		return
	}
	aConf = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	aConf.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	connexionString := fmt.Sprintf("%s:%s", host, port)
	conn, err2 = ssh.Dial("tcp", connexionString, aConf)
	if err2 != nil {
		err = cmd.Error("ssh : cannot connect to %s with user %s : %v", connexionString, user, err2)
		return
	}
	var aSession *ssh.Session
	if aSession, err2 = conn.NewSession(); err2 == nil {
		defer func() {
			if aSession != nil {
				if err3 := aSession.Close(); err3 != nil && err3 != io.EOF {
					log2.Warnf("An error occurred when closing session : %v", err3)
				}
			}
		}()
	}
	return
}
