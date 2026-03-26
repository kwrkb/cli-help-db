package db

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DB represents a directory-based help text database.
type DB struct {
	Dir string
}

// New creates a DB pointing to the given directory.
func New(dir string) *DB {
	return &DB{Dir: dir}
}

// Write saves help text for a command. Creates the directory if needed.
func (d *DB) Write(name, text string) error {
	if err := os.MkdirAll(d.Dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(d.Dir, name+".txt")
	return os.WriteFile(path, []byte(text), 0644)
}

// Read returns the help text for a command, or empty string if not found.
func (d *DB) Read(name string) (string, error) {
	path := filepath.Join(d.Dir, name+".txt")
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
func (d *DB) Has(name string) bool {
	path := filepath.Join(d.Dir, name+".txt")
	_, err := os.Stat(path)
	return err == nil
}

// List returns all command names in the database, sorted.
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
			names = append(names, strings.TrimSuffix(name, ".txt"))
		}
	}
	sort.Strings(names)
	return names, nil
}
