package newfs

import (
	"fmt"
	"io/fs"
	"time"
)

func preserveOwner(path string, srcInfo fs.FileInfo, strictOwner bool) error {
	return fmt.Errorf("preserveOwner not supported on Windows")
}

func fileTimes(info fs.FileInfo) (atime time.Time, mtime time.Time) {
	// not supported
	return info.ModTime(), info.ModTime()
}

func preserveSymlinkOwner(path string, srcInfo fs.FileInfo, strictOwner bool) error {
	return fmt.Errorf("preserveSymlinkOwner not supported on Windows")
}
