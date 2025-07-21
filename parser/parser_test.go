package parser

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/stretchr/testify/assert"
)

func TestExtractStages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := `
# Title
` + "```" + `bash {"stage":"test1"}
echo "hello"
` + "```" + `

` + "```" + `bash {"stage":"test2"}
echo "world"
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	stages, err := ExtractStages("test.md", tmpDir)
	assert.NoError(t, err, "Failed to extract stages")
	assert.Len(t, stages, 2, "Expected 2 stages")
	assert.Equal(t, "test1", stages[0][0].Stage, "Stages were not extracted correctly")
	assert.Equal(t, "test2", stages[1][0].Stage, "Stages were not extracted correctly")
	assert.Equal(t, `echo "hello"`, stages[0][0].Content[0], "Unexpected content")
}

func TestExtractStages_FileNotFound(t *testing.T) {
	_, err := ExtractStages("nonexistent.md", "anydir")
	assert.Error(t, err, "Expected an error for a nonexistent file, but got none")
}

func TestExtractStages_MalformedJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := "```bash {stage}\n```"

	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	_, err = ExtractStages("test.md", tmpDir)
	assert.Error(t, err, "Expected a JSON parsing error, but got none")
}

func TestExtractStages_SchemaValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := "```bash {\"stage\":\"test\", \"invalid_prop\":\"test\"}\n```"
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	_, err = ExtractStages("test.md", tmpDir)
	assert.Error(t, err, "Expected a schema validation error, but got none")
}

func TestUpdateChunkOutput(t *testing.T) {
	// This function is complex to test in isolation as it depends on executed chunks.
	// A more comprehensive integration test would be better suited to test this.
	// For now, we'll test the basic file operations.

	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	stages, err := ExtractStages("test.md", tmpDir)
	assert.NoError(t, err, "Failed to extract stages")
	stages[0][0].Commands = []*chunk.RunningCommand{
		{Stdout: "hello\n"},
	}

	err = UpdateChunkOutput("test.md", tmpDir, stages)
	assert.NoError(t, err, "Unexpected error")

	updatedContent, err := os.ReadFile(path.Join(tmpDir, "test.md.out"))
	assert.NoError(t, err, "Failed to read updated file")

	expectedContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
` + "```" + `shell markdown_runner
hello
` + "```" + `
`
	assert.Equal(t, strings.TrimSpace(expectedContent), strings.TrimSpace(string(updatedContent)))
}

func TestUpdateChunkOutput_FailingChunk(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := `
` + "```" + `bash {"stage":"test"}
i-will-fail
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	stages, err := ExtractStages("test.md", tmpDir)
	assert.NoError(t, err, "Failed to extract stages")
	stages[0][0].Commands = []*chunk.RunningCommand{
		{
			Stdout: "this is stdout",
			Stderr: "this is an error on stderr\n",
		},
	}

	err = UpdateChunkOutput("test.md", tmpDir, stages)
	assert.NoError(t, err, "Unexpected error")

	updatedContent, err := os.ReadFile(path.Join(tmpDir, "test.md.out"))
	assert.NoError(t, err, "Failed to read updated file")

	expectedContent := `
` + "```" + `bash {"stage":"test"}
i-will-fail
` + "```" + `
` + "```" + `shell markdown_runner
this is stdout
this is an error on stderr
` + "```" + `
`
	assert.Equal(t, strings.TrimSpace(expectedContent), strings.TrimSpace(string(updatedContent)))
}

func TestUpdateChunkOutput_MultipleCommandsOneOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := `
` + "```" + `bash {"stage":"test"}
echo "no output" > /dev/null
echo "output"
echo "no output either" > /dev/null
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	stages, err := ExtractStages("test.md", tmpDir)
	assert.NoError(t, err, "Failed to extract stages")
	stages[0][0].Commands = []*chunk.RunningCommand{
		{Stdout: ""},
		{Stdout: "output\n"},
		{Stdout: ""},
	}

	err = UpdateChunkOutput("test.md", tmpDir, stages)
	assert.NoError(t, err, "Unexpected error")

	updatedContent, err := os.ReadFile(path.Join(tmpDir, "test.md.out"))
	assert.NoError(t, err, "Failed to read updated file")

	expectedContent := `
` + "```" + `bash {"stage":"test"}
echo "no output" > /dev/null
echo "output"
echo "no output either" > /dev/null
` + "```" + `
` + "```" + `shell markdown_runner
output
` + "```" + `
`
	assert.Equal(t, strings.TrimSpace(expectedContent), strings.TrimSpace(string(updatedContent)))
}

func TestExtractStages_MismatchedFences(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := `
# Title
` + "````" + `bash {"stage":"test1"}
echo "hello"
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	stages, err := ExtractStages("test.md", tmpDir)
	assert.NoError(t, err, "Failed to extract stages")
	assert.Len(t, stages, 1, "Expected 1 stage")
	assert.Len(t, stages[0][0].Content, 2, "Should have two lines of content")
}

func TestExtractStages_NoChunks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := `
# Title
This is a markdown file with no executable chunks.
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	stages, err := ExtractStages("test.md", tmpDir)
	assert.NoError(t, err, "Should not return an error for a file with no chunks")
	assert.Len(t, stages, 0, "Expected 0 stages")
}

func TestExtractStages_MissingStage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := "```bash {\"invalid_prop\":\"test\"}\n```"
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	_, err = ExtractStages("test.md", tmpDir)
	assert.Error(t, err, "Expected a schema validation error for missing stage, but got none")
}

func TestExtractStages_ChunkWithNoContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := `
` + "```" + `bash {"stage":"test1"}
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	stages, err := ExtractStages("test.md", tmpDir)
	assert.NoError(t, err, "Should not return an error for a chunk with no content")
	assert.Len(t, stages, 1, "Expected 1 stage")
	assert.Len(t, stages[0], 1, "Expected 1 chunk in the stage")
	assert.Empty(t, stages[0][0].Content, "Expected the chunk content to be empty")
}

func TestExtractStages_InvalidChunkMetadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := "```bash {\"stage\":\"test\", \"invalid_prop\":\"test\"}\n```"
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	_, err = ExtractStages("test.md", tmpDir)
	assert.Error(t, err, "Expected a schema validation error for additional properties, but got none")
}
