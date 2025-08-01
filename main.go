package main

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/runner"
	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
)

// findMarkdownFiles recursively finds all files in a given path. If the
// provided path is a file, it will be returned directly. If the path is a
// directory and recursive is true, it will traverse into subdirectories.
func findMarkdownFiles(path string, recursive bool) ([]string, error) {
	info, statErr := os.Stat(path)
	if statErr != nil {
		return nil, statErr
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	var files []string
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range dirEntries {
		info, statErr := os.Stat(filepath.Join(path, entry.Name()))
		if statErr != nil {
			return nil, statErr
		}
		if info.IsDir() && !recursive {
			continue
		}
		subDirFiles, err := findMarkdownFiles(filepath.Join(path, entry.Name()), recursive)
		if err != nil {
			return nil, err
		}
		files = append(files, subDirFiles...)
	}
	sort.Strings(files)
	return files, nil
}

// main is the entrypoint for the markdown-runner application. It handles
// command-line flag parsing, finds the markdown files to be executed, and
// orchestrates the execution process by calling the runner.
func main() {
	if err := run(); err != nil {
		pterm.Error.PrintOnError(err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.NewConfig()
	if cfg.Help {
		pflag.Usage()
		return nil
	}
	validExtensions := []string{".md", ".MD", ".Markdown", ".markdown"}

	markdown_files, err := findMarkdownFiles(cfg.MarkdownDir, cfg.Recursive)
	if err != nil {
		return err
	}

	if cfg.NoStyling {
		pterm.DisableStyling()
	}
	if cfg.Quiet {
		pterm.DisableOutput()
	}
	cfg.Env = append(cfg.Env, os.Environ()...)
	workding_directory, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg.Env = append(cfg.Env, "WORKING_DIR="+workding_directory)

	for _, file := range markdown_files {
		// avoid parsing files that aren't markdown
		extension := path.Ext(file)
		if !slices.Contains(validExtensions, extension) {
			continue
		}

		if cfg.Filter != "" {
			matched, err := regexp.MatchString(cfg.Filter, file)
			if err != nil {
				return err
			}
			if !matched {
				pterm.Info.Println("Ignoring", file, "as it does not match the filter:", cfg.Filter)
				continue
			}
		}

		// only list the available files
		if cfg.JustList {
			pterm.Info.Println(file)
			continue
		}
		// parse and execute if possible
		err := runner.RunMD(cfg, file)
		if err != nil {
			return err
		}
	}
	return nil
}
