package chunk_test

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) {
	pterm.DisableOutput()
}

func teardown(t *testing.T) {
	pterm.EnableOutput()
}

func TestMain(m *testing.M) {
	setup(nil)
	code := m.Run()
	teardown(nil)
	os.Exit(code)
}

func TestExecutableChunk_Init(t *testing.T) {
	testChunk := chunk.ExecutableChunk{HasBreakpoint: true}
	testChunk.Init()
	assert.NotNil(t, testChunk.Content, "Expected Content to be initialized, but it was nil")
}

func TestHasOutput(t *testing.T) {
	testCases := []struct {
		name     string
		chunk    *chunk.ExecutableChunk
		expected bool
	}{
		{
			name: "Chunk with no commands has no output",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{},
			},
			expected: false,
		},
		{
			name: "Chunk with command with no output has no output",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{
					{
						Stdout: "",
						Stderr: "",
					},
				},
			},
			expected: false,
		},
		{
			name: "Chunk with command with stdout has output",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{
					{
						Stdout: "some output",
						Stderr: "",
					},
				},
			},
			expected: true,
		},
		{
			name: "Chunk with command with stderr has output",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{
					{
						Stdout: "",
						Stderr: "some error",
					},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.chunk.HasOutput())
		})
	}
}

func TestGetOrCreateRuntimeDirectory(t *testing.T) {
	tmpDirs := make(map[string]string)
	defer os.RemoveAll(tmpDirs["$tmpdir.test"])

	testChunk := chunk.ExecutableChunk{RootDir: "$initial_dir"}
	dir, err := testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
	assert.NoError(t, err)
	assert.Equal(t, config.Rootdir, dir)

	testChunk = chunk.ExecutableChunk{RootDir: "$tmpdir.test"}
	dir, err = testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(dir, "/tmp"))
	_, ok := tmpDirs["$tmpdir.test"]
	assert.True(t, ok, "Expected tmpdir to be created and stored")

	testChunk = chunk.ExecutableChunk{RootDir: "/custom/dir"}
	dir, err = testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
	assert.NoError(t, err)
	assert.Equal(t, "/custom/dir", dir)

	testChunk = chunk.ExecutableChunk{RootDir: ""}
	dir, err = testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	assert.True(t, strings.HasPrefix(dir, "/tmp"))
}

func TestAddCommandToExecute(t *testing.T) {
	tmpDirs := make(map[string]string)
	testChunk := chunk.ExecutableChunk{}
	cmdStr := "echo 'hello world'"
	cmd, err := testChunk.AddCommandToExecute(cmdStr, tmpDirs)
	assert.NoError(t, err)
	assert.Equal(t, "echo", cmd.Cmd.Args[0])
	assert.Equal(t, "hello world", cmd.Cmd.Args[1])
}

func TestBashScriptExecution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	testChunk := chunk.ExecutableChunk{
		Runtime: "bash",
		Content: []string{"export GREETING='hello from bash'", "echo $GREETING"},
	}
	cmd, err := testChunk.AddCommandToExecute("./test.sh", map[string]string{"$tmpdir.1": tmpDir})
	assert.NoError(t, err, "Failed to add command")
	cmd.Cmd.Dir = tmpDir
	err = testChunk.WriteBashScript(tmpDir, "test.sh")
	assert.NoError(t, err, "Failed to write bash script")

	err = cmd.Execute()
	assert.NoError(t, err, "Bash script execution failed")

	assert.Contains(t, cmd.Stdout, "hello from bash")

	found := false
	for _, envVar := range config.Env {
		if envVar == "GREETING=hello from bash" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected GREETING to be in the environment variables")
}

func TestWriteFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	testChunk := chunk.ExecutableChunk{
		Runtime:     "writer",
		Destination: "test.txt",
		Content:     []string{"hello", "world"},
	}

	err = testChunk.WriteFile(tmpDir)
	assert.NoError(t, err, "Failed to write file")

	content, err := os.ReadFile(tmpDir + "/test.txt")
	assert.NoError(t, err, "Failed to read file")
	assert.Equal(t, "hello\nworld\n", string(content))
}

func TestWriteFile_Error(t *testing.T) {
	testChunk := chunk.ExecutableChunk{
		Runtime:     "writer",
		Destination: "test.txt",
		Content:     []string{"hello", "world"},
	}

	err := testChunk.WriteFile("/invalid/dir")
	assert.Error(t, err, "Expected an error when writing to an invalid directory")
}

func TestAddCommandToExecute_Error(t *testing.T) {
	tmpDirs := make(map[string]string)
	testChunk := chunk.ExecutableChunk{}
	_, err := testChunk.AddCommandToExecute("", tmpDirs)
	assert.Error(t, err, "Expected an error for an empty command string")
}

func TestWriteOutputTo(t *testing.T) {
	testChunk := chunk.ExecutableChunk{
		Commands: []*chunk.RunningCommand{
			{
				Stdout: "hello world",
				Stderr: "this is an error",
			},
		},
	}
	var writer strings.Builder
	bufWriter := bufio.NewWriter(&writer)
	err := testChunk.WriteOutputTo(3, bufWriter)
	assert.NoError(t, err)
	bufWriter.Flush()
	expectedOutput := "```shell markdown_runner\nhello world\nthis is an error\n```\n"
	assert.Equal(t, expectedOutput, writer.String())
}

func TestExecutionStatus(t *testing.T) {
	// Command that has not been run
	notRunCmd := &chunk.RunningCommand{Cmd: exec.Command("echo", "not run")}

	// Command that has run and succeeded
	successCmd := &chunk.RunningCommand{Cmd: exec.Command("true")}
	_ = successCmd.Cmd.Run()

	// Command that has run and failed
	failCmd := &chunk.RunningCommand{Cmd: exec.Command("false")}
	_ = failCmd.Cmd.Run()

	testCases := []struct {
		name                      string
		chunk                     *chunk.ExecutableChunk
		expectedFinished          bool
		expectedCorrectlyExecuted bool
	}{
		{
			name: "Chunk with a not-run command",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{notRunCmd},
			},
			expectedFinished:          false,
			expectedCorrectlyExecuted: false,
		},
		{
			name: "Chunk with a successful command",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{successCmd},
			},
			expectedFinished:          true,
			expectedCorrectlyExecuted: true,
		},
		{
			name: "Chunk with a failed command",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{failCmd},
			},
			expectedFinished:          true,
			expectedCorrectlyExecuted: false,
		},
		{
			name: "Chunk with mixed commands (one not run)",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{successCmd, notRunCmd},
			},
			expectedFinished:          false,
			expectedCorrectlyExecuted: false,
		},
		{
			name: "Chunk with mixed commands (one failed)",
			chunk: &chunk.ExecutableChunk{
				Commands: []*chunk.RunningCommand{successCmd, failCmd},
			},
			expectedFinished:          true,
			expectedCorrectlyExecuted: false,
		},
		{
			name:                      "Empty chunk",
			chunk:                     &chunk.ExecutableChunk{},
			expectedFinished:          false,
			expectedCorrectlyExecuted: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedFinished, tc.chunk.HasFinishedExecution(), "Finished status was not as expected")
			assert.Equal(t, tc.expectedCorrectlyExecuted, tc.chunk.HasExecutedCorrectly(), "Correctly executed status was not as expected")
		})
	}
}
