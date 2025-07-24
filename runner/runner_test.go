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

func TestRunMD(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Unexpected error")
	})

	t.Run("teardown", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
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

		err = RunMD(cfg, mdFile)
		assert.Error(t, err, "Expected an error, but got none")
	})

	t.Run("update file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, UpdateFile: true, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"test"}
echo "hello"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
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
	})

	t.Run("start from", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, StartFrom: "stage2", MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {"stage":"stage1"}
echo "should not run"
` + "```" + `
` + "```" + `bash {"stage":"stage2"}
echo "should run"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Unexpected error")
	})

	t.Run("no stages", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
		mdContent := `
# No Stages Here
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Unexpected error")
	})

	t.Run("parser error", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
		mdContent := `
` + "```" + `bash {stage}
echo "bad json"
` + "```" + `
`
		mdFile := path.Join(tmpDir, "test.md")
		err = os.WriteFile(mdFile, []byte(mdContent), 0o644)
		assert.NoError(t, err, "Failed to write to temp file")

		err = RunMD(cfg, mdFile)
		assert.Error(t, err, "Expected a parser error")
	})

	t.Run("teardown with dependency", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)
		outputFile := path.Join(tmpDir, "teardown_check")

		cfg := &config.Config{MarkdownDir: tmpDir, MinutesToTimeout: 1}
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

		err = RunMD(cfg, mdFile)
		assert.NoError(t, err, "Expected no error")

		_, err = os.Stat(outputFile)
		assert.NoError(t, err, "Expected teardown chunk to be executed")
	})
}
