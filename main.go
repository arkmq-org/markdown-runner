package main

import (
	"flag"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/runner"
	"github.com/pterm/pterm"
)

func findMarkdownFiles(dir string, recursive bool) ([]string, error) {
	var files []string
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(dir, entry.Name()))
		} else if recursive {
			subDirFiles, err := findMarkdownFiles(filepath.Join(dir, entry.Name()), recursive)
			if err != nil {
				return nil, err
			}
			files = append(files, subDirFiles...)
		}
	}
	return files, nil
}

// main is the entrypoint for the markdown-runner application. It handles
// command-line flag parsing, finds the markdown files to be executed, and
// orchestrates the execution process by calling the runner.
func main() {
	config.ParseFlags()
	validExtensions := []string{".md", ".MD", ".Markdown", ".markdown"}

	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	markdown_files, err := findMarkdownFiles(config.MarkdownDir, config.Recursive)
	pterm.Fatal.PrintOnError(err)

	if config.NoStyling {
		pterm.DisableStyling()
	}
	if config.Quiet {
		pterm.DisableOutput()
	}
	config.Env = append(config.Env, os.Environ()...)
	workding_directory, err := os.Getwd()
	pterm.Fatal.PrintOnError(err)
	config.Env = append(config.Env, "WORKING_DIR="+workding_directory)

	for _, file := range markdown_files {
		// avoid parsing files that aren't markdown
		extension := path.Ext(file)
		if !slices.Contains(validExtensions, extension) {
			continue
		}

		if config.Filter != "" {
			matched, err := regexp.MatchString(config.Filter, file)
			if err != nil {
				pterm.Fatal.Printf("Invalid regex for -f/--filter: %v\n", err)
			}
			if !matched {
				pterm.Info.Println("Ignoring", file, "as it does not match the filter:", config.Filter)
				continue
			}
		}

		// only list the available files
		if config.JustList {
			pterm.Info.Println(file)
			continue
		}
		// parse and execute if possible
		err := runner.RunMD(file)
		if err != nil {
			pterm.Error.PrintOnError(err)
			os.Exit(1)
		}
	}
}
