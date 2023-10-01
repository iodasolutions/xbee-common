package cmd

import (
	"fmt"
	"strings"
)

func NewForceOption() *Option { return NewBooleanOption("force", "f", false) }

func Force() bool {
	if HasOption("force") {
		return OptionFrom("force").BooleanValue()
	}
	return false
}

func Confirm(message string) bool {
	if !Force() {
		for {
			var shouldDoS string
			fmt.Printf("Confirm %s ? [y,n]: ", message)
			if _, err := fmt.Scanln(&shouldDoS); err != nil {
				panic(fmt.Errorf("unexpected error while typing : %v", err))
			}
			if strings.ToLower(shouldDoS) == "y" {
				return true
			} else if strings.ToLower(shouldDoS) == "n" {
				return false
			} else {
				fmt.Println("Sorry, i do not understand")
			}
		}
	}
	return true
}
