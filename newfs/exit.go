package newfs

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
	"os"
)

func DoExitOnError(err *cmd.XbeeError) {
	if err == nil {
		panic("DoExitOnError : err param cannot be nil")
	}
	fmt.Printf("ERROR: %v\n", err)
	log2.Close()
	os.Exit(1)
}
