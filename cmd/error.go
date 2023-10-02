package cmd

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

var stackOption = NewBooleanOption("stack", "", true).WithDescription("If an unexpected error occurs, turn on this flag to display stack trace")

func init() {
	Register(stackOption)
}

func Error(format string, args ...interface{}) *XbeeError {
	var stack string
	if IsStackEnabled() {
		stack = generateStack()
	}
	return &XbeeError{
		message:            fmt.Sprintf(format, args...),
		stack:              stack,
		skipFirstLineCount: 5,
	}
}

func CauseBy(origins ...*XbeeError) *XbeeError {
	var stack string
	if IsStackEnabled() {
		stack = generateStack()
		for _, origin := range origins {
			origin.SkipLast(5)
		}
	}
	return &XbeeError{
		stack:              stack,
		origins:            origins,
		skipFirstLineCount: 5,
	}
}

func generateStack() string {
	stackSlice := make([]byte, 8192)
	n := runtime.Stack(stackSlice, false)
	stack := string(stackSlice[:n])
	return stack
}

type XbeeError struct {
	code               int
	message            string
	stack              string
	skipFirstLineCount int
	skipLastLineCount  int
	origins            []*XbeeError
}

func (xe *XbeeError) SkipFirst(lineCount int) *XbeeError {
	if lineCount > 0 {
		xe.skipFirstLineCount = lineCount
	}
	return xe
}
func (xe *XbeeError) SkipLast(lineCount int) {
	if lineCount > 0 {
		xe.skipLastLineCount = lineCount
	}
}

func (xe *XbeeError) Code() int {
	return xe.code
}

func (xe *XbeeError) Error() string {
	var buf bytes.Buffer
	if xe.message != "" {
		buf.WriteString("message=")
		buf.WriteString(xe.message)
		buf.WriteByte('\n')
	}
	if xe.stack != "" {
		buf.WriteString("stack=[\n")
		buf.WriteString(xe.trimStack())
		buf.WriteString("]\n")
	}
	if len(xe.origins) > 0 {
		buf.WriteString("CAUSED BY\n")
	}
	for _, origin := range xe.origins {
		buf.WriteString(origin.Error())
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (xe *XbeeError) trimStack() string {
	if xe.skipFirstLineCount == 0 && xe.skipLastLineCount == 0 {
		return xe.stack
	}
	lines := strings.Split(xe.stack, "\n")

	// Vérifier si le nombre de lignes à supprimer est supérieur au nombre total de lignes
	if xe.skipFirstLineCount >= len(lines) {
		return ""
	}

	// Supprimer les premières lignes
	lines = lines[xe.skipFirstLineCount:]

	if xe.skipLastLineCount >= len(lines) {
		return ""
	}
	if xe.skipLastLineCount > 0 {
		lines = lines[:len(lines)-xe.skipLastLineCount]
	}
	// Rejoindre les lignes restantes en une seule chaîne de caractères
	output := strings.Join(lines, "\n")

	return output
}

func IsStackEnabled() bool { return stackOption.BooleanValue() }
