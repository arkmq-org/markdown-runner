// Package stage defines the structure for execution stages in the markdown runner.
package stage

import (
	"strings"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/pterm/pterm"
)

// Stage represents a single stage in the execution pipeline, containing a list
// of chunks to be executed.
type Stage struct {
	Name       string
	IsParallel bool // Indicates if the stage is parallel or sequential
	Chunks     []*chunk.ExecutableChunk
}

// NewStage creates a new stage from a list of chunks. It assumes all chunks
// belong to the same stage and extracts the stage name from the first chunk.
func NewStage(chunks []*chunk.ExecutableChunk) *Stage {
	if len(chunks) == 0 {
		return nil
	}
	isParallel := false
	for _, chunk := range chunks {
		if chunk.IsParallel {
			isParallel = true
		}
	}
	return &Stage{
		Name:       chunks[0].Stage,
		Chunks:     chunks,
		IsParallel: isParallel,
	}
}

// IsParallelismConsistent checks that all chunks in a stage are either all
// parallel or all sequential. It ensures that there's no mix of execution
// modes within a single stage.
func (s *Stage) IsParallelismConsistent() bool {
	// atLeastOneParallel will be true if a chunk is set to run in parallel
	atLeastOneParallel := false
	// atLeastOneSequential will be true if a chunk is set to run sequentially
	atLeastOneSequential := false
	for _, chunk := range s.Chunks {
		atLeastOneParallel = atLeastOneParallel || chunk.IsParallel
		atLeastOneSequential = atLeastOneSequential || !chunk.IsParallel
	}
	// return true only if all the chunks are in the same parallelism mode
	if atLeastOneParallel {
		// when at least one is parallel, we don't want any sequential chunks
		return !atLeastOneSequential
	}
	// Otherwise, there's no parallel chunk, so having at least a sequential means they're all sequential
	return atLeastOneSequential
}

// Execute runs all the chunks within the stage. It first validates that the
// chunks have a consistent parallelism setting. It then iterates through each
// chunk, preparing it for execution (which includes handling dependencies) and
// then running it. If chunks are parallel, it waits for all of them to complete.
// An error is returned if any part of the execution fails.
func (s *Stage) Execute(stages []*Stage, tmpDirs map[string]string) error {
	var towait []*chunk.ExecutableChunk
	var terminatingError error

	var multiPrinter pterm.MultiPrinter
	if s.IsParallel {
		multiPrinter = pterm.DefaultMultiPrinter
	}
	for _, chunk := range s.Chunks {
		// Examine if the chunk can be executed based on previous errors
		if terminatingError != nil && chunk.Stage != "teardown" {
			continue
		}
		// Examine if the tool must be run interactively from this chunk
		if chunk.HasBreakpoint && !config.IngoreBreakpoints {
			config.Interactive = true
		}
		// Examine if the chunk has a particular dependency to another one
		if chunk.Requires != "" {
			reqStageName := strings.Split(chunk.Requires, "/")[0]
			reqChunkId := strings.Split(chunk.Requires, "/")[1]
			reqChunk := FindChunkById(stages, reqStageName, reqChunkId)
			if reqChunk == nil || !reqChunk.HasExecutedCorrectly() {
				continue
			}
		}
		terminatingError = chunk.PrepareForExecution(tmpDirs)
		if terminatingError != nil {
			continue
		}
		// All spinners spinning in parallel must be declared before the first update to their content
		// meaning that we need to differentiate the parallel case from the sequential one.
		// In the sequential case, we want to exercise the right to immediately stop executing more chunks in the stage
		// as soon as one of them fails. In the parallel case, we have to start all of them at the same time, there's
		// not really a dependency between a chunk and another. That's why the loop is broken in two pieces.
		// Either we execute, block, and wait for a chunk's result to move on.
		// Or we prepare the UI, and then execute all the chunks alongside each other.
		if chunk.IsParallel {
			// initialize the spinners
			err := chunk.Ignite(&multiPrinter)
			if err != nil {
				terminatingError = err
			}
		} else {
			err := chunk.Execute()
			if err != nil {
				terminatingError = err
			}
		}
	}
	// When In parallel, we start and wait for every chunks.
	if s.IsParallel {
		multiPrinter.Start()
		pterm.Debug.Println("Running stage " + s.Name + " with " + pterm.Green(len(s.Chunks)) + " chunks in parallel")
		for _, chunk := range s.Chunks {
			// start the chunk
			pterm.Debug.Println("Starting chunk " + pterm.Green(chunk.Id) + " in stage " + pterm.Green(s.Name))
			err := chunk.Start()
			if err != nil {
				terminatingError = err
			}
			towait = append(towait, chunk)
		}
		// wait or kill (in case an error occurred)
		for _, chunk := range towait {
			pterm.Debug.Println("Waiting for chunk " + pterm.Green(chunk.Id) + " in stage " + pterm.Green(s.Name))
			err := chunk.Wait(terminatingError != nil)
			if err != nil {
				terminatingError = err
			}
		}
		multiPrinter.Stop()
	}
	return terminatingError
}

// FindChunkById searches through a list of stages to find a chunk with a
// specific ID within a given stage.
//
// stages is the 2D slice of all chunks.
// stageName is the name of the stage to search in.
// chunkId is the ID of the chunk to find.
// It returns a pointer to the found ExecutableChunk, or nil if not found.
func FindChunkById(stages []*Stage, stageName string, chunkId string) *chunk.ExecutableChunk {
	for _, stage := range stages {
		if stage.Name != stageName {
			continue
		}
		for _, chunk := range stage.Chunks {
			if chunk.Id == chunkId {
				return chunk
			}
		}
	}
	return nil
}
