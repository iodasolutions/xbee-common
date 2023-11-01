package util

import (
	"github.com/iodasolutions/xbee-common/cmd"
)

var XbeeRelease = struct {
	Release string
	Commit  string
}{
	Release: "0.1.0-DEV",
	Commit:  "",
}

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
