//go:build windows

package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
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
	dir := t.TempDir()
	t.Setenv("PATH", dir)

	for _, name := range []string{"curl.exe", "git.exe", "jq.exe"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	testCases := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "names without extension",
			input: []string{"curl", "git", "jq", "nonexistent"},
			want:  []string{"curl", "git", "jq"},
		},
		{
			name:  "names with extension",
			input: []string{"curl.exe", "git.exe", "jq.exe", "nonexistent.exe"},
			want:  []string{"curl.exe", "git.exe", "jq.exe"},
		},
		{
			name:  "mixed names",
			input: []string{"curl", "git.exe", "nonexistent"},
			want:  []string{"curl", "git.exe"},
		},
		{
			name:  "no matches",
			input: []string{"foo", "bar"},
			want:  []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(tc.input)

			if len(got) == 0 && len(tc.want) == 0 {
				return
			}

			sort.Strings(got)
			sort.Strings(tc.want)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Filter() = %v, want %v", got, tc.want)
			}
		})
	}
}
