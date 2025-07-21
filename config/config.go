// Package config handles the parsing and storage of command-line flags and
// other application-wide configuration.
package config

import "flag"

var DryRun bool = false
var Interactive bool = false
var Verbose bool = false
var MinutesToTimeout int = 10
var StartFrom string = ""
var Rootdir = "./"
var IngoreBreakpoints bool = false
var UpdateFile bool = false
var JustList bool = false
var NoStyling bool = false
var Quiet bool = false
var MarkdownDir string = ""
var File string = ""
var Env []string

// ParseFlags initializes and parses the command-line flags from os.Args.
// It populates the global variables in this package with the values
// provided by the user.
func ParseFlags() {
	flag.BoolVar(&DryRun, "d", false, "shorthand for -dry-run")
	flag.BoolVar(&DryRun, "dry-run", false, "just list what would be executed without doing it")
	flag.BoolVar(&Interactive, "i", false, "shorthand for -interactive")
	flag.BoolVar(&Interactive, "interactive", false, "prompt to press enter between each chunk")
	flag.BoolVar(&JustList, "l", false, "shorthand for -list")
	flag.BoolVar(&JustList, "list", false, "just list the files found")
	flag.BoolVar(&Verbose, "v", false, "shorthand for -verbose")
	flag.BoolVar(&Verbose, "verbose", false, "print more logs")
	flag.BoolVar(&NoStyling, "no-styling", false, "disable spiners in cli")
	flag.BoolVar(&Quiet, "q", false, "shorthand for -quiet")
	flag.BoolVar(&Quiet, "quiet", false, "disable output")
	flag.IntVar(&MinutesToTimeout, "t", 10, "shorthand for -timeout")
	flag.IntVar(&MinutesToTimeout, "timeout", 10, "the timeout in minutes for every executed command")
	flag.StringVar(&StartFrom, "s", "", "shorthand for -start-from")
	flag.StringVar(&StartFrom, "start-from", "", "start from a specific stage name")
	flag.StringVar(&MarkdownDir, "m", "./docs", "shorthand for -markdown-dir")
	flag.StringVar(&MarkdownDir, "markdown-dir", "./docs", "where to find the markdown files to execute")
	flag.StringVar(&File, "f", "", "shorthand for -file")
	flag.StringVar(&File, "file", "", "Run only a specific markdown file")
	flag.BoolVar(&IngoreBreakpoints, "ignore-breakpoints", false, "ignore the breakpoints")
	flag.BoolVar(&UpdateFile, "u", false, "shorthand for -update-files")
	flag.BoolVar(&UpdateFile, "update-files", false, "update the chunk output section in the markdown files")
	flag.Parse()
}
