package log2

import (
	"fmt"
	"os"
	"path"
	"sync"
)

var theLevel struct {
	value level
	owner string
	once  sync.Once
}

func initLevel() {
	theLevel.once.Do(func() {
		theLevel.owner = path.Base(os.Args[0])
		s := option.StringValue()
		switch s {
		case "error":
			theLevel.value = ERROR
		case "warn":
			theLevel.value = WARN
		case "info":
			theLevel.value = INFO
		case "debug":
			theLevel.value = DEBUG
		case "off":
			theLevel.value = OFF
		default:
			panic(fmt.Sprintf("unexpected log level %s", s))
		}
	})
}
func owner() string {
	initLevel()
	return theLevel.owner
}
func value() level {
	initLevel()
	return theLevel.value
}
