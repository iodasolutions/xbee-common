package log2

import (
	"fmt"
	"strings"
	"time"
)

var ch = make(chan logElt)
var chExit = make(chan bool)

func init() {
	go consumeCh()
}

func consumeCh() {
	for elt := range ch {
		fmt.Printf(elt.message)
	}
	close(chExit)
}

type logElt struct {
	message   string
	errStream bool
}

func send(l level, format string, a ...interface{}) {
	if l >= value() {
		message := fmt.Sprintf(format, a...)
		log := buildLogLine(l, message)
		ch <- logElt{
			message: log,
		}
	}
}

func buildLogLine(l level, message string) string {
	if !strings.HasSuffix(message, "\n") {
		message = message + "\n"
	}
	return fmt.Sprintf("%s : %s : %s : %s", timestamp(), owner(), strings.ToUpper(l.String()), message)
}

func timestamp() string {
	now := time.Now()
	year, month, day := now.Date()
	h, min, sec := now.Clock()
	milli := now.Nanosecond() / 1e6
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%03d", year, month, day, h, min, sec, milli)
}
