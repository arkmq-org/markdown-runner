package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test case
	os.Args = []string{"cmd", "-dry-run", "-interactive=true", "-verbose", "-timeout=5", "-start-from=stage2", "-markdown-dir=/tmp", "-filter=test.md", "-ignore-breakpoints", "-update-files", "-list", "-no-styling", "-quiet"}

	// Reset flags to default values before parsing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ParseFlags()

	assert.True(t, DryRun, "Expected DryRun to be true, but got false")
	assert.True(t, Interactive, "Expected Interactive to be true, but got false")
	assert.True(t, Verbose, "Expected Verbose to be true, but got false")
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

func TestParseFlags_Shorthand(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test case
	os.Args = []string{"cmd", "-d", "-i=true", "-v", "-t=5", "-s=stage2", "-m=/tmp", "-f=test.md", "-u", "-l", "-q"}

	// Reset flags to default values before parsing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ParseFlags()

	assert.True(t, DryRun, "Expected DryRun to be true, but got false")
	assert.True(t, Interactive, "Expected Interactive to be true, but got false")
	assert.True(t, Verbose, "Expected Verbose to be true, but got false")
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

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ParseFlags()

	assert.False(t, DryRun, "Expected DryRun to be false by default")
	assert.False(t, Interactive, "Expected Interactive to be false by default")
	assert.False(t, Verbose, "Expected Verbose to be false by default")
	assert.Equal(t, 10, MinutesToTimeout, "Expected MinutesToTimeout to be 10 by default")
	assert.Equal(t, "", StartFrom, "Expected StartFrom to be empty by default")
	assert.Equal(t, "./docs", MarkdownDir, "Expected MarkdownDir to be './docs' by default")
	assert.Equal(t, "", Filter, "Expected filter to be empty by default")
	assert.False(t, IngoreBreakpoints, "Expected IngoreBreakpoints to be false by default")
	assert.False(t, UpdateFile, "Expected UpdateFile to be false by default")
	assert.False(t, JustList, "Expected JustList to be false by default")
	assert.False(t, NoStyling, "Expected NoStyling to be false by default")
	assert.False(t, Quiet, "Expected Quiet to be false by default")
}
