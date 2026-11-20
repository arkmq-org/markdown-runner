//go:build linux

package chunk

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// killProcessTree kills a command's process and all its descendants.
//
// Commands stay in the runner's process group (Ctrl-C and terminal reads must
// keep working), so the group can't be killed as a whole: the descendants are
// found by walking /proc instead. The walk has to happen before the kill: once
// the command dies, its children are re-parented to init and can't be told
// apart from any other process anymore. To keep the tree from changing during
// the walk, each process found is stopped (SIGSTOP) before its children are
// looked for: a stopped process can neither fork nor reap its children, so the
// pids found stay valid until they get killed. The walk is repeated until it
// finds no new descendant and every one of them is stopped, then they are all
// killed.
//
// Limit: a process that already left the tree before the walk (a daemon that
// double-forked, or a child orphaned when its own parent exited) can't be
// found. The output grace period of the timeout remains the safety net for
// those still holding the output pipes.
func killProcessTree(root *os.Process) error {
	if err := root.Signal(syscall.SIGSTOP); errors.Is(err, os.ErrProcessDone) {
		// already reaped: its pid may belong to another process by now
		return err
	}
	tree := map[int]bool{root.Pid: true}
	var procs map[int]procStat
	// bounded so that a process stuck in the kernel, which can't stop yet,
	// doesn't keep the runner waiting forever
	for pass := 0; pass < 100; pass++ {
		if pass > 0 {
			time.Sleep(time.Millisecond)
		}
		procs = readProcesses()
		grown, settled := false, true
		for pid, proc := range procs {
			if !tree[pid] && tree[proc.ppid] {
				_ = syscall.Kill(pid, syscall.SIGSTOP)
				tree[pid] = true
				grown = true
			}
		}
		for pid := range tree {
			if proc, ok := procs[pid]; ok && !proc.stopped() {
				settled = false
			}
		}
		if !grown && settled {
			break
		}
	}
	for pid := range tree {
		// a pid missing from the last walk exited before it could be stopped,
		// it may have been reused by an unrelated process: leave it alone
		if _, alive := procs[pid]; alive && pid != root.Pid {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}
	// os.Process.Kill keeps os/exec's semantics (os.ErrProcessDone once exited)
	return root.Kill()
}

// procStat is what killProcessTree needs from /proc/<pid>/stat.
type procStat struct {
	ppid  int
	state string
}

// stopped tells whether the process can't run anymore: stopped, traced, or dead.
func (p procStat) stopped() bool {
	return strings.ContainsAny(p.state, "TtZXx")
}

// readProcesses returns the parent and state of every process currently
// listed in /proc. Processes vanishing during the read are skipped.
func readProcesses() map[int]procStat {
	procs := map[int]procStat{}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return procs
	}
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		if stat, ok := readProcStat(pid); ok {
			procs[pid] = stat
		}
	}
	return procs
}

// readProcStat reads /proc/<pid>/stat, see parseProcStat.
func readProcStat(pid int) (procStat, bool) {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return procStat{}, false
	}
	return parseProcStat(string(data))
}

// parseProcStat parses a /proc/<pid>/stat line: "pid (comm) state ppid ...".
// The comm field may contain spaces and parentheses, so the fields are read
// after the last ')'.
func parseProcStat(s string) (procStat, bool) {
	end := strings.LastIndexByte(s, ')')
	if end < 0 {
		return procStat{}, false
	}
	fields := strings.Fields(s[end+1:])
	if len(fields) < 2 {
		return procStat{}, false
	}
	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return procStat{}, false
	}
	return procStat{ppid: ppid, state: fields[0]}, true
}
