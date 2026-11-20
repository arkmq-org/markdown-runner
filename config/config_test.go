package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	t.Run("flags", func(t *testing.T) {
		testCases := []struct {
			name              string
			args              []string
			dryRun            bool
			interactive       bool
			verbose           bool
			recursive         bool
			timeout           int
			startFrom         string
			debugFrom         []string
			breakpoints       []Breakpoint
			markdownDir       string
			filter            string
			ignoreBreakpoints bool
			updateFile        bool
			justList          bool
			noStyling         bool
			quiet             bool
		}{
			{
				name:              "long-form flags",
				args:              []string{"cmd", "--dry-run", "--interactive=true", "--verbose", "--recursive", "--timeout=5", "--start-from=stage2", "--break-at=stage3", "--filter=test.md", "--ignore-breakpoints", "--update-files", "--list", "--no-styling", "--quiet", "/tmp"},
				dryRun:            true,
				interactive:       true,
				verbose:           true,
				recursive:         true,
				timeout:           5,
				startFrom:         "stage2",
				debugFrom:         []string{"stage3"},
				breakpoints:       []Breakpoint{{Stage: "stage3"}},
				markdownDir:       "/tmp",
				filter:            "test.md",
				ignoreBreakpoints: true,
				updateFile:        true,
				justList:          true,
				noStyling:         true,
				quiet:             true,
			},
			{
				name:              "shorthand flags",
				args:              []string{"cmd", "-d", "-i=true", "-v", "-r", "-t=5", "-s=stage2", "-f=test.md", "-u", "-l", "-q", "/tmp"},
				dryRun:            true,
				interactive:       true,
				verbose:           true,
				recursive:         true,
				timeout:           5,
				startFrom:         "stage2",
				debugFrom:         nil,
				breakpoints:       nil,
				markdownDir:       "/tmp",
				filter:            "test.md",
				ignoreBreakpoints: false,
				updateFile:        true,
				justList:          true,
				noStyling:         false,
				quiet:             true,
			},
			{
				name:        "repeated break-at flags",
				args:        []string{"cmd", "-B", "stage1/0", "--break-at=file.md@stage2/myID", "-B", "stage3", "/tmp"},
				timeout:     10,
				debugFrom:   []string{"stage1/0", "file.md@stage2/myID", "stage3"},
				breakpoints: []Breakpoint{{Stage: "stage1", Chunk: "0"}, {File: "file.md", Stage: "stage2", Chunk: "myID"}, {Stage: "stage3"}},
				markdownDir: "/tmp",
			},
			{
				name:      "combined stage and line break flags",
				args:      []string{"cmd", "-B", "56", "-B", "stage1", "--break-before", "12", "--break-after=doc.md@34", "--break-before=7", "-B", "doc@78", "/tmp"},
				timeout:   10,
				debugFrom: []string{"56", "stage1", "doc@78"},
				breakpoints: []Breakpoint{
					{Line: 56, After: true},
					{Stage: "stage1"},
					{File: "doc", Line: 78, After: true},
					{Line: 12},
					{Line: 7},
					{File: "doc.md", Line: 34, After: true},
				},
				markdownDir: "/tmp",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				oldArgs := os.Args
				defer func() {
					os.Args = oldArgs
					pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
				}()
				os.Args = tc.args

				cfg := NewConfig()

				assert.Equal(t, tc.dryRun, cfg.DryRun)
				assert.Equal(t, tc.interactive, cfg.Interactive)
				assert.Equal(t, tc.verbose, cfg.Verbose)
				assert.Equal(t, tc.recursive, cfg.Recursive)
				assert.Equal(t, tc.timeout, cfg.MinutesToTimeout)
				assert.Equal(t, tc.startFrom, cfg.StartFrom)
				assert.Equal(t, tc.debugFrom, cfg.DebugFrom)
				assert.Equal(t, tc.breakpoints, cfg.Breakpoints)
				assert.Equal(t, tc.markdownDir, cfg.MarkdownDir)
				assert.Equal(t, tc.filter, cfg.Filter)
				assert.Equal(t, tc.ignoreBreakpoints, cfg.IgnoreBreakpoints)
				assert.Equal(t, tc.updateFile, cfg.UpdateFile)
				assert.Equal(t, tc.justList, cfg.JustList)
				assert.Equal(t, tc.noStyling, cfg.NoStyling)
				assert.Equal(t, tc.quiet, cfg.Quiet)
			})
		}
	})

	t.Run("defaults", func(t *testing.T) {
		oldArgs := os.Args
		defer func() {
			os.Args = oldArgs
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
		}()

		os.Args = []string{"cmd"}

		cfg := NewConfig()

		assert.False(t, cfg.DryRun, "Expected DryRun to be false by default")
		assert.False(t, cfg.Interactive, "Expected Interactive to be false by default")
		assert.False(t, cfg.Verbose, "Expected Verbose to be false by default")
		assert.Equal(t, 10, cfg.MinutesToTimeout, "Expected MinutesToTimeout to be 10 by default")
		assert.Equal(t, "", cfg.StartFrom, "Expected StartFrom to be empty by default")
		assert.Empty(t, cfg.DebugFrom, "Expected DebugFrom to be empty by default")
		assert.Empty(t, cfg.Breakpoints, "Expected Breakpoints to be empty by default")
		assert.Equal(t, "./", cfg.MarkdownDir, "Expected MarkdownDir to be './' by default")
		assert.Equal(t, "", cfg.Filter, "Expected filter to be empty by default")
		assert.False(t, cfg.IgnoreBreakpoints, "Expected IgnoreBreakpoints to be false by default")
		assert.False(t, cfg.UpdateFile, "Expected UpdateFile to be false by default")
		assert.False(t, cfg.JustList, "Expected JustList to be false by default")
		assert.False(t, cfg.NoStyling, "Expected NoStyling to be false by default")
		assert.False(t, cfg.Quiet, "Expected Quiet to be false by default")
		assert.False(t, cfg.Recursive, "Expected Recursive to be false by default")
	})
}

