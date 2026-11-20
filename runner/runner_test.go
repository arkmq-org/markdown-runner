package runner

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/parser"
	"github.com/arkmq-org/markdown-runner/runnercontext"
	"github.com/arkmq-org/markdown-runner/stage"
	"github.com/arkmq-org/markdown-runner/view"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	pterm.DisableOutput()
	code := m.Run()
	pterm.EnableOutput()
	os.Exit(code)
}

func TestRunMD(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Unexpected error")
	})

	t.Run("teardown", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"main", "id":"main-chunk"}
exit 1
` + "```" + `

` + "```" + `bash {"stage":"teardown", "requires":"main/main-chunk"}
echo "teardown"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.Error(t, err, "Expected an error, but got none")
	})

	t.Run("update file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, UpdateFile: true, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Unexpected error")

		updatedContent, err := os.ReadFile(mdFile)
		assert.NoError(t, err, "Failed to read updated file")
		expectedContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
` + "```" + `shell markdown_runner
hello
` + "```" + `
`
		assert.Equal(t, expectedContent, string(updatedContent), "File content is not as expected")
	})

	t.Run("start from", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, StartFromLocation: config.Location{Stage: "stage2"}, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"stage1"}
echo "should not run"
` + "```" + `
` + "```" + `bash {"stage":"stage2"}
echo "should run"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Unexpected error")
	})

	t.Run("start from chunk", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, StartFromLocation: config.Location{Stage: "stage2", Chunk: "second"}, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"stage1", "runtime":"bash"}
touch ` + path.Join(tmpDir, "stage1") + `
` + "```" + `
` + "```" + `bash {"stage":"stage2", "id":"first", "runtime":"bash"}
touch ` + path.Join(tmpDir, "first") + `
` + "```" + `
` + "```" + `bash {"stage":"stage2", "id":"second", "runtime":"bash"}
touch ` + path.Join(tmpDir, "second") + `
` + "```" + `
` + "```" + `bash {"stage":"stage3", "runtime":"bash"}
touch ` + path.Join(tmpDir, "stage3") + `
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Unexpected error")

		for name, expected := range map[string]bool{"stage1": false, "first": false, "second": true, "stage3": true} {
			_, statErr := os.Stat(path.Join(tmpDir, name))
			assert.Equal(t, expected, statErr == nil, "unexpected execution state for chunk %s", name)
		}
	})

	t.Run("start from unknown chunk", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, StartFromLocation: config.Location{Stage: "stage1", Chunk: "missing"}, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"stage1"}
echo "hello"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.Error(t, err, "Expected an error for an unknown chunk")
	})

	t.Run("no stages", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
		mdContent := `
# No Stages Here
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Unexpected error")
	})

	t.Run("parser error", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {stage}
echo "bad json"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.Error(t, err, "Expected a parser error")
	})

	t.Run("teardown with dependency", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)
		outputFile := path.Join(tmpDir, "teardown_check")

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"main", "id":"main-chunk"}
echo "main chunk executed"
` + "```" + `

` + "```" + `bash {"stage":"teardown", "requires":"main/main-chunk", "runtime":"bash"}
echo "teardown executed" > ` + outputFile + `
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Expected no error")

		_, err = os.Stat(outputFile)
		assert.NoError(t, err, "Expected teardown chunk to be executed")
	})
}

