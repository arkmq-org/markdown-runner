// Package runner is responsible for orchestrating the execution of parsed
// markdown chunks. It manages the execution flow, including stages, parallel
// execution, dependencies, and teardown logic.
package runner

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/parser"
	"github.com/arkmq-org/markdown-runner/runnercontext"
	"github.com/arkmq-org/markdown-runner/stage"
	"github.com/arkmq-org/markdown-runner/view"
)

// RunMD orchestrates the entire execution process for a single markdown file.
// It parses the file to get the stages, then iterates through them, executing
// the chunks in order. It handles parallel execution, dependencies, and teardown
// stages. If the configuration is set, it will also update the markdown file
// with the output of the executed chunks.
//
// file is the name of the markdown file to execute.
// It returns an error if any chunk fails and is not part of a teardown stage.
func RunMD(cfg *config.Config, file string) error {
	var tmpDirs map[string]string = make(map[string]string)
	var terminatingError error
	markdownDir := path.Dir(file)
	fileName := path.Base(file)
	ui := view.NewView(cfg.View)
	ctx := &runnercontext.Context{
		Cfg:   cfg,
		RView: ui,
	}

	ui.StartFile(file)

	stages, err := parser.ExtractStages(ctx, fileName, markdownDir)
	if err != nil {
		ui.EndFile(file, err)
		return err
	}
	if len(stages) == 0 {
		return nil
	}

	for _, currentStage := range stages {
		if cfg.StartFromStage != "" {
			// Check if this is the right file (if file-specific start-from is requested)
			var shouldStart bool
			if cfg.StartFromFile != "" {
				// File-specific start-from: only start if this is the matching file
				// Check various forms: full path, basename, basename without extension
				basename := path.Base(file)
				nameWithoutExt := strings.TrimSuffix(basename, path.Ext(basename))
				shouldStart = (file == cfg.StartFromFile ||
					basename == cfg.StartFromFile ||
					nameWithoutExt == cfg.StartFromFile ||
					strings.HasSuffix(file, cfg.StartFromFile))
			} else {
				// General start-from: start in any file
				shouldStart = true
			}

			if shouldStart && currentStage.Name == cfg.StartFromStage {
				cfg.StartFromStage = "" // Clear the flag so we don't keep skipping
				cfg.StartFromFile = ""  // Clear file as well
			} else {
				// Skip this stage if we haven't reached the start point yet
				continue
			}
		}

		// Handle break-at flag: start interactive mode when we reach the specified stage/chunk
		if cfg.DebugFromStage != "" {
			// Check if this is the right file (if file-specific debugging is requested)
			var shouldDebug bool
			if cfg.DebugFromFile != "" {
				// File-specific debugging: only debug if this is the matching file
				// Check various forms: full path, basename, basename without extension
				basename := path.Base(file)
				nameWithoutExt := strings.TrimSuffix(basename, path.Ext(basename))
				shouldDebug = (file == cfg.DebugFromFile ||
					basename == cfg.DebugFromFile ||
					nameWithoutExt == cfg.DebugFromFile ||
					strings.HasSuffix(file, cfg.DebugFromFile))
			} else {
				// General debugging: debug in any file
				shouldDebug = true
			}

			if shouldDebug && currentStage.Name == cfg.DebugFromStage {
				if cfg.DebugFromChunk != "" {
					// Specific chunk requested - validate it exists
					_, err := findChunkByIdOrIndex(currentStage, cfg.DebugFromChunk)
					if err != nil {
						ui.EndFile(file, err)
						return err
					}
					// Pass the chunk ID to the stage for targeted debugging
					currentStage.DebugFromChunk = cfg.DebugFromChunk
				} else {
					// Debug entire stage
					cfg.Interactive = true
				}
				cfg.DebugFromStage = "" // Clear the flag so we don't keep enabling debug mode
				cfg.DebugFromChunk = "" // Clear chunk ID as well
				cfg.DebugFromFile = ""  // Clear file as well
			}
		}

		ui.StartStage(currentStage.Name, len(currentStage.Chunks), cfg.Verbose)

		var err error

		err = currentStage.Execute(stages, tmpDirs, terminatingError)
		if err != nil {
			terminatingError = err
		}

	}

	if cfg.UpdateFile && terminatingError == nil {
		terminatingError = parser.UpdateChunkOutput(fileName, markdownDir, stages)
		if terminatingError == nil {
			os.Rename(path.Join(markdownDir, fileName+".out"), path.Join(markdownDir, fileName))
		}
	}

	for _, tmpDir := range tmpDirs {
		os.RemoveAll(tmpDir)
	}

	ui.EndFile(file, terminatingError)
	return terminatingError
}

// findChunkByIdOrIndex finds a chunk in a stage by either its ID or by index (0-based)
func findChunkByIdOrIndex(stage *stage.Stage, identifier string) (*chunk.ExecutableChunk, error) {
	// Try to parse as integer index first
	if index, err := strconv.Atoi(identifier); err == nil {
		if index < 0 || index >= len(stage.Chunks) {
			return nil, fmt.Errorf("chunk index %d is out of range (0-%d) in stage '%s'",
				index, len(stage.Chunks)-1, stage.Name)
		}
		return stage.Chunks[index], nil
	}

	// Try to find by ID
	for _, chunk := range stage.Chunks {
		if chunk.Id == identifier {
			return chunk, nil
		}
	}

	return nil, fmt.Errorf("chunk with ID '%s' not found in stage '%s'", identifier, stage.Name)
}
