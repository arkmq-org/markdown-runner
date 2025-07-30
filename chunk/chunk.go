// Package chunk provides the core data structures for the markdown runner. It
// defines the ExecutableChunk, which represents a runnable code block, and the
// RunningCommand, which represents a single command to be executed.
package chunk

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/arkmq-org/markdown-runner/runnercontext"
	"github.com/google/shlex"
	"github.com/google/uuid"
)

// ExecutableChunk represents a block of code from a markdown file that can be
// executed. It contains all the metadata and content parsed from the code fence.
type ExecutableChunk struct {
	// Stage is the name of the stage this chunk belongs to. All chunks within
	// the same stage are executed before moving to the next stage.
	Stage string `json:"stage"`
	// Id is a unique identifier for this chunk within its stage. It can be
	// referenced by other chunks using the "requires" field.
	Id string `json:"id,omitempty"`
	// Requires specifies a dependency on another chunk. The format is
	// "stageName/chunkId". The current chunk will only be executed if the
	// required chunk has been executed successfully.
	Requires string `json:"requires,omitempty"`
	// RootDir specifies the execution directory for the chunk. It can be set
	// to special values like "$initial_dir" or "$tmpdir.name" to use the
	// initial working directory or a shared temporary directory, respectively.
	RootDir string `json:"rootdir,omitempty"`
	// Runtime specifies the execution environment. Common values are "bash"
	// for shell scripts or "writer" to write content to a file.
	Runtime string `json:"runtime,omitempty"`
	// IsParallel, if true, indicates that this chunk can be run in parallel
	// with other chunks in the same stage.
	IsParallel bool `json:"parallel,omitempty"`
	// Label provides a human-readable name for the chunk, which is used in
	// logging and CLI output.
	Label string `json:"label,omitempty"`
	// HasBreakpoint, if true, will cause the runner to pause execution
	// before this chunk, waiting for user input to continue.
	HasBreakpoint bool `json:"breakpoint,omitempty"`
	// Destination is the target file path for chunks with the "writer" runtime.
	Destination string `json:"destination,omitempty"`
	// Content holds the lines of code that make up the chunk's body.
	Content []string
	// Commands is the list of RunningCommand instances generated from the Content.
	Commands []*RunningCommand
	// BackQuotes stores the number of backquotes used in the opening code fence,
	// which is needed to correctly parse the end of the chunk.
	BackQuotes int
	Context    *runnercontext.Context
	IsSkipped  bool
}

// Init initializes an ExecutableChunk after it has been unmarshalled from JSON.
// It sets up the Content slice and prints a warning if a breakpoint is set.
func (chunk *ExecutableChunk) Init() {
	chunk.Content = []string{}
	if chunk.HasBreakpoint {
		chunk.Context.RView.Warning("breakpoint in the document")
	}
	chunk.IsSkipped = false
}

// HasOutput checks if any of the commands in the chunk have produced stdout.
// It always returns true if the runner is in dry-run mode.
func (chunk *ExecutableChunk) HasOutput() bool {
	if chunk.Context.Cfg.DryRun {
		return true
	}
	for _, command := range chunk.Commands {
		if command.Stdout != "" || command.Stderr != "" {
			return true
		}
	}
	return false
}

