// Package chunk provides the core data structures for the markdown runner. It
// defines the ExecutableChunk, which represents a runnable code block, and the
// RunningCommand, which represents a single command to be executed.
package chunk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/arkmq-org/markdown-runner/runnercontext"
)

// RunningCommand represents a command that has been parsed and is ready to be
// executed. It encapsulates the command itself, its environment, and its I/O buffers.
type RunningCommand struct {
	// Env holds the environment variables for the command.
	Env []string
	// Cmd is the underlying exec.Cmd instance for the command.
	Cmd *exec.Cmd
	// Outb is the buffer used to capture the command's standard output.
	Outb bytes.Buffer
	// Errb is the buffer used to capture the command's standard error.
	Errb bytes.Buffer
	// CmdPrettyName is a human-readable name for the command, used for logging.
	CmdPrettyName string
	// CancelFunc is the function to call to cancel the command's context,
	// typically used for timeouts.
	CancelFunc context.CancelFunc
	// Stdout stores the captured standard output as a string after execution.
	Stdout string
	// Stderr stores the captured standard error as a string after execution.
	Stderr string
	// ReturnCode holds the exit code of the command after it has run.
	ReturnCode int
	// IsBash indicates whether the command is a bash script, which requires
	// special environment variable handling.
	IsBash bool
	Ctx    *runnercontext.Context
	id     string
	// Option to pass in a function for user input that will override the one from pterm
	// this is useful for testing.
	GetUserInput func(string) (string, error)
}

// InitCommandLabel sets the human-readable name for the command, using the
// chunk's label if available, and adding extra details in verbose mode.
//
// chunk is the ExecutableChunk this command belongs to.
func (command *RunningCommand) InitCommandLabel(chunk *ExecutableChunk) {
	command.CmdPrettyName = strings.Join(command.Cmd.Args, " ")
	if chunk.Label != "" {
		command.CmdPrettyName = chunk.Label
	}
	if command.Ctx.Cfg.Verbose {
		command.CmdPrettyName = fmt.Sprint(command.CmdPrettyName, " in ", command.Cmd.Dir)
		if len(command.Env) > 0 {
			command.CmdPrettyName = fmt.Sprint(command.CmdPrettyName, " with env ", command.Env)
		}
	}
}

func (command *RunningCommand) interactivePrompt() error {
	if command.Ctx.Cfg.Interactive {
		var result string
		var err error
		if command.GetUserInput != nil {
			result, err = command.GetUserInput(command.CmdPrettyName)
		} else {
			result, err = command.Ctx.RView.InteractivePromptForCommand(command.CmdPrettyName, command.CmdPrettyName, &command.Ctx.Cfg.Interactive)
		}
		if err != nil {
			return err
		}
		if result == "all" {
			command.Ctx.Cfg.Interactive = false
		}
		if result == "no" {
			command.Cmd = nil
			return nil
		}
		if result == "cancel" {
			return errors.New("User aborted")
		}
	}
	return nil
}

// Start begins the execution of the command but does not wait for it to
// complete. It sets up the I/O pipes and starts the command process.
// It returns an error if the command cannot be started.
func (command *RunningCommand) Start() error {
	if !command.Ctx.RView.HasLogger(command.id) {
		return errors.New("This command doesn't have a corresponding spinner in the UI")
	}
	if command.Cmd == nil {
		return nil // no-op
	}

	command.Cmd.Stdout = &command.Outb
	command.Cmd.Stderr = &command.Errb
	if command.Ctx.Cfg.DryRun {
		return nil
	}
	err := command.Cmd.Start()
	if err != nil {
		command.Ctx.RView.Error(fmt.Sprintf("%s: %s\n", command.CmdPrettyName, err))
	}
	return err
}

// InitializeLogger creates and configures a logger for a command.
// The logger provides visual feedback during command execution.
// It returns an error if the logger cannot be initialized.
func (command *RunningCommand) InitializeLogger() error {
	err := command.interactivePrompt()
	if err != nil {
		return err
	}
	var spinnerText string = command.CmdPrettyName
	if command.Ctx.Cfg.Interactive {
		spinnerText = "executing"
	}
	return command.Ctx.RView.StartCommand(command.id, spinnerText)
}

// Wait blocks until the command has finished execution. It captures the exit
// code, stdout, and stderr, and handles environment variable extraction for
// bash scripts. It returns an error if the command fails.
func (command *RunningCommand) Wait() error {
	// don't wait if we're in dryRun mode
	if command.Ctx.Cfg.DryRun {
		command.Ctx.RView.DryRunCommand(command.id, command.CmdPrettyName)
		return nil
	}
	defer command.CancelFunc()
	// handle the interactive scenario where the command doesn't exist because the user skipped it
	if command.Cmd == nil {
		command.Ctx.RView.SkipCommand(command.id, command.CmdPrettyName)
		return nil
	}
	// wait for the termination
	terminatingError := command.Cmd.Wait()
	command.Stdout = command.Outb.String()
	command.Stderr = command.Errb.String()

	// handle the output depending on the status of the command
	if terminatingError != nil {
		command.Ctx.RView.StopCommand(command.id, false, fmt.Sprintf("stdout:\n%s\nstderr:\n%s\nexit code:%d", command.Outb.String(), command.Errb.String(), command.Cmd.ProcessState.ExitCode()))
		return terminatingError
	}
	command.Ctx.RView.StopCommand(command.id, true, command.CmdPrettyName)

	// During a bash runtime the user might want to export new variables.
	// Our job here is to recover them to build the new environment for the next chunk
	if command.IsBash {
		stoudtLines := strings.Split(command.Stdout, "\n")
		// reinitialize the env
		command.Ctx.Cfg.Env = []string{}
		command.Ctx.Cfg.Env = append(command.Ctx.Cfg.Env, os.Environ()...)
		var newLines []string
		extractVariables := false
		for _, line := range stoudtLines {
			// the environment gets separated by the output from a special string
			if line == "### ENV ###" {
				extractVariables = true
			}
			// then all of it is variables and can be added to the env
			if extractVariables {
				parts := strings.Split(line, "=")
				if len(parts) > 1 {
					command.Ctx.Cfg.Env = append(command.Ctx.Cfg.Env, line)
				}
			} else {
				// if not it's output we want to keep for the user
				newLines = append(newLines, line)
			}
		}
		if len(newLines) > 0 {
			command.Stdout = strings.Join(newLines, "\n")
			command.Stdout = command.Stdout + "\n"
		} else {
			command.Stdout = ""
		}
	}

	// print more things while in verbose mode
	if command.Ctx.Cfg.Verbose {
		if command.Stdout != "" {
			command.Ctx.RView.Info(command.Stdout)
		}
		if command.Stderr != "" {
			command.Ctx.RView.Warning(command.Stderr)
		}
	}
	return nil
}

// Kill forcefully terminates the command's process. It's used to clean up
// running processes when a stage fails.
// It returns an error if the process cannot be killed.
func (command *RunningCommand) Kill() error {
	command.Ctx.RView.KillCommand(command.id, command.CmdPrettyName)
	return command.Cmd.Process.Kill()
}

// Execute runs a command and waits for it to complete. It's a convenience method that calls Start and then Wait.
// This should not be called on commands supposed to be executed in parallel.
func (command *RunningCommand) Execute() error {
	err := command.InitializeLogger()
	if err != nil {
		return err
	}
	err = command.Start()
	if err != nil {
		return err
	}
	return command.Wait()
}
