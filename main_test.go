package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
			if err != nil {
				t.Fatalf("findMarkdownFiles returned an error: %v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}
