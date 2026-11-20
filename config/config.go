// Package config handles the parsing and storage of command-line flags and
// other application-wide configuration.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	StartFrom         string     // raw value of the --start-from flag
	StartFromLocation Location   // parsed --start-from value, cleared once reached
	DebugFrom         []string   // raw values of every --break-at flag
	BreakBefore       []string   // raw values of every --break-before flag
	BreakAfter        []string   // raw values of every --break-after flag
	Breakpoints       []Location // parsed break flag values, consumed once reached
	MinutesToTimeout  int
	UpdateFile        bool
	Verbose           bool
	View              string
	Env               []string
	Rootdir           string
}

// Location designates a stage, a chunk or a line of a markdown file, as given
// by the --start-from flag (stage, stage/chunkID or file@stage/chunkID), the
// --break-at flag (the same, plus line or file@line) or the --break-before and
// --break-after flags (line or file@line).
type Location struct {
	// File restricts the location to a given markdown file, empty for any file.
	File string
	// Stage is the name of the targeted stage, empty for a line location.
	Stage string
	// Chunk is the ID or 0-based index of the targeted chunk, empty for the
	// whole stage.
	Chunk string
	// Line is the 1-based markdown line targeted by a line location, 0 for a
	// stage or chunk location.
	Line int
	// After selects, for a line location, the first chunk after Line instead
	// of the last chunk before it. A chunk enclosing Line is selected either
	// way.
	After bool
}

// lineNumberMatcher matches a value that designates a line in a markdown file.
var lineNumberMatcher = regexp.MustCompile("^[0-9]+$")

// ParseLocation parses a location: stage, stage/chunkID, file@stage/chunkID,
// line or file@line. A line number starts at 1; a line location selects by
// default the first chunk after the line (After is true).
// It is the single parser shared by every location flag; ParseStartFrom and
// ParseLineLocation restrict the accepted forms.
func ParseLocation(value string) (Location, error) {
	var loc Location
	formatErr := fmt.Errorf("invalid format %q. Use 'stage', 'stage/chunkID', 'file@stage/chunkID', 'line' or 'file@line'", value)
	locationPart := value
	if strings.Contains(value, "@") {
		atParts := strings.Split(value, "@")
		if len(atParts) != 2 {
			return loc, formatErr
		}
		loc.File = atParts[0]
		locationPart = atParts[1]
		if loc.File == "" {
			return loc, formatErr
		}
	}
	if lineNumberMatcher.MatchString(locationPart) {
		line, err := strconv.Atoi(locationPart)
		if err != nil || line < 1 {
			return loc, formatErr
		}
		loc.Line = line
		loc.After = true
		return loc, nil
	}
	parts := strings.Split(locationPart, "/")
	switch len(parts) {
	case 1:
		loc.Stage = parts[0]
	case 2:
		loc.Stage = parts[0]
		loc.Chunk = parts[1]
		if loc.Chunk == "" {
			return loc, formatErr
		}
	default:
		return loc, formatErr
	}
	if loc.Stage == "" {
		return loc, formatErr
	}
	return loc, nil
}

// ParseStartFrom parses a --start-from value: stage, stage/chunkID or
// file@stage/chunkID. Line numbers are rejected.
func ParseStartFrom(value string) (Location, error) {
	loc, err := ParseLocation(value)
	if err != nil {
		return loc, fmt.Errorf("invalid format %q. Use 'stage', 'stage/chunkID', or 'file@stage/chunkID'", value)
	}
	if loc.Line != 0 {
		return Location{}, fmt.Errorf("line numbers are not supported, got %q. Use 'stage', 'stage/chunkID', or 'file@stage/chunkID'", value)
	}
	return loc, nil
}

