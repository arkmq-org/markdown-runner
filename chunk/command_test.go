package chunk

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestRunningCommand(t *testing.T) {
	t.Run("init command label", func(t *testing.T) {
		cmd := exec.Command("echo", "hello")
		cfg := &config.Config{}
		command := RunningCommand{Cmd: cmd, Cfg: cfg}
		chunk := ExecutableChunk{Label: "Test Label"}
		command.InitCommandLabel(&chunk)
		if command.CmdPrettyName != "Test Label" {
			t.Errorf("Expected CmdPrettyName to be 'Test Label', but got '%s'", command.CmdPrettyName)
		}

		command.Cfg.Verbose = true
		command.InitCommandLabel(&chunk)
		if !strings.Contains(command.CmdPrettyName, "Test Label") {
			t.Error("Expected CmdPrettyName to contain 'Test Label' in verbose mode")
		}
	})

	t.Run("execute", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		cfg := &config.Config{MinutesToTimeout: 1}
		// Happy path
		chunk := ExecutableChunk{Cfg: cfg}
		cmdStr := "echo 'hello world'"
		command, err := chunk.AddCommandToExecute(cmdStr, tmpDirs)
		err = command.InitializeSequentialSpinner()
		assert.NoError(t, err, "Expected no error when initializing spinner")
		err = command.Execute()
		assert.NoError(t, err, "Expected no error when executing command")
		assert.NotEmpty(t, command.Stdout, "Expected command to produce output")
		assert.Equal(t, "hello world\n", command.Stdout, "Expected command output to be 'hello world'")
		assert.Equal(t, 0, command.ReturnCode, "Expected command to return exit code 0")

		// Failure path
		chunk = ExecutableChunk{Cfg: cfg}
		cmdStr = "nonexistent-command"
		command, err = chunk.AddCommandToExecute(cmdStr, tmpDirs)
		err = command.InitializeSequentialSpinner()
		assert.NoError(t, err, "Expected no error when initializing spinner")
		err = command.Start()
		assert.Error(t, err, "Expected an error when starting a nonexistent command")
		err = command.Wait()
		assert.Error(t, err, "Expected an error when waiting for a nonexistent command")
	})

	t.Run("execute error missing spinner", func(t *testing.T) {
		cmd := exec.Command("nonexistent-command")
		cfg := &config.Config{MinutesToTimeout: 1}
		command := RunningCommand{Cmd: cmd, Cfg: cfg}
		err := command.Execute()
		assert.Error(t, err, "Expected an error when executing a command without a spinner initialized")
	})

	t.Run("bash env extraction", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		cfg := &config.Config{MinutesToTimeout: 1}
		chunk := ExecutableChunk{
			Runtime: "bash",
			Cfg:     cfg,
		}
		cmdStr := "bash -c 'echo \"### ENV ###\" && echo \"TEST_ENV_VAR=test_value\"'"
		command, err := chunk.AddCommandToExecute(cmdStr, tmpDirs)
		assert.NoError(t, err, "Expected no error when adding command to execute")
		command.IsBash = true
		err = command.InitializeSequentialSpinner()
		assert.NoError(t, err, "Expected no error when initializing spinner")
		err = command.Execute()
		assert.NoError(t, err, "Expected no error when executing bash command")
		found := slices.ContainsFunc(command.Cfg.Env, func(env string) bool {
			return env == "TEST_ENV_VAR=test_value"
		})
		assert.True(t, found, "Expected to find TEST_ENV_VAR in the environment variables")
	})

	t.Run("kill", func(t *testing.T) {
		cmd := exec.Command("sleep", "1")
		runningCmd := &RunningCommand{
			Cmd: cmd,
			Cfg: &config.Config{MinutesToTimeout: 1},
		}
		err := runningCmd.InitializeSequentialSpinner()
		assert.NoError(t, err)
		err = runningCmd.Start()
		assert.NoError(t, err)
		err = runningCmd.Kill()
		assert.NoError(t, err)
	})

	t.Run("start error", func(t *testing.T) {
		runningCmd := &RunningCommand{
			Cfg: &config.Config{MinutesToTimeout: 1},
		}
		err := runningCmd.Start()
		assert.Error(t, err, "Expected an error when starting a command without a spinner")
	})

	t.Run("interactive should not throw an error when the value is valid", func(t *testing.T) {
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		runningCmd := &RunningCommand{
			Cfg: &config.Config{MinutesToTimeout: 1, Interactive: true},
			Cmd: exec.Command("echo", "hello"),
			GetUserInput: func(string) (string, error) {
				return "yes", nil
			},
		}
		err := runningCmd.InitializeSequentialSpinner()
		assert.NoError(t, err)
	})

	t.Run("interactive should throw an error when the user enters something wrong", func(t *testing.T) {
		runningCmd := &RunningCommand{
			Cfg: &config.Config{MinutesToTimeout: 1, Interactive: true},
			Cmd: exec.Command("echo", "hello"),
			GetUserInput: func(string) (string, error) {
				return "", assert.AnError
			},
		}
		err := runningCmd.InitializeSequentialSpinner()
		assert.Error(t, err)
	})

	t.Run("initialize parallel spiner error", func(t *testing.T) {
		runningCmd := &RunningCommand{
			Cfg: &config.Config{MinutesToTimeout: 1},
		}
		err := runningCmd.InitializeParallelSpiner(nil)
		assert.Error(t, err, "Expected an error when initializing a parallel spinner without a multi-printer")
	})

	t.Run("initialize parallel spiner nil spinners", func(t *testing.T) {
		t.Setenv("CI", "true")
		runningCmd := &RunningCommand{
			Cfg: &config.Config{MinutesToTimeout: 1},
		}
		err := runningCmd.InitializeParallelSpiner(&pterm.DefaultMultiPrinter)
		assert.NoError(t, err)
	})

	t.Run("wait dry run", func(t *testing.T) {
		runningCmd := &RunningCommand{
			Cfg: &config.Config{MinutesToTimeout: 1, DryRun: true},
			Cmd: exec.Command("echo", "hello"),
		}
		err := runningCmd.InitializeSequentialSpinner()
		assert.NoError(t, err)
		err = runningCmd.Wait()
		assert.NoError(t, err, "Expected no error in dry run mode")
	})

	t.Run("wait error", func(t *testing.T) {
		chunk := ExecutableChunk{
			Runtime: "bash",
			Cfg:     &config.Config{MinutesToTimeout: 1},
		}
		runningCmd, err := chunk.AddCommandToExecute("false", make(map[string]string))
		assert.NoError(t, err, "Expected no error when adding command to execute")
		err = runningCmd.InitializeSequentialSpinner()
		assert.NoError(t, err)
		err = runningCmd.Start()
		assert.NoError(t, err)
		err = runningCmd.Wait()
		assert.Error(t, err, "Expected an error when waiting for a failing command")
	})
}
