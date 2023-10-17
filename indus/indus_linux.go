package indus

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/newfs"
	"runtime"
)

func copyMaybeToLocalBin(cached newfs.File, goos string, goarch string) {
	if goos == "linux" && runtime.GOARCH == goarch {
		targetDir := newfs.Home.ChildFolder("go/bin")
		fmt.Printf("local executable is %s/%s\n", targetDir, cached.Base())
		cached.CopyToDir(targetDir)
	}
}
