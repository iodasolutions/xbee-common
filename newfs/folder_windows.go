package newfs

import (
	"path/filepath"
)

func (fd Folder) String() string {
	return filepath.ToSlash(string(fd))
}

func (fd Folder) Owner() (uid int, gid int) {
	return -1, -1
}
