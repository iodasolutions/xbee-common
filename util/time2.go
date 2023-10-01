package util

import (
	"fmt"
	"time"
)

type Duration struct {
	start time.Time
}

func StartDuration() Duration {
	return Duration{
		start: time.Now(),
	}
}

func (d *Duration) End(format string, data ...interface{}) {
	var message string
	if len(data) > 0 {
		message = fmt.Sprintf(format, data...)
	} else {
		message = format
	}
	fmt.Printf("%s run in %d micros\n", message, time.Now().Sub(d.start).Microseconds())
}
