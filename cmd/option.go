package cmd

import (
	"fmt"
	"strings"
)

func Register(option *Option) {
	if _, ok := globalOptions[option.Name]; !ok {
		globalOptions[option.Name] = option
	} else {
		panic(fmt.Sprintf("option with name %s already registered", option.Name))
	}
}
func AddGlobalBooleanOption(name string, shorthand string, defaultValue bool) (option *Option) {
	option = NewBooleanOption(name, shorthand, defaultValue)
	globalOptions[name] = option
	return
}

func AddGlobalOption(name string, shorthand string, defaultValue string) (option *Option) {
	option = NewOption(name, shorthand, defaultValue)
	globalOptions[name] = option
	return
}

func GlobalOption(name string) *Option {
	return globalOptions[name]
}

func NewOption(name string, shorthand string, defaultValue string) *Option {
	return &Option{
		Name:               name,
		ShortHand:          shorthand,
		defaultStringValue: defaultValue,
	}
}
func NewBooleanOption(name string, shorthand string, defaultValue bool) *Option {
	return &Option{
		Name:         name,
		ShortHand:    shorthand,
		isBool:       true,
		booleanValue: defaultValue,
	}
}

type Option struct {
	ShortHand          string //one-letter
	Name               string
	isBool             bool
	Description        string
	booleanValue       bool //set at parse time
	defaultStringValue string
	stringValues       []string //set at parse time
	isSet              bool
}

func (op *Option) BooleanValue() bool {
	return op.booleanValue
}
func (op *Option) computeShortNameNameLength() int {
	return 5 + len(op.Name)
}
func (op *Option) StringValue() string {
	if len(op.stringValues) > 0 {
		return op.stringValues[0]
	}
	return op.defaultStringValue
}
func (op *Option) StringValues() []string {
	if op == nil { // nil receiver
		return nil
	}
	switch {
	case len(op.stringValues) > 0:
		return op.stringValues
	case op.defaultStringValue != "":
		return []string{op.defaultStringValue}
	default:
		return nil
	}

}
func (op *Option) IsArgForMe(arg string) bool {
	if op.ShortHand != "" && arg == fmt.Sprintf("-%s", op.ShortHand) {
		return true
	}
	return arg == fmt.Sprintf("--%s", op.Name)
}
func (op *Option) IsSet() bool {
	return op.isSet
}
func (op *Option) IsBool() bool {
	return op.isBool
}
func (op *Option) WithDescription(description string) *Option {
	op.Description = description
	return op
}

func (op *Option) Enable(enable bool) {
	op.booleanValue = enable
	op.isSet = true
}
func (op *Option) AddValue(value string) {
	op.stringValues = append(op.stringValues, value)
	op.isSet = true
}

func (op *Option) SetStringValue(newValue string) *XbeeError {
	if op.isBool {
		return Error("cannot set string value to boolean option %s\n", op.Name)
	}
	op.stringValues = []string{newValue}
	return nil
}
func (op *Option) SetBooleanValue(newValue bool) *XbeeError {
	if !op.isBool {
		return Error("cannot set boolean value to option %s", op.Name)
	}
	op.booleanValue = newValue
	op.isSet = true
	return nil
}

func AsMap(options []*Option) map[string]*Option {
	optionMap := make(map[string]*Option)
	for _, option := range options {
		optionMap[option.Name] = option
	}
	return optionMap
}

func (op *Option) SplitColon() (map[string]string, *XbeeError) {
	result := map[string]string{}
	for _, elt := range op.StringValues() {
		splitted := strings.Split(elt, ":")
		if len(splitted) != 2 {
			return nil, Error("option %s MUST have format X:Y, actual is %s", op.Name, elt)
		}
		result[splitted[1]] = splitted[0]
	}
	return result, nil
}

func displayOptions(options []*Option) (result string) {
	columnIndexForDescription := computeShortHandNameLength(options) + 10
	for _, option := range options {
		s := "\n   "
		if option.ShortHand != "" {
			s = "\n -" + option.ShortHand
		}
		s += " --" + option.Name
		nbSpace := columnIndexForDescription - option.computeShortNameNameLength()
		for i := 0; i < nbSpace; i++ {
			s += " "
		}
		s += option.Description
		result += s
	}
	return
}
func computeShortHandNameLength(options []*Option) int {
	max := 0
	for _, option := range options {
		length := option.computeShortNameNameLength()
		if length > max {
			max = length
		}
	}
	return max
}
