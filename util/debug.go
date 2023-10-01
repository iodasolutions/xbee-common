package util

import "github.com/iodasolutions/xbee-common/cmd"

var debugOption = cmd.NewBooleanOption("debug", "", false)

func init() {
	cmd.Register(debugOption)
}

func Debug() bool { return debugOption.BooleanValue() }
func WithDelve() []string {
	return []string{"--listen=:40000", "--headless=true",
		"--api-version=2", "--accept-multiclient", "--check-go-version=false", "exec", "--"}
}
