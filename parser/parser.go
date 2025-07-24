// Package parser is responsible for parsing markdown files to find and extract
// executable code chunks. It reads the markdown content, identifies the code
// fences with the special metadata, and groups them into stages for execution.
package parser

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/stage"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	CHUNK_REGEX        = "^```+[a-zA-Z0-9_\\-. ]*\\{.*\\}.*$"
	OUTPUT_CHUNK_REGEX = "^```+shell markdown_runner$"
)

var (
	chunkMatcher, _       = regexp.Compile(CHUNK_REGEX)
	outputChunkMatcher, _ = regexp.Compile(OUTPUT_CHUNK_REGEX)
)

var schema string = `
{
    "type":"object",
    "properties":{
        "stage":{"type":"string", "pattern":"^[a-zA-Z0-9_-]*$"},
        "id":{"type":"string", "pattern":"^[a-zA-Z0-9_-]*$"},
        "requires":{"type":"string", "pattern":"^[a-zA-Z0-9_-]*/[a-zA-Z0-9_-]*$"},
        "rootdir":{"type":"string", "pattern":"^(\\$initial_dir|\\$tmpdir\\.?\\w*)?[\\w\\/\\-\\.]*$"},
        "runtime":{"enum": ["bash", "writer"]},
        "parallel":{"type":"boolean"},
        "breakpoint":{"type":"boolean"},
        "destination":{"type":"string", "pattern":"^[\\w\\/\\-\\.]*$"},
        "label":{"type":"string", "pattern":"^[a-zA-Z0-9_\\-: ]*$"}
    },
    "required":["stage"],
    "additionalProperties": false
}
`

// initChunk unmarshals the JSON metadata from a code fence into an
// ExecutableChunk struct and initializes it.
//
// params is the raw JSON string from the code fence.
// It returns the initialized ExecutableChunk and an error if unmarshalling fails.
func initChunk(cfg *config.Config, params string) (*chunk.ExecutableChunk, error) {
	var chunk chunk.ExecutableChunk
	err := json.Unmarshal([]byte(params), &chunk)
	chunk.Cfg = cfg
	chunk.Init()
	if chunk.Runtime == "writer" {
		if chunk.Destination == "" {
			return nil, errors.New("a writer runtime requires a destination property")
		}
	}
	return &chunk, err
}

// ExtractStages reads a markdown file from disk, scans it for executable
// code chunks, and groups them by their defined stage. It ignores any code
// blocks that were previously generated as output by this tool.
//
// file is the name of the markdown file to parse.
// markdownDir is the directory containing the markdown file.
// It returns a slice of Stages, where each Stage represents the chunks to be
// executed, and an error if parsing fails.
func ExtractStages(cfg *config.Config, file string, markdownDir string) ([]*stage.Stage, error) {
	var chunkStages [][]*chunk.ExecutableChunk
	var fileHandle *os.File
	filepath := path.Join(markdownDir, file)
	fileHandle, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()
	scanner := bufio.NewScanner(fileHandle)
	// regex for the chunk opening fence
	var isInChunk bool = false
	var chunkStopFend *regexp.Regexp // a regex to match when the chunk is ending
	var chunkBackQuotesCount int = 0 // with its number of back quotes

	// regex for output opening fence
	var isInFormerOutputChunk bool = false // when set to true all lines from input are ignored
	var outputChunkStopFend *regexp.Regexp // a regex to match when the chunk is ending
	var outputChunkBackQuotesCount int = 0 // with its number of back quotes

	var currentStageName string = ""
	var currentChunk *chunk.ExecutableChunk

	lineCounter := 0

	sch, err := jsonschema.CompileString("schema.json", schema)
	if err != nil {
		return nil, err
	}
	for scanner.Scan() {
		lineCounter += 1
		// when we encounter the previous output chunk, we ignore everything until the corresponding fence closing
		if !isInChunk && !isInFormerOutputChunk && outputChunkMatcher.Match(scanner.Bytes()) {
			isInFormerOutputChunk = true
			outputChunkBackQuotesCount = countOpeningBackQuotes(scanner.Text())
			outputChunkStopFend, err = regexp.Compile(fmt.Sprintf("(?m)^`{%d}$", outputChunkBackQuotesCount))
			if err != nil {
				return nil, err
			}
			continue
		}
		if !isInChunk && isInFormerOutputChunk && outputChunkStopFend.Match(scanner.Bytes()) {
			isInFormerOutputChunk = false
			continue
		}
		// When we detect a chunk we compute how many backticks are needed to find its end
		if !isInChunk && !isInFormerOutputChunk && chunkMatcher.Match(scanner.Bytes()) {
			chunkBackQuotesCount = countOpeningBackQuotes(scanner.Text())
			chunkStopFend, err = regexp.Compile(fmt.Sprintf("(?m)^`{%d}$", chunkBackQuotesCount))
			if err != nil {
				return nil, err
			}
			isInChunk = true
			raw := scanner.Text()
			params := raw[strings.Index(raw, "{"):]
			var v interface{}
			if err := json.Unmarshal([]byte(params), &v); err != nil {
				return nil, fmt.Errorf("JSON unmarshal error in %s at line %d: %w in %s", file, lineCounter, err, params)
			}
			if err = sch.Validate(v); err != nil {
				return nil, fmt.Errorf("JSON validation error in %s at line %d: %w in %s", file, lineCounter, err, params)
			}
			currentChunk, err = initChunk(cfg, params)
			if err != nil {
				return nil, fmt.Errorf("chunk initialization error in %s at line %d: %w in %s", file, lineCounter, err, params)
			}
			if currentStageName != currentChunk.Stage {
				chunkStages = append(chunkStages, []*chunk.ExecutableChunk{})
				currentStageName = currentChunk.Stage
			}
			chunkStages[len(chunkStages)-1] = append(chunkStages[len(chunkStages)-1], currentChunk)
			continue
		}
		// when the end is detected, it's time to write the new output
		if isInChunk && !isInFormerOutputChunk {
			if chunkStopFend.Match(scanner.Bytes()) {
				isInChunk = false
			} else {
				currentChunk.Content = append(currentChunk.Content, scanner.Text())
			}
		}
	}
	var stages []*stage.Stage
	for _, chunks := range chunkStages {
		if s := stage.NewStage(cfg, chunks); s != nil {
			if !s.IsParallelismConsistent() {
				return nil, errors.New("inconsistent parallelism found in stage " + s.Name)
			}
			stages = append(stages, s)
		}
	}
	return stages, nil
}

