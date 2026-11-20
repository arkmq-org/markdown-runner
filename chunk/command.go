// Package chunk provides the core data structures for the markdown runner. It
// defines the ExecutableChunk, which represents a runnable code block, and the
// RunningCommand, which represents a single command to be executed.
package chunk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

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
	// typically used for timeouts. Start re-wraps it (see armTimeout) so that
	// calling it also stops the timeout timer; Start must be called only once.
	CancelFunc context.CancelFunc
	// timeout is the maximum duration the command is allowed to run. The
	// countdown starts when the command is started, not when it is created.
	timeout time.Duration
	// execCtx is the context the command runs under. It is cancelled with
	// context.DeadlineExceeded as cause when the timeout fires, which tells
	// whether the command was killed because of it.
	execCtx context.Context
	// cancelCause cancels execCtx with a given cause.
	cancelCause context.CancelCauseFunc
	// outputPipes are the read ends of the command's stdout and stderr pipes.
	outputPipes []*os.File
	// outputDone is closed once everything written to the output pipes has been
	// copied into Outb and Errb.
	outputDone chan struct{}
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
		switch result {
		case "continue":
			// like gdb's continue: run without prompting until the next breakpoint
			command.Ctx.Cfg.Interactive = false
		case "no":
			command.Cmd = nil
			return nil
		case "quit":
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
	command.armTimeout()
	err := command.startWithOutputPipes()
	if err != nil {
		// Wait won't be called, release the timer and the context now.
		if command.CancelFunc != nil {
			command.CancelFunc()
		}
		command.Ctx.RView.Error(fmt.Sprintf("%s: %s\n", command.CmdPrettyName, err))
	}
	return err
}

// armTimeout starts the countdown after which the command's context gets
// cancelled with context.DeadlineExceeded as cause, killing the command. It is called right before the command is
// started so that the time spent before (e.g. waiting for the user to make a
// decision in interactive mode) isn't accounted for. The CancelFunc is wrapped
// so that calling it also releases the timer.
func (command *RunningCommand) armTimeout() {
	cancel := command.CancelFunc
	if cancel == nil || command.cancelCause == nil {
		return
	}
	cancelCause := command.cancelCause
	timer := time.AfterFunc(command.timeout, func() { cancelCause(context.DeadlineExceeded) })
	command.CancelFunc = func() {
		timer.Stop()
		cancel()
	}
}

// outputGracePeriod bounds how long the output of a timed out command is still
// read after the command was killed. Its descendants are killed with it (see
// killProcessTree), this is the safety net for processes that escaped the kill
// (daemons that left the process tree, platforms without /proc).
const outputGracePeriod = time.Second

// startWithOutputPipes starts the command with its stdout and stderr connected to
// pipes that are read here rather than by os/exec. With an io.Writer, os/exec's
// Wait blocks until every process holding the pipes closes them: when a timeout
// kills a bash script, its still running children (e.g. a sleep) would keep Wait
// blocked and the timeout would never be reported. Owning the pipes lets Wait
// return as soon as the command itself exits, and collectOutput decide how long
// to keep reading. The command stays in the runner's process group, so Ctrl-C
// still reaches it and it can still use the terminal.
func (command *RunningCommand) startWithOutputPipes() error {
	outR, outW, err := os.Pipe()
	if err != nil {
		return err
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		outR.Close()
		outW.Close()
		return err
	}
	command.Cmd.Stdout = outW
	command.Cmd.Stderr = errW
	err = command.Cmd.Start()
	// the child has its own copies of the write ends: closing ours lets the readers see EOF once it is done with them
	outW.Close()
	errW.Close()
	if err != nil {
		outR.Close()
		errR.Close()
		return err
	}
	command.outputPipes = []*os.File{outR, errR}
	command.outputDone = make(chan struct{})
	var copying sync.WaitGroup
	copying.Add(2)
	go func() { defer copying.Done(); _, _ = io.Copy(&command.Outb, outR) }()
	go func() { defer copying.Done(); _, _ = io.Copy(&command.Errb, errR) }()
	go func() { copying.Wait(); close(command.outputDone) }()
	return nil
}

