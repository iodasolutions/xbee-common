package util

import "github.com/iodasolutions/xbee-common/cmd"

func Debug() bool { return Contains(cmd.XbeeFlags, "--xbeeDebug") }
func WithDelve() []string {
	return []string{"--listen=:40000", "--headless=true",
		"--api-version=2", "--accept-multiclient", "--check-go-version=false", "exec", "--"}
}
