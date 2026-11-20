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
	StartFrom         string
	StartFromFile     string
	StartFromStage    string
	DebugFrom         []string     // raw values of every --break-at flag
	BreakBefore       []string     // raw values of every --break-before flag
	BreakAfter        []string     // raw values of every --break-after flag
	Breakpoints       []Breakpoint // parsed break flag values, consumed once reached
	MinutesToTimeout  int
	UpdateFile        bool
	Verbose           bool
	View              string
	Env               []string
	Rootdir           string
}

// Breakpoint is a location where the runner switches to interactive mode,
// as given by the --break-at flag (stage, stage/chunkID, file@stage/chunkID,
// line or file@line) or by the --break-before and --break-after flags (line
// or file@line).
type Breakpoint struct {
	// File restricts the breakpoint to a given markdown file, empty for any file.
	File string
	// Stage is the name of the stage to break at, empty for a line breakpoint.
	Stage string
	// Chunk is the ID or 0-based index of the chunk to break at, empty for
	// the whole stage.
	Chunk string
	// Line is the 1-based markdown line to break around, 0 for a stage
	// breakpoint.
	Line int
	// After selects, for a line breakpoint, the first chunk after Line
	// instead of the last chunk before it. A chunk enclosing Line is
	// selected either way.
	After bool
}

// lineNumberMatcher matches a value that designates a line in a markdown file.
var lineNumberMatcher = regexp.MustCompile("^[0-9]+$")

// splitFileTarget splits the optional "file@" prefix of a break flag value.
// It returns ok=false when the value contains more than one '@' or when the
// file part is empty.
func splitFileTarget(value string) (file string, rest string, ok bool) {
	if !strings.Contains(value, "@") {
		return "", value, true
	}
	parts := strings.Split(value, "@")
	if len(parts) != 2 || parts[0] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// parseLine parses the line part of a line breakpoint.
func parseLine(linePart string) (int, bool) {
	if !lineNumberMatcher.MatchString(linePart) {
		return 0, false
	}
	line, err := strconv.Atoi(linePart)
	if err != nil || line < 1 {
		return 0, false
	}
	return line, true
}

// ParseLineBreakpoint parses a --break-before (after=false) or --break-after
// (after=true) value: line or file@line, with a line number starting at 1.
func ParseLineBreakpoint(value string, after bool) (Breakpoint, error) {
	flag := "break-before"
	if after {
		flag = "break-after"
	}
	formatErr := fmt.Errorf("invalid %s value %q. Use 'line' or 'file@line' with a line number starting at 1", flag, value)
	file, linePart, ok := splitFileTarget(value)
	if !ok {
		return Breakpoint{}, formatErr
	}
	line, ok := parseLine(linePart)
	if !ok {
		return Breakpoint{}, formatErr
	}
	return Breakpoint{File: file, Line: line, After: after}, nil
}

// ParseBreakpoint parses a --break-at value: stage, stage/chunkID,
// file@stage/chunkID, line or file@line. A line number is a line breakpoint
// that behaves like --break-after.
func ParseBreakpoint(value string) (Breakpoint, error) {
	var bp Breakpoint
	formatErr := fmt.Errorf("invalid break-at format %q. Use 'stage', 'stage/chunkID', 'file@stage/chunkID', 'line' or 'file@line'", value)
	file, debugPart, ok := splitFileTarget(value)
	if !ok {
		return bp, formatErr
	}
	bp.File = file
	if lineNumberMatcher.MatchString(debugPart) {
		// A line number given to break-at defaults to break-after
		line, ok := parseLine(debugPart)
		if !ok {
			return bp, formatErr
		}
		bp.Line = line
		bp.After = true
		return bp, nil
	}
	parts := strings.Split(debugPart, "/")
	switch len(parts) {
	case 1:
		bp.Stage = parts[0]
	case 2:
		bp.Stage = parts[0]
		bp.Chunk = parts[1]
		if bp.Chunk == "" {
			return bp, formatErr
		}
	default:
		return bp, formatErr
	}
	if bp.Stage == "" {
		return bp, formatErr
	}
	return bp, nil
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
  -s, --start-from string    Start from a specific stage (stage or file@stage)
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
	pflag.StringVarP(&cfg.StartFrom, "start-from", "s", "", "Start from a specific stage (stage or file@stage)")
	pflag.StringArrayVarP(&cfg.DebugFrom, "break-at", "B", nil, "Start debugging from a specific stage, chunk or line (stage, stage/chunkID, file@stage/chunkID, line or file@line), can be repeated")
	pflag.StringArrayVar(&cfg.BreakBefore, "break-before", nil, "Start debugging from the last chunk before a line (line or file@line), can be repeated")
	pflag.StringArrayVar(&cfg.BreakAfter, "break-after", nil, "Start debugging from the first chunk after a line (line or file@line), can be repeated")
	pflag.IntVarP(&cfg.MinutesToTimeout, "timeout", "t", 10, "The timeout in minutes for every executed command")
	pflag.BoolVarP(&cfg.UpdateFile, "update-files", "u", false, "Update the chunk output section in the markdown files")
	pflag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Print more logs")
	pflag.StringVar(&cfg.View, "view", "default", "UI to be used, can be 'default' or 'ci'")

	pflag.Parse()

	// Parse start-from format: stage or file@stage
	if cfg.StartFrom != "" {
		// Check for file@stage format
		if strings.Contains(cfg.StartFrom, "@") {
			atParts := strings.Split(cfg.StartFrom, "@")
			if len(atParts) != 2 {
				pterm.Fatal.Println("Invalid start-from format. Use 'stage' or 'file@stage'.")
			}
			cfg.StartFromFile = atParts[0]
			cfg.StartFromStage = atParts[1]
		} else {
			cfg.StartFromStage = cfg.StartFrom
		}
	}

	// Parse break-at format: stage, stage/chunkID, file@stage/chunkID, line or file@line
	for _, value := range cfg.DebugFrom {
		bp, err := ParseBreakpoint(value)
		if err != nil {
			pterm.Fatal.Println(err)
		}
		cfg.Breakpoints = append(cfg.Breakpoints, bp)
	}
	// Parse break-before and break-after format: line or file@line
	for _, value := range cfg.BreakBefore {
		bp, err := ParseLineBreakpoint(value, false)
		if err != nil {
			pterm.Fatal.Println(err)
		}
		cfg.Breakpoints = append(cfg.Breakpoints, bp)
	}
	for _, value := range cfg.BreakAfter {
		bp, err := ParseLineBreakpoint(value, true)
		if err != nil {
			pterm.Fatal.Println(err)
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
