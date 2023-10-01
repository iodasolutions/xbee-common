package util

import (
	"io"
)

func CloseWithError(s io.Closer, err error) *XbeeError {
	if s != nil {
		err2 := s.Close()
		if err2 != nil && err2.Error() == "EOF" { //skip this kind of error, which is caused by server closing first.
			err2 = nil
		}
		if err2 != nil {
			if err == nil {
				return Error("cannot close : %v", err2)
			} else {
				return Error("close operation failed: %v. First error was : %v", err2, err)
			}
		}
	}
	if err == nil {
		return nil
	}
	return Error("%v", err)
}
