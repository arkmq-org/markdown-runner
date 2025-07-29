// Package view provides a layer of abstraction for all UI operations,
// decoupling the core logic from the presentation layer (e.g., pterm).
package view

// MockRunnerView provides a mock implementation of the UI interface for testing.
type MockRunnerView struct {
	Calls    map[string][][]any
	Spinners map[string]string
}

// newMockView returns a new MockUI instance.
func newMockView() *MockRunnerView {
	return &MockRunnerView{
		Calls:    make(map[string][][]any),
		Spinners: make(map[string]string),
	}
}

func (m *MockRunnerView) logCall(name string, args ...any) {
	m.Calls[name] = append(m.Calls[name], args)
}

func (m *MockRunnerView) StartFile(file string) {
	m.logCall("StartRun", file)
}

func (v *MockRunnerView) EndFile(file string, err error) {
}

func (m *MockRunnerView) StartStage(stageName string, chunkCount int, verbose bool) {
	m.logCall("StartStage", stageName, chunkCount, verbose)
}

func (m *MockRunnerView) Warning(message string) {
	m.logCall("Warning", message)
}

func (m *MockRunnerView) Error(message string) {
	m.logCall("Error", message)
}

func (m *MockRunnerView) Info(message string) {
	m.logCall("Info", message)
}

func (m *MockRunnerView) InteractivePromptForCommand(prompt string, commandName string, isInteractive *bool) (string, error) {
	m.logCall("InteractivePrompt", prompt, commandName, isInteractive)
	if isInteractive != nil && *isInteractive {
		return "yes", nil
	}
	return "", nil
}

func (m *MockRunnerView) StartCommand(id, text string) error {
	m.logCall("StartSpinner", id, text)
	m.Spinners[id] = "ok"
	return nil
}

func (m *MockRunnerView) StopCommand(id string, success bool, message string) error {
	m.logCall("StopSpinner", id, success, message)
	delete(m.Spinners, id)
	return nil
}

// HasLogger checks if a mock logger exists.
func (m *MockRunnerView) HasLogger(id string) bool {
	m.logCall("HasSpinner", id)
	_, ok := m.Spinners[id]
	return ok
}

func (m *MockRunnerView) DryRunCommand(id, text string) error {
	m.logCall("DryRun", id, text)
	return nil
}

func (m *MockRunnerView) SkipCommand(id, text string) error {
	m.logCall("Skipped", id, text)
	return nil
}

func (m *MockRunnerView) KillCommand(id, text string) error {
	m.logCall("Killed", id, text)
	return nil
}

func (m *MockRunnerView) DeclareParallelMode() {
	m.logCall("DeclareParallel")
}

func (m *MockRunnerView) StartParallelMode() error {
	m.logCall("StartParallel")
	return nil
}

func (m *MockRunnerView) QuitParallelMode() error {
	m.logCall("QuitParallel")
	return nil
}
