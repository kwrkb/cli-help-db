package db

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DB represents a directory-based help text database.
// Command names may contain spaces for subcommands (e.g. "docker container ls"),
// which are stored as "docker__container__ls.txt" on disk.
type DB struct {
	Dir string
}

// New creates a DB pointing to the given directory.
func New(dir string) *DB {
	return &DB{Dir: dir}
}

// NameToKey converts a command name (possibly with spaces for subcommands)
// to a filesystem-safe key using "__" as separator.
// e.g. "docker container ls" -> "docker__container__ls"
func NameToKey(name string) string {
	return strings.ReplaceAll(name, " ", "__")
}

// KeyToName converts a filesystem key back to a display name with spaces.
// e.g. "docker__container__ls" -> "docker container ls"
func KeyToName(key string) string {
	return strings.ReplaceAll(key, "__", " ")
}

// Write saves help text for a command. Creates the directory if needed.
// name can contain spaces for subcommands (e.g. "docker container ls").
func (d *DB) Write(name, text string) error {
	if err := os.MkdirAll(d.Dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(d.Dir, NameToKey(name)+".txt")
	return os.WriteFile(path, []byte(text), 0644)
}

// Read returns the help text for a command, or empty string if not found.
// name can contain spaces for subcommands (e.g. "docker container ls").
func (d *DB) Read(name string) (string, error) {
	path := filepath.Join(d.Dir, NameToKey(name)+".txt")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// Has checks if a command exists in the database.
// name can contain spaces for subcommands (e.g. "docker container ls").
func (d *DB) Has(name string) bool {
	path := filepath.Join(d.Dir, NameToKey(name)+".txt")
	_, err := os.Stat(path)
	return err == nil
}

// List returns all command names in the database, sorted.
// Subcommand keys (e.g. "docker__container__ls") are returned as
// display names with spaces (e.g. "docker container ls").
func (d *DB) List() ([]string, error) {
	entries, err := os.ReadDir(d.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".txt") {
			key := strings.TrimSuffix(name, ".txt")
			names = append(names, KeyToName(key))
		}
	}
	sort.Strings(names)
	return names, nil
}