func TestParseBreakpoint(t *testing.T) {
	testCases := []struct {
		value    string
		expected Breakpoint
		isError  bool
	}{
		{value: "stage", expected: Breakpoint{Stage: "stage"}},
		{value: "stage/0", expected: Breakpoint{Stage: "stage", Chunk: "0"}},
		{value: "stage/myID", expected: Breakpoint{Stage: "stage", Chunk: "myID"}},
		{value: "file.md@stage", expected: Breakpoint{File: "file.md", Stage: "stage"}},
		{value: "file@stage/myID", expected: Breakpoint{File: "file", Stage: "stage", Chunk: "myID"}},
		{value: "a@b@c", isError: true},
		{value: "stage/chunk/extra", isError: true},
		{value: "file@", isError: true},
		{value: "/chunk", isError: true},
		{value: "stage/", isError: true},
		{value: "@stage", isError: true},
		{value: "102", expected: Breakpoint{Line: 102, After: true}},
		{value: "doc@42", expected: Breakpoint{File: "doc", Line: 42, After: true}},
		{value: "0", isError: true},
		{value: "@42", isError: true},
		{value: "stage/3", expected: Breakpoint{Stage: "stage", Chunk: "3"}},
	}
	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			bp, err := ParseBreakpoint(tc.value)
			if tc.isError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, bp)
		})
	}
}

func TestParseLineBreakpoint(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		after    bool
		expected Breakpoint
		isError  bool
	}{
		{name: "before line", value: "12", expected: Breakpoint{Line: 12}},
		{name: "before file@line", value: "doc.md@12", expected: Breakpoint{File: "doc.md", Line: 12}},
		{name: "after line", value: "7", after: true, expected: Breakpoint{Line: 7, After: true}},
		{name: "after file@line", value: "sub/doc.md@7", after: true, expected: Breakpoint{File: "sub/doc.md", Line: 7, After: true}},
		{name: "not a number", value: "setup", after: true, isError: true},
		{name: "stage/chunk", value: "setup/0", isError: true},
		{name: "negative", value: "-3", isError: true},
		{name: "zero", value: "0", isError: true},
		{name: "empty line", value: "doc@", isError: true},
		{name: "empty file", value: "@3", isError: true},
		{name: "too many @", value: "a@b@3", isError: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bp, err := ParseLineBreakpoint(tc.value, tc.after)
			if tc.isError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, bp)
		})
	}
}