// collectOutput waits until the command's output has been read, then releases
// the pipes. Normally it waits until every process holding the pipes (including
// background children of a successful command) has closed them. For a timed out
// command it only waits outputGracePeriod: its children may still hold them.
func (command *RunningCommand) collectOutput(timedOut bool) {
	if command.outputDone == nil {
		return
	}
	if timedOut {
		select {
		case <-command.outputDone:
		case <-time.After(outputGracePeriod):
		}
	} else {
		<-command.outputDone
	}
	command.closeOutputPipes()
	<-command.outputDone
}

// closeOutputPipes closes the read ends of the output pipes, which stops the
// goroutines copying them.
func (command *RunningCommand) closeOutputPipes() {
	for _, pipe := range command.outputPipes {
		pipe.Close()
	}
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
	if command.CancelFunc != nil {
		defer command.CancelFunc()
	}
	// don't wait if we're in dryRun mode
	if command.Ctx.Cfg.DryRun {
		command.Ctx.RView.DryRunCommand(command.id, command.CmdPrettyName)
		return nil
	}
	// handle the interactive scenario where the command doesn't exist because the user skipped it
	if command.Cmd == nil {
		command.Ctx.RView.SkipCommand(command.id, command.CmdPrettyName)
		return nil
	}
	// wait for the termination
	terminatingError := command.Cmd.Wait()
	timedOut := terminatingError != nil && command.hasTimedOut()
	command.collectOutput(timedOut)
	command.Stdout = command.Outb.String()
	command.Stderr = command.Errb.String()

	// handle the output depending on the status of the command
	if terminatingError != nil {
		message := fmt.Sprintf("stdout:\n%s\nstderr:\n%s\nexit code:%d", command.Stdout, command.Stderr, command.Cmd.ProcessState.ExitCode())
		if timedOut {
			// the process was killed by the context deadline, make it explicit which command it was
			message = fmt.Sprintf("%s: timed out after %s\n%s", command.CmdPrettyName, command.timeout, message)
			terminatingError = fmt.Errorf("%s: timed out after %s: %w", command.CmdPrettyName, command.timeout, terminatingError)
		}
		command.Ctx.RView.StopCommand(command.id, false, message)
		return terminatingError
	}
	command.Ctx.RView.StopCommand(command.id, true, command.CmdPrettyName)

	// During a bash runtime the user might want to export new variables or unset existing ones.
	// Our job here is to recover them to build the new environment for the next chunk
	if command.IsBash {
		stoudtLines := strings.Split(command.Stdout, "\n")
		var newLines []string
		extractVariables := false
		var bashEnvVars []string

		for _, line := range stoudtLines {
			// the environment gets separated by the output from a special string
			if line == "### ENV ###" {
				extractVariables = true
			}
			// then all of it is variables and can be added to the env
			if extractVariables {
				parts := strings.Split(line, "=")
				if len(parts) > 1 {
					bashEnvVars = append(bashEnvVars, line)
				}
			} else {
				// if not it's output we want to keep for the user
				newLines = append(newLines, line)
			}
		}

		// This includes all environment variables that should be available to subsequent chunks.
		command.Ctx.Cfg.Env = bashEnvVars

		if len(newLines) > 0 {
			// Remove the trailing empty line that precedes the ENV marker (added by our \n in the echo)
			// This prevents an extra blank line when the ENV section is stripped
			if len(newLines) > 0 && newLines[len(newLines)-1] == "" {
				newLines = newLines[:len(newLines)-1]
			}
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

// hasTimedOut reports whether the process was killed because it ran longer than
// the timeout: its context was cancelled by the timeout timer and it was terminated by a
// signal. Checking the signal matters for parallel chunks, which are waited for
// one after the other: a command that failed on its own may only be waited for
// after its deadline has passed, and must not be reported as a timeout.
func (command *RunningCommand) hasTimedOut() bool {
	if command.execCtx == nil || !errors.Is(context.Cause(command.execCtx), context.DeadlineExceeded) {
		return false
	}
	if command.Cmd.ProcessState == nil {
		return false
	}
	status, ok := command.Cmd.ProcessState.Sys().(syscall.WaitStatus)
	return ok && status.Signaled()
}

// Kill forcefully terminates the command's process and its descendants. It's
// used to clean up running processes when a stage fails.
// It returns an error if the process cannot be killed.
func (command *RunningCommand) Kill() error {
	command.Ctx.RView.KillCommand(command.id, command.CmdPrettyName)
	err := killProcessTree(command.Cmd.Process)
	// nobody reads the output of a killed command, release the pipes
	command.closeOutputPipes()
	return err
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
