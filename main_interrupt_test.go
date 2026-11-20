//go:build unix

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

const interruptHelperEnv = "MARKDOWN_RUNNER_INTERRUPT_HELPER_DIR"

// TestInterruptHelper is not a test on its own: TestInterruptStopsRunningChunk
// runs the test binary through it to get a markdown-runner process.
func TestInterruptHelper(t *testing.T) {
	dir := os.Getenv(interruptHelperEnv)
	if dir == "" {
		t.Skip("helper process for TestInterruptStopsRunningChunk")
	}
	os.Args = []string{"markdown-runner", "--view", "ci", dir}
	main()
}

// TestInterruptStopsRunningChunk sends SIGINT to the process group of a running
// markdown-runner, as Ctrl-C does, and checks that the command of the running
// chunk dies with it instead of being left behind.
func TestInterruptStopsRunningChunk(t *testing.T) {
	if signal.Ignored(os.Interrupt) {
		t.Skip("SIGINT is ignored by this process, the helper would inherit it")
	}
	if _, err := exec.LookPath("pgrep"); err != nil {
		t.Skip("pgrep is needed to look for the chunk's process")
	}
	sleepArg := fmt.Sprintf("4243.%06d", time.Now().UnixNano()%1000000)
	pattern := "^sleep " + sleepArg + "$"
	isRunning := func() bool { return exec.Command("pgrep", "-f", pattern).Run() == nil }
	t.Cleanup(func() { _ = exec.Command("pkill", "-KILL", "-f", pattern).Run() })

	dir := t.TempDir()
	content := "```bash {\"stage\":\"test\"}\nsleep " + sleepArg + "\n```\n"
	if err := os.WriteFile(filepath.Join(dir, "interrupt.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	runner := exec.Command(os.Args[0], "-test.run=^TestInterruptHelper$")
	runner.Env = append(os.Environ(), interruptHelperEnv+"="+dir)
	// its own process group, so that the interrupt, like Ctrl-C, reaches the runner and its commands but not go test
	runner.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := runner.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- runner.Wait() }()
	t.Cleanup(func() { _ = syscall.Kill(-runner.Process.Pid, syscall.SIGKILL) })

	waitFor := func(condition func() bool, timeout time.Duration) bool {
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			if condition() {
				return true
			}
			time.Sleep(50 * time.Millisecond)
		}
		return condition()
	}
	if !waitFor(isRunning, 20*time.Second) {
		t.Fatal("the chunk's command never started")
	}

	if err := syscall.Kill(-runner.Process.Pid, syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	select {
	case <-exited:
	case <-time.After(10 * time.Second):
		t.Fatal("the runner did not stop on interrupt")
	}
	if !waitFor(func() bool { return !isRunning() }, 5*time.Second) {
		t.Fatal("the chunk's command survived the interrupt")
	}
}
