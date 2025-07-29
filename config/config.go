// Package config handles the parsing and storage of command-line flags and
// other application-wide configuration.
package config

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
)

// Config holds all the configuration for the markdown-runner.
type Config struct {
	DryRun            bool
	Help              bool
	IgnoreBreakpoints bool
	Interactive       bool
	JustList          bool
	MarkdownDir       string
	Filter            string
	NoStyling         bool
	Quiet             bool
	Recursive         bool
	StartFrom         string
	MinutesToTimeout  int
	UpdateFile        bool
	Verbose           bool
	View              string
	Env               []string
	Rootdir           string
}

// NewConfig creates a new Config object and parses the command-line flags.
func NewConfig() *Config {
	cfg := &Config{}

	pflag.Usage = func() {
		helpText := `Usage: markdown-runner [options] [path]

Executes markdown files as scripts.
The default path is the current directory.

Modes:
  -d, --dry-run              Just list what would be executed without doing it
  -l, --list                 Just list the files found

Execution Control:
  -i, --interactive          Prompt to press enter between each chunk
  -s, --start-from string    Start from a specific stage name
  -t, --timeout int          The timeout in minutes for every executed command (default 10)
  -u, --update-files         Update the chunk output section in the markdown files
      --ignore-breakpoints   Ignore the breakpoints

File Selection:
  -f, --filter string        Run only the files matching the regex
  -r, --recursive            Search for markdown files recursively

Output & Logging:
      --view string          UI to be used, can be 'default' or 'ci'
  -v, --verbose              Print more logs
  -q, --quiet                Disable output
      --no-styling           Disable spinners in CLI

Help:
  -h, --help                 Show this help message
`
		fmt.Fprint(os.Stderr, helpText)
	}

	pflag.BoolVarP(&cfg.DryRun, "dry-run", "d", false, "Just list what would be executed without doing it")
	pflag.BoolVarP(&cfg.Help, "help", "h", false, "Show this help message")
	pflag.BoolVarP(&cfg.IgnoreBreakpoints, "ignore-breakpoints", "", false, "Ignore the breakpoints")
	pflag.BoolVarP(&cfg.Interactive, "interactive", "i", false, "Prompt to press enter between each chunk")
	pflag.BoolVarP(&cfg.JustList, "list", "l", false, "Just list the files found")
	pflag.StringVarP(&cfg.Filter, "filter", "f", "", "Run only the files matching the regex")
	pflag.BoolVarP(&cfg.NoStyling, "no-styling", "", false, "Disable spinners in CLI")
	pflag.BoolVarP(&cfg.Quiet, "quiet", "q", false, "Disable output")
	pflag.BoolVarP(&cfg.Recursive, "recursive", "r", false, "Search for markdown files recursively")
	pflag.StringVarP(&cfg.StartFrom, "start-from", "s", "", "Start from a specific stage name")
	pflag.IntVarP(&cfg.MinutesToTimeout, "timeout", "t", 10, "The timeout in minutes for every executed command")
	pflag.BoolVarP(&cfg.UpdateFile, "update-files", "u", false, "Update the chunk output section in the markdown files")
	pflag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Print more logs")
	pflag.StringVar(&cfg.View, "view", "default", "UI to be used, can be 'default' or 'ci'")

	pflag.Parse()

	if len(pflag.Args()) > 1 {
		pterm.Fatal.Println("Too many positional arguments, please specify only one directory.")
	}

	if len(pflag.Args()) == 1 {
		cfg.MarkdownDir = pflag.Arg(0)
	} else {
		cfg.MarkdownDir = "./"
	}
	cfg.Rootdir = "./"

	return cfg
}
