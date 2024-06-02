package main

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"log"
)

func main() {
	_, err := cmd.Setup(BuildPackCmdTree)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("xbee options=%v\n", cmd.XbeeFlags)
	fmt.Printf("args=%v\n", cmd.Args)
}

func BuildPackCmdTree(root *cmd.Command) *cmd.XbeeError {
	return nil
}
