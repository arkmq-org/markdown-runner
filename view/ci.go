// Package view provides a layer of abstraction for all UI operations,
// decoupling the core logic from the presentation layer (e.g., pterm).
package view

import (
	"fmt"

	"github.com/pterm/pterm"
)

// CiView is a view for CI environments, with no spinners and simplified outptut
// it ignores the interactive mode completely and prints only the output of the crashing command (if the case may arise)
type CiView struct {
	hasPrintedResult bool
}

// newCiView returns a new CiView
func newCiView() RunnerView {
	return &CiView{}
}

// StartFile implements RunnerView
func (v *CiView) StartFile(file string) {
	pterm.Info.Printf("%s:", file)
}

func (v *CiView) EndFile(file string, err error) {
	if v.hasPrintedResult {
		return
	}
	if err != nil {
		fmt.Println("❌")
	} else {
		fmt.Println("✅")
	}
}

// StartStage implements RunnerView
func (v *CiView) StartStage(stageName string, chunkCount int, verbose bool) {
	if verbose {
		pterm.Info.Printf("Stage %s with %d chunks\n", stageName, chunkCount)
	}
}

// DeclareParallelMode implements RunnerView
func (v *CiView) DeclareParallelMode() {}

// StartParallelMode implements RunnerView
func (v *CiView) StartParallelMode() error {
	return nil
}

// QuitParallelMode implements RunnerView
func (v *CiView) QuitParallelMode() error {
	return nil
}

// StartCommand implements RunnerView
func (v *CiView) StartCommand(id, text string) error {
	return nil
}

// InteractivePromptForCommand implements RunnerView
func (v *CiView) InteractivePromptForCommand(prompt string, commandName string, isInteractive *bool) (string, error) {
	*isInteractive = false
	return "y", nil
}

// DryRunCommand implements RunnerView
func (v *CiView) DryRunCommand(id, text string) error {
	return nil
}

// SkipCommand implements RunnerView
func (v *CiView) SkipCommand(id, text string) error {
	return nil
}

// StopCommand implements RunnerView
func (v *CiView) StopCommand(id string, success bool, message string) error {
	if !success {
		if !v.hasPrintedResult {
			fmt.Println("❌")
			v.hasPrintedResult = true
		}
		pterm.Error.Println(id, message)
	}
	return nil
}

// KillCommand implements RunnerView
func (v *CiView) KillCommand(id, text string) error {
	return nil
}

// Info implements RunnerView
func (v *CiView) Info(message string) {
}

// Error implements RunnerView
func (v *CiView) Error(message string) {
	if !v.hasPrintedResult {
		fmt.Println("❌")
		v.hasPrintedResult = true
	}
	pterm.Error.Println(message)
}

// Warning implements RunnerView
func (v *CiView) Warning(message string) {
}

// HasLogger implements RunnerView
func (v *CiView) HasLogger(id string) bool {
	return true
}
