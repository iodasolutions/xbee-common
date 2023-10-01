package exec2

import (
	"context"
	"fmt"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/util"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"time"
)

type SSHClient struct {
	conn *ssh.Client
}

func Connect(host string, port string, user string) (client *SSHClient, err *util.XbeeError) {
	var aConf *ssh.ClientConfig
	rg := newfs.NewRsaGen("")
	xbeeKey := rg.RootKeyPEM().Content()
	pemBytes := []byte(xbeeKey)
	signer, err2 := ssh.ParsePrivateKey(pemBytes)
	if err2 != nil {
		err = util.Error("ssh : parse key %s failed for host %s : %v", xbeeKey, host, err2)
		return
	}
	aConf = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	aConf.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	connexionString := fmt.Sprintf("%s:%s", host, port)
	conn, err3 := ssh.Dial("tcp", connexionString, aConf)
	if err3 != nil {
		err = util.Error("ssh : cannot connect to %s with user %s : %v", connexionString, user, err3)
		return
	}
	if aSession, err4 := conn.NewSession(); err4 == nil {
		defer func() {
			err = util.CloseWithError(aSession, err)
		}()
		client = &SSHClient{conn: conn}
	}
	return
}
func (c *SSHClient) Close() *util.XbeeError {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			return util.Error("%v", err)
		}
	}
	return nil
}

func (c *SSHClient) RunCommand(command string) *util.XbeeError {
	return c.run(command, true)
}

func (c *SSHClient) RunCommandToOut(command string) (out string, err *util.XbeeError) {
	sess, err2 := c.conn.NewSession()
	if err2 != nil {
		return "", util.Error("cannot create session : %v", err2)
	}
	defer func() {
		err = util.CloseWithError(sess, err)
	}()
	w := NewMachineOnlyReadableWriter()
	we := NewMachineOnlyReadableWriter()
	sess.Stdout = w
	sess.Stderr = we
	if err2 = sess.Run(command); err2 != nil {
		err = util.Error("This command [%s] failed : %s", command, we.String())
	} else {
		out = w.String()
	}
	return
}
func (c *SSHClient) RunCommandQuiet(command string) *util.XbeeError {
	return c.run(command, false)
}

func (c *SSHClient) run(command string, redirectStd bool) (err *util.XbeeError) {
	sess, err2 := c.conn.NewSession()
	if err2 != nil {
		return util.Error("cannot create session : %v", err2)
	}
	defer func() {
		err = util.CloseWithError(sess, err)
	}()
	if redirectStd {
		sess.Stdout = os.Stdout
		sess.Stdin = os.Stdin
		sess.Stderr = os.Stderr
	}
	err2 = sess.Run(command)
	if err2 != nil {
		err = util.Error("run command in session failed : %v", err2)
	}
	return
}

func (c *SSHClient) RunScript(script string) error {
	return c.runScript(script, true)
}
func (c *SSHClient) RunScriptQuiet(script string) error {
	return c.runScript(script, false)
}
func (c *SSHClient) runScript(script string, redirectStd bool) (err error) {
	f := newfs.EnsureTmpDir().RandomFile()
	defer func() {
		err2 := f.EnsureDelete()
		if err2 != nil {
			if err == nil {
				err = err2
			} else {
				err = fmt.Errorf("closing session failed: %v. First error was : %v", err2, err)
			}
		}
	}()
	f.SetContent(script)
	err = c.Upload(f, "/tmp")
	if err != nil {
		return
	}
	err = c.run(fmt.Sprintf("sudo bash /tmp/%s", f.Base()), redirectStd)
	return
}

func (c *SSHClient) Upload(path newfs.File, todir newfs.Folder) (err error) {
	if err = c.RunCommandQuiet(fmt.Sprintf("sudo mkdir -p %s", todir)); err != nil {
		return
	}
	var session *ssh.Session
	session, err = c.conn.NewSession()
	if err != nil {
		return
	}
	defer func() {
		err = util.CloseWithError(session, err)
	}()
	go func() {
		w, _ := session.StdinPipe()
		defer func() {
			err = util.CloseWithError(w, err)
		}()
		//		fmt.Fprintln(w, "D0755", 0, "xbee") // mkdir
		if err = uploadInternScp(string(path), w); err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		if _, err = fmt.Fprint(w, "\x00"); err != nil {
			return
		} // transfer end with \x00

	}()
	command := fmt.Sprintf("sudo /usr/bin/scp -tr %s", todir)
	err = session.Run(command)
	return err
}

func uploadInternScp(path string, w io.Writer) (err error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return
	}
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		err = util.CloseWithError(file, err)
	}()
	length := fileInfo.Size()
	mode := fmt.Sprintf("%#o", fileInfo.Mode())
	_, name := filepath.Split(path)
	if _, err = fmt.Fprintln(w, "C"+mode, length, name); err != nil {
		return
	}
	_, err = io.Copy(w, file)
	return
}

func CheckSSH(ctx context.Context, host string, port string, user string) bool {
	for {
		select {
		case <-time.After(time.Second):
			if aClient, err := Connect(host, port, user); err == nil {
				if err := aClient.Close(); err != nil {
					log2.Warnf("connection to host %s succeeded, but closing failed : %v", host, err)
				}
				return true
			}
		case <-ctx.Done():
			return false
		}
	}
}