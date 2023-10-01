package constants

var State = struct {
	NotExisting  string
	Up           string
	Down         string
	Pending      string
	Stopping     string
	ShuttingDown string
}{
	NotExisting:  "not existing",
	Up:           "up",
	Down:         "down",
	Pending:      "pending",
	Stopping:     "stopping",
	ShuttingDown: "shutting down",
}
