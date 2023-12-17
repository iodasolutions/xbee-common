package exec2

import (
	"context"
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/newfs"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"time"
)

type SSHClient struct {
	conn     *ssh.Client
	hostPort string
}

func Connect(host string, port string, user string) (client *SSHClient, err *cmd.XbeeError) {
	var aConf *ssh.ClientConfig
	rg := newfs.NewRsaGen("")
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
	conn, err3 := ssh.Dial("tcp", connexionString, aConf)
	if err3 != nil {
		err = cmd.Error("ssh : cannot connect to %s with user %s : %v", connexionString, user, err3)
		return
	}
	if aSession, err4 := conn.NewSession(); err4 == nil {
		defer func() {
			if aSession != nil {
				if err3 := aSession.Close(); err3 != nil && err3 != io.EOF {
					log2.Warnf("An error occurred when closing session : %v", err3)
				}
			}
		}()
		client = &SSHClient{
			conn:     conn,
			hostPort: connexionString,
		}
	}
	return
}
func (c *SSHClient) Close() *cmd.XbeeError {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			return cmd.Error("%v", err)
		}
	}
	return nil
}

func (c *SSHClient) RunCommand(command string) *cmd.XbeeError {
	return c.run(command, true)
}

func (c *SSHClient) RunCommandToOut(command string) (out string, err *cmd.XbeeError) {
	sess, err2 := c.conn.NewSession()
	if err2 != nil {
		return "", cmd.Error("cannot create session : %v", err2)
	}
	defer func() {
		if sess != nil {
			if err3 := sess.Close(); err3 != nil && err3 != io.EOF {
				log2.Warnf("An error occurred when closing session : %v", err3)
			}
		}
	}()
	w := NewMachineOnlyReadableWriter()
	we := NewMachineOnlyReadableWriter()
	sess.Stdout = w
	sess.Stderr = we
	if err2 = sess.Run(command); err2 != nil {
		err = cmd.Error("This command [%s] failed : %s", command, we.String())
	} else {
		out = w.String()
	}
	return
}
func (c *SSHClient) RunCommandQuiet(command string) *cmd.XbeeError {
	return c.run(command, false)
}

func (c *SSHClient) run(command string, redirectStd bool) (err *cmd.XbeeError) {
	sess, err2 := c.conn.NewSession()
	if err2 != nil {
		return cmd.Error("cannot create session : %v", err2)
	}
	defer func() {
		if sess != nil {
			if err3 := sess.Close(); err3 != nil && err3 != io.EOF {
				log2.Warnf("An error occurred when closing session : %v", err3)
			}
		}
	}()
	if redirectStd {
		sess.Stdout = os.Stdout
		sess.Stdin = os.Stdin
		sess.Stderr = os.Stderr
	}
	err2 = sess.Run(command)
	if err2 != nil {
		err = cmd.Error("run command in session failed : %v", err2)
	}
	return
}

func (c *SSHClient) RunScript(script string) *cmd.XbeeError {
	return c.runScript(script, true)
}
func (c *SSHClient) RunScriptQuiet(script string) *cmd.XbeeError {
	return c.runScript(script, false)
}
func (c *SSHClient) runScript(script string, redirectStd bool) (err *cmd.XbeeError) {
	f := newfs.TmpDir().RandomFile()
	defer func() {
		err2 := f.EnsureDelete()
		if err2 != nil {
			if err == nil {
				err = err2
			} else {
				err = cmd.Error("cannot delete tmp file: %v. First error was : %v", err2, err)
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

func (c *SSHClient) Upload(path newfs.File, todir newfs.Folder) (err *cmd.XbeeError) {
	if err = c.RunCommandQuiet(fmt.Sprintf("sudo mkdir -p %s", todir)); err != nil {
		return
	}
	sess, err2 := c.conn.NewSession()
	if err2 != nil {
		err = cmd.Error("cannot create a session for connection %s: %v", c.hostPort, err)
		return
	}
	defer func() {
		if sess != nil {
			if err3 := sess.Close(); err3 != nil && err3 != io.EOF {
				log2.Warnf("An error occurred when closing session : %v", err3)
			}
		}
	}()
	go func() {
		w, _ := sess.StdinPipe()
		defer func() {
			if err2 := w.Close(); err2 != nil {
				err = cmd.Error("cannot close writer: %v", err)
			}
		}()
		//		fmt.Fprintln(w, "D0755", 0, "xbee") // mkdir
		if err = uploadInternScp(string(path), w); err != nil {
			return
		}
		if _, err2 := fmt.Fprint(w, "\x00"); err2 != nil {
			return
		} // transfer end with \x00

	}()
	command := fmt.Sprintf("sudo /usr/bin/scp -tr %s", todir)
	if err := sess.Run(command); err != nil {
		return cmd.Error("command [%s] for session [%s] failed: %v", command, c.hostPort, err)
	}
	return nil
}

func uploadInternScp(path string, w io.Writer) (err *cmd.XbeeError) {
	fileInfo, err2 := os.Stat(path)
	if err2 != nil {
		err = cmd.Error("cannot stat %s : %v", path, err2)
		return
	}
	file, err3 := os.Open(path)
	if err3 != nil {
		err = cmd.Error("cannot open %s : %v", path, err3)
		return
	}
	defer func() {
		if file != nil {
			if err4 := file.Close(); err != nil {
				err = cmd.Error("cannot close f %s: %v", file, err4)
			}
		}
	}()
	length := fileInfo.Size()
	mode := fmt.Sprintf("%#o", fileInfo.Mode())
	_, name := filepath.Split(path)
	if _, err5 := fmt.Fprintln(w, "C"+mode, length, name); err != nil {
		err = cmd.Error("unexpected error : %v", err5)
		return
	}
	_, err6 := io.Copy(w, file)
	if err6 != nil {
		err = cmd.Error("unexpected error : %v", err6)
	}
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
