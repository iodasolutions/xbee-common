package newfs

import (
	"github.com/iodasolutions/xbee-common/cmd"
	"os"
	"syscall"
)

func (fd Folder) String() string {
	return string(fd)
}

func (fd Folder) Owner() (uid int, gid int) {
	si, err := os.Stat(string(fd))
	if err != nil {
		panic(cmd.Error("unexpected error when finding owner for folder %s : %v", fd, err))
	}
	stat := si.Sys().(*syscall.Stat_t)
	return int(stat.Uid), int(stat.Gid)
}
