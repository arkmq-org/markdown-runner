// Package config handles the parsing and storage of command-line flags and
// other application-wide configuration.
package config

import (
	"fmt"
	"os"
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
	Breakpoints       []Location // parsed --break-at values, consumed once reached
	MinutesToTimeout  int
	UpdateFile        bool
	Verbose           bool
	View              string
	Env               []string
	Rootdir           string
}

// Location designates a stage or a chunk, as given by the --start-from and
// --break-at flags: stage, stage/chunkID or file@stage/chunkID.
type Location struct {
	// File restricts the location to a given markdown file, empty for any file.
	File string
	// Stage is the name of the targeted stage.
	Stage string
	// Chunk is the ID or 0-based index of the targeted chunk, empty for the
	// whole stage.
	Chunk string
}

// ParseLocation parses a --start-from or --break-at value: stage,
// stage/chunkID or file@stage/chunkID.
func ParseLocation(value string) (Location, error) {
	var loc Location
	formatErr := fmt.Errorf("invalid format %q. Use 'stage', 'stage/chunkID', or 'file@stage/chunkID'", value)
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
  -B, --break-at string      Start debugging from a specific stage or chunk (stage, stage/chunkID, or file@stage/chunkID), can be repeated
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
	pflag.StringArrayVarP(&cfg.DebugFrom, "break-at", "B", nil, "Start debugging from a specific stage or chunk (stage, stage/chunkID, or file@stage/chunkID), can be repeated")
	pflag.IntVarP(&cfg.MinutesToTimeout, "timeout", "t", 10, "The timeout in minutes for every executed command")
	pflag.BoolVarP(&cfg.UpdateFile, "update-files", "u", false, "Update the chunk output section in the markdown files")
	pflag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Print more logs")
	pflag.StringVar(&cfg.View, "view", "default", "UI to be used, can be 'default' or 'ci'")

	pflag.Parse()

	// Parse start-from format: stage, stage/chunkID, or file@stage/chunkID
	if cfg.StartFrom != "" {
		loc, err := ParseLocation(cfg.StartFrom)
		if err != nil {
			pterm.Fatal.Println("start-from: " + err.Error())
		}
		cfg.StartFromLocation = loc
	}

	// Parse break-at format: stage, stage/chunkID, or file@stage/chunkID
	for _, value := range cfg.DebugFrom {
		bp, err := ParseLocation(value)
		if err != nil {
			pterm.Fatal.Println("break-at: " + err.Error())
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
