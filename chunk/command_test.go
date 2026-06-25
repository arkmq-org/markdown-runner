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

	t.Run("bash env export preserves original environment", func(t *testing.T) {
		tmpDirs := make(map[string]string)

		// Set up initial environment with some original variables
		originalEnv := []string{
			"ORIGINAL_VAR=original_value",
			"PATH=/usr/bin:/bin",
			"HOME=/home/test",
			"WORKING_DIR=/test/dir",
		}

		cfg := &config.Config{MinutesToTimeout: 1, Env: originalEnv}
		ui := view.NewView("mock")

		// Export a new variable using the real bash chunk approach
		chunk := ExecutableChunk{
			Runtime: "bash",
			Content: []string{"export NEW_VAR=new_value"},
			Context: &runnercontext.Context{
				Cfg:   cfg,
				RView: ui,
			},
		}
		err := chunk.PrepareForExecution(tmpDirs)
		assert.NoError(t, err, "Expected no error when preparing chunk for execution")

		err = chunk.ExecuteSequential()
		assert.NoError(t, err, "Expected no error when executing bash chunk")

		// Verify original environment variables are preserved
		foundOriginal := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return env == "ORIGINAL_VAR=original_value"
		})
		assert.True(t, foundOriginal, "Expected ORIGINAL_VAR to be preserved")

		foundPath := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return strings.HasPrefix(env, "PATH=")
		})
		assert.True(t, foundPath, "Expected PATH to be preserved")

		foundHome := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return env == "HOME=/home/test"
		})
		assert.True(t, foundHome, "Expected HOME to be preserved")

		foundWorkingDir := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return env == "WORKING_DIR=/test/dir"
		})
		assert.True(t, foundWorkingDir, "Expected WORKING_DIR to be preserved")

		// Verify new variable was added
		foundNew := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return strings.HasPrefix(env, "NEW_VAR=")
		})
		assert.True(t, foundNew, "Expected NEW_VAR to be added")
	})

	t.Run("bash env unset removes variables selectively", func(t *testing.T) {
		tmpDirs := make(map[string]string)

		// Set up environment with variables to test unset behavior
		initialEnv := []string{
			"KEEP_VAR=keep_this",
			"REMOVE_VAR=remove_this",
			"PATH=/usr/bin:/bin",
		}

		cfg := &config.Config{MinutesToTimeout: 1, Env: initialEnv}
		ui := view.NewView("mock")

		// Unset one variable while keeping others
		chunk := ExecutableChunk{
			Runtime: "bash",
			Content: []string{"unset REMOVE_VAR"},
			Context: &runnercontext.Context{
				Cfg:   cfg,
				RView: ui,
			},
		}
		err := chunk.PrepareForExecution(tmpDirs)
		assert.NoError(t, err, "Expected no error when preparing chunk for execution")

		err = chunk.ExecuteSequential()
		assert.NoError(t, err, "Expected no error when executing bash chunk")

		// Verify the targeted variable was unset
		foundRemoved := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return strings.HasPrefix(env, "REMOVE_VAR=")
		})
		assert.False(t, foundRemoved, "Expected REMOVE_VAR to be removed after unset")

		// Verify other variables are still preserved
		foundKeep := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return env == "KEEP_VAR=keep_this"
		})
		assert.True(t, foundKeep, "Expected KEEP_VAR to be preserved after unset")

		foundPath := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return strings.HasPrefix(env, "PATH=")
		})
		assert.True(t, foundPath, "Expected PATH to be preserved after unset")
	})

	t.Run("bash env unset critical variables", func(t *testing.T) {
		tmpDirs := make(map[string]string)

		// Set up environment with critical variables
		initialEnv := []string{
			"HOME=/home/test",
			"PATH=/usr/bin:/bin",
			"KEEP_VAR=keep_this",
		}

		cfg := &config.Config{MinutesToTimeout: 1, Env: initialEnv}
		ui := view.NewView("mock")

		// Unset HOME to verify critical environment variables can be unset
		chunk := ExecutableChunk{
			Runtime: "bash",
			Content: []string{"unset HOME"},
			Context: &runnercontext.Context{
				Cfg:   cfg,
				RView: ui,
			},
		}
		err := chunk.PrepareForExecution(tmpDirs)
		assert.NoError(t, err, "Expected no error when preparing chunk for execution")

		err = chunk.ExecuteSequential()
		assert.NoError(t, err, "Expected no error when executing bash chunk")

		// Verify HOME was unset
		foundHome := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return strings.HasPrefix(env, "HOME=")
		})
		assert.False(t, foundHome, "Expected HOME to be removed after unset")

		// Verify other variables are still preserved
		foundKeep := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return env == "KEEP_VAR=keep_this"
		})
		assert.True(t, foundKeep, "Expected KEEP_VAR to be preserved after unsetting HOME")

		foundPath := slices.ContainsFunc(chunk.Context.Cfg.Env, func(env string) bool {
			return strings.HasPrefix(env, "PATH=")
		})
		assert.True(t, foundPath, "Expected PATH to be preserved after unsetting HOME")
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

	t.Run("bash env extraction with echo -n command", func(t *testing.T) {
		// Regression test for issue #25: ensure ENV marker doesn't get appended
		// to command output when the command doesn't end with a newline
		tmpDirs := make(map[string]string)

		cfg := &config.Config{MinutesToTimeout: 1, Env: []string{}}
		ui := view.NewView("mock")

		chunk := ExecutableChunk{
			Runtime: "bash",
			Content: []string{"echo -n 'TEST'"},
			Context: &runnercontext.Context{
				Cfg:   cfg,
				RView: ui,
			},
		}
		err := chunk.PrepareForExecution(tmpDirs)
		assert.NoError(t, err, "Expected no error when preparing chunk for execution")

		err = chunk.ExecuteSequential()
		assert.NoError(t, err, "Expected no error when executing bash chunk")

		// Verify the output is just "TEST\n" without the ENV marker appended to it
		// The trailing \n is added by the command processing logic
		assert.Equal(t, "TEST\n", chunk.Commands[0].Stdout, "Expected stdout to be 'TEST\\n' without ENV marker concatenated")
		// Crucially, verify it's NOT "TEST### ENV ###" which was the bug
		assert.NotContains(t, chunk.Commands[0].Stdout, "### ENV ###", "Expected ENV marker to be stripped from output")

		// Verify ENV section was still extracted (should have environment variables)
		assert.NotEmpty(t, chunk.Context.Cfg.Env, "Expected environment variables to be extracted")
	})

	t.Run("bash env extraction with normal echo command", func(t *testing.T) {
		// Ensure normal commands (with newline) don't get extra blank lines
		tmpDirs := make(map[string]string)

		cfg := &config.Config{MinutesToTimeout: 1, Env: []string{}}
		ui := view.NewView("mock")

		chunk := ExecutableChunk{
			Runtime: "bash",
			Content: []string{"echo 'NORMAL'"},
			Context: &runnercontext.Context{
				Cfg:   cfg,
				RView: ui,
			},
		}
		err := chunk.PrepareForExecution(tmpDirs)
		assert.NoError(t, err, "Expected no error when preparing chunk for execution")

		err = chunk.ExecuteSequential()
		assert.NoError(t, err, "Expected no error when executing bash chunk")

		// Verify the output doesn't have extra blank lines
		assert.Equal(t, "NORMAL\n", chunk.Commands[0].Stdout, "Expected stdout to be 'NORMAL\\n' without extra blank lines")

		// Verify ENV section was still extracted
		assert.NotEmpty(t, chunk.Context.Cfg.Env, "Expected environment variables to be extracted")
	})
}
