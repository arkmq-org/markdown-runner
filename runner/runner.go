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

	// Handle line breakpoints: hand the chunk around each requested line over
	// to its stage. Only the chunks that --start-from lets run are considered,
	// so that a target located in a skipped stage or chunk is not silently
	// consumed.
	if err := resolveLineBreakpoints(cfg, file, stages); err != nil {
		ui.EndFile(file, err)
		return err
	}

	for _, currentStage := range stages {
		// Handle start-from flag: skip everything before the specified stage/chunk
		if skip, err := applyStartFrom(cfg, file, currentStage); err != nil {
			ui.EndFile(file, err)
			return err
		} else if skip {
			continue
		}

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

// applyStartFrom tells whether the given stage of the given file precedes the
// --start-from location and must therefore be skipped. Once the location is
// reached, it is cleared so that the following stages run normally and, when a
// chunk is given, it is handed over to the stage so that it skips the chunks
// preceding it.
// It returns an error if the targeted chunk doesn't exist in the stage.
func applyStartFrom(cfg *config.Config, file string, currentStage *stage.Stage) (bool, error) {
	loc := cfg.StartFromLocation
	if loc.Stage == "" {
		return false, nil
	}
	if loc.Stage != currentStage.Name || !matchesFile(file, loc.File) {
		return true, nil
	}
	if loc.Chunk != "" {
		// Specific chunk requested - validate it exists
		if _, err := findChunkByIdOrIndex(currentStage, loc.Chunk); err != nil {
			return false, err
		}
		currentStage.StartFromChunk = loc.Chunk
	}
	cfg.StartFromLocation = config.Location{}
	return false, nil
}

// runnableChunk designates a chunk by its stage and its index in that stage.
type runnableChunk struct {
	stage *stage.Stage
	index int
}

// runnableChunks lists, in document order, the chunks of the given file that
// the --start-from location does not skip: the chunks of the stages preceding
// the start stage and, when a chunk is given, the chunks preceding it in the
// start stage are left out. The location is not consumed, applyStartFrom does
// it when the stages run.
// It returns an error if the targeted chunk doesn't exist in the start stage.
func runnableChunks(cfg *config.Config, file string, stages []*stage.Stage) ([]runnableChunk, error) {
	loc := cfg.StartFromLocation
	started := loc.Stage == ""
	var chunks []runnableChunk
	for _, s := range stages {
		first := 0
		if !started {
			if loc.Stage != s.Name || !matchesFile(file, loc.File) {
				continue
			}
			started = true
			if loc.Chunk != "" {
				if _, err := findChunkByIdOrIndex(s, loc.Chunk); err != nil {
					return nil, err
				}
				first = s.IndexOfChunk(loc.Chunk)
			}
		}
		for i := first; i < len(s.Chunks); i++ {
			chunks = append(chunks, runnableChunk{stage: s, index: i})
		}
	}
	return chunks, nil
}

// resolveLineBreakpoints consumes the line breakpoints (--break-before,
// --break-after, -B line) targeting the given file. Each one is resolved,
// among the chunks that will run (see runnableChunks), to one chunk whose
// index is handed over to its stage as a chunk breakpoint. A breakpoint
// without file part that matches no chunk is kept for the next files.
// It returns an error if a breakpoint restricted to this file matches no
// chunk.
func resolveLineBreakpoints(cfg *config.Config, file string, stages []*stage.Stage) error {
	hasLine := false
	for _, bp := range cfg.Breakpoints {
		hasLine = hasLine || (bp.Line != 0 && matchesFile(file, bp.File))
	}
	if !hasLine {
		return nil
	}
	chunks, err := runnableChunks(cfg, file, stages)
	if err != nil {
		return err
	}
	total := 0
	for _, s := range stages {
		total += len(s.Chunks)
	}
	var remaining []config.Location
	for _, bp := range cfg.Breakpoints {
		if bp.Line == 0 || !matchesFile(file, bp.File) {
			remaining = append(remaining, bp)
			continue
		}
		target, ok := findChunkByLine(chunks, bp.Line, bp.After)
		if !ok {
			if bp.File == "" {
				remaining = append(remaining, bp)
				continue
			}
			err := fmt.Errorf("no chunk found %s line %d in %s", lineDirection(bp.After), bp.Line, file)
			if len(chunks) != total {
				err = fmt.Errorf("%w (chunks skipped by --start-from are ignored)", err)
			}
			return err
		}
		target.stage.DebugFromChunks = append(target.stage.DebugFromChunks, strconv.Itoa(target.index))
	}
	cfg.Breakpoints = remaining
	return nil
}

// findChunkByLine returns the chunk to break at for a given line of the
// markdown file, among the given chunks in document order. A chunk whose
// fences enclose the line is always selected. Otherwise, when after is true,
// the first chunk opening after the line is selected, and when after is false,
// the last chunk opening before the line.
// It returns false when no chunk matches.
func findChunkByLine(chunks []runnableChunk, line int, after bool) (runnableChunk, bool) {
	var before runnableChunk
	found := false
	for _, rc := range chunks {
		c := rc.stage.Chunks[rc.index]
		if c.Line <= line && line <= c.EndLine {
			return rc, true
		}
		if after && c.Line > line {
			return rc, true
		}
		if !after && c.Line < line {
			before, found = rc, true
		}
	}
	return before, found
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
	var remaining []config.Location
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
