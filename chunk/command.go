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

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/pterm/pterm"
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
	if config.Verbose {
		command.CmdPrettyName = fmt.Sprint(command.CmdPrettyName, " in ", command.Cmd.Dir)
		if len(command.Env) > 0 {
			command.CmdPrettyName = fmt.Sprint(command.CmdPrettyName, " with env ", command.Env)
		}
	}
}

// Start begins the execution of the command but does not wait for it to
// complete. It sets up the I/O pipes and starts the command process.
// It returns an error if the command cannot be started.
func (command *RunningCommand) Start() error {
	if config.Interactive {
		result, _ := pterm.DefaultInteractiveContinue.WithDefaultText(command.CmdPrettyName).Show()
		if result == "all" {
			config.Interactive = false
		}
		if result == "no" {
			command.Cmd = nil
			return nil
		}
		if result == "cancel" {
			return errors.New("User aborted")
		}
	}

	command.Cmd.Stdout = &command.Outb
	command.Cmd.Stderr = &command.Errb
	if config.DryRun {
		return nil
	}
	err := command.Cmd.Start()
	if err != nil {
		pterm.Error.Printf("%s: %s\n", command.CmdPrettyName, err)
	}
	return err
}

// Wait blocks until the command has finished execution. It captures the exit
// code, stdout, and stderr, and handles environment variable extraction for
// bash scripts. It returns an error if the command fails.
func (command *RunningCommand) Wait() error {
	var spiner *pterm.SpinnerPrinter
	if config.Interactive {
		spiner, _ = pterm.DefaultSpinner.Start("executing")
	} else {
		spiner, _ = pterm.DefaultSpinner.Start(command.CmdPrettyName)
	}
	// don't wait if we're in dryRun mode
	if config.DryRun {
		spiner.InfoPrinter = &pterm.PrefixPrinter{
			MessageStyle: &pterm.Style{pterm.FgLightBlue},
			Prefix: pterm.Prefix{
				Style: &pterm.Style{pterm.FgBlack, pterm.BgLightBlue},
				Text:  " DRY-RUN ",
			},
		}
		spiner.Warning(command.CmdPrettyName)
		return nil
	}
	defer command.CancelFunc()
	// handle the interactive scenario where the command doesn't exist because the user skipped it
	if command.Cmd == nil {
		spiner.InfoPrinter = &pterm.PrefixPrinter{
			MessageStyle: &pterm.Style{pterm.FgLightBlue},
			Prefix: pterm.Prefix{
				Style: &pterm.Style{pterm.FgBlack, pterm.BgLightBlue},
				Text:  " SKIPPED ",
			},
		}
		spiner.Warning(command.CmdPrettyName)
		return nil
	}
	// wait for the termination
	terminatingError := command.Cmd.Wait()
	command.Stdout = command.Outb.String()
	command.Stderr = command.Errb.String()

	// handle the output depending on the status of the command
	if terminatingError != nil {
		spiner.Fail("stdout:\n", command.Outb.String(), "\nstderr:\n", command.Errb.String(), "\nexit code:", command.Cmd.ProcessState.ExitCode())
		return terminatingError
	}
	spiner.Success(command.CmdPrettyName)

	// During a bash runtime the user might want to export new variables.
	// Our job here is to recover them to build the new environment for the next chunk
	if command.IsBash {
		stoudtLines := strings.Split(command.Stdout, "\n")
		// reinitialize the env
		config.Env = []string{}
		config.Env = append(config.Env, os.Environ()...)
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
					config.Env = append(config.Env, line)
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
	if config.Verbose {
		if command.Stdout != "" {
			pterm.Info.Println(command.Stdout)
		}
		if command.Stderr != "" {
			pterm.Warning.Println(command.Stderr)
		}
	}
	return nil
}

// Execute runs a command and waits for it to complete. It's a convenience
// method that calls Start and then Wait.
// It returns an error if the command fails at any stage.
func (command *RunningCommand) Execute() error {
	err := command.Start()
	if err != nil {
		return err
	}
	return command.Wait()
}
