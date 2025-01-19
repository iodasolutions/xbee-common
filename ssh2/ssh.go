package ssh2

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/exec2"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/newfs"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"strings"
)

type SSHClient struct {
	*ssh.Client
}

func Connect(host string, port string, user string) (*SSHClient, *cmd.XbeeError) {
	var aConf *ssh.ClientConfig
	rg := newfs.NewRsaGen(newfs.NewFolder(""))
	xbeeKey := rg.RootKeyPEM().Content()
	pemBytes := []byte(xbeeKey)
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, cmd.Error("ssh : parse key %s failed for host %s : %v", xbeeKey, host, err)
	}
	aConf = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	aConf.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	connexionString := fmt.Sprintf("%s:%s", host, port)
	var conn *ssh.Client
	conn, err = ssh.Dial("tcp", connexionString, aConf)
	if err != nil {
		return nil, cmd.Error("ssh : cannot connect to %s with user %s : %v", connexionString, user, err)
	}
	var aSession *ssh.Session
	if aSession, err = conn.NewSession(); err == nil {
		defer func() {
			if aSession != nil {
				if err3 := aSession.Close(); err3 != nil && err3 != io.EOF {
					log2.Warnf("An error occurred when closing session : %v", err3)
				}
			}
		}()
	} else {
		return nil, cmd.Error("ssh : cannot create session to %s : %v", connexionString, err)
	}
	return &SSHClient{conn}, nil
}

func (hr *SSHClient) RunCommand(command string) *cmd.XbeeError {
	return hr.run(command, true)
}

func (hr *SSHClient) RunCommandToOut(command string) (out string, err *cmd.XbeeError) {
	sess, err2 := hr.NewSession()
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
	w := exec2.NewMachineOnlyReadableWriter()
	we := exec2.NewMachineOnlyReadableWriter()
	sess.Stdout = w
	sess.Stderr = we
	if err2 = sess.Run(command); err2 != nil {
		err = cmd.Error("This command [%s] failed : %s", command, we.String())
	} else {
		out = w.String()
	}
	return
}
func (hr *SSHClient) RunCommandQuiet(command string) *cmd.XbeeError {
	return hr.run(command, false)
}

func (hr *SSHClient) run(command string, redirectStd bool) (err *cmd.XbeeError) {
	sess, err2 := hr.NewSession()
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

func (hr *SSHClient) RunScript(script string) *cmd.XbeeError {
	return hr.runScript(script, true)
}
func (hr *SSHClient) RunScriptQuiet(script string) *cmd.XbeeError {
	return hr.runScript(script, false)
}
func (hr *SSHClient) runScript(script string, redirectStd bool) (err *cmd.XbeeError) {
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
	err = hr.UploadFile(f, newfs.NewFolder("/tmp"))
	if err != nil {
		return
	}
	err = hr.run(fmt.Sprintf("sudo bash /tmp/%s", f.Base()), redirectStd)
	return
}

func (hr *SSHClient) UploadFile(path newfs.File, todir newfs.Folder) (err *cmd.XbeeError) {
	fileInfo, err2 := os.Stat(path.String())
	if err2 != nil {
		err = cmd.Error("cannot stat %s : %v", path, err2)
		return
	}
	var file *os.File
	file, err2 = os.Open(path.String())
	if err2 != nil {
		err = cmd.Error("cannot open %s : %v", path, err2)
	}
	defer func() {
		if file != nil {
			if err4 := file.Close(); err != nil {
				err = cmd.Error("cannot close f %s: %v", file, err4)
			}
		}
	}()
	length := fileInfo.Size()
	return hr.upload(file, length, path.Base(), todir)
}

func (hr *SSHClient) upload(r io.Reader, length int64, name string, todir newfs.Folder) (err *cmd.XbeeError) {
	if err = hr.RunCommandQuiet(fmt.Sprintf("sudo mkdir -p %s", todir)); err != nil {
		return
	}
	sess, err2 := hr.NewSession()
	if err2 != nil {
		err = cmd.Error("cannot create a session for connection %s: %v", hr.RemoteAddr().String(), err)
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
		if _, err5 := fmt.Fprintln(w, "C0644", length, name); err5 != nil {
			err = cmd.Error("unexpected error : %v", err5)
			return
		}
		_, err6 := io.Copy(w, r)
		if err6 != nil {
			err = cmd.Error("unexpected error : %v", err6)
		}
		if _, err2 := fmt.Fprint(w, "\x00"); err2 != nil {
			return
		} // transfer end with \x00

	}()
	command := fmt.Sprintf("sudo /usr/bin/scp -tr %s", todir)
	if err2 = sess.Run(command); err2 != nil {
		err = cmd.Error("command [%s] for session [%s] failed: %v", command, hr.RemoteAddr().String(), err)
	}
	return
}

func (hr *SSHClient) UploadContent(content string, path newfs.File) (err *cmd.XbeeError) {
	r := strings.NewReader(content)
	length := int64(len(content))
	return hr.upload(r, length, path.Base(), path.Dir())
}

func (hr *SSHClient) Download(remoteFile newfs.File, todir newfs.Folder) (err *cmd.XbeeError) {
	todir.EnsureExists()
	sess, err2 := hr.NewSession()
	if err2 != nil {
		return cmd.Error("cannot create a session for connection: %v", err2)
	}
	defer func() {
		if sess != nil {
			if err3 := sess.Close(); err3 != nil && err3 != io.EOF {
				log2.Warnf("An error occurred when closing session : %v", err3)
			}
		}
	}()

	localPath := todir.ChildFile(remoteFile.Base())
	var f *os.File
	f, err = localPath.OpenFileForCreation()
	if err != nil {
		return
	}
	defer f.Close()

	go func() {
		r, _ := sess.StdoutPipe()
		if err4 := downloadInternScp(r, f); err4 != nil {
			err = err4
		}
	}()

	command := fmt.Sprintf("sudo /usr/bin/scp -f %s", remoteFile)
	if err := sess.Run(command); err != nil {
		return cmd.Error("command [%s] for session failed: %v", command, err)
	}
	return nil
}

func downloadInternScp(r io.Reader, w io.Writer) (err *cmd.XbeeError) {
	buf := make([]byte, 1)
	if _, err2 := r.Read(buf); err2 != nil {
		err = cmd.Error("error reading from remote: %v", err2)
		return
	}

	for {
		_, err2 := r.Read(buf)
		if err2 != nil {
			err = cmd.Error("error reading from remote: %v", err2)
			return
		}
		if buf[0] == 'C' {
			break
		}
	}

	_, err2 := fmt.Fscanln(r)
	if err2 != nil {
		err = cmd.Error("error reading file details: %v", err2)
		return
	}

	_, err3 := io.Copy(w, r)
	if err3 != nil {
		err = cmd.Error("error copying data to local file: %v", err3)
		return err
	}

	return nil
}
