package log2

func Debugf(format string, a ...interface{}) {
	send(DEBUG, format, a...)
}
func Infof(format string, a ...interface{}) {
	send(INFO, format, a...)
}
func Warnf(format string, a ...interface{}) {
	send(WARN, format, a...)
}
func Errorf(format string, a ...interface{}) {
	send(ERROR, format, a...)
}

func Level() string {
	return theLevel.value.String()
}

func TurnOffIfNotSet() {
	initLevel()
	if !option.IsSet() {
		theLevel.value = OFF
	}
}

func Close() {
	close(ch)
	<-chExit
}
