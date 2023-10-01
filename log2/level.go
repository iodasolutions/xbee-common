package log2

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
)

type level int

const (
	DEBUG level = iota
	INFO
	WARN
	ERROR
	OFF
)

var option = cmd.NewOption("log", "l", INFO.String()).WithDescription("Control log verbosity (info,debug,off)")

func init() {
	cmd.Register(option)
}

func (l level) String() string {
	switch l {
	case ERROR:
		return "error"
	case WARN:
		return "warn"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	case OFF:
		return "off"
	default:
		panic(fmt.Sprintf("unexpected log level %d", l))
	}
}
