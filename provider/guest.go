package provider

import "github.com/iodasolutions/xbee-common/newfs"

func VmPack() newfs.Folder       { return newfs.Folder("/var/xbee/pack") }
func VmSystemPack() newfs.Folder { return newfs.Folder("/var/xbee/systempack") }
