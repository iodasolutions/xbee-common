package newfs

import (
	"github.com/iodasolutions/xbee-common/util"
	"os"
	"syscall"
)

func (f File) Owner() (uid int, gid int) {
	si, err := os.Stat(string(f))
	if err != nil {
		panic(util.Error("unexpected error when finding owner for folder %s : %v", f, err))
	}
	stat := si.Sys().(*syscall.Stat_t)
	return int(stat.Uid), int(stat.Gid)
}
