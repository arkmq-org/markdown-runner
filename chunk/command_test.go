package chunk

import (
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/runnercontext"
	"github.com/arkmq-org/markdown-runner/view"
	"github.com/stretchr/testify/assert"
)

func TestRunningCommand(t *testing.T) {
	t.Run("init command label", func(t *testing.T) {
		cmd := exec.Command("echo", "hello")
		cfg := &config.Config{}
		ctx := &runnercontext.Context{Cfg: cfg}
		command := RunningCommand{Cmd: cmd, Ctx: ctx}
		chunk := ExecutableChunk{Label: "Test Label"}
		command.InitCommandLabel(&chunk)
		if command.CmdPrettyName != "Test Label" {
			t.Errorf("Expected CmdPrettyName to be 'Test Label', but got '%s'", command.CmdPrettyName)
		}

		command.Ctx.Cfg.Verbose = true
		command.InitCommandLabel(&chunk)
		if !strings.Contains(command.CmdPrettyName, "Test Label") {
			t.Error("Expected CmdPrettyName to contain 'Test Label' in verbose mode")
		}
	})

	t.Run("execute", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		cfg := &config.Config{MinutesToTimeout: 1}
		// Happy path
		ui := view.NewView("mock")
		chunk := ExecutableChunk{
			Context: &runnercontext.Context{
				Cfg:   cfg,
				RView: ui,
			},
		}
		cmdStr := "echo 'hello world'"
		command, err := chunk.AddCommandToExecute(cmdStr, tmpDirs)
		err = command.Execute()
		assert.NoError(t, err, "Expected no error when executing command")
		assert.NotEmpty(t, command.Stdout, "Expected command to produce output")
		assert.Equal(t, "hello world\n", command.Stdout, "Expected command output to be 'hello world'")
		assert.Equal(t, 0, command.ReturnCode, "Expected command to return exit code 0")

		// Failure path
		chunk = ExecutableChunk{
			Context: &runnercontext.Context{
				Cfg:   cfg,
				RView: ui,
			},
		}
		cmdStr = "nonexistent-command"
		command, err = chunk.AddCommandToExecute(cmdStr, tmpDirs)
		err = command.InitializeLogger()
		assert.NoError(t, err, "Expected no error when initializing spinner")
		err = command.Start()
		assert.Error(t, err, "Expected an error when starting a nonexistent command")
		err = command.Wait()
		assert.Error(t, err, "Expected an error when waiting for a nonexistent command")
	})

	t.Run("execute error missing spinner", func(t *testing.T) {
		cmd := exec.Command("nonexistent-command")
		cfg := &config.Config{MinutesToTimeout: 1}
		ctx := &runnercontext.Context{Cfg: cfg, RView: view.NewView("mock")}
		command := RunningCommand{Cmd: cmd, Ctx: ctx}
		err := command.Execute()
		assert.Error(t, err, "Expected an error when executing a command without a spinner initialized")
	})

	t.Run("bash env extraction", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		cfg := &config.Config{MinutesToTimeout: 1}
		ui := view.NewView("mock")
		chunk := ExecutableChunk{
			Runtime: "bash",
			Context: &runnercontext.Context{
				Cfg:   cfg,
				RView: ui,
			},
		}
		cmdStr := "bash -c 'echo \"### ENV ###\" && echo \"TEST_ENV_VAR=test_value\"'"
		command, err := chunk.AddCommandToExecute(cmdStr, tmpDirs)
		assert.NoError(t, err, "Expected no error when adding command to execute")
		command.IsBash = true
		err = command.Execute()
		assert.NoError(t, err, "Expected no error when executing bash command")
		found := slices.ContainsFunc(command.Ctx.Cfg.Env, func(env string) bool {
			return env == "TEST_ENV_VAR=test_value"
		})
		assert.True(t, found, "Expected to find TEST_ENV_VAR in the environment variables")
	})

	t.Run("kill", func(t *testing.T) {
		cmd := exec.Command("sleep", "1")
		ui := view.NewView("mock")
		runningCmd := &RunningCommand{
			Cmd: cmd,
			Ctx: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: ui,
			},
			id: "test-kill",
		}
		err := runningCmd.InitializeLogger()
		assert.NoError(t, err)
		err = runningCmd.Start()
		assert.NoError(t, err)
		err = runningCmd.Kill()
		assert.NoError(t, err)
	})

	t.Run("start error", func(t *testing.T) {
		ui := view.NewView("mock")
		runningCmd := &RunningCommand{
			Ctx: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: ui,
			},
		}
		err := runningCmd.Start()
		assert.Error(t, err, "Expected an error when starting a command without a spinner")
	})

	t.Run("interactive should not throw an error when the value is valid", func(t *testing.T) {
		ui := view.NewView("mock")
		runningCmd := &RunningCommand{
			Ctx: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1, Interactive: true},
				RView: ui,
			},
			Cmd: exec.Command("echo", "hello"),
			GetUserInput: func(string) (string, error) {
				return "yes", nil
			},
		}
		err := runningCmd.InitializeLogger()
		assert.NoError(t, err)
	})

	t.Run("interactive should throw an error when the user enters something wrong", func(t *testing.T) {
		ui := view.NewView("mock")
		runningCmd := &RunningCommand{
			Ctx: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1, Interactive: true},
				RView: ui,
			},
			Cmd: exec.Command("echo", "hello"),
			GetUserInput: func(string) (string, error) {
				return "", assert.AnError
			},
		}
		err := runningCmd.InitializeLogger()
		assert.Error(t, err)
	})

	t.Run("initialize parallel spiner error", func(t *testing.T) {
		ui := view.NewView("mock")
		runningCmd := &RunningCommand{
			Ctx: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: ui,
			},
		}
		runningCmd.Ctx.RView.DeclareParallelMode()
		err := runningCmd.InitializeLogger()
		assert.NoError(t, err)
		runningCmd.Ctx.RView.QuitParallelMode()
	})

	t.Run("initialize parallel spiner nil spinners", func(t *testing.T) {
		ui := view.NewView("mock")
		runningCmd := &RunningCommand{
			Ctx: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: ui,
			},
		}
		err := runningCmd.InitializeLogger()
		assert.NoError(t, err)
	})

	t.Run("wait dry run", func(t *testing.T) {
		ui := view.NewView("mock")
		runningCmd := &RunningCommand{
			Ctx: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1, DryRun: true},
				RView: ui,
			},
			Cmd: exec.Command("echo", "hello"),
			id:  "test-dry-run",
		}
		err := runningCmd.InitializeLogger()
		assert.NoError(t, err)
		err = runningCmd.Wait()
		assert.NoError(t, err, "Expected no error in dry run mode")
	})

	t.Run("wait error", func(t *testing.T) {
		ui := view.NewView("mock")
		chunk := ExecutableChunk{
			Runtime: "bash",
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: ui,
			},
		}
		runningCmd, err := chunk.AddCommandToExecute("false", make(map[string]string))
		assert.NoError(t, err, "Expected no error when adding command to execute")
		err = runningCmd.InitializeLogger()
		assert.NoError(t, err)
		err = runningCmd.Start()
		assert.NoError(t, err)
		err = runningCmd.Wait()
		assert.Error(t, err, "Expected an error when waiting for a failing command")
	})
}
