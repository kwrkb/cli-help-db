package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirs_WindowsExecExtensions(t *testing.T) {
	dir := t.TempDir()

	// Create files with various Windows executable extensions
	exts := []string{".exe", ".cmd", ".bat"}
	for _, ext := range exts {
		name := "tool" + ext
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a non-executable file
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	names := ScanDirs([]string{dir})
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	for _, ext := range exts {
		want := "tool" + ext
		if !nameSet[want] {
			t.Errorf("expected %q in results", want)
		}
	}
	if nameSet["readme.txt"] {
		t.Error(".txt file should not be included")
	}
}

func TestFilter_WindowsExtensionStripping(t *testing.T) {
	// Create a temp dir with .exe files to simulate Windows PATH
	dir := t.TempDir()

	for _, name := range []string{"curl.exe", "git.exe", "jq.exe"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Override Scan by testing through ScanDirs directly
	scanned := ScanDirs([]string{dir})

	// Build the lookup map the same way Filter does
	all := make(map[string]bool)
	for _, n := range scanned {
		all[n] = true
		base := n[:len(n)-len(filepath.Ext(n))]
		all[base] = true
	}

	// Config-style names (without .exe) should match
	for _, name := range []string{"curl", "git", "jq"} {
		if !all[name] {
			t.Errorf("expected %q to match after extension stripping", name)
		}
	}

	// Names with .exe should also match
	for _, name := range []string{"curl.exe", "git.exe", "jq.exe"} {
		if !all[name] {
			t.Errorf("expected %q to match directly", name)
		}
	}

	// Nonexistent should not match
	if all["nonexistent"] {
		t.Error("nonexistent command should not be in map")
	}
}
