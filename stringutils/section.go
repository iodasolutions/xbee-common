package stringutils

import (
	"fmt"
	"strings"
)

const (
	xbeeToken = "#### %s XBEE AREA ####\n"
)

var xbeeStart = fmt.Sprintf(xbeeToken, "BEGIN") + "#Do not edit in this AREA\n"
var xbeeEnd = fmt.Sprintf(xbeeToken, "END")

func ExtractSection(content string) (xbeeSection string, otherSection string, index int, err error) {
	indexStart := strings.Index(content, xbeeStart)
	indexEnd := strings.Index(content, xbeeEnd)
	if indexStart == -1 && indexEnd == -1 {
		otherSection = content
		index = -1
	} else if indexStart != -1 && indexEnd != -1 {
		index = indexStart
		xbeeSection = content[indexStart+len(xbeeStart) : indexEnd]
		otherSection = content[:indexStart] + content[indexEnd+len(xbeeEnd):]
	} else {
		err = fmt.Errorf("content %s has an unbounded XBEE AREA", content)
	}
	return
}

func InsertSection(content string, section string, index int) string {
	xbeeSection := xbeeStart + section + xbeeEnd
	if index > -1 {
		return content[:index] + xbeeSection + content[index:]
	} else {
		return content + xbeeSection
	}
}
