// Package runner is responsible for orchestrating the execution of parsed
// markdown chunks. It manages the execution flow, including stages, parallel
// execution, dependencies, and teardown logic.
package runner

import (
	"os"
	"path"
	"strings"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/parser"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

// findChunkById searches through a list of stages to find a chunk with a
// specific ID within a given stage.
//
// stages is the 2D slice of all chunks.
// stageName is the name of the stage to search in.
// chunkId is the ID of the chunk to find.
// It returns a pointer to the found ExecutableChunk, or nil if not found.
func findChunkById(stages [][]*chunk.ExecutableChunk, stageName string, chunkId string) *chunk.ExecutableChunk {
	for _, stage := range stages {
		if stage[0].Stage != stageName {
			continue
		}
		for _, chunk := range stage {
			if chunk.Id == chunkId {
				return chunk
			}
		}
	}
	return nil
}

// RunMD orchestrates the entire execution process for a single markdown file.
// It parses the file to get the stages, then iterates through them, executing
// the chunks in order. It handles parallel execution, dependencies, and teardown
// stages. If the configuration is set, it will also update the markdown file
// with the output of the executed chunks.
//
// file is the name of the markdown file to execute.
// It returns an error if any chunk fails and is not part of a teardown stage.
func RunMD(file string) error {
	var tmpDirs map[string]string = make(map[string]string)
	var terminatingError error

	stages, err := parser.ExtractStages(file, config.MarkdownDir)
	if err != nil {
		return err
	}
	if len(stages) == 0 {
		return nil
	}
	pterm.DefaultSection.Println("Testing " + file)

	for _, chunks := range stages {
		if config.StartFrom != "" {
			if chunks[0].Stage != config.StartFrom {
				continue
			} else {
				config.StartFrom = ""
			}
		}
		if config.Verbose {
			pterm.DefaultSection.WithLevel(2).Printf("stage %s with %d chunks\n", chunks[0].Stage, len(chunks))
		}
		var towait []*chunk.ExecutableChunk
		for _, chunk := range chunks {
			if terminatingError != nil && chunk.Stage != "teardown" {
				continue
			}
			if chunk.HasBreakpoint && !config.IngoreBreakpoints {
				config.Interactive = true
			}

			if chunk.Requires != "" {
				reqStageName := strings.Split(chunk.Requires, "/")[0]
				reqChunkId := strings.Split(chunk.Requires, "/")[1]
				reqChunk := findChunkById(stages, reqStageName, reqChunkId)
				if reqChunk == nil || !reqChunk.HasExecutedCorrectly() {
					continue
				}
			}

			var hasCreationError bool = false
			if chunk.Runtime == "writer" {
				writerString := "writing " + chunk.Destination + " on disk"
				if chunk.Label != "" {
					writerString += " for " + chunk.Label
				}
				spiner, _ := pterm.DefaultSpinner.Start(writerString)
				directory, err := chunk.GetOrCreateRuntimeDirectory(tmpDirs)
				if err != nil {
					spiner.Fail(err.Error())
					hasCreationError = true
					terminatingError = err
					continue
				}
				err = chunk.WriteFile(directory)
				if err != nil {
					spiner.Fail(err.Error())
					hasCreationError = true
					terminatingError = err
					continue
				}
				spiner.Success()
			} else if chunk.Runtime == "bash" {
				uuid := uuid.New()
				cmd := "./" + uuid.String() + ".sh"
				command, err := chunk.AddCommandToExecute(cmd, tmpDirs)
				if err != nil {
					hasCreationError = true
					terminatingError = err
					continue
				}
				err = chunk.WriteBashScript(command.Cmd.Dir, cmd)
				if err != nil {
					hasCreationError = true
					terminatingError = err
					continue
				}
			} else {
				for _, command := range chunk.Content {
					_, err := chunk.AddCommandToExecute(command, tmpDirs)
					if err != nil {
						terminatingError = err
						hasCreationError = true
						break
					}
				}
			}
			if hasCreationError {
				continue
			}
			for commandIndex, command := range chunk.Commands {
				if chunk.IsParallel && commandIndex == len(chunk.Commands)-1 {
					err := command.Start()
					if err != nil {
						terminatingError = err
						break
					}
					towait = append(towait, chunk)
				} else {
					err := command.Execute()
					if err != nil {
						terminatingError = err
						break
					}
				}
			}
		}

		for _, chunk := range towait {
			lastCommand := chunk.Commands[len(chunk.Commands)-1]
			if terminatingError != nil {
				pterm.Warning.Printf("Killing %s\n", lastCommand.CmdPrettyName)
				lastCommand.Cmd.Process.Kill()
			} else {
				err := lastCommand.Wait()
				if err != nil {
					terminatingError = err
				}
			}
		}
	}

	if config.UpdateFile && terminatingError == nil {
		terminatingError = parser.UpdateChunkOutput(file, config.MarkdownDir, stages)
		if terminatingError == nil {
			os.Rename(path.Join(config.MarkdownDir, file+".out"), path.Join(config.MarkdownDir, file))
		}
	}

	for _, tmpDir := range tmpDirs {
		os.RemoveAll(tmpDir)
	}
	return terminatingError
}
