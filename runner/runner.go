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
	return runMD(cfg, file, view.NewView(cfg.View))
}

// runMD is RunMD with the UI given by the caller.
func runMD(cfg *config.Config, file string, ui view.RunnerView) error {
	var tmpDirs map[string]string = make(map[string]string)
	var terminatingError error
	markdownDir := path.Dir(file)
	fileName := path.Base(file)
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

	// Handle start-from flag: drop the stages located before the start point
	runnableStages := filterStartFrom(cfg, file, stages)

	// Handle line breakpoints: hand the chunk around each requested line over
	// to its stage. Only the stages that will run are considered, so that a
	// target located in a stage skipped by --start-from is not silently consumed.
	if err := resolveLineBreakpoints(cfg, file, runnableStages, len(runnableStages) != len(stages)); err != nil {
		ui.EndFile(file, err)
		return err
	}

	for _, currentStage := range runnableStages {
		// Handle break-at flags: start interactive mode when we reach the specified stages/chunks
		if err := applyBreakpoints(cfg, file, currentStage); err != nil {
			ui.EndFile(file, err)
			return err
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

// matchesFile reports whether the markdown file being executed matches the
// file part of a file@stage specification. An empty specification matches any file.
// Check various forms: full path, basename, basename without extension.
func matchesFile(file string, spec string) bool {
	if spec == "" {
		return true
	}
	basename := path.Base(file)
	nameWithoutExt := strings.TrimSuffix(basename, path.Ext(basename))
	return file == spec ||
		basename == spec ||
		nameWithoutExt == spec ||
		strings.HasSuffix(file, spec)
}

// filterStartFrom returns the stages to execute according to the start-from
// flag. Stages located before the requested start stage are dropped. Once the
// start stage is found, the flag is cleared so that the next files run fully.
func filterStartFrom(cfg *config.Config, file string, stages []*stage.Stage) []*stage.Stage {
	if cfg.StartFromStage == "" {
		return stages
	}
	// Check if this is the right file (if file-specific start-from is requested)
	if !matchesFile(file, cfg.StartFromFile) {
		return nil
	}
	for i, s := range stages {
		if s.Name == cfg.StartFromStage {
			cfg.StartFromStage = "" // Clear the flag so we don't keep skipping
			cfg.StartFromFile = ""  // Clear file as well
			return stages[i:]
		}
	}
	// Skip the whole file if we haven't reached the start point yet
	return nil
}

// resolveLineBreakpoints consumes the line breakpoints (--break-before,
// --break-after, -B line) targeting the given file. Each one is resolved to a
// chunk of the given stages, whose index is handed over to its stage as a
// chunk breakpoint. A breakpoint without file part that matches no chunk is
// kept for the next files.
// It returns an error if a breakpoint restricted to this file matches no
// chunk; skipped tells whether stages were dropped by --start-from.
func resolveLineBreakpoints(cfg *config.Config, file string, stages []*stage.Stage, skipped bool) error {
	var remaining []config.Breakpoint
	for _, bp := range cfg.Breakpoints {
		if bp.Line == 0 || !matchesFile(file, bp.File) {
			remaining = append(remaining, bp)
			continue
		}
		targetStage, index := findChunkByLine(stages, bp.Line, bp.After)
		if targetStage == nil {
			if bp.File == "" {
				remaining = append(remaining, bp)
				continue
			}
			err := fmt.Errorf("no chunk found %s line %d in %s", lineDirection(bp.After), bp.Line, file)
			if skipped {
				err = fmt.Errorf("%w (stages skipped by --start-from are ignored)", err)
			}
			return err
		}
		targetStage.DebugFromChunks = append(targetStage.DebugFromChunks, strconv.Itoa(index))
	}
	cfg.Breakpoints = remaining
	return nil
}

// findChunkByLine returns the stage and the index in that stage of the chunk
// to break at for a given line of the markdown file. A chunk whose fences
// enclose the line is always selected. Otherwise, when after is true, the
// first chunk opening after the line is selected, and when after is false,
// the last chunk closing before the line.
// It returns a nil stage when no chunk matches.
func findChunkByLine(stages []*stage.Stage, line int, after bool) (*stage.Stage, int) {
	var beforeStage *stage.Stage
	beforeIndex := 0
	for _, s := range stages {
		for i, c := range s.Chunks {
			if c.Line <= line && line <= c.EndLine {
				return s, i
			}
			if after && c.Line > line {
				return s, i
			}
			if !after && c.Line < line {
				beforeStage, beforeIndex = s, i
			}
		}
	}
	if after {
		return nil, 0
	}
	return beforeStage, beforeIndex
}

// lineDirection returns a human readable description of a line break mode.
func lineDirection(after bool) string {
	if after {
		return "after"
	}
	return "before"
}

// applyBreakpoints consumes the --break-at breakpoints targeting the given
// stage of the given file. A stage breakpoint turns the interactive mode on
// right away, a chunk breakpoint is handed over to the stage so that it turns
// the interactive mode on when reaching that chunk. Each breakpoint is
// consumed once.
// It returns an error if a targeted chunk doesn't exist in the stage.
func applyBreakpoints(cfg *config.Config, file string, currentStage *stage.Stage) error {
	var remaining []config.Breakpoint
	for _, bp := range cfg.Breakpoints {
		if bp.Line != 0 || bp.Stage != currentStage.Name || !matchesFile(file, bp.File) {
			remaining = append(remaining, bp)
			continue
		}
		if bp.Chunk == "" {
			// Debug entire stage
			cfg.Interactive = true
			continue
		}
		// Specific chunk requested - validate it exists
		if _, err := findChunkByIdOrIndex(currentStage, bp.Chunk); err != nil {
			return err
		}
		currentStage.DebugFromChunks = append(currentStage.DebugFromChunks, bp.Chunk)
	}
	cfg.Breakpoints = remaining
	return nil
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
