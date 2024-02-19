package provider

import "github.com/iodasolutions/xbee-common/newfs"

var XbeeVar = newfs.Folder("/var/xbee")

func XbeePath() string { return "/usr/bin/xbee" }
