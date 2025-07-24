package chunk

import (
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/stretchr/testify/assert"
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
	err = command.InitializeSpiner()
	assert.NoError(t, err, "Expected no error when initializing spinner")
	err = command.Execute()
	assert.NoError(t, err, "Expected no error when executing command")
	assert.NotEmpty(t, command.Stdout, "Expected command to produce output")
	assert.Equal(t, "hello world\n", command.Stdout, "Expected command output to be 'hello world'")
	assert.Equal(t, 0, command.ReturnCode, "Expected command to return exit code 0")

	// Failure path
	chunk = ExecutableChunk{}
	cmdStr = "nonexistent-command"
	command, err = chunk.AddCommandToExecute(cmdStr, tmpDirs)
	err = command.InitializeSpiner()
	assert.NoError(t, err, "Expected no error when initializing spinner")
	err = command.Start()
	assert.Error(t, err, "Expected an error when starting a nonexistent command")
	err = command.Wait()
	assert.Error(t, err, "Expected an error when waiting for a nonexistent command")
}

func TestRunningCommand_Execute_Error_Missing_Spinner(t *testing.T) {
	cmd := exec.Command("nonexistent-command")
	command := RunningCommand{Cmd: cmd}
	err := command.Execute()
	assert.Error(t, err, "Expected an error when executing a command without a spinner initialized")
}

func TestRunningCommand_BashEnvExtraction(t *testing.T) {
	tmpDirs := make(map[string]string)
	chunk := ExecutableChunk{
		Runtime: "bash",
	}
	cmdStr := "bash -c 'echo \"### ENV ###\" && echo \"TEST_ENV_VAR=test_value\"'"
	command, err := chunk.AddCommandToExecute(cmdStr, tmpDirs)
	assert.NoError(t, err, "Expected no error when adding command to execute")
	command.IsBash = true
	err = command.InitializeSpiner()
	assert.NoError(t, err, "Expected no error when initializing spinner")
	err = command.Execute()
	assert.NoError(t, err, "Expected no error when executing bash command")
	found := slices.ContainsFunc(config.Env, func(env string) bool {
		return env == "TEST_ENV_VAR=test_value"
	})
	assert.True(t, found, "Expected to find TEST_ENV_VAR in the environment variables")
}