// ParseLineLocation parses a --break-before (after=false) or --break-after
// (after=true) value: line or file@line, with a line number starting at 1.
func ParseLineLocation(value string, after bool) (Location, error) {
	loc, err := ParseLocation(value)
	if err != nil || loc.Line == 0 {
		return Location{}, fmt.Errorf("invalid format %q. Use 'line' or 'file@line' with a line number starting at 1", value)
	}
	loc.After = after
	return loc, nil
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
  -i, --interactive          Prompt before each chunk (yes/no/continue/quit)
  -s, --start-from string    Start from a specific stage or chunk (stage, stage/chunkID, or file@stage/chunkID)
  -B, --break-at string      Start debugging from a specific stage, chunk or line (stage, stage/chunkID, file@stage/chunkID, line or file@line), can be repeated
      --break-before string  Start debugging from the last chunk before a line (line or file@line), can be repeated
      --break-after string   Start debugging from the first chunk after a line (line or file@line), can be repeated
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
	pflag.BoolVarP(&cfg.Interactive, "interactive", "i", false, "Prompt before each chunk (yes/no/continue/quit)")
	pflag.BoolVarP(&cfg.JustList, "list", "l", false, "Just list the files found")
	pflag.StringVarP(&cfg.Filter, "filter", "f", "", "Run only the files matching the regex")
	pflag.BoolVarP(&cfg.NoStyling, "no-styling", "", false, "Disable spinners in CLI")
	pflag.BoolVarP(&cfg.Quiet, "quiet", "q", false, "Disable output")
	pflag.BoolVarP(&cfg.Recursive, "recursive", "r", false, "Search for markdown files recursively")
	pflag.StringVarP(&cfg.StartFrom, "start-from", "s", "", "Start from a specific stage or chunk (stage, stage/chunkID, or file@stage/chunkID)")
	pflag.StringArrayVarP(&cfg.DebugFrom, "break-at", "B", nil, "Start debugging from a specific stage, chunk or line (stage, stage/chunkID, file@stage/chunkID, line or file@line), can be repeated")
	pflag.StringArrayVar(&cfg.BreakBefore, "break-before", nil, "Start debugging from the last chunk before a line (line or file@line), can be repeated")
	pflag.StringArrayVar(&cfg.BreakAfter, "break-after", nil, "Start debugging from the first chunk after a line (line or file@line), can be repeated")
	pflag.IntVarP(&cfg.MinutesToTimeout, "timeout", "t", 10, "The timeout in minutes for every executed command")
	pflag.BoolVarP(&cfg.UpdateFile, "update-files", "u", false, "Update the chunk output section in the markdown files")
	pflag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Print more logs")
	pflag.StringVar(&cfg.View, "view", "default", "UI to be used, can be 'default' or 'ci'")

	pflag.Parse()

	// Parse start-from format: stage, stage/chunkID, or file@stage/chunkID
	if cfg.StartFrom != "" {
		loc, err := ParseStartFrom(cfg.StartFrom)
		if err != nil {
			pterm.Fatal.Println("start-from: " + err.Error())
		}
		cfg.StartFromLocation = loc
	}

	// Parse break-at format: stage, stage/chunkID, file@stage/chunkID, line or file@line
	for _, value := range cfg.DebugFrom {
		bp, err := ParseLocation(value)
		if err != nil {
			pterm.Fatal.Println("break-at: " + err.Error())
		}
		cfg.Breakpoints = append(cfg.Breakpoints, bp)
	}
	// Parse break-before and break-after format: line or file@line
	for _, value := range cfg.BreakBefore {
		bp, err := ParseLineLocation(value, false)
		if err != nil {
			pterm.Fatal.Println("break-before: " + err.Error())
		}
		cfg.Breakpoints = append(cfg.Breakpoints, bp)
	}
	for _, value := range cfg.BreakAfter {
		bp, err := ParseLineLocation(value, true)
		if err != nil {
			pterm.Fatal.Println("break-after: " + err.Error())
		}
		cfg.Breakpoints = append(cfg.Breakpoints, bp)
	}

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
