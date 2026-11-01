package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// secureDestPath resolves dest to an absolute path confined to baseDir. A dest
// (absolute or relative) that resolves outside baseDir via "../" traversal is
// rejected, so a destination taken from an untrusted source cannot redirect a
// write to an arbitrary location such as ~/.ssh/authorized_keys or
// ~/.kube/config. See CWE-22.
//
// Containment is checked twice: once lexically, then again after resolving
// symlinks, so a symlinked subdirectory (or a symlinked dest itself) inside the
// base cannot redirect the write outside it. Callers that create the file
// should open it with createNoFollow to narrow the remaining TOCTOU window.
func secureDestPath(baseDir, dest string) (string, error) {
	if dest == "" {
		return "", errors.New("empty destination")
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base directory: %w", err)
	}
	// Resolve symlinks in the base itself, so the comparison below is between
	// real paths. The base may legitimately be a symlink (e.g. /tmp on macOS,
	// which is a link to /private/tmp).
	if resolvedBase, err := filepath.EvalSymlinks(absBase); err == nil {
		absBase = resolvedBase
	}
	full := dest
	if filepath.IsAbs(full) {
		full = filepath.Clean(full)
	} else {
		full = filepath.Join(absBase, full)
	}
	if !pathContained(absBase, full) {
		return "", fmt.Errorf("destination %q escapes the base directory %q", dest, absBase)
	}
	resolvedFull, err := resolveSymlinkedPath(full)
	if err != nil {
		return "", fmt.Errorf("cannot resolve destination %q: %w", dest, err)
	}
	if !pathContained(absBase, resolvedFull) {
		return "", fmt.Errorf("destination %q resolves through a symlink to %q, outside the base directory %q", dest, resolvedFull, absBase)
	}
	return full, nil
}

// pathContained reports whether target is base itself or lies beneath it. Both
// arguments must already be absolute and symlink-resolved.
func pathContained(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// resolveSymlinkedPath resolves symlinks in the deepest existing ancestor of p
// and re-appends the trailing components that do not exist yet.
// filepath.EvalSymlinks fails outright on a path that does not exist, which is
// the normal case for a download destination, hence the walk upwards.
func resolveSymlinkedPath(p string) (string, error) {
	cur := filepath.Clean(p)
	rest := ""
	for {
		resolved, err := filepath.EvalSymlinks(cur)
		if err == nil {
			if rest == "" {
				return resolved, nil
			}
			return filepath.Join(resolved, rest), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			// Walked up to the root without finding an existing ancestor.
			return filepath.Clean(p), nil
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}

// isPlainFileName reports whether name is usable as a single path component:
// non-empty, not "." or "..", no path separator and not absolute. Names
// reported by a remote peer are checked with this before they are joined to a
// local destination directory or to a remote source directory.
func isPlainFileName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if filepath.IsAbs(name) || strings.ContainsRune(name, '/') || strings.ContainsRune(name, os.PathSeparator) {
		return false
	}
	return name == filepath.Base(name)
}