func TestApplyBreakpoints(t *testing.T) {
	newStage := func(name string) *stage.Stage {
		return &stage.Stage{
			Name: name,
			Chunks: []*chunk.ExecutableChunk{
				{Id: "first", Stage: name},
				{Id: "second", Stage: name},
			},
		}
	}

	t.Run("a later stage breakpoint prompts again after continue", func(t *testing.T) {
		cfg := &config.Config{Breakpoints: []config.Location{{Stage: "s1"}, {Stage: "s2"}}}
		assert.NoError(t, applyBreakpoints(cfg, "dir/test.md", newStage("s1")))
		assert.True(t, cfg.Interactive)
		// the user answers "continue" during s1
		cfg.Interactive = false
		assert.NoError(t, applyBreakpoints(cfg, "dir/test.md", newStage("s2")))
		assert.True(t, cfg.Interactive, "Expected the s2 breakpoint to turn prompting back on")
		assert.Empty(t, cfg.Breakpoints)
	})

	t.Run("consumes only the breakpoints of the stage", func(t *testing.T) {
		cfg := &config.Config{Breakpoints: []config.Location{
			{Stage: "s1", Chunk: "0"},
			{Stage: "s2", Chunk: "second"},
			{Stage: "s1", Chunk: "second"},
			{File: "other.md", Stage: "s1", Chunk: "1"},
		}}
		s1 := newStage("s1")
		err := applyBreakpoints(cfg, "dir/test.md", s1)
		assert.NoError(t, err)
		assert.Equal(t, []string{"0", "second"}, s1.DebugFromChunks)
		assert.False(t, cfg.Interactive)
		assert.Equal(t, []config.Location{
			{Stage: "s2", Chunk: "second"},
			{File: "other.md", Stage: "s1", Chunk: "1"},
		}, cfg.Breakpoints)

		s2 := newStage("s2")
		err = applyBreakpoints(cfg, "dir/test.md", s2)
		assert.NoError(t, err)
		assert.Equal(t, []string{"second"}, s2.DebugFromChunks)
		assert.Equal(t, []config.Location{{File: "other.md", Stage: "s1", Chunk: "1"}}, cfg.Breakpoints)
	})

	t.Run("stage breakpoint turns interactive mode on", func(t *testing.T) {
		cfg := &config.Config{Breakpoints: []config.Location{{File: "test", Stage: "s1"}}}
		s1 := newStage("s1")
		err := applyBreakpoints(cfg, "dir/test.md", s1)
		assert.NoError(t, err)
		assert.True(t, cfg.Interactive)
		assert.Empty(t, s1.DebugFromChunks)
		assert.Empty(t, cfg.Breakpoints)
	})

	t.Run("unknown chunk is an error", func(t *testing.T) {
		cfg := &config.Config{Breakpoints: []config.Location{{Stage: "s1", Chunk: "0"}, {Stage: "s1", Chunk: "2"}}}
		err := applyBreakpoints(cfg, "dir/test.md", newStage("s1"))
		assert.Error(t, err)
	})
}

func TestApplyStartFrom(t *testing.T) {
	newStage := func(name string) *stage.Stage {
		return &stage.Stage{
			Name: name,
			Chunks: []*chunk.ExecutableChunk{
				{Id: "first", Stage: name},
				{Id: "second", Stage: name},
			},
		}
	}

	t.Run("no start-from runs every stage", func(t *testing.T) {
		cfg := &config.Config{}
		skip, err := applyStartFrom(cfg, "dir/test.md", newStage("s1"))
		assert.NoError(t, err)
		assert.False(t, skip)
	})

	t.Run("skips stages until the location is reached", func(t *testing.T) {
		cfg := &config.Config{StartFromLocation: config.Location{File: "test", Stage: "s2", Chunk: "second"}}
		skip, err := applyStartFrom(cfg, "dir/test.md", newStage("s1"))
		assert.NoError(t, err)
		assert.True(t, skip)
		skip, err = applyStartFrom(cfg, "dir/other.md", newStage("s2"))
		assert.NoError(t, err)
		assert.True(t, skip, "Expected s2 of another file to be skipped")
		s2 := newStage("s2")
		skip, err = applyStartFrom(cfg, "dir/test.md", s2)
		assert.NoError(t, err)
		assert.False(t, skip)
		assert.Equal(t, "second", s2.StartFromChunk)
		assert.Equal(t, config.Location{}, cfg.StartFromLocation)
		// Once reached, the following stages run normally
		skip, err = applyStartFrom(cfg, "dir/test.md", newStage("s1"))
		assert.NoError(t, err)
		assert.False(t, skip)
	})

	t.Run("unknown chunk is an error", func(t *testing.T) {
		cfg := &config.Config{StartFromLocation: config.Location{Stage: "s1", Chunk: "2"}}
		_, err := applyStartFrom(cfg, "dir/test.md", newStage("s1"))
		assert.Error(t, err)
	})
}

func TestRunMDWithSeveralBreakpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	mdContent := `
` + "```" + `bash {"stage":"stage1"}
echo "one"
` + "```" + `
` + "```" + `bash {"stage":"stage2"}
echo "two"
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	t.Run("every breakpoint is reached", func(t *testing.T) {
		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1, View: "mock",
			Breakpoints: []config.Location{{Stage: "stage1", Chunk: "0"}, {Stage: "stage2", Chunk: "0"}}}
		err := RunMD(cfg, mdFile)
		assert.NoError(t, err)
		assert.Empty(t, cfg.Breakpoints)
	})

	t.Run("the second breakpoint is validated", func(t *testing.T) {
		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1, View: "mock",
			Breakpoints: []config.Location{{Stage: "stage1", Chunk: "0"}, {Stage: "stage2", Chunk: "5"}}}
		err := RunMD(cfg, mdFile)
		assert.Error(t, err)
	})
}

// breakLineDoc is a markdown document used by the line based break tests.
// Line numbers of the fences are:
//
//	stage1/a: 3-5, stage1/b: 7-9, stage2/c: 13-15, stage2/d: 17-19
var breakLineDoc = "# Title\n" + // 1
	"\n" + // 2
	"```bash {\"stage\":\"stage1\", \"id\":\"a\"}\n" + // 3
	"echo a\n" + // 4
	"```\n" + // 5
	"\n" + // 6
	"```bash {\"stage\":\"stage1\", \"id\":\"b\"}\n" + // 7
	"echo b\n" + // 8
	"```\n" + // 9
	"\n" + // 10
	"Some text\n" + // 11
	"\n" + // 12
	"```bash {\"stage\":\"stage2\", \"id\":\"c\"}\n" + // 13
	"echo c\n" + // 14
	"```\n" + // 15
	"\n" + // 16
	"```bash {\"stage\":\"stage2\", \"id\":\"d\"}\n" + // 17
	"echo d\n" + // 18
	"```\n" // 19

// writeBreakLineDoc writes breakLineDoc into a temporary test.md file.
func writeBreakLineDoc(t *testing.T) string {
	mdFile := path.Join(t.TempDir(), "test.md")
	assert.NoError(t, os.WriteFile(mdFile, []byte(breakLineDoc), 0o644))
	return mdFile
}

func TestFindChunkByLine(t *testing.T) {
	mdFile := writeBreakLineDoc(t)
	ctx := &runnercontext.Context{Cfg: &config.Config{}, RView: view.NewView("mock")}
	stages, err := parser.ExtractStages(ctx, path.Base(mdFile), path.Dir(mdFile))
	assert.NoError(t, err)
	chunks, err := runnableChunks(&config.Config{}, mdFile, stages)
	assert.NoError(t, err)
	assert.Len(t, chunks, 4)

	testCases := []struct {
		name     string
		line     int
		after    bool
		expected string // chunk id, "" for no chunk
	}{
		{name: "after, before any chunk", line: 1, after: true, expected: "a"},
		{name: "before, before any chunk", line: 1, after: false, expected: ""},
		{name: "after, on an opening fence", line: 7, after: true, expected: "b"},
		{name: "before, on an opening fence", line: 7, after: false, expected: "b"},
		{name: "after, inside a chunk", line: 8, after: true, expected: "b"},
		{name: "before, on a closing fence", line: 9, after: false, expected: "b"},
		{name: "after, between chunks across stages", line: 11, after: true, expected: "c"},
		{name: "before, between chunks across stages", line: 11, after: false, expected: "b"},
		{name: "after, between chunks of a stage", line: 6, after: true, expected: "b"},
		{name: "before, between chunks of a stage", line: 6, after: false, expected: "a"},
		{name: "after, past the last chunk", line: 100, after: true, expected: ""},
		{name: "before, past the last chunk", line: 100, after: false, expected: "d"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rc, ok := findChunkByLine(chunks, tc.line, tc.after)
			if tc.expected == "" {
				assert.False(t, ok)
				return
			}
			if assert.True(t, ok) {
				assert.Equal(t, tc.expected, rc.stage.Chunks[rc.index].Id)
			}
		})
	}
}

