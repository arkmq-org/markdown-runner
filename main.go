package main

import (
	"flag"
	"os"
	"path"
	"slices"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/runner"
	"github.com/pterm/pterm"
)

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
	markdown_files, err := os.ReadDir(config.MarkdownDir)
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

	for _, e := range markdown_files {
		// avoid parsing files that aren't markdown
		extension := path.Ext(e.Name())
		if !slices.Contains(validExtensions, extension) {
			continue
		}
		// only list the available files
		if config.JustList {
			pterm.Info.Println(e.Name())
			continue
		}
		// if a single markdown file is selected, only execute this one
		if config.File != "" {
			if path.Base(config.File) != e.Name() {
				pterm.Info.Println("Ignoring", e.Name())
				continue
			}
		}
		// parse and execute if possible
		err := runner.RunMD(e.Name())
		if err != nil {
			pterm.Error.PrintOnError(err)
			os.Exit(1)
		}
	}
}