// WriteOutputTo writes the captured stdout and stderr of all commands in the
// chunk to a new code block in the provided writer.
//
// bqNumber is the number of backquotes to use for the output code fence.
// writer is the bufio.Writer to write the output to.
func (chunk *ExecutableChunk) WriteOutputTo(bqNumber int, writer *bufio.Writer) error {
	var err error
	// print the start of the chunk
	for i := 0; i < bqNumber; i++ {
		_, err = writer.WriteString("`")
		if err != nil {
			return err
		}
	}
	_, err = writer.WriteString("shell markdown_runner\n")
	if err != nil {
		return err
	}
	for _, command := range chunk.Commands {
		if command.Stdout != "" {
			_, err = writer.WriteString(command.Stdout)
			if err != nil {
				return err
			}
			// make sure to only have one carriage return at the end
			if command.Stdout[len(command.Stdout)-1] != '\n' {
				_, err = writer.WriteString("\n")
				if err != nil {
					return err
				}
			}
		}
		if command.Stderr != "" {
			_, err = writer.WriteString(command.Stderr)
			if err != nil {
				return err
			}
			// make sure to only have one carriage return at the end
			if command.Stderr[len(command.Stderr)-1] != '\n' {
				_, err = writer.WriteString("\n")
				if err != nil {
					return err
				}
			}
		}
	}
	// print the end of the chunk
	for i := 0; i < bqNumber; i++ {
		_, err = writer.WriteString("`")
		if err != nil {
			return err
		}
	}
	// add a final carriage return
	_, err = writer.WriteString("\n")
	if err != nil {
		return err
	}
	return nil
}

// GetOrCreateRuntimeDirectory determines the correct execution directory for
// a chunk based on its RootDir property. It supports special values like
// "$initial_dir" and "$tmpdir.name" and creates temporary directories as needed.
//
// tmpDirs is a map used to cache and reuse temporary directories.
// It returns the path to the directory and an error if one occurred.
func (chunk *ExecutableChunk) GetOrCreateRuntimeDirectory(tmpDirs map[string]string) (string, error) {
	var err error
	// if nothing was passed in, create a unique temporary directory anyway to keep the system clean
	if chunk.RootDir == "" {
		uuid := uuid.New()
		chunk.RootDir = "$tmpdir." + uuid.String()
	}
	// Initialize where the command is getting executed
	if chunk.RootDir == "$initial_dir" {
		return chunk.Context.Cfg.Rootdir, nil
	}
	// tmpdirs are reusable between commands.
	if strings.HasPrefix(chunk.RootDir, "$tmpdir") {
		dirselector := strings.Split(chunk.RootDir, string(os.PathSeparator))[0]
		// so they are stored in a map.
		tmpdir, exists := tmpDirs[dirselector]
		if !exists {
			tmpdir, err = os.MkdirTemp("/tmp", "*")
			if err != nil {
				chunk.Context.RView.Error(err.Error())
				return "", err
			}
			tmpDirs[dirselector] = tmpdir
		}
		return strings.Replace(chunk.RootDir, dirselector, tmpdir, 1), nil
	}
	// if the directory doesn't start with a $ sign the user wanted its own custom value, let's use it directly
	if !strings.HasPrefix(chunk.RootDir, "$") {
		return chunk.RootDir, nil
	}
	return "", errors.New("Impossible to figure out the directory to run in: " + chunk.RootDir)
}