func TestResolveLineBreakpoints(t *testing.T) {
	newStages := func() []*stage.Stage {
		return []*stage.Stage{
			{Name: "s1", Chunks: []*chunk.ExecutableChunk{
				{Id: "a", Stage: "s1", Line: 3, EndLine: 5},
				{Id: "b", Stage: "s1", Line: 7, EndLine: 9},
			}},
			{Name: "s2", Chunks: []*chunk.ExecutableChunk{
				{Id: "c", Stage: "s2", Line: 13, EndLine: 15},
			}},
		}
	}

	t.Run("each line breakpoint becomes a chunk breakpoint of its stage", func(t *testing.T) {
		stages := newStages()
		cfg := &config.Config{Breakpoints: []config.Location{
			{Line: 6, After: true},
			{Stage: "s1"},
			{Line: 11},
			{File: "test", Line: 14},
			{File: "other.md", Line: 1, After: true},
		}}
		assert.NoError(t, resolveLineBreakpoints(cfg, "dir/test.md", stages))
		assert.Equal(t, []string{"1", "1"}, stages[0].DebugFromChunks)
		assert.Equal(t, []string{"0"}, stages[1].DebugFromChunks)
		assert.Equal(t, []config.Location{
			{Stage: "s1"},
			{File: "other.md", Line: 1, After: true},
		}, cfg.Breakpoints, "Expected stage and other file breakpoints to be kept")
		assert.False(t, cfg.Interactive)
	})

	t.Run("an unmatched line without file is kept for the next files", func(t *testing.T) {
		cfg := &config.Config{Breakpoints: []config.Location{{Line: 100, After: true}}}
		assert.NoError(t, resolveLineBreakpoints(cfg, "dir/test.md", newStages()))
		assert.Equal(t, []config.Location{{Line: 100, After: true}}, cfg.Breakpoints)
	})

	t.Run("a chunk skipped by start-from stage/chunk is never selected", func(t *testing.T) {
		stages := newStages()
		// Line 4 is inside s1/a, skipped because execution starts at s1/b
		cfg := &config.Config{
			StartFromLocation: config.Location{Stage: "s1", Chunk: "b"},
			Breakpoints:       []config.Location{{Line: 4, After: true}, {File: "test", Line: 4}},
		}
		err := resolveLineBreakpoints(cfg, "dir/test.md", stages)
		if assert.Error(t, err, "Expected the break-before with no executed chunk before line 4 to fail") {
			assert.Contains(t, err.Error(), "start-from")
		}
		// The break-after moved to s1/b, the first executed chunk after line 4
		assert.Equal(t, []string{"1"}, stages[0].DebugFromChunks)
		assert.Equal(t, config.Location{Stage: "s1", Chunk: "b"}, cfg.StartFromLocation, "Expected start-from not to be consumed")
	})

	t.Run("an unknown start-from chunk is an error", func(t *testing.T) {
		cfg := &config.Config{
			StartFromLocation: config.Location{Stage: "s1", Chunk: "missing"},
			Breakpoints:       []config.Location{{Line: 4, After: true}},
		}
		assert.Error(t, resolveLineBreakpoints(cfg, "dir/test.md", newStages()))
	})

	t.Run("an unmatched file@line is an error", func(t *testing.T) {
		cfg := &config.Config{Breakpoints: []config.Location{{File: "test.md", Line: 1}}}
		assert.Error(t, resolveLineBreakpoints(cfg, "dir/test.md", newStages()))
	})
}

