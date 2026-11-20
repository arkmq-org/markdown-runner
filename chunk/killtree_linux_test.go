//go:build linux

package chunk

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arkmq-org/markdown-runner/config"
	"github.com/arkmq-org/markdown-runner/runnercontext"
	"github.com/arkmq-org/markdown-runner/view"
	"github.com/stretchr/testify/assert"
)

// waitForPid polls the file in which a test script writes a pid.
func waitForPid(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil && strings.HasSuffix(string(data), "\n") {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err == nil {
				return pid
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no pid written to %s", path)
	return 0
}

// assertProcessGone polls until the process doesn't exist anymore, a zombie
// waiting to be reaped counting as gone.
func assertProcessGone(t *testing.T, pid int, what string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		stat, ok := readProcStat(pid)
		if !ok || strings.ContainsAny(stat.state, "ZXx") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("the %s (pid %d) is still running", what, pid)
}

// killOnCleanup makes sure a test doesn't leave a process behind when it fails.
func killOnCleanup(t *testing.T, pids ...int) {
	t.Cleanup(func() {
		for _, pid := range pids {
			if p, err := os.FindProcess(pid); err == nil {
				_ = p.Kill()
			}
		}
	})
}

func TestKillProcessTree(t *testing.T) {
	t.Run("timeout kills the background and foreground children of a bash chunk", func(t *testing.T) {
		useTimeoutUnit(t, 2*time.Second)
		dir := t.TempDir()
		bg, fg := filepath.Join(dir, "bg"), filepath.Join(dir, "fg")
		chunk := ExecutableChunk{
			Label:   "tree",
			Runtime: "bash",
			Content: []string{
				"sleep 30 & echo $! > " + bg,
				// exec keeps the pid written by sh for the foreground sleep
				"sh -c 'echo $$ > " + fg + "; exec sleep 30'",
			},
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1, Env: []string{}},
				RView: view.NewView("mock"),
			},
		}
		assert.NoError(t, chunk.PrepareForExecution(map[string]string{}))
		start := time.Now()
		err := chunk.ExecuteSequential()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tree: timed out after 2s")
		assert.Less(t, time.Since(start), 20*time.Second)

		bgPid, fgPid := waitForPid(t, bg), waitForPid(t, fg)
		killOnCleanup(t, bgPid, fgPid)
		assertProcessGone(t, bgPid, "background sleep")
		assertProcessGone(t, fgPid, "foreground sleep")
	})

	t.Run("timeout kills nested descendants", func(t *testing.T) {
		useTimeoutUnit(t, 2*time.Second)
		dir := t.TempDir()
		inner, leaf := filepath.Join(dir, "inner"), filepath.Join(dir, "leaf")
		chunk := ExecutableChunk{
			Runtime: "bash",
			Content: []string{
				// bash -> bash -> sleep
				"bash -c 'echo $$ > " + inner + "; sleep 30 & echo $! > " + leaf + "; wait'",
			},
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1, Env: []string{}},
				RView: view.NewView("mock"),
			},
		}
		assert.NoError(t, chunk.PrepareForExecution(map[string]string{}))
		err := chunk.ExecuteSequential()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out after 2s")

		innerPid, leafPid := waitForPid(t, inner), waitForPid(t, leaf)
		killOnCleanup(t, innerPid, leafPid)
		assertProcessGone(t, innerPid, "inner bash")
		assertProcessGone(t, leafPid, "nested sleep")
	})

	t.Run("kill kills the descendants", func(t *testing.T) {
		dir := t.TempDir()
		bg, fg := filepath.Join(dir, "bg"), filepath.Join(dir, "fg")
		chunk := ExecutableChunk{
			Runtime: "bash",
			Content: []string{
				"sleep 30 & echo $! > " + bg,
				"sh -c 'echo $$ > " + fg + "; exec sleep 30'",
			},
			Context: &runnercontext.Context{
				Cfg:   &config.Config{MinutesToTimeout: 1, Env: []string{}},
				RView: view.NewView("mock"),
			},
		}
		assert.NoError(t, chunk.PrepareForExecution(map[string]string{}))
		command := chunk.Commands[0]
		assert.NoError(t, command.InitializeLogger())
		assert.NoError(t, command.Start())
		bgPid, fgPid := waitForPid(t, bg), waitForPid(t, fg)
		killOnCleanup(t, bgPid, fgPid)

		assert.NoError(t, command.Kill())
		assertProcessGone(t, bgPid, "background sleep")
		assertProcessGone(t, fgPid, "foreground sleep")
		_ = command.Cmd.Wait()
		command.CancelFunc()
	})

	t.Run("stat parsing reads the fields after the last parenthesis", func(t *testing.T) {
		stat, ok := parseProcStat("1234 (a) b (c)) T 42 1234 0 0\n")
		assert.True(t, ok)
		assert.Equal(t, procStat{ppid: 42, state: "T"}, stat)
		_, ok = parseProcStat("1234 (truncated")
		assert.False(t, ok)
		stat, ok = readProcStat(os.Getpid())
		assert.True(t, ok)
		assert.Equal(t, os.Getppid(), stat.ppid)
	})
}
