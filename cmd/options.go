package cmd

import (
	"sync"
)

var options struct {
	Map  map[string]*Option
	once sync.Once
}

func initOptions() {
	options.once.Do(func() {
		options.Map = make(map[string]*Option)
		if leaf != nil {
			for _, option := range leaf.Options {
				options.Map[option.Name] = option
			}
		}
	})
}

func HasOption(key string) bool {
	initOptions()
	_, ok := options.Map[key]
	return ok
}

func OptionFrom(key string) *Option {
	initOptions()
	return options.Map[key]
}
