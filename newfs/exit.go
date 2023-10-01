package newfs

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/log2"
	"os"
)

func DoExitOnError(err error) {
	if err == nil {
		panic("DoExitOnError : err param cannot be nil")
	}
	fmt.Printf("ERROR: %v\n", err)
	log2.Close()
	DeleteTmp()
	os.Exit(1)
}