// UpdateChunkOutput rewrites the given markdown file, inserting the captured
// output of each executed chunk directly after its corresponding code block.
// It writes to a temporary ".out" file first and expects the caller to rename it.
//
// file is the name of the markdown file to update.
// markdownDir is the directory where the file is located.
// stages contains the executed chunks with the output to be written.
// It returns an error if any file operations fail.
func UpdateChunkOutput(file string, markdownDir string, stages []*stage.Stage) error {
	// get the path for the input file
	inFPath := path.Join(markdownDir, file)
	inFile, err := os.Open(inFPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	// write to a temporary file
	outFPath := path.Join(markdownDir, file+".out")
	outFile, err := os.Create(outFPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	scanner := bufio.NewScanner(inFile)

	// regex for the chunk opening fence
	var isInChunk bool = false
	var chunkStopFend *regexp.Regexp // a regex to match when the chunk is ending
	var chunkBackQuotesCount int = 0 // with its number of back quotes

	// regex for output opening fence
	var isInFormerOutputChunk bool = false // when set to true all lines from input are ignored
	var outputChunkStopFend *regexp.Regexp // a regex to match when the chunk is ending
	var outputChunkBackQuotesCount int = 0 // with its number of back quotes

	var writeNewOutput bool = false // when set to true, the stdout is written in a new output chunk

	var currentStageIndex int = 0
	var currentChunkIndex int = 0

	for scanner.Scan() {
		// when we encounter the previous output chunk, we ignore everything until the corresponding fence closing
		if !isInChunk && !isInFormerOutputChunk && outputChunkMatcher.Match(scanner.Bytes()) {
			isInFormerOutputChunk = true
			outputChunkBackQuotesCount = countOpeningBackQuotes(scanner.Text())
			outputChunkStopFend, err = regexp.Compile(fmt.Sprintf("(?m)^`{%d}$", outputChunkBackQuotesCount))
			if err != nil {
				return err
			}
			continue
		}
		if !isInFormerOutputChunk {
			_, err = writer.WriteString(scanner.Text() + "\n")
			if err != nil {
				return err
			}
		}
		if !isInChunk && isInFormerOutputChunk && outputChunkStopFend.Match(scanner.Bytes()) {
			isInFormerOutputChunk = false
		}
		// When we detect a chunk we compute how many backticks are needed to find its end
		if !isInChunk && !isInFormerOutputChunk && chunkMatcher.Match(scanner.Bytes()) {
			chunkBackQuotesCount = countOpeningBackQuotes(scanner.Text())
			chunkStopFend, err = regexp.Compile(fmt.Sprintf("(?m)^`{%d}$", chunkBackQuotesCount))
			if err != nil {
				return err
			}
			isInChunk = true
			continue
		}
		// when the end is detected, it's time to write the new output
		if isInChunk && chunkStopFend.Match(scanner.Bytes()) {
			isInChunk = false
			writeNewOutput = true
		}
		if writeNewOutput {
			writeNewOutput = false
			// get the current chunk to write
			stage := stages[currentStageIndex]
			chunk := stage.Chunks[currentChunkIndex]
			if chunk.HasOutput() {
				chunk.WriteOutputTo(chunkBackQuotesCount, writer)
			}
			// Compute the next chunk and stage
			currentChunkIndex += 1
			if currentChunkIndex == len(stages[currentStageIndex].Chunks) {
				currentChunkIndex = 0
				currentStageIndex += 1
			}
		}
	}
	return writer.Flush()
}

// countOpeningBackQuotes is a helper function that counts the number of
// backticks at the beginning of a string. This is used to find the matching
// closing fence for a code block.
func countOpeningBackQuotes(str string) int {
	total := 0
	// count the total amount of backquotes to be able to find the closing fence
	after, found := strings.CutPrefix(str, "`")
	for found {
		total += 1
		after, found = strings.CutPrefix(after, "`")
	}
	return total
}
