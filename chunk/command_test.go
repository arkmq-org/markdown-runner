package chunk

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/config"
)

func TestRunningCommand_InitCommandLabel(t *testing.T) {
	cmd := exec.Command("echo", "hello")
	command := RunningCommand{Cmd: cmd}
	chunk := ExecutableChunk{Label: "Test Label"}
	command.InitCommandLabel(&chunk)
	if command.CmdPrettyName != "Test Label" {
		t.Errorf("Expected CmdPrettyName to be 'Test Label', but got '%s'", command.CmdPrettyName)
	}

	config.Verbose = true
	command.InitCommandLabel(&chunk)
	if !strings.Contains(command.CmdPrettyName, "Test Label") {
		t.Error("Expected CmdPrettyName to contain 'Test Label' in verbose mode")
	}
}

func TestRunningCommand_Execute(t *testing.T) {
	tmpDirs := make(map[string]string)
	// Happy path
	chunk := ExecutableChunk{}
	cmdStr := "echo 'hello world'"
	command, err := chunk.AddCommandToExecute(cmdStr, tmpDirs)
	err = command.Execute()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if !strings.Contains(command.Stdout, "hello world") {
		t.Errorf("Expected stdout to contain 'hello world', but got '%s'", command.Stdout)
	}

	// Failure path
	chunk = ExecutableChunk{}
	cmdStr = "nonexistent-command"
	command, err = chunk.AddCommandToExecute(cmdStr, tmpDirs)
	err = command.Start()
	if err == nil {
		t.Error("Expected an error, but got none")
	}
	err = command.Wait()
	if err == nil {
		t.Error("Expected an error, but got none")
	}
}

func TestRunningCommand_Execute_Error(t *testing.T) {
	cmd := exec.Command("nonexistent-command")
	command := RunningCommand{Cmd: cmd}
	err := command.Execute()
	if err == nil {
		t.Error("Expected an error when executing a nonexistent command, but got none")
	}
}

func TestRunningCommand_BashEnvExtraction(t *testing.T) {
	tmpDirs := make(map[string]string)
	chunk := ExecutableChunk{
		Runtime: "bash",
	}
	cmdStr := "bash -c 'echo \"### ENV ###\" && echo \"TEST_ENV_VAR=test_value\"'"
	command, err := chunk.AddCommandToExecute(cmdStr, tmpDirs)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
	command.IsBash = true
	err = command.Execute()
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
	found := false
	for _, envVar := range config.Env {
		if envVar == "TEST_ENV_VAR=test_value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find TEST_ENV_VAR in the environment, but it was not there")
	}
}
