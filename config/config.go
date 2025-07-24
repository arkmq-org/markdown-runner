// Package config handles the parsing and storage of command-line flags and
// other application-wide configuration.
package config

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
)

var (
	DryRun            bool   = false
	Interactive       bool   = false
	Verbose           bool   = false
	MinutesToTimeout  int    = 10
	StartFrom         string = ""
	Rootdir                  = "./"
	IngoreBreakpoints bool   = false
	UpdateFile        bool   = false
	JustList          bool   = false
	NoStyling         bool   = false
	Quiet             bool   = false
	Recursive         bool   = false
	MarkdownDir       string
	Filter            string = ""
	Env               []string
	Help              bool
)

// ParseFlags initializes and parses the command-line flags from os.Args.
// It populates the global variables in this package with the values
// provided by the user.
func ParseFlags() {
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
  -v, --verbose              Print more logs
  -q, --quiet                Disable output
      --no-styling           Disable spinners in CLI

Help:
  -h, --help                 Show this help message
`
		fmt.Fprint(os.Stderr, helpText)
	}
	pflag.BoolVarP(&Help, "help", "h", false, "show this help message")
	pflag.BoolVarP(&DryRun, "dry-run", "d", false, "just list what would be executed without doing it")
	pflag.BoolVarP(&Interactive, "interactive", "i", false, "prompt to press enter between each chunk")
	pflag.BoolVarP(&JustList, "list", "l", false, "just list the files found")
	pflag.BoolVarP(&Verbose, "verbose", "v", false, "print more logs")
	pflag.BoolVar(&NoStyling, "no-styling", false, "disable spiners in cli")
	pflag.BoolVarP(&Quiet, "quiet", "q", false, "disable output")
	pflag.IntVarP(&MinutesToTimeout, "timeout", "t", 10, "the timeout in minutes for every executed command")
	pflag.StringVarP(&StartFrom, "start-from", "s", "", "start from a specific stage name")
	pflag.StringVarP(&Filter, "filter", "f", "", "Run only the files matching the regex")
	pflag.BoolVarP(&Recursive, "recursive", "r", false, "search for markdown files recursively")
	pflag.BoolVar(&IngoreBreakpoints, "ignore-breakpoints", false, "ignore the breakpoints")
	pflag.BoolVarP(&UpdateFile, "update-files", "u", false, "update the chunk output section in the markdown files")
	pflag.Parse()

	if len(pflag.Args()) > 1 {
		pterm.Fatal.Println("Too many positional arguments, please specify only one directory.")
	}

	if len(pflag.Args()) == 1 {
		MarkdownDir = pflag.Arg(0)
	} else {
		MarkdownDir = "./"
	}
}
