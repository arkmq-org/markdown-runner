package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("should show help", func(t *testing.T) {
		os.Args = []string{"markdown-runner", "-h"}
		err := run()
		assert.NoError(t, err)
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	})

	t.Run("should list files", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		file1 := filepath.Join(tmpDir, "test.md")
		err = os.WriteFile(file1, []byte(""), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		os.Args = []string{"markdown-runner", "-l", tmpDir}
		err = run()
		assert.NoError(t, err)
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	})

	t.Run("should filter files", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		file1 := filepath.Join(tmpDir, "test.md")
		err = os.WriteFile(file1, []byte(""), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		file2 := filepath.Join(tmpDir, "other.md")
		err = os.WriteFile(file2, []byte(""), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		os.Args = []string{"markdown-runner", "-f", "other.md", tmpDir}
		err = run()
		assert.NoError(t, err)
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	})

	t.Run("should return error for invalid path", func(t *testing.T) {
		os.Args = []string{"markdown-runner", "/invalid/path"}
		err := run()
		assert.Error(t, err, "Expected an error for an invalid path")
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	})

	t.Run("should return error for invalid regex", func(t *testing.T) {
		os.Args = []string{"markdown-runner", "-f", "["}
		err := run()
		assert.Error(t, err, "Expected an error for an invalid regex")
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	})

	t.Run("should not fail with no filter match", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		file1 := filepath.Join(tmpDir, "test.md")
		err = os.WriteFile(file1, []byte(""), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		os.Args = []string{"markdown-runner", "-f", "no-match", tmpDir}
		err = run()
		assert.NoError(t, err)
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	})

	t.Run("should not fail with invalid extension", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		file1 := filepath.Join(tmpDir, "test.txt")
		err = os.WriteFile(file1, []byte(""), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		os.Args = []string{"markdown-runner", tmpDir}
		err = run()
		assert.NoError(t, err)
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	})
}

func TestMainFunc(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	}()

	os.Args = []string{"markdown-runner", "-h"}
	main()
}

func TestFindMarkdownFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files and directories
	//-tmpDir
	//  |-test.md
	//  |-subdir
	//    |-subtest.md

	file1 := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(file1, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	file2 := filepath.Join(subDir, "subtest.md")
	if err := os.WriteFile(file2, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Test cases
	testCases := []struct {
		name      string
		path      string
		recursive bool
		expected  []string
	}{
		{
			name:      "Single file",
			path:      file1,
			recursive: false,
			expected:  []string{file1},
		},
		{
			name:      "Directory with recursion",
			path:      tmpDir,
			recursive: true,
			expected:  []string{file2, file1},
		},
		{
			name:      "Directory without recursion",
			path:      tmpDir,
			recursive: false,
			expected:  []string{file1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := findMarkdownFiles(tc.path, tc.recursive)
			assert.NoError(t, err, "findMarkdownFiles returned an error")
			assert.ElementsMatch(t, tc.expected, result, "The returned files are not as expected")
		})
	}

	t.Run("should return error for invalid path", func(t *testing.T) {
		_, err := findMarkdownFiles("/invalid/path", false)
		assert.Error(t, err, "Expected an error for an invalid path")
	})
}
