package stringutils

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	uuid "github.com/nu7hatch/gouuid"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString() string {
	u4, err := uuid.NewV4()
	if err != nil {
		panic(fmt.Errorf("Unexpected error : %v", err))
	}
	return u4.String()
}

func ValidatePortableFilenameCharacterSet(s string) *cmd.XbeeError {
	for i, c := range s {
		if !isCharacterPortableFilename(c) {
			return cmd.Error("Character %c at index %d in string %s is not valid", c, i, s)
		}
	}
	return nil
}

func isCharacterPortableFilename(c rune) bool {
	return ('A' <= c && c <= 'Z') ||
		('a' <= c && c <= 'z') ||
		c == '.' || c == '_' || c == '-'
}

func ToString(o interface{}) string {
	if o == nil {
		return ""
	}
	switch x := o.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "true"
		} else {
			return "false"
		}
	default:
		return "???"
	}
}
