package util

import (
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/template"
)

// GitCommit set at build time
var GitCommit string

// GitRelease set at build time
var GitRelease string

const DevRelease = "0.1.0-DEV"

type Closer func() *cmd.XbeeError

func CloseWithError(close Closer, err error) *cmd.XbeeError {
	if close != nil {
		err2 := close()
		if err2 != nil && err2.Error() == "EOF" { //skip this kind of error, which is caused by server closing first.
			err2 = nil
		}
		if err2 != nil {
			if err == nil {
				return cmd.Error("cannot close : %v", err2)
			} else {
				return cmd.Error("close operation failed: %v. First error was : %v", err2, err)
			}
		}
	}
	if err == nil {
		return nil
	}
	return cmd.Error("%v", err)
}

// release: X.Y.Z for a released version, or X.Y.Z-DEV for a development version.
// commit: hash of the commit of xbee repository, on which this binary was built.

// CurrentVersion should be used by cli command version.
func CurrentVersion() (string, *cmd.XbeeError) {
	s := `
Release: {{ .Release }}
Commit: {{ .Commit }}
`
	release := GitRelease
	if release == "" {
		release = DevRelease
	}
	data := struct {
		Release string
		Commit  string
	}{
		Release: release,
		Commit:  GitCommit,
	}
	err := template.Output(&s, data, nil)
	return s, err

}
