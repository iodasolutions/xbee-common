package cmd

type ArgsParser struct {
	options map[string]*Option
}

func NewArgsParser(options []*Option) *ArgsParser {
	optionMap := make(map[string]*Option)
	for _, option := range options {
		optionMap[option.Name] = option
	}
	return &ArgsParser{
		options: optionMap,
	}
}


func (ap *ArgsParser) ParseArgs(args ...string) (realArgs []string) {
	var skipIteration bool
	for i, arg := range args {
		if skipIteration {
			skipIteration = false
		} else {
			option := ap.optionFor(arg)
			if option != nil {
				if option.IsBool() {
					option.Enable(true)
				} else if i < len(args) - 1{
					option.AddValue(args[i+1])
					skipIteration = true
				} else { //degenerated case.
					realArgs = append(realArgs, args[i:]...)
					break
				}
			} else {
				realArgs = append(realArgs, args[i:]...)
				break
			}
		}
	}
	return
}

func (ap *ArgsParser) optionFor(arg string) *Option {
	for _, option := range ap.options {
		if option.IsArgForMe(arg) {
			return option
		}
	}
	for _, option := range globalOptions {
		if option.IsArgForMe(arg) {
			return option
		}
	}
	return nil
}
