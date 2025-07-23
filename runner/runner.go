// Package runner is responsible for orchestrating the execution of parsed
// markdown chunks. It manages the execution flow, including stages, parallel
// execution, dependencies, and teardown logic.
package runner

import (
	"os"
	"path"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/parser"
	"github.com/pterm/pterm"
)

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

	for _, currentStage := range stages {
		if config.StartFrom != "" {
			if currentStage.Name != config.StartFrom {
				continue
			} else {
				config.StartFrom = ""
			}
		}
		if config.Verbose {
			pterm.DefaultSection.WithLevel(2).Printf("stage %s with %d chunks\n", currentStage.Name, len(currentStage.Chunks))
		}
		var err error

		err = currentStage.Execute(stages, tmpDirs)
		if err != nil {
			terminatingError = err
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
