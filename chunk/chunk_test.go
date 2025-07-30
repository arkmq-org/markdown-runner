package chunk_test

import (
	"bufio"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/arkmq-org/markdown-runner/chunk"
	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/runnercontext"
	"github.com/arkmq-org/markdown-runner/view"
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

func TestExecutableChunk(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		testChunk := chunk.ExecutableChunk{HasBreakpoint: true, Context: &runnercontext.Context{RView: view.NewView("mock")}}
		testChunk.Init()
		assert.NotNil(t, testChunk.Content, "Expected Content to be initialized, but it was nil")
	})
	t.Run("has output", func(t *testing.T) {
		testCases := []struct {
			name     string
			chunk    *chunk.ExecutableChunk
			expected bool
		}{
			{
				name: "Chunk with no commands has no output",
				chunk: &chunk.ExecutableChunk{
					Commands: []*chunk.RunningCommand{},
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
					},
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
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
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
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
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
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
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
	})
	t.Run("get or create runtime directory", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		defer os.RemoveAll(tmpDirs["$tmpdir.test"])

		testChunk := chunk.ExecutableChunk{RootDir: "$initial_dir", Context: &runnercontext.Context{Cfg: &config.Config{MinutesToTimeout: 1, Rootdir: "/tmp"}, RView: view.NewView("mock")}}
		dir, err := testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
		assert.NoError(t, err)
		assert.Equal(t, "/tmp", dir)

		testChunk = chunk.ExecutableChunk{RootDir: "$tmpdir.test", Context: &runnercontext.Context{Cfg: &config.Config{MinutesToTimeout: 1}, RView: view.NewView("mock")}}
		dir, err = testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(dir, "/tmp"))
		_, ok := tmpDirs["$tmpdir.test"]
		assert.True(t, ok, "Expected tmpdir to be created and stored")

		testChunk = chunk.ExecutableChunk{RootDir: "/custom/dir", Context: &runnercontext.Context{Cfg: &config.Config{MinutesToTimeout: 1}, RView: view.NewView("mock")}}
		dir, err = testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
		assert.NoError(t, err)
		assert.Equal(t, "/custom/dir", dir)

		testChunk = chunk.ExecutableChunk{RootDir: "", Context: &runnercontext.Context{Cfg: &config.Config{MinutesToTimeout: 1}, RView: view.NewView("mock")}}
		dir, err = testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
		assert.NoError(t, err)
		defer os.RemoveAll(dir)
		assert.True(t, strings.HasPrefix(dir, "/tmp"))
	})
	t.Run("add command to execute", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		testChunk := chunk.ExecutableChunk{Context: &runnercontext.Context{Cfg: &config.Config{MinutesToTimeout: 1}, RView: view.NewView("mock")}}
		cmdStr := "echo 'hello world'"
		cmd, err := testChunk.AddCommandToExecute(cmdStr, tmpDirs)
		assert.NoError(t, err)
		assert.Equal(t, "echo", cmd.Cmd.Args[0])
		assert.Equal(t, "hello world", cmd.Cmd.Args[1])
	})
	t.Run("execute error", func(t *testing.T) {
		testChunk := chunk.ExecutableChunk{
			IsParallel: true,
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: view.NewView("mock"),
			},
		}
		err := testChunk.ExecuteSequential()
		assert.Error(t, err, "Expected an error when executing a parallel chunk with Execute")
	})
	t.Run("prepare for execution", func(t *testing.T) {
		t.Run("it should return an error if a parallel non-bash chunk has more than one command", func(t *testing.T) {
			c := &chunk.ExecutableChunk{
				IsParallel: true,
				Content:    []string{"echo 1", "echo 2"},
				Context: &runnercontext.Context{
					Cfg:   &config.Config{MinutesToTimeout: 1},
					RView: view.NewView("mock"),
				},
			}
			err := c.PrepareForExecution(make(map[string]string))
			assert.Error(t, err)
		})

		t.Run("it should not return an error if a parallel bash chunk has more than one command", func(t *testing.T) {
			c := &chunk.ExecutableChunk{
				IsParallel: true,
				Runtime:    "bash",
				Content:    []string{"echo 1", "echo 2"},
				Context: &runnercontext.Context{
					Cfg:   &config.Config{MinutesToTimeout: 1},
					RView: view.NewView("mock"),
				},
			}
			err := c.PrepareForExecution(make(map[string]string))
			assert.NoError(t, err)
		})
	})
	t.Run("execute", func(t *testing.T) {
		t.Run("it should execute a command", func(t *testing.T) {
			c := &chunk.ExecutableChunk{
				Content: []string{"echo 1"},
				Context: &runnercontext.Context{
					Cfg:   &config.Config{MinutesToTimeout: 1},
					RView: view.NewView("mock"),
				},
			}
			err := c.PrepareForExecution(make(map[string]string))
			assert.NoError(t, err)
			err = c.ExecuteSequential()
			assert.NoError(t, err)
		})
	})
	t.Run("wait", func(t *testing.T) {
		t.Run("it should wait for a command", func(t *testing.T) {
			ui := view.NewView("mock")
			c := &chunk.ExecutableChunk{
				IsParallel: true,
				Content:    []string{"sleep 0.1"},
				Context: &runnercontext.Context{
					Cfg:   &config.Config{MinutesToTimeout: 1},
					RView: ui,
				},
			}
			err := c.PrepareForExecution(make(map[string]string))
			assert.NoError(t, err)
			err = c.DeclareParallelLoggers()
			assert.NoError(t, err)
			err = c.StartParallel()
			assert.NoError(t, err)
			err = c.WaitParallel(false)
			assert.NoError(t, err)
		})

		t.Run("it should kill a command", func(t *testing.T) {
			ui := view.NewView("mock")
			c := &chunk.ExecutableChunk{
				IsParallel: true,
				Content:    []string{"sleep 1"},
				Context: &runnercontext.Context{
					Cfg:   &config.Config{MinutesToTimeout: 1},
					RView: ui,
				},
			}
			err := c.PrepareForExecution(make(map[string]string))
			assert.NoError(t, err)
			err = c.DeclareParallelLoggers()
			assert.NoError(t, err)
			err = c.StartParallel()
			assert.NoError(t, err)
			err = c.WaitParallel(true)
			assert.NoError(t, err)
		})
	})
	t.Run("bash script execution", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		testChunk := chunk.ExecutableChunk{
			Runtime: "bash",
			Content: []string{"export GREETING='hello from bash'", "echo $GREETING"},
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: view.NewView("mock"),
			},
		}
		err = testChunk.PrepareForExecution(make(map[string]string))
		assert.NoError(t, err, "Failed to write the bash script to disk or prepare for execution")

		err = testChunk.ExecuteSequential()
		assert.NoError(t, err, "Bash script execution failed")

		assert.Contains(t, testChunk.Commands[0].Stdout, "hello from bash")

		found := slices.ContainsFunc(testChunk.Context.Cfg.Env, func(env string) bool {
			return env == "GREETING=hello from bash"
		})
		assert.True(t, found, "Expected GREETING to be in the environment variables")
	})
	t.Run("write bash script read only dir", func(t *testing.T) {
		readOnlyDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(readOnlyDir)

		err = os.Chmod(readOnlyDir, 0o444)
		assert.NoError(t, err, "Failed to change directory permissions")

		testChunk := chunk.ExecutableChunk{Context: &runnercontext.Context{RView: view.NewView("mock")}}
		err = testChunk.WriteBashScript(readOnlyDir, "test.sh")
		assert.Error(t, err, "Expected an error when writing to a read-only directory")
	})
	t.Run("apply writer", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		testChunk := chunk.ExecutableChunk{
			Runtime:     "writer",
			Destination: "test.txt",
			Content:     []string{"hello", "world"},
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: view.NewView("mock"),
			},
		}
		err := testChunk.PrepareForExecution(tmpDirs)
		assert.NoError(t, err, "Failed to apply writer")

		dir, err := testChunk.GetOrCreateRuntimeDirectory(tmpDirs)
		assert.NoError(t, err, "Failed to get runtime directory")
		content, err := os.ReadFile(path.Join(dir, "test.txt"))
		assert.NoError(t, err, "Failed to read file")
		assert.Equal(t, "hello\nworld\n", string(content))
	})
	t.Run("write output to no newline", func(t *testing.T) {
		testChunk := chunk.ExecutableChunk{
			Commands: []*chunk.RunningCommand{
				{
					Stdout: "hello world",
				},
			},
			Context: &runnercontext.Context{RView: view.NewView("mock")},
		}
		var writer strings.Builder
		bufWriter := bufio.NewWriter(&writer)
		err := testChunk.WriteOutputTo(3, bufWriter)
		assert.NoError(t, err)
		bufWriter.Flush()
		expectedOutput := "```shell markdown_runner\nhello world\n```\n"
		assert.Equal(t, expectedOutput, writer.String())
	})
	t.Run("write file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err, "Failed to create temp dir")
		defer os.RemoveAll(tmpDir)

		testChunk := chunk.ExecutableChunk{
			Runtime:     "writer",
			Destination: "test.txt",
			Content:     []string{"hello", "world"},
			Context:     &runnercontext.Context{RView: view.NewView("mock")},
		}

		err = testChunk.WriteFile(tmpDir)
		assert.NoError(t, err, "Failed to write file")

		content, err := os.ReadFile(tmpDir + "/test.txt")
		assert.NoError(t, err, "Failed to read file")
		assert.Equal(t, "hello\nworld\n", string(content))
	})
	t.Run("write file error", func(t *testing.T) {
		testChunk := chunk.ExecutableChunk{
			Runtime:     "writer",
			Destination: "test.txt",
			Content:     []string{"hello", "world"},
			Context:     &runnercontext.Context{RView: view.NewView("mock")},
		}

		err := testChunk.WriteFile("/invalid/dir")
		assert.Error(t, err, "Expected an error when writing to an invalid directory")
	})
	t.Run("apply writer error", func(t *testing.T) {
		testChunk := chunk.ExecutableChunk{
			Runtime:     "writer",
			Destination: "test.txt",
			Content:     []string{"hello", "world"},
			RootDir:     "/invalid/dir",
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: view.NewView("mock"),
			},
		}
		err := testChunk.PrepareForExecution(make(map[string]string))
		assert.Error(t, err, "Expected an error when writing to an invalid directory")
	})
	t.Run("prepare classical error", func(t *testing.T) {
		testChunk := chunk.ExecutableChunk{
			Content: []string{"'"},
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1},
				RView: view.NewView("mock"),
			},
		}
		err := testChunk.PrepareForExecution(make(map[string]string))
		assert.Error(t, err, "Expected an error for an invalid command")
	})
	t.Run("add command to execute error", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		testChunk := chunk.ExecutableChunk{Context: &runnercontext.Context{Cfg: &config.Config{MinutesToTimeout: 1}, RView: view.NewView("mock")}}
		_, err := testChunk.AddCommandToExecute("", tmpDirs)
		assert.Error(t, err, "Expected an error for an empty command string")
	})
	t.Run("prepare bash chunk for execution error", func(t *testing.T) {
		tmpDirs := make(map[string]string)
		testChunk := chunk.ExecutableChunk{Context: &runnercontext.Context{Cfg: &config.Config{MinutesToTimeout: 1}, RView: view.NewView("mock")}}

		// This is not a valid command, so it should fail.
		testChunk.Content = []string{"'"}
		err := testChunk.PrepareForExecution(tmpDirs)
		assert.Error(t, err, "Expected an error for an invalid command")
	})
	t.Run("write output to", func(t *testing.T) {
		testChunk := chunk.ExecutableChunk{
			Commands: []*chunk.RunningCommand{
				{
					Stdout: "hello world",
					Stderr: "this is an error",
				},
			},
			Context: &runnercontext.Context{RView: view.NewView("mock")},
		}
		var writer strings.Builder
		bufWriter := bufio.NewWriter(&writer)
		err := testChunk.WriteOutputTo(3, bufWriter)
		assert.NoError(t, err)
		bufWriter.Flush()
		expectedOutput := "```shell markdown_runner\nhello world\nthis is an error\n```\n"
		assert.Equal(t, expectedOutput, writer.String())
	})
	t.Run("has output dry run", func(t *testing.T) {
		testChunk := &chunk.ExecutableChunk{
			Commands: []*chunk.RunningCommand{},
			Context: &runnercontext.Context{
				Cfg:   &config.Config{DryRun: true},
				RView: view.NewView("mock"),
			},
		}
		assert.True(t, testChunk.HasOutput(), "Expected HasOutput to be true in dry run mode")
	})
	t.Run("parallel functions on non parallel chunk", func(t *testing.T) {
		c := &chunk.ExecutableChunk{IsParallel: false, Context: &runnercontext.Context{RView: view.NewView("mock")}}
		err := c.DeclareParallelLoggers()
		assert.Error(t, err)

		err = c.StartParallel()
		assert.Error(t, err)

		err = c.WaitParallel(false)
		assert.Error(t, err)
	})
	t.Run("execution status", func(t *testing.T) {
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
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
					},
				},
				expectedFinished:          false,
				expectedCorrectlyExecuted: false,
			},
			{
				name: "Chunk with a successful command",
				chunk: &chunk.ExecutableChunk{
					Commands: []*chunk.RunningCommand{successCmd},
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
					},
				},
				expectedFinished:          true,
				expectedCorrectlyExecuted: true,
			},
			{
				name: "Chunk with a failed command",
				chunk: &chunk.ExecutableChunk{
					Commands: []*chunk.RunningCommand{failCmd},
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
					},
				},
				expectedFinished:          true,
				expectedCorrectlyExecuted: false,
			},
			{
				name: "Chunk with mixed commands (one not run)",
				chunk: &chunk.ExecutableChunk{
					Commands: []*chunk.RunningCommand{successCmd, notRunCmd},
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
					},
				},
				expectedFinished:          false,
				expectedCorrectlyExecuted: false,
			},
			{
				name: "Chunk with mixed commands (one failed)",
				chunk: &chunk.ExecutableChunk{
					Commands: []*chunk.RunningCommand{successCmd, failCmd},
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
					},
				},
				expectedFinished:          true,
				expectedCorrectlyExecuted: false,
			},
			{
				name: "Empty chunk",
				chunk: &chunk.ExecutableChunk{
					Context: &runnercontext.Context{
						Cfg:   &config.Config{MinutesToTimeout: 1},
						RView: view.NewView("mock"),
					},
				},
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
	})
}
