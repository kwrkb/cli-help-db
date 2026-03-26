package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Scan returns a deduplicated list of executable command names found on $PATH.
// First occurrence wins (matches shell resolution order).
func Scan() []string {
	return ScanDirs(pathDirs())
}

// ScanDirs returns executable command names from the given directories.
func ScanDirs(dirs []string) []string {
	seen := make(map[string]bool)
	var names []string

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !seen[name] && isExecutable(e) {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	return names
}

func pathDirs() []string {
	sep := string(os.PathListSeparator)
	return strings.Split(os.Getenv("PATH"), sep)
}

func isExecutable(entry os.DirEntry) bool {
	if runtime.GOOS == "windows" {
		return hasWindowsExecExt(entry.Name())
	}
	info, err := entry.Info()
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

func hasWindowsExecExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".exe", ".cmd", ".bat", ".com", ".ps1":
		return true
	}
	// Check PATHEXT env var for additional extensions
	for _, pe := range strings.Split(os.Getenv("PATHEXT"), ";") {
		if strings.EqualFold(ext, pe) {
			return true
		}
	}
	return false
}

// Exists checks if a command name exists on $PATH.
func Exists(name string) bool {
	for _, dir := range pathDirs() {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if !info.IsDir() && info.Mode()&0111 != 0 {
			return true
		}
	}
	return false
}

// Filter returns only the names that exist as executables on $PATH.
// On Windows, matches are tried both with and without executable extensions
// (e.g., "curl" matches "curl.exe").
func Filter(names []string) []string {
	all := make(map[string]bool)
	for _, n := range Scan() {
		all[n] = true
		// On Windows, also index without extension so "curl" matches "curl.exe"
		if runtime.GOOS == "windows" {
			base := strings.TrimSuffix(n, filepath.Ext(n))
			all[base] = true
		}
	}
	var result []string
	for _, n := range names {
		if all[n] {
			result = append(result, n)
		}
	}
	return result
}
