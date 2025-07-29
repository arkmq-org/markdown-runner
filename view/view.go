// Package view provides a layer of abstraction for all UI operations,
// decoupling the core logic from the presentation layer (e.g., pterm).
package view

// The RunnerView takes care of visual feedback to the user about what's happening to the file being executed
// This interface allows for several implementation, such as one for the tests.
type RunnerView interface {
	// When running a new file
	StartFile(file string)
	EndFile(file string, err error)
	// When entering a new stage
	StartStage(stageName string, chunkCount int, verbose bool)
	// When entering a parallel section with multiple writers, you need to declare it
	DeclareParallelMode()
	// Start the parallel writers
	StartParallelMode() error
	// Quit the parallel mode to resume sequential operation
	QuitParallelMode() error
	// Gives feedback that a command has started
	StartCommand(id, text string) error
	// Prompts the user for what to do for a given command
	InteractivePromptForCommand(prompt string, commandName string, isInteractive *bool) (string, error)
	// Gives feedback that the command is in dry-run mode
	DryRunCommand(id, text string) error
	// Gives feedback that the command was skipped
	SkipCommand(id, text string) error
	// Gives feedback that the command was done
	StopCommand(id string, success bool, message string) error
	// Gives feedback that the command was killed
	KillCommand(id, text string) error
	// Info logger
	Info(message string)
	// Error logger
	Error(message string)
	// Warning logger
	Warning(message string)
	// Looks if a command was correctly declared to have its own logger
	HasLogger(id string) bool
}

// NewView returns the interface chosen by the user. The kind can be "mock" "ci" of "default"
func NewView(kind string) RunnerView {
	switch kind {
	case "mock":
		return newMockView()
	case "ci":
		return newCiView()
	default:
		return newDefaultView()
	}
}
