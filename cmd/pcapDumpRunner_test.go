package cmd

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestFilterPcapFiles_RejectsUnusableNames(t *testing.T) {
	files := []string{
		"worker-20260101-120000.pcap",
		"../escaped.pcap",
		"../../home/user/.bashrc",
		"sub/dir.pcap",
		"/etc/passwd",
		"",
		".",
		"..",
	}

	kept, errs := filterPcapFiles("worker-abc", files, nil)

	if len(kept) != 1 || kept[0] != "worker-20260101-120000.pcap" {
		t.Fatalf("Expected only the plain file name to survive, got %v", kept)
	}
	if len(errs) != len(files)-1 {
		t.Errorf("Expected %d rejection errors, got %d: %v", len(files)-1, len(errs), errs)
	}
}

func TestFilterPcapFiles_CutoffShortTimestamp(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// A name the pod controls whose trailing segment is shorter than the
	// timestamp the filter slices out of it.
	kept, _ := filterPcapFiles("worker-abc", []string{"a-b", "worker-x"}, &cutoff)

	if len(kept) != 0 {
		t.Errorf("Expected no files to pass the cutoff, got %v", kept)
	}
}

func TestFilterPcapFiles_CutoffKeepsRecent(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	kept, _ := filterPcapFiles("worker-abc", []string{
		"worker-20251231-235959.pcap",
		"worker-20260102-010101.pcap",
	}, &cutoff)

	if len(kept) != 1 || kept[0] != "worker-20260102-010101.pcap" {
		t.Errorf("Expected only the file after the cutoff, got %v", kept)
	}
}

func TestIsPlainFileName(t *testing.T) {
	valid := []string{"a.pcap", "worker-20260101-120000.pcap", "..hidden", "a..b"}
	for _, name := range valid {
		if !isPlainFileName(name) {
			t.Errorf("Expected %q to be accepted", name)
		}
	}

	invalid := []string{"", ".", "..", "/abs.pcap", "a/b.pcap", "../b.pcap", "./b.pcap"}
	for _, name := range invalid {
		if isPlainFileName(name) {
			t.Errorf("Expected %q to be rejected", name)
		}
	}
}

func TestSecureDestPath_ConfinesToBase(t *testing.T) {
	base := t.TempDir()
	outside := t.TempDir()

	ok, err := secureDestPath(base, "data.pcap")
	if err != nil {
		t.Fatalf("Expected a plain name to resolve, got %v", err)
	}
	if filepath.Dir(ok) != filepath.Clean(base) && !pathContained(base, ok) {
		t.Errorf("Resolved path %q is not inside %q", ok, base)
	}

	for _, dest := range []string{
		"../escaped.pcap",
		"../../escaped.pcap",
		filepath.Join(outside, "escaped.pcap"),
		"",
	} {
		if _, err := secureDestPath(base, dest); err == nil {
			t.Errorf("Expected %q to be rejected", dest)
		}
	}
}

func TestSecureDestPath_RejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation on Windows requires elevated rights")
	}

	base := t.TempDir()
	outside := t.TempDir()

	if err := os.Symlink(outside, filepath.Join(base, "linkdir")); err != nil {
		t.Fatalf("Failed to create dir symlink: %v", err)
	}
	victim := filepath.Join(outside, "victim.pcap")
	if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
		t.Fatalf("Failed to seed victim file: %v", err)
	}
	if err := os.Symlink(victim, filepath.Join(base, "linkfile.pcap")); err != nil {
		t.Fatalf("Failed to create file symlink: %v", err)
	}

	for _, dest := range []string{"linkdir/escaped.pcap", "linkfile.pcap"} {
		if _, err := secureDestPath(base, dest); err == nil {
			t.Errorf("Expected %q to be rejected", dest)
		}
	}
}

// copyFileFromPod must refuse a traversing source name before it builds the
// in-pod path and execs, which is why a nil client is enough here.
func TestCopyFileFromPod_RejectsTraversingSrcFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.pcap")

	err := copyFileFromPod(context.Background(), nil, nil, &PodFileInfo{}, "../../etc/shadow", dest)
	if err == nil {
		t.Fatal("Expected a traversing srcFile to be refused")
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Error("Destination file was created despite the refusal")
	}
}

func TestCreateNoFollow_RefusesSymlinkedDest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("O_NOFOLLOW is not available on Windows")
	}

	base := t.TempDir()
	outside := t.TempDir()
	victim := filepath.Join(outside, "victim.pcap")
	if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
		t.Fatalf("Failed to seed victim file: %v", err)
	}
	link := filepath.Join(base, "link.pcap")
	if err := os.Symlink(victim, link); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	if f, err := createNoFollow(link); err == nil {
		f.Close()
		t.Fatal("Expected createNoFollow to refuse a symlinked destination")
	}
	if content, err := os.ReadFile(victim); err != nil || string(content) != "original" {
		t.Errorf("Victim file was modified: %q, err %v", content, err)
	}
}
