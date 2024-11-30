package main

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/newfs"
)

func main() {
	f := newfs.NewFile("/home/eric/xbee-tosiam/packs/maven/xbee-pack.yaml")
	fmt.Println(f.String())
}
