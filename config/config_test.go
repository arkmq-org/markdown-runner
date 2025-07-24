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
				args:              []string{"cmd", "--dry-run", "--interactive=true", "--verbose", "--recursive", "--timeout=5", "--start-from=stage2", "--filter=test.md", "--ignore-breakpoints", "--update-files", "--list", "--no-styling", "--quiet", "/tmp"},
				dryRun:            true,
				interactive:       true,
				verbose:           true,
				recursive:         true,
				timeout:           5,
				startFrom:         "stage2",
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
				markdownDir:       "/tmp",
				filter:            "test.md",
				ignoreBreakpoints: false,
				updateFile:        true,
				justList:          true,
				noStyling:         false,
				quiet:             true,
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
