package newfs

import (
	"strings"
)

func (fd Folder) String() string {
	return strings.ReplaceAll(string(fd), "\\", "\\\\")
}

func (fd Folder) Owner() (uid int, gid int) {
	return -1, -1
}
