package util

import (
	"regexp"
	"strconv"
)

var suffixNumberRegexp = regexp.MustCompile(`^(.*?)(\d+)$`)

func SplitSuffixNumber(value string) (prefix string, number int, hasNumber bool) {

	matches := suffixNumberRegexp.FindStringSubmatch(value)
	if matches == nil {
		return value, 0, false
	}

	number, _ = strconv.Atoi(matches[2])

	return matches[1], number, true
}
