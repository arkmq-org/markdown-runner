package runner

import (
	"os"
	"path"
	"testing"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	pterm.DisableOutput()
	code := m.Run()
	pterm.EnableOutput()
	os.Exit(code)
}

func TestRunMD_HappyPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	config.MarkdownDir = tmpDir
	mdContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	err = RunMD(mdFile)
	assert.NoError(t, err, "Unexpected error")
}

func TestRunMD_Teardown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	config.MarkdownDir = tmpDir
	mdContent := `
` + "```" + `bash {"stage":"main", "id":"main-chunk"}
exit 1
` + "```" + `

` + "```" + `bash {"stage":"teardown", "requires":"main/main-chunk"}
echo "teardown"
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	err = RunMD(mdFile)
	assert.Error(t, err, "Expected an error, but got none")
}

func TestRunMD_UpdateFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	config.MarkdownDir = tmpDir
	config.UpdateFile = true
	defer func() { config.UpdateFile = false }()
	mdContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	err = RunMD(mdFile)
	assert.NoError(t, err, "Unexpected error")

	updatedContent, err := os.ReadFile(mdFile)
	assert.NoError(t, err, "Failed to read updated file")
	expectedContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
` + "```" + `shell markdown_runner
hello
` + "```" + `
`
	assert.Equal(t, expectedContent, string(updatedContent), "File content is not as expected")
}

func TestRunMD_TeardownWithDependency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)
	outputFile := path.Join(tmpDir, "teardown_check")

	config.MarkdownDir = tmpDir
	mdContent := `
` + "```" + `bash {"stage":"main", "id":"main-chunk"}
echo "main chunk executed"
` + "```" + `

` + "```" + `bash {"stage":"teardown", "requires":"main/main-chunk", "runtime":"bash"}
echo "teardown executed" > ` + outputFile + `
` + "```" + `
`
	mdFile := path.Join(tmpDir, "test.md")
	err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
	assert.NoError(t, err, "Failed to write to temp file")

	err = RunMD(mdFile)
	assert.NoError(t, err, "Expected no error")

	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "Expected teardown chunk to be executed")
}
