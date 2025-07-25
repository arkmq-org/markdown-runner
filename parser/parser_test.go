package parser

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/runnercontext"
	"github.com/arkmq-org/markdown-runner/view"
	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	t.Run("extract stages", func(t *testing.T) {
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

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Failed to extract stages")
		assert.Len(t, stages, 2, "Expected 2 stages")
		assert.Equal(t, "test1", stages[0].Name, "Stages were not extracted correctly")
		assert.Equal(t, "test2", stages[1].Name, "Stages were not extracted correctly")
		assert.Equal(t, `echo "hello"`, stages[0].Chunks[0].Content[0], "Unexpected content")
	})
	t.Run("extract stages with existing output", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		mdContent := `
` + "```" + `bash {"stage":"test1"}
echo "hello"
` + "```" + `
` + "```" + `shell markdown_runner
previous output
` + "```" + `
` + "```" + `bash {"stage":"test1"}
echo "world"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Failed to extract stages")
		assert.Len(t, stages, 1, "Expected 1 stage")
		assert.Len(t, stages[0].Chunks, 2, "Expected 2 chunks in the stage")
		assert.Equal(t, `echo "hello"`, stages[0].Chunks[0].Content[0])
		assert.Equal(t, `echo "world"`, stages[0].Chunks[1].Content[0])
	})
	t.Run("extract stages errors", func(t *testing.T) {
		testCases := []struct {
			name        string
			mdContent   string
			expectError bool
		}{
			{
				name:        "Malformed JSON",
				mdContent:   "```bash {stage}\n```",
				expectError: true,
			},
			{
				name:        "Schema validation error",
				mdContent:   "```bash {\"stage\":\"test\", \"invalid_prop\":\"test\"}\n```",
				expectError: true,
			},
			{
				name:        "Inconsistent parallelism",
				mdContent:   "```bash {\"stage\":\"test\", \"parallel\":true}\n```\n```bash {\"stage\":\"test\"}\n```",
				expectError: true,
			},
			{
				name:        "Invalid JSON",
				mdContent:   "```bash {invalid-json}\n```",
				expectError: true,
			},
			{
				name:        "Writer with no destination",
				mdContent:   "```bash {\"stage\":\"test\", \"runtime\":\"writer\"}\n```",
				expectError: true,
			},
			{
				name:        "Missing stage",
				mdContent:   "```bash {\"invalid_prop\":\"test\"}\n```",
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tmpDir, err := os.MkdirTemp("", "test")
				assert.NoError(t, err, "Failed to create temp dir")
				defer os.RemoveAll(tmpDir)

				mdFile := path.Join(tmpDir, "test.md")
				err = os.WriteFile(mdFile, []byte(tc.mdContent), 0o644)
				assert.NoError(t, err, "Failed to write to temp file")

				cfg := &config.Config{}
				ui := view.NewMock()
				ctx := &runnercontext.Context{
					Cfg: cfg,
					UI:  ui,
				}
				_, err = ExtractStages(ctx, "test.md", tmpDir)
				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
	t.Run("extract stages file not found", func(t *testing.T) {
		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		_, err := ExtractStages(ctx, "nonexistent.md", "anydir")
		assert.Error(t, err, "Expected an error for a nonexistent file, but got none")
	})
	t.Run("update chunk output", func(t *testing.T) {
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

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Failed to extract stages")
		stages[0].Chunks[0].Commands = []*chunk.RunningCommand{
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
	})
	t.Run("update chunk output failing chunk", func(t *testing.T) {
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

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Failed to extract stages")
		stages[0].Chunks[0].Commands = []*chunk.RunningCommand{
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
	})
	t.Run("update chunk output stderr no newline", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		mdContent := "```bash {\"stage\":\"test\"}\ni-will-fail\n```"
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Failed to extract stages")
		stages[0].Chunks[0].Commands = []*chunk.RunningCommand{
			{
				Stderr: "this is an error on stderr",
			},
		}

		err = UpdateChunkOutput("test.md", tmpDir, stages)
		assert.NoError(t, err, "Unexpected error")

		updatedContent, err := os.ReadFile(path.Join(tmpDir, "test.md.out"))
		assert.NoError(t, err, "Failed to read updated file")

		expectedContent := "```shell markdown_runner\nthis is an error on stderr\n```"
		assert.Contains(t, string(updatedContent), expectedContent)
	})
	t.Run("update chunk output multiple commands one output", func(t *testing.T) {
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

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Failed to extract stages")
		stages[0].Chunks[0].Commands = []*chunk.RunningCommand{
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
	})
	t.Run("update chunk output with existing output", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		mdContent := `
# Title
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
` + "```" + `shell markdown_runner
old output
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Failed to extract stages")

		// Simulate command execution
		stages[0].Chunks[0].Commands = []*chunk.RunningCommand{
			{Stdout: "new output\n"},
		}

		err = UpdateChunkOutput("test.md", tmpDir, stages)
		assert.NoError(t, err, "Unexpected error")

		updatedContent, err := os.ReadFile(path.Join(tmpDir, "test.md.out"))
		assert.NoError(t, err, "Failed to read updated file")

		expectedContent := `
# Title
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
` + "```" + `shell markdown_runner
new output
` + "```" + `
`
		assert.Equal(t, strings.TrimSpace(expectedContent), strings.TrimSpace(string(updatedContent)))
	})
	t.Run("update chunk output read only dir", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(""), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = os.Chmod(tmpDir, 0o444)
		assert.NoError(t, err, "Failed to change directory permissions")

		err = UpdateChunkOutput("test.md", tmpDir, nil)
		assert.Error(t, err, "Expected an error when writing to a read-only directory")
	})
	t.Run("init chunk error", func(t *testing.T) {
		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		_, err := initChunk(ctx, `{"stage":"test", "runtime":"writer"}`)
		assert.Error(t, err, "Expected an error for a writer chunk without a destination")
	})
	t.Run("extract stages inconsistent parallelism", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		mdContent := `
` + "```" + `bash {"stage":"test", "parallel":true}
echo "hello"
` + "```" + `
` + "```" + `bash {"stage":"test"}
echo "world"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		_, err = ExtractStages(ctx, "test.md", tmpDir)
		assert.Error(t, err, "Expected an error for inconsistent parallelism")
	})
	t.Run("extract stages mismatched fences", func(t *testing.T) {
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

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Failed to extract stages")
		assert.Len(t, stages, 1, "Expected 1 stage")
		assert.Len(t, stages[0].Chunks[0].Content, 2, "Should have two lines of content")
	})
	t.Run("count opening back quotes", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected int
		}{
			{
				name:     "Three backticks",
				input:    "```bash {\"stage\":\"test\"}",
				expected: 3,
			},
			{
				name:     "Four backticks",
				input:    "````bash {\"stage\":\"test\"}",
				expected: 4,
			},
			{
				name:     "No backticks",
				input:    "bash {\"stage\":\"test\"}",
				expected: 0,
			},
			{
				name:     "Backticks in the middle",
				input:    "bash ``` {\"stage\":\"test\"}",
				expected: 0,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				actual := countOpeningBackQuotes(tc.input)
				assert.Equal(t, tc.expected, actual)
			})
		}
	})
	t.Run("extract stages no chunks", func(t *testing.T) {
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

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Should not return an error for a file with no chunks")
		assert.Len(t, stages, 0, "Expected 0 stages")
	})
	t.Run("extract stages chunk with no content", func(t *testing.T) {
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

		cfg := &config.Config{}
		ui := view.NewMock()
		ctx := &runnercontext.Context{
			Cfg: cfg,
			UI:  ui,
		}
		stages, err := ExtractStages(ctx, "test.md", tmpDir)
		assert.NoError(t, err, "Should not return an error for a chunk with no content")
		assert.Len(t, stages, 1, "Expected 1 stage")
		assert.Len(t, stages[0].Chunks, 1, "Expected 1 chunk in the stage")
		assert.Empty(t, stages[0].Chunks[0].Content, "Expected the chunk content to be empty")
	})
}