// WriteFile writes the content of a chunk with the "writer" runtime to its
// specified destination file.
//
// basedir is the root directory where the destination file will be created.
func (chunk *ExecutableChunk) WriteFile(basedir string) error {
	scriptPath := path.Join(basedir, chunk.Destination)
	f, err := os.Create(scriptPath)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	// the actual content the user wants in the file
	for _, line := range chunk.Content {
		_, err = writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

// WriteBashScript writes the content of a chunk with the "bash" runtime to a
// temporary shell script on disk. The script is made executable and includes
// standard shell boilerplate.
//
// basedir is the directory where the script will be created.
// script_name is the name of the script file.
func (chunk *ExecutableChunk) WriteBashScript(basedir string, script_name string) error {
	scriptPath := path.Join(basedir, script_name)
	f, err := os.Create(scriptPath)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	_, err = writer.WriteString("#!/bin/bash\n")
	if err != nil {
		return err
	}

	// fail fast
	_, err = writer.WriteString("set -euo pipefail\n")
	if err != nil {
		return err
	}

	// the actual script the user wants to execute
	for _, line := range chunk.Content {
		_, err = writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	// bubble up the env after the script execution
	_, err = writer.WriteString("echo \"### ENV ###\"\n")
	_, err = writer.WriteString("printenv\n")
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return os.Chmod(scriptPath, 0o770)
}

// HasFinishedExecution checks if all commands within the chunk have completed
// their execution, regardless of their exit code.
func (chunk *ExecutableChunk) HasFinishedExecution() bool {
	if chunk.Context.Cfg.DryRun {
		return true
	}
	if len(chunk.Commands) == 0 {
		return false
	}
	for _, command := range chunk.Commands {
		if command.Cmd.ProcessState == nil {
			return false
		}
	}
	return true
}

// HasExecutedCorrectly checks if all commands within the chunk completed with a
// zero exit code. It always returns true in dry-run mode.
func (chunk *ExecutableChunk) HasExecutedCorrectly() bool {
	if !chunk.HasFinishedExecution() {
		return false
	}
	var allOk bool = true
	for _, command := range chunk.Commands {
		allOk = allOk && command.Cmd.ProcessState.ExitCode() == 0
	}
	return allOk
}

// AddCommandToExecute parses a command string, creates a new RunningCommand,
// and adds it to the chunk's list of commands.
//
// trimedCommand is the raw command string to be parsed.
// tmpDirs is the map of temporary directories for runtime directory resolution.
// It returns the newly created RunningCommand and an error if parsing fails.
func (chunk *ExecutableChunk) AddCommandToExecute(trimedCommand string, tmpDirs map[string]string) (*RunningCommand, error) {
	var command RunningCommand
	command.Ctx = chunk.Context
	command.id = uuid.New().String()

	// create the cancel background function, the command.cancelFunc has to get called eventually to avoid leaking
	// memory
	ctx := context.Background()
	ctx, command.CancelFunc = context.WithTimeout(context.Background(), time.Duration(chunk.Context.Cfg.MinutesToTimeout)*time.Minute)

	if trimedCommand == "" {
		return nil, errors.New("empty command string provided")
	}

	// smart split the command in several pieces
	splited, err := shlex.Split(trimedCommand)
	if err != nil {
		return nil, err
	}
	executable := splited[0]
	args := splited[1:]
	command.Cmd = exec.CommandContext(ctx, executable, args...)

	// set the runtime directory for the command
	command.Cmd.Dir, err = chunk.GetOrCreateRuntimeDirectory(tmpDirs)
	if err != nil {
		return nil, err
	}

	// Copy the environment before calling the command
	command.Cmd.Env = append(command.Cmd.Env, chunk.Context.Cfg.Env...)

	// give a pretty name to the command for the cli output
	command.CmdPrettyName = trimedCommand

	// set the bash flag for the command
	command.IsBash = chunk.Runtime == "bash"

	chunk.Commands = append(chunk.Commands, &command)
	return &command, nil
}

// ExecuteSequential runs the commands in a sequential chunk. It's a convenience method
// that initializes a spinner for each command and then executes it. This method
// should only be used for chunks where IsParallel is false.
func (chunk *ExecutableChunk) ExecuteSequential() error {
	if chunk.IsParallel {
		return errors.New("Cannot execute a parallel chunk with Execute, use Start instead")
	}
	for _, command := range chunk.Commands {
		err := command.Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

func (chunk *ExecutableChunk) DeclareParallelLoggers() error {
	if !chunk.IsParallel {
		return errors.New("Cannot declare parallel loggers on a sequential chunk")
	}
	return chunk.Commands[0].InitializeLogger()
}

// StartParallel begins the execution of a parallel chunk's command without waiting for
// it to complete. This method should only be used for chunks where IsParallel
// is true.
func (chunk *ExecutableChunk) StartParallel() error {
	if !chunk.IsParallel {
		return errors.New("Cannot start a non-parallel chunk with Start, use Execute instead")
	}
	return chunk.Commands[0].Start()
}

// Wait waits for the last command of a parallel chunk to complete. If
// shouldKill is true, it terminates the command process instead of waiting for

// it to finish gracefully. This is used to clean up parallel processes when an
// error has occurred elsewhere in the stage.
func (chunk *ExecutableChunk) WaitParallel(shouldKill bool) error {
	if !chunk.IsParallel {
		return errors.New("Cannot wait for a non-parallel chunk with Wait, use Execute instead")
	}
	command := chunk.Commands[0]
	if shouldKill {
		return command.Kill()
	}
	return command.Wait()
}

// applyWriter handles the execution for a chunk with the "writer" runtime. It
// creates the destination file and writes the chunk's content to it, showing a
// spinner in the CLI during the process.
func (chunk *ExecutableChunk) applyWriter(tmpDirs map[string]string) error {
	writerString := "writing " + chunk.Destination + " on disk"
	if chunk.Label != "" {
		writerString += " for " + chunk.Label
	}
	id := uuid.New().String()
	chunk.Context.RView.StartCommand(id, writerString)
	directory, err := chunk.GetOrCreateRuntimeDirectory(tmpDirs)
	if err != nil {
		chunk.Context.RView.StopCommand(id, false, err.Error())
		return err
	}
	err = chunk.WriteFile(directory)
	if err != nil {
		chunk.Context.RView.StopCommand(id, false, err.Error())
		return err
	}
	chunk.Context.RView.StopCommand(id, true, "")
	return nil
}

// prepareClassical prepares a standard chunk for execution by converting each
// line of its content into a RunningCommand. This is the default behavior for
// chunks without a specific runtime.
func (chunk *ExecutableChunk) prepareClassical(tmpDirs map[string]string) error {
	// In the case the chunk is a parallel classical chunk, make sure the user hasn't
	// specified multiple commands, as the behavior would be hard to keep in check.
	if chunk.IsParallel && len(chunk.Content) > 1 {
		return errors.New("Multiple commands for non bash runtime is not supported when parallel is set, update the chunk to a bash runtime")
	}
	for _, command := range chunk.Content {
		_, err := chunk.AddCommandToExecute(command, tmpDirs)
		if err != nil {
			return err
		}
	}
	return nil
}

// prepareBashChunkForExecution prepares a "bash" runtime chunk for execution.
// It writes the chunk's content to a temporary script file and creates a
// RunningCommand to execute that script.
func (chunk *ExecutableChunk) prepareBashChunkForExecution(tmpDirs map[string]string) error {
	uuid := uuid.New()
	cmd := "./" + uuid.String() + ".sh"
	command, err := chunk.AddCommandToExecute(cmd, tmpDirs)
	if err != nil {
		return err
	}
	err = chunk.WriteBashScript(command.Cmd.Dir, cmd)
	if err != nil {
		return err
	}
	return nil
}

// PrepareForExecution sets up the chunk for execution based on its runtime.
// It dispatches to the appropriate helper function (e.g., for "writer" or
// "bash" runtimes) to create the necessary commands and files.
func (chunk *ExecutableChunk) PrepareForExecution(tmpDirs map[string]string) error {
	switch chunk.Runtime {
	case "writer":
		return chunk.applyWriter(tmpDirs)
	case "bash":
		return chunk.prepareBashChunkForExecution(tmpDirs)
	default:
		return chunk.prepareClassical(tmpDirs)
	}
}

func (chunk *ExecutableChunk) Skip() {
	chunk.IsSkipped = true
	switch chunk.Runtime {
	case "writer":
		chunk.Context.RView.Info("Skip writer chunk '" + chunk.Label + "' due to previous errors")
	case "bash":
		chunk.Context.RView.Info("Skip bash chunk '" + chunk.Label + "' due to previous errors")
	default:
		for _, command := range chunk.Content {
			chunk.Context.RView.Info("Skip command '" + command + "' due to previous errors")
		}
	}
}
