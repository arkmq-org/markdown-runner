// Package config handles the parsing and storage of command-line flags and
// other application-wide configuration.
package config

import "github.com/spf13/pflag"

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
	MarkdownDir       string = ""
	Filter            string = ""
	Env               []string
	Help              bool
)

// ParseFlags initializes and parses the command-line flags from os.Args.
// It populates the global variables in this package with the values
// provided by the user.
func ParseFlags() {
	pflag.BoolVarP(&Help, "help", "h", false, "show this help message")
	pflag.BoolVarP(&DryRun, "dry-run", "d", false, "just list what would be executed without doing it")
	pflag.BoolVarP(&Interactive, "interactive", "i", false, "prompt to press enter between each chunk")
	pflag.BoolVarP(&JustList, "list", "l", false, "just list the files found")
	pflag.BoolVarP(&Verbose, "verbose", "v", false, "print more logs")
	pflag.BoolVar(&NoStyling, "no-styling", false, "disable spiners in cli")
	pflag.BoolVarP(&Quiet, "quiet", "q", false, "disable output")
	pflag.IntVarP(&MinutesToTimeout, "timeout", "t", 10, "the timeout in minutes for every executed command")
	pflag.StringVarP(&StartFrom, "start-from", "s", "", "start from a specific stage name")
	pflag.StringVarP(&MarkdownDir, "markdown-dir", "m", "./", "where to find the markdown files to execute")
	pflag.StringVarP(&Filter, "filter", "f", "", "Run only the files matching the regex")
	pflag.BoolVarP(&Recursive, "recursive", "r", false, "search for markdown files recursively")
	pflag.BoolVar(&IngoreBreakpoints, "ignore-breakpoints", false, "ignore the breakpoints")
	pflag.BoolVarP(&UpdateFile, "update-files", "u", false, "update the chunk output section in the markdown files")
	pflag.Parse()
}
