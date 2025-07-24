package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test case
	os.Args = []string{"cmd", "--dry-run", "--interactive=true", "--verbose", "--recursive", "--timeout=5", "--start-from=stage2", "--filter=test.md", "--ignore-breakpoints", "--update-files", "--list", "--no-styling", "--quiet", "/tmp"}

	// Reset flags to default values before parsing
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	ParseFlags()

	assert.True(t, DryRun, "Expected DryRun to be true, but got false")
	assert.True(t, Interactive, "Expected Interactive to be true, but got false")
	assert.True(t, Verbose, "Expected Verbose to be true, but got false")
	assert.True(t, Recursive, "Expected Recursive to be true, but got false")
	assert.Equal(t, 5, MinutesToTimeout, "Expected MinutesToTimeout to be 5")
	assert.Equal(t, "stage2", StartFrom, "Expected StartFrom to be 'stage2'")
	assert.Equal(t, "/tmp", MarkdownDir, "Expected MarkdownDir to be '/tmp'")
	assert.Equal(t, "test.md", Filter, "Expected File to be 'test.md'")
	assert.True(t, IngoreBreakpoints, "Expected IngoreBreakpoints to be true, but got false")
	assert.True(t, UpdateFile, "Expected UpdateFile to be true, but got false")
	assert.True(t, JustList, "Expected JustList to be true, but got false")
	assert.True(t, NoStyling, "Expected NoStyling to be true, but got false")
	assert.True(t, Quiet, "Expected Quiet to be true, but got false")
}

func TestParseFlags_PositionalDir(t *testing.T) {
	t.Run("flag before positional", func(t *testing.T) {
		// Save original os.Args
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		// Test case
		os.Args = []string{"cmd", "-f=test.md", "/tmp"}

		// Reset flags to default values before parsing
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
		ParseFlags()

		assert.Equal(t, "/tmp", MarkdownDir, "Expected MarkdownDir to be '/tmp'")
		assert.Equal(t, "test.md", Filter, "Expected filter to be 'test.md'")
	})

	t.Run("positional before flag", func(t *testing.T) {
		// Save original os.Args
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		// Test case
		os.Args = []string{"cmd", "/tmp", "-f=test.md"}

		// Reset flags to default values before parsing
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
		ParseFlags()

		assert.Equal(t, "/tmp", MarkdownDir, "Expected MarkdownDir to be '/tmp'")
		assert.Equal(t, "test.md", Filter, "Expected filter to be 'test.md'")
	})

	t.Run("no positional before or after flag", func(t *testing.T) {
		// Save original os.Args
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		// Test case
		os.Args = []string{"cmd", "-f=test.md"}

		// Reset flags to default values before parsing
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
		ParseFlags()

		assert.Equal(t, "./", MarkdownDir, "Expected MarkdownDir to be '/tmp'")
		assert.Equal(t, "test.md", Filter, "Expected filter to be 'test.md'")
	})
}

func TestParseFlags_Shorthand(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test case
	os.Args = []string{"cmd", "-d", "-i=true", "-v", "-r", "-t=5", "-s=stage2", "-f=test.md", "-u", "-l", "-q", "/tmp"}

	// Reset flags to default values before parsing
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	ParseFlags()

	assert.True(t, DryRun, "Expected DryRun to be true, but got false")
	assert.True(t, Interactive, "Expected Interactive to be true, but got false")
	assert.True(t, Verbose, "Expected Verbose to be true, but got false")
	assert.True(t, Recursive, "Expected Recursive to be true, but got false")
	assert.Equal(t, 5, MinutesToTimeout, "Expected MinutesToTimeout to be 5")
	assert.Equal(t, "stage2", StartFrom, "Expected StartFrom to be 'stage2'")
	assert.Equal(t, "/tmp", MarkdownDir, "Expected MarkdownDir to be '/tmp'")
	assert.Equal(t, "test.md", Filter, "Expected filter to be 'test.md'")
	assert.True(t, UpdateFile, "Expected UpdateFile to be true, but got false")
	assert.True(t, JustList, "Expected JustList to be true, but got false")
	assert.True(t, Quiet, "Expected Quiet to be true, but got false")
}

func TestParseFlags_Defaults(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cmd"}

	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	ParseFlags()

	assert.False(t, DryRun, "Expected DryRun to be false by default")
	assert.False(t, Interactive, "Expected Interactive to be false by default")
	assert.False(t, Verbose, "Expected Verbose to be false by default")
	assert.Equal(t, 10, MinutesToTimeout, "Expected MinutesToTimeout to be 10 by default")
	assert.Equal(t, "", StartFrom, "Expected StartFrom to be empty by default")
	assert.Equal(t, "./", MarkdownDir, "Expected MarkdownDir to be './' by default")
	assert.Equal(t, "", Filter, "Expected filter to be empty by default")
	assert.False(t, IngoreBreakpoints, "Expected IngoreBreakpoints to be false by default")
	assert.False(t, UpdateFile, "Expected UpdateFile to be false by default")
	assert.False(t, JustList, "Expected JustList to be false by default")
	assert.False(t, NoStyling, "Expected NoStyling to be false by default")
	assert.False(t, Quiet, "Expected Quiet to be false by default")
	assert.False(t, Recursive, "Expected Recursive to be false by default")
}
