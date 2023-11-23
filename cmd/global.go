package cmd

var globalOptions = make(map[string]*Option)

func init() {
	AddGlobalBooleanOption("help", "h", false).WithDescription("Display usage information for this command")
}
