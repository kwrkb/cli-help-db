package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirs_Basic(t *testing.T) {
	dir := t.TempDir()

	// Create executable files
	for _, name := range []string{"foo", "bar", "baz"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create a non-executable file
	if err := os.WriteFile(filepath.Join(dir, "noexec"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a directory (should be skipped)
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	names := ScanDirs([]string{dir})

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	for _, want := range []string{"foo", "bar", "baz"} {
		if !nameSet[want] {
			t.Errorf("expected %q in results", want)
		}
	}
	if nameSet["noexec"] {
		t.Error("non-executable file should not be included")
	}
	if nameSet["subdir"] {
		t.Error("directory should not be included")
	}
}

func TestScanDirs_Dedup(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Same name in both dirs
	for _, dir := range []string{dir1, dir2} {
		path := filepath.Join(dir, "dup")
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	names := ScanDirs([]string{dir1, dir2})
	count := 0
	for _, n := range names {
		if n == "dup" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 occurrence of 'dup', got %d", count)
	}
}

func TestScanDirs_NonexistentDir(t *testing.T) {
	names := ScanDirs([]string{"/nonexistent/path/xyz"})
	if len(names) != 0 {
		t.Errorf("expected empty result for nonexistent dir, got %v", names)
	}
}