func TestRunMDBreakLine(t *testing.T) {
	t.Run("break after a line enables interactive mode", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", Breakpoints: []config.Location{{Line: 11, After: true}}}
		err := RunMD(cfg, mdFile)
		assert.NoError(t, err)
		assert.True(t, cfg.Interactive, "Expected the break to enable interactive mode")
		assert.Empty(t, cfg.Breakpoints, "Expected the line breakpoint to be consumed")
	})

	t.Run("break with file@line on the matching file", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", Breakpoints: []config.Location{{File: "test", Line: 11}}}
		err := RunMD(cfg, mdFile)
		assert.NoError(t, err)
		assert.True(t, cfg.Interactive)
	})

	t.Run("break with file@line on another file is ignored", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		bps := []config.Location{{File: "other.md", Line: 11}}
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", Breakpoints: bps}
		err := RunMD(cfg, mdFile)
		assert.NoError(t, err)
		assert.False(t, cfg.Interactive)
		assert.Equal(t, bps, cfg.Breakpoints, "Expected the line breakpoint to be kept for the next files")
	})

	t.Run("no chunk before the line in a file-specific break is an error", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", Breakpoints: []config.Location{{File: "test.md", Line: 1}}}
		err := RunMD(cfg, mdFile)
		assert.Error(t, err)
		assert.False(t, cfg.Interactive)
	})

	t.Run("no chunk after the line without file is not an error", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", Breakpoints: []config.Location{{Line: 100, After: true}}}
		err := RunMD(cfg, mdFile)
		assert.NoError(t, err)
		assert.False(t, cfg.Interactive)
	})

	t.Run("line target in a stage skipped by start-from is not consumed", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		// Line 4 is inside stage1/a, skipped because execution starts at stage2
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", StartFromLocation: config.Location{Stage: "stage2"}, Breakpoints: []config.Location{{Line: 4, After: true}}}
		err := RunMD(cfg, mdFile)
		assert.NoError(t, err)
		// The first executed chunk after line 4 is stage2/c, so the break happens there
		assert.True(t, cfg.Interactive, "Expected the break to move to the first executed chunk")
		assert.Empty(t, cfg.Breakpoints)
		assert.Equal(t, config.Location{}, cfg.StartFromLocation)
	})

	t.Run("line target on a chunk skipped by start-from stage/chunk moves to the next executed chunk", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		ui := view.NewView("mock").(*view.MockRunnerView)
		ui.PromptAnswers = []string{"continue"}
		// Line 4 is inside stage1/a, skipped because execution starts at stage1/b
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", StartFromLocation: config.Location{Stage: "stage1", Chunk: "b"}, Breakpoints: []config.Location{{Line: 4, After: true}}}
		err := runMD(cfg, mdFile, ui)
		assert.NoError(t, err)
		assert.Empty(t, cfg.Breakpoints)
		var prompted []string
		for _, call := range ui.Calls["InteractivePrompt"] {
			prompted = append(prompted, strings.TrimPrefix(call[1].(string), "echo "))
		}
		assert.Equal(t, []string{"b"}, prompted, "Expected the break to move to stage1/b, the first executed chunk")
	})

	t.Run("line target in a stage skipped by start-from moves to the next executed chunk", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		ui := view.NewView("mock").(*view.MockRunnerView)
		ui.PromptAnswers = []string{"continue"}
		// Line 8 is inside stage1/b, skipped because execution starts at stage2
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", StartFromLocation: config.Location{Stage: "stage2"}, Breakpoints: []config.Location{{Line: 8, After: true}}}
		err := runMD(cfg, mdFile, ui)
		assert.NoError(t, err)
		var prompted []string
		for _, call := range ui.Calls["InteractivePrompt"] {
			prompted = append(prompted, strings.TrimPrefix(call[1].(string), "echo "))
		}
		assert.Equal(t, []string{"c"}, prompted, "Expected the break to move to stage2/c, the first executed chunk")
	})

	t.Run("break before a line only in skipped stages fails for a given file", func(t *testing.T) {
		mdFile := writeBreakLineDoc(t)
		cfg := &config.Config{MinutesToTimeout: 1, View: "mock", StartFromLocation: config.Location{Stage: "stage2"}, Breakpoints: []config.Location{{File: "test", Line: 11}}}
		err := RunMD(cfg, mdFile)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "start-from")
		}
		assert.False(t, cfg.Interactive)
	})
}

func TestRunMDContinueAtLineBreakpoints(t *testing.T) {
	testCases := []struct {
		name        string
		breakpoints []config.Location
		// promptedChunks are the IDs of the chunks expected to be prompted
		promptedChunks []string
	}{
		{
			name:           "continue stops at a second line breakpoint",
			breakpoints:    []config.Location{{Line: 2, After: true}, {Line: 16, After: true}},
			promptedChunks: []string{"a", "d"},
		},
		{
			name:           "continue stops at a before and an after line breakpoint",
			breakpoints:    []config.Location{{Line: 10}, {Line: 16, After: true}},
			promptedChunks: []string{"b", "d"},
		},
		{
			name:           "continue stops at a line breakpoint then at a stage breakpoint",
			breakpoints:    []config.Location{{Line: 4, After: true}, {Stage: "stage2"}},
			promptedChunks: []string{"a", "c"},
		},
		{
			name:           "continue stops at a stage breakpoint then at a line breakpoint",
			breakpoints:    []config.Location{{Stage: "stage1"}, {File: "test", Line: 18}},
			promptedChunks: []string{"a", "d"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mdFile := writeBreakLineDoc(t)
			ui := view.NewView("mock").(*view.MockRunnerView)
			// answer "continue" at every prompt
			ui.PromptAnswers = []string{"continue", "continue", "continue", "continue"}
			cfg := &config.Config{MinutesToTimeout: 1, View: "mock", Breakpoints: tc.breakpoints}
			err := runMD(cfg, mdFile, ui)
			assert.NoError(t, err)
			assert.Empty(t, cfg.Breakpoints, "Expected every breakpoint to be consumed")
			assert.False(t, cfg.Interactive, "Expected the last continue to stop prompting")
			var prompted []string
			for _, call := range ui.Calls["InteractivePrompt"] {
				// the command of chunk <id> is "echo <id>"
				prompted = append(prompted, strings.TrimPrefix(call[1].(string), "echo "))
			}
			assert.Equal(t, tc.promptedChunks, prompted, "Expected one prompt per breakpoint")
		})
	}
}
