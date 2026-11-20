//go:build !linux

package chunk

import "os"

// killProcessTree kills the command's process only: finding its descendants
// relies on /proc, which is Linux specific (see killtree_linux.go).
func killProcessTree(root *os.Process) error {
	return root.Kill()
}
