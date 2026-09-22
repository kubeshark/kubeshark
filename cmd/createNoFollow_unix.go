//go:build !windows

package cmd

import (
	"os"
	"syscall"
)

// createNoFollow creates or truncates name for writing and fails if the final
// path component is a symlink. Used for MCP download destinations so a symlink
// cannot redirect the write outside the confined download directory.
func createNoFollow(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|syscall.O_NOFOLLOW, 0o644)
}
