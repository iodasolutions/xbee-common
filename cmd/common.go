package cmd

func UpdateOption() *Option { return NewBooleanOption("update", "u", false) }
func Update() bool {
	if HasOption("update") {
		return OptionFrom("update").BooleanValue()
	}
	return false
}
